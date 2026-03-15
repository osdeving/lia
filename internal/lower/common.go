package lower

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// WriteProjectFiles writes all files in a project to disk under dir.
// This is the shared implementation used by all backends.
func WriteProjectFiles(dir string, files []File) error {
	for _, file := range files {
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

// SanitizeIdentifier makes a valid identifier by removing non-alphanumeric chars.
func SanitizeIdentifier(name string) string {
	var b strings.Builder
	hasLetter := false
	for _, r := range name {
		if unicode.IsLetter(r) || r == '_' {
			b.WriteRune(r)
			hasLetter = true
		} else if hasLetter && unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "_"
	}
	return b.String()
}

// ToUpperCamel converts a name to UpperCamelCase.
func ToUpperCamel(s string) string {
	if s == "" {
		return s
	}
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == '.' || r == ' '
	})
	var b strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		b.WriteString(strings.ToUpper(part[:1]) + part[1:])
	}
	result := b.String()
	if result == "" {
		return s
	}
	return result
}

// ToLowerCamel converts a name to lowerCamelCase.
func ToLowerCamel(s string) string {
	upper := ToUpperCamel(s)
	if upper == "" {
		return upper
	}
	return strings.ToLower(upper[:1]) + upper[1:]
}

// BuiltinTypeMap provides common IR type → language type mappings.
// Each backend can extend or override these.
var BuiltinTypeMap = map[string]string{
	"String":  "String",
	"Int":     "int",
	"Integer": "int",
	"Long":    "long",
	"Float":   "float",
	"Double":  "double",
	"Bool":    "boolean",
	"Boolean": "boolean",
	"Byte":    "byte",
	"Void":    "void",
}

// IsBuiltinType checks if a type name is a known primitive/builtin.
func IsBuiltinType(name string) bool {
	_, ok := BuiltinTypeMap[name]
	return ok
}

// SanitizePackageSegment cleans a name for use in package paths.
func SanitizePackageSegment(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "unnamed"
	}
	return b.String()
}
