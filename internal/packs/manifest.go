package packs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Manifest describes a pack in a machine-readable way for planning/inference.
type Manifest struct {
	Name            string           `json:"name"`
	Version         string           `json:"version,omitempty"`
	Summary         string           `json:"summary,omitempty"`
	Keywords        []string         `json:"keywords,omitempty"`
	Targets         []string         `json:"targets,omitempty"`
	RepoSignals     []string         `json:"repo_signals,omitempty"`
	DefaultSelected bool             `json:"default_selected,omitempty"`
	ModuleTemplates []ModuleTemplate `json:"module_templates,omitempty"`
}

// ModuleTemplate suggests a module layout implied by a pack.
type ModuleTemplate struct {
	Suffix  string `json:"suffix"`
	Role    string `json:"role"`
	Context string `json:"context"`
}

// LoadManifests scans pack search directories and loads all pack manifests found.
func LoadManifests(searchDirs []string) ([]Manifest, error) {
	seen := map[string]bool{}
	var manifests []Manifest

	for _, dir := range searchDirs {
		if strings.TrimSpace(dir) == "" {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".pack.json") {
				continue
			}
			path := filepath.Join(dir, entry.Name())
			manifest, err := LoadManifestFile(path)
			if err != nil {
				return nil, err
			}
			key := manifest.Name + "@" + manifest.Version
			if key == "@" || seen[key] {
				continue
			}
			seen[key] = true
			manifests = append(manifests, *manifest)
		}
	}

	sort.Slice(manifests, func(i, j int) bool {
		if manifests[i].Name == manifests[j].Name {
			return manifests[i].Version < manifests[j].Version
		}
		return manifests[i].Name < manifests[j].Name
	})
	return manifests, nil
}

// LoadManifestFile loads one pack manifest from disk.
func LoadManifestFile(path string) (*Manifest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var manifest Manifest
	if err := json.Unmarshal(b, &manifest); err != nil {
		return nil, err
	}
	manifest.normalize()
	return &manifest, nil
}

func (m *Manifest) normalize() {
	if m == nil {
		return
	}
	m.Name = strings.TrimSpace(m.Name)
	m.Version = strings.TrimSpace(m.Version)
	m.Summary = strings.TrimSpace(m.Summary)
	for i := range m.Keywords {
		m.Keywords[i] = strings.TrimSpace(m.Keywords[i])
	}
	for i := range m.Targets {
		m.Targets[i] = strings.TrimSpace(m.Targets[i])
	}
	for i := range m.RepoSignals {
		m.RepoSignals[i] = strings.TrimSpace(m.RepoSignals[i])
	}
	for i := range m.ModuleTemplates {
		m.ModuleTemplates[i].Suffix = strings.TrimSpace(m.ModuleTemplates[i].Suffix)
		m.ModuleTemplates[i].Role = strings.TrimSpace(m.ModuleTemplates[i].Role)
		m.ModuleTemplates[i].Context = strings.TrimSpace(m.ModuleTemplates[i].Context)
	}
}
