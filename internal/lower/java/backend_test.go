package java

import (
	"testing"

	"github.com/willams/lia/internal/ir"
	"github.com/willams/lia/internal/lower"
)

func TestJavaBackend_Name(t *testing.T) {
	b := &JavaBackend{}
	if b.Name() != "java" {
		t.Errorf("expected java, got %s", b.Name())
	}
}

func TestJavaBackend_SupportedProfiles(t *testing.T) {
	b := &JavaBackend{}
	profiles := b.SupportedProfiles()
	if len(profiles) != 3 {
		t.Fatalf("expected 3 profiles, got %d", len(profiles))
	}
	expected := map[string]bool{"plain": true, "spring-boot": true, "quarkus": true}
	for _, p := range profiles {
		if !expected[p] {
			t.Errorf("unexpected profile: %s", p)
		}
	}
}

func TestJavaBackend_MapType(t *testing.T) {
	b := &JavaBackend{}
	tests := []struct {
		irType   string
		wantName string
		wantZero string
	}{
		{"String", "String", "\"\""},
		{"Int", "int", "0"},
		{"Bool", "boolean", "false"},
		{"Long", "long", "0L"},
		{"Float", "float", "0.0f"},
		{"Double", "double", "0.0"},
		{"Byte", "byte", "0"},
		{"Void", "void", ""},
		{"Custom", "Custom", "null"},
	}
	for _, tc := range tests {
		got := b.MapType(tc.irType)
		if got.Name != tc.wantName {
			t.Errorf("MapType(%q).Name = %q, want %q", tc.irType, got.Name, tc.wantName)
		}
		if got.ZeroValue != tc.wantZero {
			t.Errorf("MapType(%q).ZeroValue = %q, want %q", tc.irType, got.ZeroValue, tc.wantZero)
		}
	}
}

func TestJavaBackend_LowerProject(t *testing.T) {
	b := &JavaBackend{}
	prog := &ir.Program{
		Version: "0.1",
		Projects: []ir.Project{{Name: "Test"}},
		Modules: []ir.Module{
			{
				Name: "test.core",
				Role: "domain",
				Types: []ir.TypeDecl{{Name: "Id", Base: "String"}},
				Enums: []ir.EnumDecl{{Name: "Status", Values: []string{"NEW", "DONE"}}},
			},
		},
	}
	project, err := b.LowerProject(prog, lower.Options{})
	if err != nil {
		t.Fatalf("LowerProject failed: %v", err)
	}
	if project.Name != "Test" {
		t.Errorf("expected project name Test, got %s", project.Name)
	}
	if project.Language != "java" {
		t.Errorf("expected language java, got %s", project.Language)
	}
	if len(project.Files) == 0 {
		t.Error("expected at least one file")
	}
}

func TestJavaBackend_LowerProject_WithProfile(t *testing.T) {
	b := &JavaBackend{}
	prog := &ir.Program{
		Version: "0.1",
		Projects: []ir.Project{{Name: "Demo"}},
		Modules: []ir.Module{
			{Name: "demo.core", Role: "domain"},
		},
	}
	project, err := b.LowerProject(prog, lower.Options{Profile: "spring-boot"})
	if err != nil {
		t.Fatalf("LowerProject spring-boot failed: %v", err)
	}
	if project.Language != "java" {
		t.Errorf("expected java language, got %s", project.Language)
	}
}

func TestJavaBackend_LowerProject_InvalidProfile(t *testing.T) {
	b := &JavaBackend{}
	prog := &ir.Program{Version: "0.1"}
	_, err := b.LowerProject(prog, lower.Options{Profile: "invalid"})
	if err == nil {
		t.Fatal("expected error for invalid profile")
	}
}

func TestJavaBackend_WriteProject(t *testing.T) {
	b := &JavaBackend{}
	dir := t.TempDir()
	project := &lower.Project{
		Name:     "test",
		Language: "java",
		Files: []lower.File{
			{Path: "Test.java", Content: []byte("class Test {}")},
		},
	}
	err := b.WriteProject(dir, project)
	if err != nil {
		t.Fatalf("WriteProject failed: %v", err)
	}
}

func TestJavaBackend_RegisteredInRegistry(t *testing.T) {
	b, err := lower.Get("java")
	if err != nil {
		t.Fatalf("Java backend not registered: %v", err)
	}
	if b.Name() != "java" {
		t.Errorf("expected java, got %s", b.Name())
	}
}

func TestResolveTypeSimple(t *testing.T) {
	tests := []struct {
		input    string
		wantName string
	}{
		{"String", "String"},
		{"Integer", "int"},
		{"Boolean", "boolean"},
		{"Unknown", "Unknown"},
	}
	for _, tc := range tests {
		got := resolveTypeSimple(tc.input)
		if got.Name != tc.wantName {
			t.Errorf("resolveTypeSimple(%q).Name = %q, want %q", tc.input, got.Name, tc.wantName)
		}
	}
}
