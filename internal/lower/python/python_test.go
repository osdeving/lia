package python

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/willams/lia/internal/ir"
	"github.com/willams/lia/internal/lower"
)

func TestLowerProject_Basic(t *testing.T) {
	prog := &ir.Program{
		Version:  "0.1",
		Projects: []ir.Project{{Name: "TestApp"}},
		Modules: []ir.Module{
			{
				Name: "orders.core",
				Role: "domain",
				Types: []ir.TypeDecl{
					{Name: "OrderId", Base: "String", Predicate: "nonEmpty"},
				},
				Enums: []ir.EnumDecl{
					{Name: "Status", Values: []string{"NEW", "PAID", "SHIPPED"}},
				},
			},
		},
	}
	project, err := LowerProject(prog)
	if err != nil {
		t.Fatalf("LowerProject failed: %v", err)
	}
	if project.Name != "testapp" {
		t.Errorf("expected project name testapp, got %s", project.Name)
	}
	if len(project.Files) == 0 {
		t.Fatal("expected files")
	}
	// Check for expected files
	fileSet := map[string]bool{}
	for _, f := range project.Files {
		fileSet[f.Path] = true
	}
	expectedFiles := []string{
		"README.md",
		"requirements.txt",
		"__init__.py",
		"orders/core/__init__.py",
		"orders/core/order_id.py",
		"orders/core/status.py",
	}
	for _, ef := range expectedFiles {
		if !fileSet[ef] {
			t.Errorf("missing expected file: %s", ef)
		}
	}
}

func TestLowerProject_NilProgram(t *testing.T) {
	_, err := LowerProject(nil)
	if err == nil {
		t.Fatal("expected error for nil program")
	}
}

func TestLowerProject_WithPorts(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name: "orders.port",
				Role: "port",
				Ports: []ir.PortDecl{
					{
						Name: "OrderRepository",
						Methods: []ir.FuncDecl{
							{
								Name:    "FindById",
								Params:  []ir.Field{{Name: "id", Type: "String"}},
								Returns: []ir.Field{{Name: "order", Type: "String"}},
							},
						},
					},
				},
			},
		},
	}
	project, err := LowerProject(prog)
	if err != nil {
		t.Fatalf("LowerProject failed: %v", err)
	}
	// Find the port file
	var portContent string
	for _, f := range project.Files {
		if strings.Contains(f.Path, "order_repository.py") {
			portContent = string(f.Content)
			break
		}
	}
	if portContent == "" {
		t.Fatal("expected order_repository.py")
	}
	if !strings.Contains(portContent, "ABC") {
		t.Error("expected ABC import")
	}
	if !strings.Contains(portContent, "abstractmethod") {
		t.Error("expected abstractmethod")
	}
	if !strings.Contains(portContent, "def find_by_id") {
		t.Error("expected find_by_id method")
	}
}

func TestLowerProject_WithUsecase(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name: "orders.app",
				Role: "usecase",
				Usecases: []ir.UsecaseDecl{
					{
						Name:    "CreateOrder",
						Inputs:  []ir.Field{{Name: "id", Type: "String"}},
						Outputs: []ir.Field{{Name: "ok", Type: "Bool"}},
						Effects: []string{"io", "tx"},
					},
				},
			},
		},
	}
	project, err := LowerProject(prog)
	if err != nil {
		t.Fatalf("LowerProject failed: %v", err)
	}
	var ucContent string
	for _, f := range project.Files {
		if strings.Contains(f.Path, "create_order.py") {
			ucContent = string(f.Content)
			break
		}
	}
	if ucContent == "" {
		t.Fatal("expected create_order.py")
	}
	if !strings.Contains(ucContent, "CreateOrderInput") {
		t.Error("expected CreateOrderInput dataclass")
	}
	if !strings.Contains(ucContent, "CreateOrderOutput") {
		t.Error("expected CreateOrderOutput dataclass")
	}
	if !strings.Contains(ucContent, "def execute") {
		t.Error("expected execute method")
	}
}

func TestLowerProject_WithAdapter(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name: "orders.infra",
				Role: "adapter",
				Adapters: []ir.AdapterDecl{
					{
						Name:       "PgOrderRepo",
						Implements: "orders.port::port:OrderRepository",
					},
				},
			},
		},
	}
	project, err := LowerProject(prog)
	if err != nil {
		t.Fatalf("LowerProject failed: %v", err)
	}
	var adapterContent string
	for _, f := range project.Files {
		if strings.Contains(f.Path, "pg_order_repo.py") {
			adapterContent = string(f.Content)
			break
		}
	}
	if adapterContent == "" {
		t.Fatal("expected pg_order_repo.py")
	}
	if !strings.Contains(adapterContent, "orders.port::port:OrderRepository") {
		t.Error("expected implements reference")
	}
}

func TestLowerProject_WithRecords(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name: "data.core",
				Role: "domain",
				Records: []ir.RecordDecl{
					{
						Name: "UserDTO",
						Fields: []ir.Field{
							{Name: "id", Type: "String"},
							{Name: "name", Type: "String"},
							{Name: "age", Type: "Int"},
						},
					},
				},
			},
		},
	}
	project, err := LowerProject(prog)
	if err != nil {
		t.Fatalf("LowerProject failed: %v", err)
	}
	var recContent string
	for _, f := range project.Files {
		if strings.Contains(f.Path, "user_d_t_o.py") {
			recContent = string(f.Content)
			break
		}
	}
	if recContent == "" {
		t.Fatal("expected user_d_t_o.py (or similar)")
	}
	if !strings.Contains(recContent, "@dataclass") {
		t.Error("expected @dataclass decorator")
	}
}

