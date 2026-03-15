package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// goldenTest is a golden test case for parser validation.
type goldenTest struct {
	Name    string `json:"name"`
	Input   string `json:"input"`
	WantErr bool   `json:"want_err,omitempty"`
}

func loadGoldenTests(t *testing.T) []goldenTest {
	t.Helper()
	path := filepath.Join("testdata", "golden.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("no golden testdata: %v", err)
	}
	var tests []goldenTest
	if err := json.Unmarshal(data, &tests); err != nil {
		t.Fatalf("invalid golden.json: %v", err)
	}
	return tests
}

func TestGolden_FullGrammar(t *testing.T) {
	tests := loadGoldenTests(t)
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			_, err := ParseString(tc.Input)
			if tc.WantErr && err == nil {
				t.Errorf("expected parse error")
			}
			if !tc.WantErr && err != nil {
				t.Errorf("unexpected parse error: %v", err)
			}
		})
	}
}

// TestGolden_Inline tests various grammar constructs inline.
func TestGolden_Inline(t *testing.T) {
	cases := []goldenTest{
		{Name: "minimal_module", Input: "module test.core as domain {}"},
		{Name: "type_alias", Input: "module t as domain { type Id = String; }"},
		{Name: "type_with_where", Input: "module t as domain { type Email = String where nonEmpty; }"},
		{Name: "enum_basic", Input: "module t as domain { enum Color { RED, GREEN, BLUE }; }"},
		{Name: "port_with_method", Input: `module t as port {
			port Repo {
				fn Get(id: String) -> (val: String);
			}
		}`},
		{Name: "usecase_full", Input: `module t as usecase {
			usecase Create {
				input { id: String };
				output { ok: Bool };
				effects [io];
				return true;
			}
		}`},
		{Name: "adapter_with_impl", Input: `module t as adapter {
			adapter PgRepo implements port:Repo {
				input { conn: String };
			}
		}`},
		{Name: "wiring_basic", Input: `module t as wiring {
			wiring AppWiring {
				bind port:Repo -> adapter:PgRepo;
			}
		}`},
		{Name: "constraint", Input: `module t as domain {
			constraint no_io: forbid(effect io);
		}`},
		{Name: "prefer_basic", Input: `module t as domain {
			prefer fast: latency < 100;
		}`},
		{Name: "hole_basic", Input: `module t as domain {
			hole paymentGateway: payment gateway contract;
		}`},
		{Name: "project_with_modules", Input: `project Demo {
			repro strict;
			module demo.core as domain {
				type Id = String;
			}
		}`},
		{Name: "pack_basic", Input: `pack TestPack@1.0 {
			policy p: allow_roles ["domain"];
		}`},
		{Name: "gen_meta", Input: `@gen { prompt_ref:"p1", model_id:"m1" }
module gen.core as domain { type Id = String; }`},
		{Name: "invalid_missing_brace", Input: `module bad { type X =`, WantErr: true},
		{Name: "invalid_unknown_keyword", Input: `foobar { }`, WantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			_, err := ParseString(tc.Input)
			if tc.WantErr && err == nil {
				t.Errorf("expected parse error")
			}
			if !tc.WantErr && err != nil {
				t.Errorf("unexpected parse error: %v", err)
			}
		})
	}
}
