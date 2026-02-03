package parser

import "testing"

func TestParseStringList_SkipsComma(t *testing.T) {
	tokens, err := lex(`["a",, "b"]`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	p := &parser{tokens: tokens}
	list, err := p.parseStringList()
	if err != nil {
		t.Fatalf("parseStringList failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list))
	}
}

func TestParseKVList_Semicolons(t *testing.T) {
	tokens, err := lex(`{ a: 1; b: 2 }`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	p := &parser{tokens: tokens}
	list, err := p.parseKVList()
	if err != nil {
		t.Fatalf("parseKVList failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 kv items, got %d", len(list))
	}
}

func TestParseMethodDecl_Direct(t *testing.T) {
	tokens, err := lex(`Ping() -> (ok: Bool);`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	p := &parser{tokens: tokens}
	decl, err := p.parseMethodDecl()
	if err != nil {
		t.Fatalf("parseMethodDecl failed: %v", err)
	}
	if decl.Name != "Ping" {
		t.Fatalf("unexpected method name")
	}
}

func TestParseHelpers_ErrorPaths(t *testing.T) {
	cases := []struct {
		name string
		fn   func() error
	}{
		{
			name: "type decl missing name",
			fn: func() error {
				p := &parser{tokens: []token{{typ: tokenAssign}, {typ: tokenEOF}}}
				_, err := p.parseTypeDecl()
				return err
			},
		},
		{
			name: "method decl missing name",
			fn: func() error {
				p := &parser{tokens: []token{{typ: tokenLParen}, {typ: tokenEOF}}}
				_, err := p.parseMethodDecl()
				return err
			},
		},
		{
			name: "wiring decl missing name",
			fn: func() error {
				p := &parser{tokens: []token{{typ: tokenLBrace}, {typ: tokenEOF}}}
				_, err := p.parseWiringDecl()
				return err
			},
		},
		{
			name: "policy decl missing name",
			fn: func() error {
				p := &parser{tokens: []token{{typ: tokenColon}, {typ: tokenEOF}}}
				_, err := p.parsePolicyDecl()
				return err
			},
		},
		{
			name: "constraint decl missing name",
			fn: func() error {
				p := &parser{tokens: []token{{typ: tokenColon}, {typ: tokenEOF}}}
				_, err := p.parseConstraintDecl()
				return err
			},
		},
		{
			name: "hole decl missing name",
			fn: func() error {
				p := &parser{tokens: []token{{typ: tokenColon}, {typ: tokenEOF}}}
				_, err := p.parseHoleDecl()
				return err
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.fn(); err == nil {
				t.Fatalf("expected error")
			}
		})
	}
}

func TestParseEffects_Error(t *testing.T) {
	p := &parser{tokens: []token{{typ: tokenIdent, lexeme: "oops"}, {typ: tokenEOF}}}
	if _, err := p.parseEffects(); err == nil {
		t.Fatalf("expected error for effects without '['")
	}
}

func TestParseFieldBlock_Error(t *testing.T) {
	p := &parser{tokens: []token{{typ: tokenIdent, lexeme: "oops"}, {typ: tokenEOF}}}
	if _, err := p.parseFieldBlock(); err == nil {
		t.Fatalf("expected error for field block without '{'")
	}
}

func TestParseCandidateDecl_Error(t *testing.T) {
	tokens, err := lex(`sym { bad }`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	p := &parser{tokens: tokens}
	if _, err := p.parseCandidateDecl(); err == nil {
		t.Fatalf("expected error for bad candidate block")
	}
}

func TestParseCandidateDecl_InvalidScore(t *testing.T) {
	tokens, err := lex(`sym score bad;`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	p := &parser{tokens: tokens}
	if _, err := p.parseCandidateDecl(); err == nil {
		t.Fatalf("expected error for invalid score")
	}
}

func TestParseProject_Error(t *testing.T) {
	tokens, err := lex(`Demo`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	p := &parser{tokens: tokens}
	if _, _, err := p.parseProject(); err == nil {
		t.Fatalf("expected error for project without '{'")
	}
}

func TestParsePack_Error(t *testing.T) {
	tokens, err := lex(`Pack`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	p := &parser{tokens: tokens}
	if _, err := p.parsePack(); err == nil {
		t.Fatalf("expected error for pack without '{'")
	}
}

func TestParseEnumDecl_Error(t *testing.T) {
	tokens, err := lex(`{`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	p := &parser{tokens: tokens}
	if _, err := p.parseEnumDecl(); err == nil {
		t.Fatalf("expected error for enum without name")
	}
}

func TestParseEffects_MissingBracket(t *testing.T) {
	tokens, err := lex(`[io`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	p := &parser{tokens: tokens}
	if _, err := p.parseEffects(); err == nil {
		t.Fatalf("expected error for effects missing ']'")
	}
}

func TestParseStmt_DirectReturn(t *testing.T) {
	tokens, err := lex(`return;`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	p := &parser{tokens: tokens}
	stmt, err := p.parseStmt()
	if err != nil {
		t.Fatalf("parseStmt failed: %v", err)
	}
	if stmt.Kind != "return" {
		t.Fatalf("expected return stmt")
	}
}

func TestParseStmt_DirectAssign(t *testing.T) {
	tokens, err := lex(`x = 1;`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	p := &parser{tokens: tokens}
	stmt, err := p.parseStmt()
	if err != nil {
		t.Fatalf("parseStmt failed: %v", err)
	}
	if stmt.Kind != "assign" {
		t.Fatalf("expected assign stmt")
	}
}

func TestParseFieldList_Semicolons(t *testing.T) {
	tokens, err := lex(`a: Int; b: Int }`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	p := &parser{tokens: tokens}
	fields, err := p.parseFieldList(tokenRBrace)
	if err != nil {
		t.Fatalf("parseFieldList failed: %v", err)
	}
	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}
}

func TestParseEffects_LeadingComma(t *testing.T) {
	tokens, err := lex(`[ , io ]`)
	if err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	p := &parser{tokens: tokens}
	list, err := p.parseEffects()
	if err != nil {
		t.Fatalf("parseEffects failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 effect, got %d", len(list))
	}
}
