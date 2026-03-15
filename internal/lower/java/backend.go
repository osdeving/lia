package java

import (
	"github.com/willams/lia/internal/ir"
	"github.com/willams/lia/internal/lower"
)

// JavaBackend implements the lower.Backend interface for Java code generation.
type JavaBackend struct{}

// Ensure JavaBackend implements lower.Backend at compile time.
var _ lower.Backend = (*JavaBackend)(nil)

func init() {
	lower.Register(&JavaBackend{})
}

// Name returns "java".
func (j *JavaBackend) Name() string { return "java" }

// LowerProject lowers an IR program into a Java project.
func (j *JavaBackend) LowerProject(p *ir.Program, opts lower.Options) (*lower.Project, error) {
	javaOpts := Options{
		Strict: opts.Strict,
	}
	if opts.Profile != "" {
		profile, err := ParseProfile(opts.Profile)
		if err != nil {
			return nil, err
		}
		javaOpts.Profile = profile
	}
	javaProject, err := LowerProjectWithOptions(p, javaOpts)
	if err != nil {
		return nil, err
	}
	// Convert java.Project → lower.Project
	files := make([]lower.File, len(javaProject.Files))
	for i, f := range javaProject.Files {
		files[i] = lower.File{Path: f.Path, Content: f.Content}
	}
	return &lower.Project{
		Name:        javaProject.Name,
		BasePackage: javaProject.BasePackage,
		Language:    "java",
		Files:       files,
	}, nil
}

// WriteProject writes a Java project to disk.
func (j *JavaBackend) WriteProject(dir string, project *lower.Project) error {
	return lower.WriteProjectFiles(dir, project.Files)
}

// MapType maps an IR type to a Java type.
func (j *JavaBackend) MapType(irType string) lower.MappedType {
	jt := resolveTypeSimple(irType)
	return lower.MappedType{
		Name:      jt.Name,
		Imports:   jt.Imports,
		ZeroValue: jt.ZeroValue,
	}
}

// SupportedProfiles returns Java-specific profile names.
func (j *JavaBackend) SupportedProfiles() []string {
	return []string{"plain", "spring-boot", "quarkus"}
}

// resolveTypeSimple is a standalone type resolver for simple type names
// (without needing the full javaIndex).
func resolveTypeSimple(irType string) javaType {
	switch irType {
	case "String":
		return javaType{Name: "String", ZeroValue: "\"\""}
	case "Int", "Integer":
		return javaType{Name: "int", ZeroValue: "0"}
	case "Long":
		return javaType{Name: "long", ZeroValue: "0L"}
	case "Float":
		return javaType{Name: "float", ZeroValue: "0.0f"}
	case "Double":
		return javaType{Name: "double", ZeroValue: "0.0"}
	case "Bool", "Boolean":
		return javaType{Name: "boolean", ZeroValue: "false"}
	case "Byte":
		return javaType{Name: "byte", ZeroValue: "0"}
	case "Void":
		return javaType{Name: "void", ZeroValue: ""}
	default:
		return javaType{Name: irType, ZeroValue: "null"}
	}
}
