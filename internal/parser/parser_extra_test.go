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

func TestTokenName(t *testing.T) {
	if tokenName(tokenRParen) == "" || tokenName(tokenRBrace) == "" || tokenName(tokenRBracket) == "" {
		t.Fatalf("expected token names")
	}
	if tokenName(tokenSemicolon) != ";" {
		t.Fatalf("expected semicolon token name")
	}
	if tokenName(tokenEOF) == "" {
		t.Fatalf("expected default token name")
	}
}

func TestPeekNextAndPrevious(t *testing.T) {
	p := &parser{tokens: []token{{typ: tokenIdent, lexeme: "x"}, {typ: tokenEOF}}}
	_ = p.previous()
	if p.peekNext().typ != tokenEOF {
		t.Fatalf("expected peekNext to return EOF")
	}
	p.pos = 1
	if p.peekNext().typ != tokenEOF {
		t.Fatalf("expected peekNext fallback EOF")
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

func TestParseFile_InvalidTypeMissingAssign(t *testing.T) {
	content := `module bad.type3 as domain {
  type Email String;
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for missing '='")
	}
}

func TestParseFile_InvalidMethodMissingSemicolon(t *testing.T) {
	content := `module bad.port2 as port {
  port P {
    fn Ping()
  }
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for missing semicolon")
	}
}

func TestParseFile_WiringWithSemicolon(t *testing.T) {
	content := `module wire.ok as wiring {
  wiring W {
    ;
    bind a::port:P -> b::adapter:A;
  }
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
}

func TestParseFile_GenMetaInvalid(t *testing.T) {
	content := `@gen { prompt_ref "p" }
module bad.gen as domain {
  usecase Noop { return true; }
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for invalid @gen")
	}
}

func TestParseFile_StringListInvalid(t *testing.T) {
	content := `@gen { context_refs: "x" }
module bad.gen as domain {
  usecase Noop { return true; }
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for invalid context_refs")
	}
}

func TestParseFile_InvalidPortDecl(t *testing.T) {
	content := `module bad.port3 as port {
  port P {
    bad;
  }
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for invalid port decl")
	}
}

func TestParseFile_InvalidEnumDecl(t *testing.T) {
	content := `module bad.enum as domain {
  enum E { A;
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for invalid enum")
	}
}

func TestParseFile_TypeMissingBase(t *testing.T) {
	content := `module bad.type4 as domain {
  type X = ;
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for missing type base")
	}
}

func TestParseFile_WiringMissingBind(t *testing.T) {
	content := `module bad.wiring2 as wiring {
  wiring W {
    nope;
  }
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for missing bind")
	}
}

func TestParseFile_ReturnWithoutExpr(t *testing.T) {
	content := `module ret.none as usecase {
  usecase R {
    return;
  }
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
}

func TestParseFile_InvalidProjectItem(t *testing.T) {
	content := `project Bad {
  foo;
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for invalid project item")
	}
}

func TestParseFile_InvalidPackItem(t *testing.T) {
	content := `pack BadPack {
  foo;
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for invalid pack item")
	}
}

func TestParseFile_InvalidModelParams(t *testing.T) {
	content := `@gen { model_params: { bad } }
module bad.gen2 as domain {
  usecase Noop { return true; }
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for invalid model_params")
	}
}

func TestUnquoteString_Invalid(t *testing.T) {
	_, err := unquoteString("\"")
	if err == nil {
		t.Fatalf("expected error for invalid string")
	}
}

func TestParseFile_InvalidTypeWhereMissingSemicolon(t *testing.T) {
	content := `module bad.type5 as domain {
  type Email = String where nonEmpty
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for missing semicolon after where")
	}
}

func TestParseFile_InvalidMethodMissingReturnParen(t *testing.T) {
	content := `module bad.port4 as port {
  port P {
    fn Ping() -> ok: Bool);
  }
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for invalid return clause")
	}
}

func TestParseFile_CandidateNoScore(t *testing.T) {
	content := `module cand.core as domain {
  candidate cand.core::adapter:Impl;
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	prog, err := ParseFile(tmpfile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(prog.Modules[0].Candidates) != 1 {
		t.Fatalf("expected 1 candidate")
	}
}

func TestParseFile_IfBlockMissingBrace(t *testing.T) {
	content := `module bad.block as usecase {
  usecase Bad {
    if true { return true;
  }
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for missing brace")
	}
}

func TestParseFile_IfMissingBrace(t *testing.T) {
	content := `module bad.if as usecase {
  usecase Bad {
    if true return true;
  }
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for missing block brace")
	}
}

func TestParseFile_InvalidMethodMissingParen(t *testing.T) {
	content := `module bad.port5 as port {
  port P {
    fn Ping;
  }
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	_, err := ParseFile(tmpfile)
	if err == nil {
		t.Fatalf("expected parse error for missing paren")
	}
}

func TestParseFile_PackRefIdentVersion(t *testing.T) {
	content := `project Demo2 {
  use pack Foo@beta;
  module demo.mod as domain {
  }
}`
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)
	prog, err := ParseFile(tmpfile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(prog.Projects) != 1 || len(prog.Projects[0].Uses) != 1 {
		t.Fatalf("expected pack ref")
	}
	if prog.Projects[0].Uses[0].Version != "beta" {
		t.Fatalf("expected version beta")
	}
}
