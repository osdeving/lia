package python

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/willams/lia/internal/ir"
	"github.com/willams/lia/internal/lower"
)

// PythonBackend implements the lower.Backend interface for Python code generation.
type PythonBackend struct{}

var _ lower.Backend = (*PythonBackend)(nil)

func init() {
	lower.Register(&PythonBackend{})
}

func (p *PythonBackend) Name() string { return "python" }

func (p *PythonBackend) SupportedProfiles() []string {
	return []string{"plain"}
}

func (p *PythonBackend) MapType(irType string) lower.MappedType {
	py := mapIRType(irType)
	return lower.MappedType{Name: py, ZeroValue: pythonZeroValue(py)}
}

func (p *PythonBackend) WriteProject(dir string, project *lower.Project) error {
	return lower.WriteProjectFiles(dir, project.Files)
}

func (p *PythonBackend) LowerProject(prog *ir.Program, opts lower.Options) (*lower.Project, error) {
	project, err := LowerProject(prog)
	if err != nil {
		return nil, err
	}
	files := make([]lower.File, len(project.Files))
	for i, f := range project.Files {
		files[i] = lower.File{Path: f.Path, Content: f.Content}
	}
	return &lower.Project{
		Name:     project.Name,
		Language: "python",
		Files:    files,
	}, nil
}

// File is one lowered Python file.
type File struct {
	Path    string
	Content []byte
}

// Project is a lowered Python project.
type Project struct {
	Name  string
	Files []File
}

