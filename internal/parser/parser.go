package parser

import (
	"bufio"
	"os"
	"strings"

	"github.com/willams/lia/internal/ir"
)

// ParseFile parses a .lia source file into a minimal IR program.
func ParseFile(path string) (*ir.Program, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	p := &ir.Program{Version: "0.1"}
	scanner := bufio.NewScanner(f)
	var currentProject *ir.Project
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		if strings.HasPrefix(line, "project ") {
			name := strings.Fields(strings.TrimPrefix(line, "project "))
			if len(name) > 0 {
				p.Projects = append(p.Projects, ir.Project{Name: name[0]})
				currentProject = &p.Projects[len(p.Projects)-1]
			}
			continue
		}

		if line == "}" {
			currentProject = nil
			continue
		}

		if strings.HasPrefix(line, "repro ") && currentProject != nil {
			value := strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(line, "repro ")), ";")
			currentProject.Repro = ir.ReproProfile(value)
			continue
		}

		if strings.HasPrefix(line, "tape ") && currentProject != nil {
			// e.g. tape prompt_tape "./prompt-tape.json";
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				path := strings.TrimSuffix(parts[len(parts)-1], ";")
				path = strings.Trim(path, "\"")
				currentProject.Tape = path
			}
			continue
		}

		if strings.HasPrefix(line, "use pack ") && currentProject != nil {
			ref := strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(line, "use pack ")), ";")
			name, version := parsePackRef(ref)
			currentProject.Uses = append(currentProject.Uses, ir.PackRef{Name: name, Version: version})
			continue
		}

		if strings.HasPrefix(line, "module ") {
			modLine := strings.TrimPrefix(line, "module ")
			modName := strings.Fields(modLine)
			if len(modName) == 0 {
				continue
			}
			role := ""
			if strings.Contains(line, " as ") {
				parts := strings.Split(line, " as ")
				if len(parts) == 2 {
					role = strings.Fields(parts[1])[0]
				}
			}
			p.Modules = append(p.Modules, ir.Module{Name: modName[0], Role: role})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return p, nil
}

func parsePackRef(ref string) (string, string) {
	ref = strings.TrimSpace(ref)
	if strings.Contains(ref, "@") {
		parts := strings.SplitN(ref, "@", 2)
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return ref, ""
}
