package lower

import (
	"testing"

	"github.com/willams/lia/internal/ir"
)

// mockBackend is a test implementation of Backend.
type mockBackend struct {
	name     string
	profiles []string
}

func (m *mockBackend) Name() string { return m.name }
func (m *mockBackend) LowerProject(p *ir.Program, opts Options) (*Project, error) {
	return &Project{Name: "mock", Language: m.name, Files: []File{{Path: "test.txt", Content: []byte("test")}}}, nil
}
func (m *mockBackend) WriteProject(dir string, project *Project) error { return nil }
func (m *mockBackend) MapType(irType string) MappedType {
	return MappedType{Name: irType, ZeroValue: "nil"}
}
func (m *mockBackend) SupportedProfiles() []string { return m.profiles }

func TestRegister_And_Get(t *testing.T) {
	// Clean up registry for test isolation
	oldRegistry := Registry
	Registry = map[string]Backend{}
	defer func() { Registry = oldRegistry }()

	mock := &mockBackend{name: "test-lang", profiles: []string{"default"}}
	Register(mock)

	got, err := Get("test-lang")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Name() != "test-lang" {
		t.Errorf("expected test-lang, got %s", got.Name())
	}
}

func TestGet_NotFound(t *testing.T) {
	oldRegistry := Registry
	Registry = map[string]Backend{}
	defer func() { Registry = oldRegistry }()

	_, err := Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent backend")
	}
}

func TestBackends(t *testing.T) {
	oldRegistry := Registry
	Registry = map[string]Backend{}
	defer func() { Registry = oldRegistry }()

	Register(&mockBackend{name: "a"})
	Register(&mockBackend{name: "b"})

	names := Backends()
	if len(names) != 2 {
		t.Errorf("expected 2 backends, got %d", len(names))
	}
}

func TestMockBackend_LowerProject(t *testing.T) {
	mock := &mockBackend{name: "mock", profiles: []string{"default"}}
	prog := &ir.Program{Version: "0.1"}
	proj, err := mock.LowerProject(prog, Options{})
	if err != nil {
		t.Fatalf("LowerProject failed: %v", err)
	}
	if proj.Name != "mock" {
		t.Errorf("expected project name mock, got %s", proj.Name)
	}
	if proj.Language != "mock" {
		t.Errorf("expected language mock, got %s", proj.Language)
	}
	if len(proj.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(proj.Files))
	}
}

func TestMockBackend_MapType(t *testing.T) {
	mock := &mockBackend{name: "mock"}
	mapped := mock.MapType("String")
	if mapped.Name != "String" {
		t.Errorf("expected String, got %s", mapped.Name)
	}
	if mapped.ZeroValue != "nil" {
		t.Errorf("expected nil zero value, got %s", mapped.ZeroValue)
	}
}

func TestOptions_Fields(t *testing.T) {
	opts := Options{Profile: "spring-boot", Strict: true}
	if opts.Profile != "spring-boot" {
		t.Errorf("unexpected profile: %s", opts.Profile)
	}
	if !opts.Strict {
		t.Error("expected strict")
	}
}

func TestProject_Fields(t *testing.T) {
	proj := Project{
		Name:        "test",
		BasePackage: "com.test",
		Language:    "java",
		Files: []File{
			{Path: "a.java", Content: []byte("class A {}")},
		},
	}
	if proj.Name != "test" {
		t.Errorf("unexpected name: %s", proj.Name)
	}
	if len(proj.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(proj.Files))
	}
}

func TestMappedType_Fields(t *testing.T) {
	mt := MappedType{
		Name:      "String",
		Imports:   []string{"java.lang.String"},
		ZeroValue: "\"\"",
	}
	if mt.Name != "String" || len(mt.Imports) != 1 || mt.ZeroValue != "\"\"" {
		t.Error("unexpected MappedType fields")
	}
}

func TestRenderContext_Fields(t *testing.T) {
	ctx := RenderContext{CurrentModule: "orders.app"}
	if ctx.CurrentModule != "orders.app" {
		t.Errorf("unexpected module: %s", ctx.CurrentModule)
	}
}
