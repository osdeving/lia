package packs

import (
	"fmt"
	"os"
	"strings"

	"github.com/willams/lia/internal/ir"
)

// ParsePackFile parses a pseudo-LIA pack file into an IR Pack.
func ParsePackFile(path string) (*ir.Pack, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(b), "\n")
	pack := &ir.Pack{}

	inPack := false
	for i := 0; i < len(lines); i++ {
		line := cleanLine(lines[i])
		if line == "" {
			continue
		}

		if !inPack {
			if strings.HasPrefix(line, "pack ") {
				name, version, err := parsePackHeader(line)
				if err != nil {
					return nil, err
				}
				pack.Name = name
				pack.Version = version
				inPack = true
			}
			continue
		}

		if line == "}" {
			break
		}

		if strings.HasPrefix(line, "constraint ") {
			c, err := parseConstraintLine(line)
			if err != nil {
				return nil, err
			}
			pack.Constraints = append(pack.Constraints, *c)
			continue
		}

		if strings.HasPrefix(line, "policy ") {
			name, remainder, err := parsePolicyHeader(line)
			if err != nil {
				return nil, err
			}
			var body []string
			if remainder != "" {
				body = append(body, trimSemicolon(remainder))
			}
			for j := i + 1; j < len(lines); j++ {
				next := cleanLine(lines[j])
				if next == "" {
					continue
				}
				if strings.HasPrefix(next, "policy ") || strings.HasPrefix(next, "constraint ") || next == "}" {
					i = j - 1
					break
				}
				body = append(body, trimSemicolon(next))
				i = j
			}
			pack.Policies = append(pack.Policies, ir.PolicyDecl{Name: name, Body: strings.Join(body, "\n")})
			continue
		}
	}

	if !inPack {
		return nil, fmt.Errorf("missing pack header")
	}
	return pack, nil
}

func parsePackHeader(line string) (string, string, error) {
	trim := strings.TrimSpace(strings.TrimPrefix(line, "pack "))
	trim = strings.TrimSuffix(trim, "{")
	trim = strings.TrimSpace(trim)
	if trim == "" {
		return "", "", fmt.Errorf("empty pack header")
	}
	name := trim
	version := ""
	if strings.Contains(trim, "@") {
		parts := strings.SplitN(trim, "@", 2)
		name = strings.TrimSpace(parts[0])
		version = strings.TrimSpace(parts[1])
	}
	return name, version, nil
}

func parseConstraintLine(line string) (*ir.ConstraintDecl, error) {
	trim := strings.TrimSpace(strings.TrimPrefix(line, "constraint "))
	parts := strings.SplitN(trim, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid constraint: %s", line)
	}
	name := strings.TrimSpace(parts[0])
	expr := trimSemicolon(strings.TrimSpace(parts[1]))
	return &ir.ConstraintDecl{Name: name, Expr: expr}, nil
}

func parsePolicyHeader(line string) (string, string, error) {
	trim := strings.TrimSpace(strings.TrimPrefix(line, "policy "))
	parts := strings.SplitN(trim, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid policy: %s", line)
	}
	name := strings.TrimSpace(parts[0])
	remainder := strings.TrimSpace(parts[1])
	return name, remainder, nil
}

func trimSemicolon(s string) string {
	return strings.TrimSpace(strings.TrimSuffix(s, ";"))
}

func cleanLine(line string) string {
	trim := strings.TrimSpace(line)
	if trim == "" {
		return ""
	}
	trim = stripInlineComment(trim)
	return strings.TrimSpace(trim)
}

func stripInlineComment(line string) string {
	var out strings.Builder
	inQuote := false
	for i := 0; i < len(line); i++ {
		if line[i] == '"' {
			inQuote = !inQuote
		}
		if !inQuote && i+1 < len(line) && line[i] == '/' && line[i+1] == '/' {
			break
		}
		out.WriteByte(line[i])
	}
	return out.String()
}
