package lower

import (
	"fmt"

	"github.com/willams/lia/internal/ir"
)

// File is one lowered output file.
type File struct {
	Path    string
	Content []byte
}

// Project is a lowered project with multiple output files.
type Project struct {
	Name        string
	BasePackage string
	Files       []File
	Language    string // e.g. "java", "python"
}

// Options configures lowering behavior.
type Options struct {
	Profile string
	Strict  bool
}

// MappedType is a type mapping result from IR → target language.
type MappedType struct {
	Name      string
	Imports   []string
	ZeroValue string
}

// RenderContext provides rendering context including module/index data.
type RenderContext struct {
	CurrentModule string
	ExtraData     interface{} // backend-specific data
}

// Backend defines the interface for a lowering target.
// Each target language (Java, Python, Go, etc.) implements this interface.
type Backend interface {
	// Name returns the target name (e.g. "java", "python").
	Name() string

	// LowerProject lowers an IR program into a multi-file project.
	LowerProject(p *ir.Program, opts Options) (*Project, error)

	// WriteProject writes a lowered project to disk at the given directory.
	WriteProject(dir string, project *Project) error

	// MapType maps an IR type name to the target language type.
	MapType(irType string) MappedType

	// SupportedProfiles returns the list of supported profile names.
	SupportedProfiles() []string
}

// Registry holds registered backends by name.
var Registry = map[string]Backend{}

// Register adds a backend to the global registry.
func Register(b Backend) {
	Registry[b.Name()] = b
}

// Get returns a registered backend by name, or an error if not found.
func Get(name string) (Backend, error) {
	b, ok := Registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown lower backend: %s", name)
	}
	return b, nil
}

// Backends returns the names of all registered backends.
func Backends() []string {
	names := make([]string, 0, len(Registry))
	for name := range Registry {
		names = append(names, name)
	}
	return names
}
