package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/willams/lia/internal/ir"
)

func TestParseFile_ProjectModuleCore(t *testing.T) {
	content := `project Demo {
  repro strict;
  tape prompt_tape "./tape.json";
  use pack ArchBaseline@0.1.0;

  constraint safe: forbid(effect io) when role == "domain";

  @gen { prompt_ref:"p-001", prompt_hash:"abc", model_id:"gpt-x" }
  module orders.port as port {
    port OrderRepository {
      fn Get(id: String) -> (order: String);
    }
  }

  module orders.app as usecase {
    type OrderId = String where nonEmpty;
    enum Status { NEW, PAID };

    usecase CreateOrder {
      input { id: OrderId };
      output { ok: Bool };
      effects [io, tx];
      let ok = true;
      return ok;
    }

    adapter OrdersDb implements orders.port::port:OrderRepository {
      effects [io];
      return true;
    }

    wiring OrdersWiring {
      bind orders.port::port:OrderRepository -> orders.app::adapter:OrdersDb;
    }

    prefer prefer_repo: orders.port::port:OrderRepository weight 0.7;
  }
}
`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	prog, err := ParseFile(tmpfile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if prog.Version != "0.1" {
		t.Errorf("expected version 0.1, got %s", prog.Version)
	}
	if len(prog.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(prog.Projects))
	}
	proj := prog.Projects[0]
	if proj.Repro != ir.ReproStrict {
		t.Errorf("expected repro strict, got %s", proj.Repro)
	}
	if proj.Tape != "./tape.json" {
		t.Errorf("expected tape ./tape.json, got %s", proj.Tape)
	}
	if len(proj.Uses) != 1 {
		t.Fatalf("expected 1 pack ref, got %d", len(proj.Uses))
	}
	if proj.Uses[0].Name != "ArchBaseline" || proj.Uses[0].Version != "0.1.0" {
		t.Errorf("unexpected pack ref: %s@%s", proj.Uses[0].Name, proj.Uses[0].Version)
	}

	if len(prog.Modules) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(prog.Modules))
	}

	var app *ir.Module
	var port *ir.Module
	for i := range prog.Modules {
		if prog.Modules[i].Name == "orders.app" {
			app = &prog.Modules[i]
		}
		if prog.Modules[i].Name == "orders.port" {
			port = &prog.Modules[i]
		}
	}
	if app == nil || port == nil {
		t.Fatalf("expected orders.app and orders.port modules")
	}

	if app.Gen != nil {
		t.Errorf("expected no @gen on orders.app")
	}
	if port.Gen == nil || port.Gen.PromptRef != "p-001" {
		t.Fatalf("expected @gen on orders.port")
	}

	if len(app.Types) != 1 {
		t.Errorf("expected 1 type, got %d", len(app.Types))
	}
	if len(app.Enums) != 1 {
		t.Errorf("expected 1 enum, got %d", len(app.Enums))
	}
	if len(app.Usecases) != 1 {
		t.Fatalf("expected 1 usecase, got %d", len(app.Usecases))
	}
	if len(app.Usecases[0].Body) == 0 {
		t.Errorf("expected usecase body statements")
	}
	if len(app.Adapters) != 1 {
		t.Fatalf("expected 1 adapter, got %d", len(app.Adapters))
	}
	if app.Adapters[0].Implements != "orders.port::port:OrderRepository" {
		t.Errorf("unexpected adapter implements: %s", app.Adapters[0].Implements)
	}
	if len(app.Wirings) != 1 {
		t.Fatalf("expected 1 wiring, got %d", len(app.Wirings))
	}
	if len(app.Wirings[0].Binds) != 1 {
		t.Errorf("expected 1 bind, got %d", len(app.Wirings[0].Binds))
	}

	if len(port.Ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(port.Ports))
	}
	if len(port.Ports[0].Methods) != 1 {
		t.Fatalf("expected 1 port method, got %d", len(port.Ports[0].Methods))
	}

	if len(app.Requires) == 0 {
		t.Fatalf("expected derived requires")
	}
	foundReq := false
	for _, req := range app.Requires {
		if req.QName == "orders.port::port:OrderRepository" {
			foundReq = true
			break
		}
	}
	if !foundReq {
		t.Errorf("expected requires to include orders.port::port:OrderRepository")
	}

	if len(port.Provides) == 0 {
		t.Fatalf("expected derived provides")
	}
	foundProv := false
	for _, prov := range port.Provides {
		if prov.QName == "orders.port::port:OrderRepository" {
			foundProv = true
			break
		}
	}
	if !foundProv {
		t.Errorf("expected provides to include orders.port::port:OrderRepository")
	}
}

