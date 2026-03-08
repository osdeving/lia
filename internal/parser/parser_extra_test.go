package parser

import (
	"os"
	"testing"
)

func TestParseFile_EmptyBlocks(t *testing.T) {
	content := `module empty.core as usecase {
  usecase Empty {
    input { };
    output { };
    effects [];
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
	uc := prog.Modules[0].Usecases[0]
	if len(uc.Inputs) != 0 || len(uc.Outputs) != 0 {
		t.Fatalf("expected empty input/output")
	}
	if len(uc.Effects) != 0 {
		t.Fatalf("expected empty effects")
	}
}

func TestParseFile_TypeRefs(t *testing.T) {
	content := `module types.core as domain {
  type Result = result<String, Error>;
  type List = []String;
}
`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	prog, err := ParseFile(tmpfile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(prog.Modules[0].Types) != 2 {
		t.Fatalf("expected 2 types")
	}
	if prog.Modules[0].Types[0].Base == "" {
		t.Fatalf("expected base type")
	}
}

func TestParseFile_ProjectTapeIdentAndPackStringVersion(t *testing.T) {
	content := `project Demo {
  tape tapejson;
  use pack Foo@"1.0.0";
  module demo.mod as domain {
  }
}
`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	prog, err := ParseFile(tmpfile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(prog.Projects) != 1 {
		t.Fatalf("expected 1 project")
	}
	proj := prog.Projects[0]
	if proj.Tape != "tapejson" {
		t.Fatalf("unexpected tape value: %s", proj.Tape)
	}
	if len(proj.Uses) != 1 || proj.Uses[0].Version != "1.0.0" {
		t.Fatalf("unexpected pack version")
	}
}

func TestParseFile_InvalidTypeDecl(t *testing.T) {
	content := `module bad.type as domain {
  type = String;
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for invalid type decl")
	}
}

func TestParseFile_InvalidInputBlock(t *testing.T) {
	content := `module bad.input as usecase {
  usecase Bad {
    input;
  }
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for invalid input block")
	}
}

func TestParseFile_InvalidWiring(t *testing.T) {
	content := `module bad.wiring as wiring {
  wiring W {
    bind Foo;
  }
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for invalid wiring")
	}
}

func TestParseFile_InvalidMethodDecl(t *testing.T) {
	content := `module bad.port as port {
  port P {
    fn Ping(;
  }
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for invalid method")
	}
}

func TestParseFile_InvalidTypeMissingSemicolon(t *testing.T) {
	content := `module bad.type2 as domain {
  type Email = String
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for missing semicolon")
	}
}

func TestParseFile_InvalidPolicy(t *testing.T) {
	content := `project Bad {
  policy roles allow_roles ["domain"];
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for invalid policy")
	}
}

func TestParseFile_InvalidConstraint(t *testing.T) {
	content := `module bad.constraint as domain {
  constraint c1 forbid(effect io);
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for invalid constraint")
	}
}

func TestParseFile_InvalidHole(t *testing.T) {
	content := `module bad.hole as domain {
  hole Missing;
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for invalid hole")
	}
}

func TestParseFile_GenMetaEmptyLists(t *testing.T) {
	content := `@gen { prompt_ref:"p3", prompt_hash:"h3", context_refs: [], tools_trace_refs: [], model_params: {} }
module top.empty as domain {
  usecase Noop {
    return true;
  }
}
`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
}
