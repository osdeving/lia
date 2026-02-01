package packs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/willams/lia/internal/ir"
)

// Loader resolves and parses pack files.
type Loader struct {
	SearchDirs []string
}

// DefaultSearchDirs returns default pack lookup paths that exist.
func DefaultSearchDirs() []string {
	candidates := []string{"./packs", "./docs/spec/packs"}
	var dirs []string
	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			dirs = append(dirs, dir)
		}
	}
	return dirs
}

// LoadAll loads all referenced packs and returns diagnostics.
func (l Loader) LoadAll(refs []ir.PackRef) ([]ir.Pack, []ir.Diagnostic) {
	var packs []ir.Pack
	var diags []ir.Diagnostic

	seen := map[string]bool{}
	for _, ref := range refs {
		key := ref.Name + "@" + ref.Version
		if seen[key] {
			continue
		}
		seen[key] = true

		path, err := l.resolvePackPath(ref)
		if err != nil {
			diags = append(diags, ir.Diagnostic{
				Severity: "error",
				Message:  "pack not found: " + key,
			})
			continue
		}
		p, err := ParsePackFile(path)
		if err != nil {
			diags = append(diags, ir.Diagnostic{
				Severity: "error",
				Message:  "pack parse failed: " + path + ": " + err.Error(),
			})
			continue
		}
		if p.Name == "" {
			p.Name = ref.Name
		}
		if p.Version == "" {
			p.Version = ref.Version
		}
		packs = append(packs, *p)
	}

	return packs, diags
}

func (l Loader) resolvePackPath(ref ir.PackRef) (string, error) {
	if len(l.SearchDirs) == 0 {
		return "", errors.New("no pack search dirs")
	}
	candidates := packFileCandidates(ref.Name, ref.Version)
	for _, dir := range l.SearchDirs {
		for _, name := range candidates {
			path := filepath.Join(dir, name)
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				return path, nil
			}
		}
	}
	return "", os.ErrNotExist
}

func packFileCandidates(name, version string) []string {
	var candidates []string
	kebab := toKebab(name)
	lower := strings.ToLower(name)
	upperUnderscore := strings.ToUpper(strings.ReplaceAll(kebab, "-", "_"))

	base := []string{name, lower, kebab, upperUnderscore}
	for _, b := range base {
		if b == "" {
			continue
		}
		candidates = append(candidates, b+".lia")
		if version != "" {
			candidates = append(candidates, b+"-"+version+".lia")
			candidates = append(candidates, b+"@"+version+".lia")
			candidates = append(candidates, b+"_"+version+".lia")
		}
	}

	return candidates
}

func toKebab(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r == '_' {
			b.WriteByte('-')
			continue
		}
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}
