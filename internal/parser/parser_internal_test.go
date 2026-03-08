package parser

import (
	"os"
	"testing"
)

func TestParseGenMeta_StringListSkipsComma(t *testing.T) {
	content := `@gen { context_refs: ["a",, "b"] }
module demo.core as domain { }
`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	prog, err := ParseFile(tmpfile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(prog.Modules) != 1 || prog.Modules[0].Gen == nil {
		t.Fatalf("expected module with @gen")
	}
	if len(prog.Modules[0].Gen.ContextRefs) != 2 {
		t.Fatalf("expected 2 context refs, got %d", len(prog.Modules[0].Gen.ContextRefs))
	}
}

func TestParseGenMeta_KVListSemicolons(t *testing.T) {
	content := `@gen { model_params: { a: 1; b: 2 } }
module demo.core as domain { }
`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	prog, err := ParseFile(tmpfile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(prog.Modules) != 1 || prog.Modules[0].Gen == nil {
		t.Fatalf("expected module with @gen")
	}
	if len(prog.Modules[0].Gen.ModelParams) != 2 {
		t.Fatalf("expected 2 model params, got %d", len(prog.Modules[0].Gen.ModelParams))
	}
}

func TestParseEffects_LeadingComma(t *testing.T) {
	content := `module effects.core as usecase {
  usecase Test {
    effects [ , io ];
    return true;
  }
}
`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	prog, err := ParseFile(tmpfile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(prog.Modules) != 1 || len(prog.Modules[0].Usecases) != 1 {
		t.Fatalf("expected 1 usecase")
	}
	if len(prog.Modules[0].Usecases[0].Effects) != 1 {
		t.Fatalf("expected 1 effect, got %d", len(prog.Modules[0].Usecases[0].Effects))
	}
}

func TestParseMethodDecl_Direct(t *testing.T) {
	content := `module port.core as port {
  port P {
    fn Ping() -> (ok: Bool);
  }
}
`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	prog, err := ParseFile(tmpfile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(prog.Modules) != 1 || len(prog.Modules[0].Ports) != 1 {
		t.Fatalf("expected 1 port")
	}
	meth := prog.Modules[0].Ports[0].Methods
	if len(meth) != 1 || meth[0].Name != "Ping" {
		t.Fatalf("unexpected method name")
	}
}

func TestParsePrefer_Weight(t *testing.T) {
	content := `module pref.core as domain {
  prefer prefer_repo: orders.port::port:Repo weight 0.7;
}
`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	prog, err := ParseFile(tmpfile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(prog.Modules) != 1 || len(prog.Modules[0].Preferences) != 1 {
		t.Fatalf("expected 1 preference")
	}
	pref := prog.Modules[0].Preferences[0]
	if pref.Weight < 0.69 || pref.Weight > 0.71 {
		t.Fatalf("unexpected weight: %f", pref.Weight)
	}
}