func TestParseFile_Statements(t *testing.T) {
	content := `module game.logic as usecase {
  usecase Guess {
    input { guess: Int };
    output { ok: Bool };
    effects [io];
    let target = 7;
    if guess == target { return true; } else { return false; }
    while target > 0 { target = target - 1; }
    for i in [1,2,3] { target = target + i; }
    callMe(target);
  }
}
`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	prog, err := ParseFile(tmpfile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(prog.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(prog.Modules))
	}
	uc := prog.Modules[0].Usecases
	if len(uc) != 1 {
		t.Fatalf("expected 1 usecase, got %d", len(uc))
	}
	if len(uc[0].Body) < 4 {
		t.Errorf("expected multiple statements, got %d", len(uc[0].Body))
	}
}

func TestParseFile_InvalidSyntax(t *testing.T) {
	content := `module bad.example { type X = String `
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestParseFile_FullGrammar(t *testing.T) {
	content := `project Demo {
  repro pinned;
  tape "./tape.json";
  use pack ArchBaseline@0.1.0;
  policy roles: allow_roles ["domain", "usecase"];

  @gen { prompt_ref:"p1", prompt_hash:"h1", model_id:"m1", context_refs:["a", "b"], tools_trace_refs:["t1"], model_params:{ temperature: 0.1, top_p: 0.9 } }
  module demo.core as usecase {
    type Email = String;
    enum Color { RED, GREEN };

    port Mailer {
      fn Send(to: String);
    }

    hole Missing: need.adapter;
    candidate demo.core::adapter:Impl score 0.5;
    candidate demo.core::adapter:Alt { score 1.2; constraint c1: forbid(effect io); };
    prefer choose_alt: demo.core::adapter:Alt weight 1.1;

    usecase Run {
      input { user: String };
      output { ok: Bool };
      effects ["io"];
      let i = 0;
      i = i + 1;
      if i > 0 { continue; } else { break; }
      while i < 3 { i = i + 1; }
      for x in [1,2,3] { i = i + x; }
      return !false;
    }

    adapter Impl implements demo.core::port:Mailer {
      effects [io];
      Send(user);
    }

    wiring Wiring {
      bind demo.core::port:Mailer -> demo.core::adapter:Impl;
    }
  }
}

pack DemoPack@1.2.3 {
  policy roles: allow_roles ["domain"];
  constraint c1: forbid(effect io);
}
`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	prog, err := ParseFile(tmpfile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(prog.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(prog.Projects))
	}
	if len(prog.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(prog.Modules))
	}
	if len(prog.Packs) != 1 {
		t.Fatalf("expected 1 pack, got %d", len(prog.Packs))
	}
	mod := prog.Modules[0]
	if mod.Gen == nil || mod.Gen.ModelID != "m1" {
		t.Fatalf("expected @gen model_id")
	}
	if len(mod.Holes) != 1 {
		t.Fatalf("expected 1 hole, got %d", len(mod.Holes))
	}
	if len(mod.Candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(mod.Candidates))
	}
	if len(mod.Preferences) != 1 {
		t.Fatalf("expected 1 preference, got %d", len(mod.Preferences))
	}
	if len(mod.Usecases) != 1 || len(mod.Usecases[0].Body) == 0 {
		t.Fatalf("expected usecase body statements")
	}
	if len(mod.Adapters) != 1 {
		t.Fatalf("expected adapter declaration")
	}
	if len(mod.Wirings) != 1 {
		t.Fatalf("expected wiring declaration")
	}
	if len(prog.Packs[0].Policies) != 1 {
		t.Fatalf("expected pack policy")
	}
}

