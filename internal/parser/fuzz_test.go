package parser

import (
	"testing"
)

// FuzzParser tests the robust parser against random inputs to ensure it doesn't crash.
func FuzzParser(f *testing.F) {
	// Seed with valid and invalid grammar constructs
	seeds := []string{
		"module test as domain {}",
		"project Demo { repro strict; }",
		"pack Test@1.0 { policy p: allow_roles [\"domain\"]; }",
		"module test { type Id = String; }",
		"module test { enum Status { A, B }; }",
		"module test { port Repo { fn Get(id: String) -> (val: String); } }",
		"module test { usecase Create { input { id: String }; } }",
		"module test { adapter Pg implements port:Repo {} }",
		"module test { wiring App { bind port:Repo -> adapter:Pg; } }",
		"module test { type X = ",
		"project { ",
		"// Comment only",
		"@gen { model_id: \"gpt-4\" } module test {}",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Just ensure that parsing doesn't panic
		ParseString(input)
		ParseWithRecovery(input)
	})
}
