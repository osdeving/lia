package parser

import "testing"

func TestLIAKeywords_AllExpectedPresent(t *testing.T) {
	expected := []string{
		"project", "module", "pack", "as",
		"type", "enum", "port", "usecase", "adapter", "wiring", "record",
		"input", "output", "effects",
		"implements", "where", "use",
		"repro", "tape", "policy", "constraint", "prefer", "hole", "candidate",
		"let", "return", "if", "else", "while", "for", "in", "break", "continue",
		"fn", "method", "bind", "score", "weight",
	}
	for _, kw := range expected {
		if !IsLIAKeyword(kw) {
			t.Errorf("expected %q to be a keyword", kw)
		}
	}
}

func TestIsLIAKeyword_NegativeCases(t *testing.T) {
	negatives := []string{
		"", "string", "int", "bool", "foo", "OrderId", "main", "var", "func",
		"import", "package", "class", "interface", "struct",
	}
	for _, s := range negatives {
		if IsLIAKeyword(s) {
			t.Errorf("expected %q NOT to be a keyword", s)
		}
	}
}

func TestLIATypeTerminators_Completeness(t *testing.T) {
	terms := LIATypeTerminators()

	// All type terminators must also be valid keywords
	for kw := range terms {
		if !IsLIAKeyword(kw) {
			t.Errorf("type terminator %q is not in LIAKeywords", kw)
		}
	}

	// The original hardcoded list from TypeRef.Parse must be a subset
	originalHardcoded := []string{
		"as", "usecase", "adapter", "port", "wiring", "use", "project", "module",
		"pack", "repro", "tape", "policy", "constraint", "prefer", "hole",
		"candidate", "input", "output", "effects", "implements", "type", "enum", "where",
	}
	for _, kw := range originalHardcoded {
		if !terms[kw] {
			t.Errorf("original terminator %q is missing from LIATypeTerminators()", kw)
		}
	}
}

func TestLIATypeTerminators_IsFresh(t *testing.T) {
	// Each call returns a fresh map (can be modified without affecting others)
	a := LIATypeTerminators()
	b := LIATypeTerminators()
	a["fake_keyword"] = true
	if b["fake_keyword"] {
		t.Error("LIATypeTerminators() should return a fresh map each time")
	}
}

func TestLIAKeywords_MinimumCount(t *testing.T) {
	if len(LIAKeywords) < 35 {
		t.Errorf("expected at least 35 keywords, got %d", len(LIAKeywords))
	}
}