func TestParseFile_ExpressionOperators(t *testing.T) {
	content := `module ops.core as usecase {
  // comment line
  /* block comment */
  usecase Ops {
    input { a: Int, b: Int };
    output { ok: Bool };
    effects [pure];
    let x = (a + b) * 2 - 3 / 4 % 5;
    let y = a == b || a != b && a <= b && a >= b && a < b && a > b;
    let z = -a;
    let w = !false;
    let s = "hi";
    let list = [1,2,3];
    let idx = list[0];
    let member = obj.field;
    let call = fnc(a, b).field;
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
	if len(prog.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(prog.Modules))
	}
	uc := prog.Modules[0].Usecases
	if len(uc) != 1 {
		t.Fatalf("expected 1 usecase, got %d", len(uc))
	}
	if len(uc[0].Body) < 8 {
		t.Fatalf("expected multiple statements, got %d", len(uc[0].Body))
	}
}

func TestParseFile_PackWithModule(t *testing.T) {
	content := `pack DemoPack@1.2.3 {
  module pack.mod as policy {
    constraint c1: forbid(effect io);
  }
  policy rules: allow_roles ["domain"];
  constraint c2: forbid(effect tx);
}
`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	prog, err := ParseFile(tmpfile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(prog.Packs) != 1 {
		t.Fatalf("expected 1 pack, got %d", len(prog.Packs))
	}
	if len(prog.Packs[0].Modules) != 1 {
		t.Fatalf("expected 1 pack module, got %d", len(prog.Packs[0].Modules))
	}
}

func TestParseFile_GenMetaTopLevel(t *testing.T) {
	content := `@gen { prompt_ref:"p2", prompt_hash:"h2", model_id:"m2", generator_pass:"pass-1", timestamp:"2026-01-01T00:00:00Z", model_params:"temp=0.1", unknown:"x" }
module top.gen as domain {
  usecase Noop {
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
	if len(prog.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(prog.Modules))
	}
	if prog.Modules[0].Gen == nil || prog.Modules[0].Gen.PromptRef != "p2" {
		t.Fatalf("expected @gen prompt_ref")
	}
}

func TestParseFile_AdapterFieldsNoImplements(t *testing.T) {
	content := `module adapter.test as adapter {
  port PingPort {
    fn Ping() -> (ok: Bool);
  }
  adapter FileAdapter {
    input { path: String };
    output { ok: Bool };
    effects [];
    return true;
  }
  wiring Empty {
  }
}
`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	prog, err := ParseFile(tmpfile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(prog.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(prog.Modules))
	}
	mod := prog.Modules[0]
	if len(mod.Adapters) != 1 {
		t.Fatalf("expected adapter")
	}
	if len(mod.Adapters[0].Inputs) != 1 || len(mod.Adapters[0].Outputs) != 1 {
		t.Fatalf("expected adapter input/output")
	}
}

func TestParseFile_FieldListError(t *testing.T) {
	content := `module bad.fields as usecase {
  usecase Bad {
    let x = fnc(1 2);
    return true;
  }
}
`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for invalid field list")
	}
}

// Helper: createTempFile creates a temporary .lia file for testing.
func createTempFile(t *testing.T, content string) string {
	t.Helper()
	tmpdir := t.TempDir()
	tmpfile := filepath.Join(tmpdir, "test.lia")
	if err := os.WriteFile(tmpfile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return tmpfile
}