// Lower emits a manifest for the lowered Python project (backward compat).
func Lower(p *ir.Program) ([]byte, error) {
	project, err := LowerProject(p)
	if err != nil {
		return nil, err
	}
	sort.Slice(project.Files, func(i, j int) bool {
		return project.Files[i].Path < project.Files[j].Path
	})
	var buf bytes.Buffer
	buf.WriteString("# LIA lower (python)\n")
	buf.WriteString("# project: ")
	buf.WriteString(project.Name)
	buf.WriteByte('\n')
	for _, file := range project.Files {
		buf.WriteString(file.Path)
		buf.WriteByte('\n')
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

// LowerProject lowers a linked IR program into a multi-file Python project.
func LowerProject(p *ir.Program) (*Project, error) {
	if p == nil {
		return nil, fmt.Errorf("nil program")
	}
	name := "lia_generated"
	if len(p.Projects) > 0 && strings.TrimSpace(p.Projects[0].Name) != "" {
		name = sanitizePythonModule(p.Projects[0].Name)
	}

	var files []File
	files = append(files, File{
		Path:    "README.md",
		Content: []byte(renderReadme(name, p)),
	})
	files = append(files, File{
		Path:    "requirements.txt",
		Content: []byte("# LIA generated Python project\n"),
	})

	// Root __init__.py
	files = append(files, File{
		Path:    "__init__.py",
		Content: []byte(fmt.Sprintf("\"\"\"LIA generated project: %s\"\"\"\n", name)),
	})

	for _, mod := range p.Modules {
		pkg := modulePythonPackage(mod.Name)
		// Module __init__.py
		files = append(files, File{
			Path:    filepath.ToSlash(filepath.Join(pkg, "__init__.py")),
			Content: []byte(fmt.Sprintf("\"\"\"Module %s\"\"\"\n", mod.Name)),
		})

		// Types → dataclasses
		for _, decl := range mod.Types {
			files = append(files, File{
				Path:    filepath.ToSlash(filepath.Join(pkg, snakeCase(decl.Name)+".py")),
				Content: []byte(renderTypeDecl(decl)),
			})
		}

		// Enums → Python Enum
		for _, decl := range mod.Enums {
			files = append(files, File{
				Path:    filepath.ToSlash(filepath.Join(pkg, snakeCase(decl.Name)+".py")),
				Content: []byte(renderEnumDecl(decl)),
			})
		}

		// Records → @dataclass
		for _, decl := range mod.Records {
			files = append(files, File{
				Path:    filepath.ToSlash(filepath.Join(pkg, snakeCase(decl.Name)+".py")),
				Content: []byte(renderRecordDecl(decl)),
			})
		}

		// Ports → ABC
		for _, decl := range mod.Ports {
			files = append(files, File{
				Path:    filepath.ToSlash(filepath.Join(pkg, snakeCase(decl.Name)+".py")),
				Content: []byte(renderPortDecl(decl)),
			})
		}

		// Usecases → classes with execute()
		for _, decl := range mod.Usecases {
			files = append(files, File{
				Path:    filepath.ToSlash(filepath.Join(pkg, snakeCase(decl.Name)+".py")),
				Content: []byte(renderUsecaseDecl(decl)),
			})
		}

		// Adapters → implementations
		for _, decl := range mod.Adapters {
			files = append(files, File{
				Path:    filepath.ToSlash(filepath.Join(pkg, snakeCase(decl.Name)+".py")),
				Content: []byte(renderAdapterDecl(decl)),
			})
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	return &Project{Name: name, Files: files}, nil
}

// WriteProject writes a lowered Python project to disk.
func WriteProject(dir string, project *Project) error {
	if project == nil {
		return fmt.Errorf("nil project")
	}
	for _, file := range project.Files {
		target := filepath.Join(dir, filepath.FromSlash(file.Path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, file.Content, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func renderReadme(name string, p *ir.Program) string {
	var b strings.Builder
	b.WriteString("# " + name + "\n\n")
	b.WriteString("Projeto Python gerado pelo lower do LIA.\n\n")
	b.WriteString("- Modulos LIA: " + fmt.Sprintf("%d", len(p.Modules)) + "\n")
	b.WriteString("- Run: `python -m " + name + "`\n")
	return b.String()
}

func renderTypeDecl(decl ir.TypeDecl) string {
	pyType := mapIRType(decl.Base)
	var b strings.Builder
	b.WriteString("from dataclasses import dataclass\n\n\n")
	b.WriteString("@dataclass(frozen=True)\n")
	b.WriteString("class " + decl.Name + ":\n")
	b.WriteString("    \"\"\"Type alias: " + decl.Name + " = " + decl.Base + "\"\"\"\n")
	b.WriteString("    value: " + pyType + "\n")
	if decl.Predicate != "" {
		b.WriteString("\n    def __post_init__(self):\n")
		if strings.EqualFold(strings.TrimSpace(decl.Predicate), "nonEmpty") && strings.EqualFold(decl.Base, "String") {
			b.WriteString("        if not self.value or not self.value.strip():\n")
			b.WriteString("            raise ValueError(\"" + decl.Name + " must not be empty\")\n")
		} else {
			b.WriteString("        # Predicate: " + strings.TrimSpace(decl.Predicate) + "\n")
			b.WriteString("        pass\n")
		}
	}
	return b.String()
}

func renderEnumDecl(decl ir.EnumDecl) string {
	var b strings.Builder
	b.WriteString("from enum import Enum\n\n\n")
	b.WriteString("class " + decl.Name + "(Enum):\n")
	if len(decl.Values) == 0 {
		b.WriteString("    pass\n")
		return b.String()
	}
	for _, v := range decl.Values {
		b.WriteString("    " + v + " = \"" + v + "\"\n")
	}
	return b.String()
}

func renderRecordDecl(decl ir.RecordDecl) string {
	var b strings.Builder
	b.WriteString("from dataclasses import dataclass\n\n\n")
	b.WriteString("@dataclass\n")
	b.WriteString("class " + decl.Name + ":\n")
	if len(decl.Fields) == 0 {
		b.WriteString("    pass\n")
		return b.String()
	}
	for _, f := range decl.Fields {
		b.WriteString("    " + snakeCase(f.Name) + ": " + mapIRType(f.Type) + "\n")
	}
	return b.String()
}

func renderPortDecl(decl ir.PortDecl) string {
	var b strings.Builder
	b.WriteString("from abc import ABC, abstractmethod\n\n\n")
	b.WriteString("class " + decl.Name + "(ABC):\n")
	if len(decl.Methods) == 0 {
		b.WriteString("    pass\n")
		return b.String()
	}
	for _, m := range decl.Methods {
		params := []string{"self"}
		for _, p := range m.Params {
			params = append(params, snakeCase(p.Name)+": "+mapIRType(p.Type))
		}
		returnType := "None"
		if len(m.Returns) == 1 {
			returnType = mapIRType(m.Returns[0].Type)
		} else if len(m.Returns) > 1 {
			types := make([]string, len(m.Returns))
			for i, r := range m.Returns {
				types[i] = mapIRType(r.Type)
			}
			returnType = "tuple[" + strings.Join(types, ", ") + "]"
		}
		b.WriteString("    @abstractmethod\n")
		b.WriteString("    def " + snakeCase(m.Name) + "(" + strings.Join(params, ", ") + ") -> " + returnType + ":\n")
		b.WriteString("        ...\n\n")
	}
	return b.String()
}

func renderUsecaseDecl(decl ir.UsecaseDecl) string {
	var b strings.Builder
	b.WriteString("from dataclasses import dataclass\n\n\n")

	// Input dataclass
	inputName := decl.Name + "Input"
	b.WriteString("@dataclass\n")
	b.WriteString("class " + inputName + ":\n")
	if len(decl.Inputs) == 0 {
		b.WriteString("    pass\n")
	} else {
		for _, f := range decl.Inputs {
			b.WriteString("    " + snakeCase(f.Name) + ": " + mapIRType(f.Type) + "\n")
		}
	}
	b.WriteString("\n\n")

	// Output dataclass
	outputName := decl.Name + "Output"
	b.WriteString("@dataclass\n")
	b.WriteString("class " + outputName + ":\n")
	if len(decl.Outputs) == 0 {
		b.WriteString("    pass\n")
	} else {
		for _, f := range decl.Outputs {
			b.WriteString("    " + snakeCase(f.Name) + ": " + mapIRType(f.Type) + "\n")
		}
	}
	b.WriteString("\n\n")

	// Usecase class
	b.WriteString("class " + decl.Name + ":\n")
	b.WriteString("    def __init__(self):\n")
	b.WriteString("        pass\n\n")
	b.WriteString("    def execute(self, input: " + inputName + ") -> " + outputName + ":\n")
	b.WriteString("        # Effects: " + strings.Join(decl.Effects, ", ") + "\n")
	b.WriteString("        raise NotImplementedError\n")
	return b.String()
}

func renderAdapterDecl(decl ir.AdapterDecl) string {
	var b strings.Builder
	if decl.Implements != "" {
		b.WriteString("# Implements: " + decl.Implements + "\n\n\n")
	}
	b.WriteString("class " + decl.Name + ":\n")
	b.WriteString("    \"\"\"Adapter generated from LIA.\"\"\"\n")
	b.WriteString("    def __init__(self):\n")
	b.WriteString("        pass\n")
	return b.String()
}

// mapIRType maps IR types to Python types.
func mapIRType(irType string) string {
	switch irType {
	case "String":
		return "str"
	case "Int", "Integer", "Long":
		return "int"
	case "Float", "Double", "Decimal":
		return "float"
	case "Bool", "Boolean":
		return "bool"
	case "Byte":
		return "bytes"
	case "Void":
		return "None"
	default:
		if strings.HasPrefix(irType, "list<") || strings.HasPrefix(irType, "List<") {
			inner := irType[5 : len(irType)-1]
			return "list[" + mapIRType(inner) + "]"
		}
		if strings.HasPrefix(irType, "option<") || strings.HasPrefix(irType, "Option<") {
			inner := irType[7 : len(irType)-1]
			return mapIRType(inner) + " | None"
		}
		return irType
	}
}

func pythonZeroValue(pyType string) string {
	switch pyType {
	case "str":
		return "\"\""
	case "int":
		return "0"
	case "float":
		return "0.0"
	case "bool":
		return "False"
	case "None":
		return "None"
	default:
		return "None"
	}
}

func sanitizePythonModule(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		} else if r == '-' || r == '.' || r == ' ' {
			b.WriteRune('_')
		}
	}
	if b.Len() == 0 {
		return "lia_generated"
	}
	return b.String()
}

func modulePythonPackage(moduleName string) string {
	return strings.ReplaceAll(moduleName, ".", "/")
}

func snakeCase(name string) string {
	var b strings.Builder
	for i, r := range name {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteRune('_')
			}
			b.WriteRune(r - 'A' + 'a')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
