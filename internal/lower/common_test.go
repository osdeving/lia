package lower

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteProjectFiles(t *testing.T) {
	dir := t.TempDir()
	files := []File{
		{Path: "src/main/Hello.java", Content: []byte("class Hello {}")},
		{Path: "README.md", Content: []byte("# Test")},
	}
	err := WriteProjectFiles(dir, files)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		path := filepath.Join(dir, filepath.FromSlash(f.Path))
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read %s: %v", f.Path, err)
		}
		if string(data) != string(f.Content) {
			t.Errorf("unexpected content for %s", f.Path)
		}
	}
}

func TestWriteProjectFiles_Empty(t *testing.T) {
	dir := t.TempDir()
	err := WriteProjectFiles(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSanitizeIdentifier(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"Hello_World", "Hello_World"},
		{"123abc", "abc"},
		{"hello-world", "helloworld"},
		{"", "_"},
		{"test.name", "testname"},
		{"_valid", "_valid"},
	}
	for _, tc := range tests {
		got := SanitizeIdentifier(tc.input)
		if got != tc.want {
			t.Errorf("SanitizeIdentifier(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestToUpperCamel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "Hello"},
		{"hello_world", "HelloWorld"},
		{"hello-world", "HelloWorld"},
		{"hello.world", "HelloWorld"},
		{"", ""},
		{"already", "Already"},
		{"a_b_c", "ABC"},
	}
	for _, tc := range tests {
		got := ToUpperCamel(tc.input)
		if got != tc.want {
			t.Errorf("ToUpperCamel(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestToLowerCamel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"Hello_World", "helloWorld"},
		{"hello-world", "helloWorld"},
		{"", ""},
	}
	for _, tc := range tests {
		got := ToLowerCamel(tc.input)
		if got != tc.want {
			t.Errorf("ToLowerCamel(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestIsBuiltinType(t *testing.T) {
	builtins := []string{"String", "Int", "Bool", "Float", "Double", "Long", "Byte", "Void", "Integer", "Boolean"}
	for _, b := range builtins {
		if !IsBuiltinType(b) {
			t.Errorf("expected %q to be builtin", b)
		}
	}
	nonBuiltins := []string{"OrderId", "UserDTO", "Custom", ""}
	for _, nb := range nonBuiltins {
		if IsBuiltinType(nb) {
			t.Errorf("expected %q NOT to be builtin", nb)
		}
	}
}

func TestSanitizePackageSegment(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello", "hello"},
		{"my-app", "myapp"},
		{"test_123", "test_123"},
		{"", "unnamed"},
		{"...!!!", "unnamed"},
	}
	for _, tc := range tests {
		got := SanitizePackageSegment(tc.input)
		if got != tc.want {
			t.Errorf("SanitizePackageSegment(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
