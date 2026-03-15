package parser

import (
	"strings"
	"testing"
)

func TestNormalizeLLMOutput_StripMarkdownFences(t *testing.T) {
	input := "```lia\nmodule test as domain {}\n```\n"
	result := NormalizeLLMOutput(input)
	if strings.Contains(result, "```") {
		t.Errorf("expected fences stripped, got: %s", result)
	}
	if !strings.Contains(result, "module test as domain {}") {
		t.Errorf("expected content preserved, got: %s", result)
	}
}

func TestNormalizeLLMOutput_StripMarkdownFencesLiaLang(t *testing.T) {
	input := "```lia-lang\nmodule test as domain {}\n```\n"
	result := NormalizeLLMOutput(input)
	if strings.Contains(result, "```") {
		t.Errorf("expected fences stripped, got: %s", result)
	}
}

func TestNormalizeLLMOutput_StripBareFences(t *testing.T) {
	input := "```\nmodule test as domain {}\n```\n"
	result := NormalizeLLMOutput(input)
	if strings.Contains(result, "```") {
		t.Errorf("expected bare fences stripped, got: %s", result)
	}
}

func TestNormalizeLLMOutput_FixFunctionToFn(t *testing.T) {
	input := "port MyPort {\n  function Get(id: String) -> (ok: Bool);\n}\n"
	result := NormalizeLLMOutput(input)
	if !strings.Contains(result, "fn Get") {
		t.Errorf("expected function→fn, got: %s", result)
	}
}

func TestNormalizeLLMOutput_FixDefToFn(t *testing.T) {
	input := "port MyPort {\n  def Get(id: String);\n}\n"
	result := NormalizeLLMOutput(input)
	if !strings.Contains(result, "fn Get") {
		t.Errorf("expected def→fn, got: %s", result)
	}
}

func TestNormalizeLLMOutput_FixClassToModule(t *testing.T) {
	input := "class OrderService {\n  type Id = String;\n}\n"
	result := NormalizeLLMOutput(input)
	if !strings.Contains(result, "module OrderService") {
		t.Errorf("expected class→module, got: %s", result)
	}
}

func TestNormalizeLLMOutput_FixStructToRecord(t *testing.T) {
	input := "struct UserDTO {\n  id: String;\n}\n"
	result := NormalizeLLMOutput(input)
	if !strings.Contains(result, "record UserDTO {") {
		t.Errorf("expected struct→record, got: %s", result)
	}
}

func TestNormalizeLLMOutput_FixInterfaceToPort(t *testing.T) {
	input := "interface Repository {\n  fn Get(id: String);\n}\n"
	result := NormalizeLLMOutput(input)
	if !strings.Contains(result, "port Repository {") {
		t.Errorf("expected interface→port, got: %s", result)
	}
}

func TestNormalizeLLMOutput_InsertMissingSemicolon(t *testing.T) {
	input := "module test as domain {\n  type Id = String\n}\n"
	result := NormalizeLLMOutput(input)
	if !strings.Contains(result, "String;") {
		t.Errorf("expected semicolon inserted before }, got: %s", result)
	}
}

func TestNormalizeLLMOutput_NoDoubleSemicolon(t *testing.T) {
	input := "module test as domain {\n  type Id = String;\n}\n"
	result := NormalizeLLMOutput(input)
	if strings.Contains(result, ";;") {
		t.Errorf("expected no double semicolons, got: %s", result)
	}
}

func TestNormalizeLLMOutput_StripTrailingComma(t *testing.T) {
	input := "enum Status { NEW, PAID, }\n"
	result := NormalizeLLMOutput(input)
	if strings.Contains(result, ", }") || strings.Contains(result, ",}") {
		t.Errorf("expected trailing comma removed, got: %s", result)
	}
}

func TestNormalizeLLMOutput_NormalizeLineEndings(t *testing.T) {
	input := "module test as domain {\r\n  type Id = String;\r\n}\r\n"
	result := NormalizeLLMOutput(input)
	if strings.Contains(result, "\r") {
		t.Errorf("expected no \\r in output, got: %s", result)
	}
}

func TestNormalizeLLMOutput_CollapseBlankLines(t *testing.T) {
	input := "module test as domain {\n\n\n\n  type Id = String;\n\n\n}\n"
	result := NormalizeLLMOutput(input)
	if strings.Contains(result, "\n\n\n") {
		t.Errorf("expected consecutive blank lines collapsed, got: %s", result)
	}
}

