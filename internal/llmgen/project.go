package llmgen

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/willams/lia/internal/ir"
)

// ProjectSpec defines a project bootstrap request driven by AI generation.
type ProjectSpec struct {
	Name        string          `json:"name"`
	Brief       string          `json:"brief,omitempty"`
	Repro       string          `json:"repro,omitempty"`
	TapeFile    string          `json:"tape_file,omitempty"`
	Packs       []ir.PackRef    `json:"packs,omitempty"`
	Policies    []NamedExpr     `json:"policies,omitempty"`
	Constraints []NamedExpr     `json:"constraints,omitempty"`
	Modules     []ProjectModule `json:"modules"`
}

// ProjectModule defines one module to generate inside a project.
type ProjectModule struct {
	Name        string  `json:"name"`
	Role        string  `json:"role"`
	Context     string  `json:"context"`
	Model       string  `json:"model,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// NamedExpr is used for policy/constraint declarations rendered into LIA.
type NamedExpr struct {
	Name string `json:"name"`
	Expr string `json:"expr"`
}

// GeneratedModule is a generated source module plus parsed metadata.
type GeneratedModule struct {
	Source string
	Module *ir.Module
	Meta   *GenMetadata
}

// LoadProjectSpec loads a JSON project spec from disk.
func LoadProjectSpec(path string) (*ProjectSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var spec ProjectSpec
	if err := json.Unmarshal(b, &spec); err != nil {
		return nil, err
	}
	if err := spec.Normalize(); err != nil {
		return nil, err
	}
	return &spec, nil
}

// Normalize applies defaults and validates required fields.
func (s *ProjectSpec) Normalize() error {
	if s == nil {
		return fmt.Errorf("nil project spec")
	}
	s.Name = strings.TrimSpace(s.Name)
	s.Brief = strings.TrimSpace(s.Brief)
	if s.Name == "" {
		return fmt.Errorf("project spec name is required")
	}
	if s.Repro == "" {
		s.Repro = string(ir.ReproPinned)
	}
	if s.TapeFile == "" {
		s.TapeFile = "prompt-tape.json"
	}
	if len(s.Modules) == 0 {
		return fmt.Errorf("project spec requires at least one module")
	}

	seen := map[string]bool{}
	for i := range s.Modules {
		mod := &s.Modules[i]
		mod.Name = strings.TrimSpace(mod.Name)
		mod.Role = strings.TrimSpace(mod.Role)
		mod.Context = strings.TrimSpace(mod.Context)
		mod.Model = strings.TrimSpace(mod.Model)
		if mod.Name == "" {
			return fmt.Errorf("module[%d].name is required", i)
		}
		if mod.Role == "" {
			return fmt.Errorf("module[%d].role is required", i)
		}
		if seen[mod.Name] {
			return fmt.Errorf("duplicate module name in project spec: %s", mod.Name)
		}
		seen[mod.Name] = true
	}

	return nil
}