func TestLower_BackwardCompat(t *testing.T) {
	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{Name: "test.core", Role: "domain"},
		},
	}
	data, err := Lower(prog)
	if err != nil {
		t.Fatalf("Lower failed: %v", err)
	}
	if !strings.Contains(string(data), "# LIA lower (python)") {
		t.Error("expected header")
	}
}

func TestMapIRType(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"String", "str"},
		{"Int", "int"},
		{"Integer", "int"},
		{"Long", "int"},
		{"Float", "float"},
		{"Double", "float"},
		{"Decimal", "float"},
		{"Bool", "bool"},
		{"Boolean", "bool"},
		{"Byte", "bytes"},
		{"Void", "None"},
		{"Custom", "Custom"},
		{"list<String>", "list[str]"},
		{"option<Int>", "int | None"},
	}
	for _, tc := range tests {
		got := mapIRType(tc.input)
		if got != tc.want {
			t.Errorf("mapIRType(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestPythonZeroValue(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"str", "\"\""},
		{"int", "0"},
		{"float", "0.0"},
		{"bool", "False"},
		{"None", "None"},
		{"Custom", "None"},
	}
	for _, tc := range tests {
		got := pythonZeroValue(tc.input)
		if got != tc.want {
			t.Errorf("pythonZeroValue(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"orderId", "order_id"},
		{"OrderId", "order_id"},
		{"ABC", "a_b_c"},
		{"simple", "simple"},
		{"", ""},
	}
	for _, tc := range tests {
		got := snakeCase(tc.input)
		if got != tc.want {
			t.Errorf("snakeCase(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestSanitizePythonModule(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"MyApp", "myapp"},
		{"my-app", "my_app"},
		{"my.app", "my_app"},
		{"", "lia_generated"},
	}
	for _, tc := range tests {
		got := sanitizePythonModule(tc.input)
		if got != tc.want {
			t.Errorf("sanitizePythonModule(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestModulePythonPackage(t *testing.T) {
	got := modulePythonPackage("orders.core")
	if got != "orders/core" {
		t.Errorf("expected orders/core, got %s", got)
	}
}

func TestWriteProject(t *testing.T) {
	project := &Project{
		Name: "test",
		Files: []File{
			{Path: "test.py", Content: []byte("print('hello')")},
		},
	}
	dir := t.TempDir()
	err := WriteProject(dir, project)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := filepath.Glob(filepath.Join(dir, "test.py")); err != nil {
		t.Fatal(err)
	}
}

func TestWriteProject_Nil(t *testing.T) {
	err := WriteProject(t.TempDir(), nil)
	if err == nil {
		t.Fatal("expected error for nil project")
	}
}

func TestPythonBackend_Name(t *testing.T) {
	b := &PythonBackend{}
	if b.Name() != "python" {
		t.Errorf("expected python, got %s", b.Name())
	}
}

func TestPythonBackend_MapType(t *testing.T) {
	b := &PythonBackend{}
	mt := b.MapType("String")
	if mt.Name != "str" {
		t.Errorf("expected str, got %s", mt.Name)
	}
}

func TestPythonBackend_LowerProject(t *testing.T) {
	b := &PythonBackend{}
	prog := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{{Name: "test.core", Role: "domain"}},
	}
	proj, err := b.LowerProject(prog, lower.Options{})
	if err != nil {
		t.Fatalf("LowerProject failed: %v", err)
	}
	if proj.Language != "python" {
		t.Errorf("expected python, got %s", proj.Language)
	}
}

func TestPythonBackend_RegisteredInRegistry(t *testing.T) {
	b, err := lower.Get("python")
	if err != nil {
		t.Fatalf("Python backend not registered: %v", err)
	}
	if b.Name() != "python" {
		t.Errorf("expected python, got %s", b.Name())
	}
}

func TestRenderEnumDecl_Empty(t *testing.T) {
	decl := ir.EnumDecl{Name: "Empty"}
	result := renderEnumDecl(decl)
	if !strings.Contains(result, "pass") {
		t.Error("expected pass for empty enum")
	}
}

func TestRenderPortDecl_Empty(t *testing.T) {
	decl := ir.PortDecl{Name: "EmptyPort"}
	result := renderPortDecl(decl)
	if !strings.Contains(result, "pass") {
		t.Error("expected pass for empty port")
	}
}

func TestRenderPortDecl_MultiReturn(t *testing.T) {
	decl := ir.PortDecl{
		Name: "MultiPort",
		Methods: []ir.FuncDecl{
			{
				Name:    "Get",
				Returns: []ir.Field{{Name: "a", Type: "String"}, {Name: "b", Type: "Int"}},
			},
		},
	}
	result := renderPortDecl(decl)
	if !strings.Contains(result, "tuple[str, int]") {
		t.Errorf("expected tuple return, got: %s", result)
	}
}

func TestRenderRecordDecl_Empty(t *testing.T) {
	decl := ir.RecordDecl{Name: "Empty"}
	result := renderRecordDecl(decl)
	if !strings.Contains(result, "pass") {
		t.Error("expected pass for empty record")
	}
}

func TestRenderTypeDecl_WithPredicate(t *testing.T) {
	decl := ir.TypeDecl{Name: "Email", Base: "String", Predicate: "custom_pred"}
	result := renderTypeDecl(decl)
	if !strings.Contains(result, "Predicate: custom_pred") {
		t.Error("expected predicate comment")
	}
}

func TestRenderUsecaseDecl_Empty(t *testing.T) {
	decl := ir.UsecaseDecl{Name: "Noop"}
	result := renderUsecaseDecl(decl)
	if !strings.Contains(result, "NoopInput") {
		t.Error("expected NoopInput")
	}
	if !strings.Contains(result, "NoopOutput") {
		t.Error("expected NoopOutput")
	}
}