func TestNormalizeLLMOutput_EmptyInput(t *testing.T) {
	result := NormalizeLLMOutput("")
	if result != "\n" {
		t.Errorf("expected newline for empty input, got: %q", result)
	}
}

func TestNormalizeLLMOutput_PreservesValidCode(t *testing.T) {
	input := `module orders.app as usecase {
  type OrderId = String where nonEmpty;
  enum Status { NEW, PAID };

  usecase CreateOrder {
    input { id: OrderId };
    output { ok: Bool };
    effects [io];
    return true;
  }
}
`
	result := NormalizeLLMOutput(input)
	// Should be parseable
	_, err := ParseString(result)
	if err != nil {
		t.Fatalf("normalized valid code should still parse: %v", err)
	}
}

func TestNormalizeLLMOutput_ComplexLLMOutput(t *testing.T) {
	input := "```lia\nclass OrderService {\r\n  type OrderId = String\r\n  interface OrderRepo {\r\n    function Get(id: String) -> (ok: Bool);\n  }\n}\n```\n"
	result := NormalizeLLMOutput(input)
	// Should have: no fences, class→module, interface→port, function→fn, CRLF→LF
	if strings.Contains(result, "```") {
		t.Error("fences not stripped")
	}
	if strings.Contains(result, "\r") {
		t.Error("CRLF not normalized")
	}
	if strings.Contains(result, "class ") {
		t.Error("class not replaced")
	}
	if strings.Contains(result, "interface ") {
		t.Error("interface not replaced")
	}
	if strings.Contains(result, "function ") {
		t.Error("function not replaced")
	}
}

func TestNormalizeLineEndings(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"abc\ndef", "abc\ndef"},
		{"abc\r\ndef", "abc\ndef"},
		{"abc\rdef", "abc\ndef"},
		{"abc\r\ndef\r\n", "abc\ndef\n"},
	}
	for _, tc := range tests {
		got := normalizeLineEndings(tc.input)
		if got != tc.want {
			t.Errorf("normalizeLineEndings(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestStripMarkdownFences(t *testing.T) {
	tests := []struct {
		input    string
		contains string
		excludes string
	}{
		{"```lia\nhello\n```\n", "hello", "```"},
		{"```\nhello\n```\n", "hello", "```"},
		{"no fences here", "no fences here", "```"},
		{"```lia-lang\nhello\n```\n", "hello", "```"},
	}
	for _, tc := range tests {
		got := stripMarkdownFences(tc.input)
		if !strings.Contains(got, tc.contains) {
			t.Errorf("stripMarkdownFences(%q): missing %q", tc.input, tc.contains)
		}
		if tc.excludes != "" && strings.Contains(got, tc.excludes) {
			t.Errorf("stripMarkdownFences(%q): should not contain %q", tc.input, tc.excludes)
		}
	}
}

func TestFixKeywordAliases(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"function foo", "fn foo"},
		{"def bar", "fn bar"},
		{"class Svc {\n}", "module Svc {\n}"},
		{"struct DTO {", "record DTO {"},
		{"interface Repo {", "port Repo {"},
	}
	for _, tc := range tests {
		got := fixKeywordAliases(tc.input)
		if !strings.Contains(got, tc.want) {
			t.Errorf("fixKeywordAliases(%q) = %q, want to contain %q", tc.input, got, tc.want)
		}
	}
}

func TestInsertMissingSemicolons(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"  type Id = String\n}", "  type Id = String;\n}"},
		{"  type Id = String;\n}", "  type Id = String;\n}"},
		{"  {\n}", "  {\n}"},
	}
	for _, tc := range tests {
		got := insertMissingSemicolons(tc.input)
		if got != tc.want {
			t.Errorf("insertMissingSemicolons(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestStripTrailingCommasBeforeBraces(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"[a, b, ]", "[a, b ]"},
		{"{x, }", "{x }"},
		{"[a, b]", "[a, b]"},
	}
	for _, tc := range tests {
		got := stripTrailingCommasBeforeBraces(tc.input)
		if got != tc.want {
			t.Errorf("stripTrailingCommasBeforeBraces(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestNormalizeWhitespace(t *testing.T) {
	input := "  line1  \n\n\n\nline2  \n"
	got := normalizeWhitespace(input)
	if strings.Contains(got, "  \n") {
		t.Error("trailing spaces not trimmed")
	}
	if strings.Contains(got, "\n\n\n") {
		t.Error("consecutive blank lines not collapsed")
	}
}
