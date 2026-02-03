package parser

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/willams/lia/internal/ir"
	"github.com/willams/lia/internal/symbols"
)

// ParseFile parses a .lia source file into an IR program.
func ParseFile(path string) (*ir.Program, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	tokens, err := lex(string(b))
	if err != nil {
		return nil, err
	}
	p := &parser{tokens: tokens}
	prog, err := p.parseProgram()
	if err != nil {
		return nil, err
	}
	_ = symbols.DeriveProgramSymbols(prog)
	return prog, nil
}

type parser struct {
	tokens []token
	pos    int
}

func (p *parser) parseProgram() (*ir.Program, error) {
	prog := &ir.Program{Version: "0.1"}
	for !p.check(tokenEOF) {
		if p.matchIdent("project") {
			proj, mods, err := p.parseProject()
			if err != nil {
				return nil, err
			}
			prog.Projects = append(prog.Projects, proj)
			prog.Modules = append(prog.Modules, mods...)
			continue
		}
		if p.matchIdent("pack") {
			pack, err := p.parsePack()
			if err != nil {
				return nil, err
			}
			prog.Packs = append(prog.Packs, pack)
			continue
		}
		if p.match(tokenAt) {
			if !p.matchIdent("gen") {
				return nil, p.errorf(p.peek(), "expected gen after @")
			}
			meta, err := p.parseGenMeta()
			if err != nil {
				return nil, err
			}
			if !p.matchIdent("module") {
				return nil, p.errorf(p.peek(), "expected module after @gen")
			}
			mod, err := p.parseModule(meta)
			if err != nil {
				return nil, err
			}
			prog.Modules = append(prog.Modules, mod)
			continue
		}
		if p.matchIdent("module") {
			mod, err := p.parseModule(nil)
			if err != nil {
				return nil, err
			}
			prog.Modules = append(prog.Modules, mod)
			continue
		}
		return nil, p.errorf(p.peek(), "unexpected token")
	}
	return prog, nil
}

func (p *parser) parseProject() (ir.Project, []ir.Module, error) {
	name, err := p.parseQName()
	if err != nil {
		return ir.Project{}, nil, err
	}
	proj := ir.Project{Name: name}
	if err := p.expect(tokenLBrace, "expected '{' after project name"); err != nil {
		return ir.Project{}, nil, err
	}
	var modules []ir.Module
	for !p.check(tokenRBrace) && !p.check(tokenEOF) {
		if p.match(tokenSemicolon) {
			continue
		}
		if p.matchIdent("repro") {
			reproTok, err := p.expectIdent("expected repro value")
			if err != nil {
				return ir.Project{}, nil, err
			}
			proj.Repro = ir.ReproProfile(reproTok.lexeme)
			if err := p.expect(tokenSemicolon, "expected ';' after repro"); err != nil {
				return ir.Project{}, nil, err
			}
			continue
		}
		if p.matchIdent("tape") {
			tapePath, err := p.parseTapePath()
			if err != nil {
				return ir.Project{}, nil, err
			}
			proj.Tape = tapePath
			if err := p.expect(tokenSemicolon, "expected ';' after tape"); err != nil {
				return ir.Project{}, nil, err
			}
			continue
		}
		if p.matchIdent("use") {
			if !p.matchIdent("pack") {
				return ir.Project{}, nil, p.errorf(p.peek(), "expected 'pack' after use")
			}
			ref, err := p.parsePackRef()
			if err != nil {
				return ir.Project{}, nil, err
			}
			proj.Uses = append(proj.Uses, ref)
			if err := p.expect(tokenSemicolon, "expected ';' after use pack"); err != nil {
				return ir.Project{}, nil, err
			}
			continue
		}
		if p.matchIdent("policy") {
			pol, err := p.parsePolicyDecl()
			if err != nil {
				return ir.Project{}, nil, err
			}
			proj.Policies = append(proj.Policies, pol)
			continue
		}
		if p.matchIdent("constraint") {
			c, err := p.parseConstraintDecl()
			if err != nil {
				return ir.Project{}, nil, err
			}
			proj.Constraints = append(proj.Constraints, c)
			continue
		}
		if p.match(tokenAt) {
			if !p.matchIdent("gen") {
				return ir.Project{}, nil, p.errorf(p.peek(), "expected gen after @")
			}
			meta, err := p.parseGenMeta()
			if err != nil {
				return ir.Project{}, nil, err
			}
			if !p.matchIdent("module") {
				return ir.Project{}, nil, p.errorf(p.peek(), "expected module after @gen")
			}
			mod, err := p.parseModule(meta)
			if err != nil {
				return ir.Project{}, nil, err
			}
			modules = append(modules, mod)
			continue
		}
		if p.matchIdent("module") {
			mod, err := p.parseModule(nil)
			if err != nil {
				return ir.Project{}, nil, err
			}
			modules = append(modules, mod)
			continue
		}
		return ir.Project{}, nil, p.errorf(p.peek(), "unexpected token in project")
	}
	if err := p.expect(tokenRBrace, "expected '}' to close project"); err != nil {
		return ir.Project{}, nil, err
	}
	return proj, modules, nil
}

func (p *parser) parsePack() (ir.Pack, error) {
	nameTok, err := p.expectIdent("expected pack name")
	if err != nil {
		return ir.Pack{}, err
	}
	pack := ir.Pack{Name: nameTok.lexeme}
	if p.match(tokenAt) {
		ver, err := p.parseVersion()
		if err != nil {
			return ir.Pack{}, err
		}
		pack.Version = ver
	}
	if err := p.expect(tokenLBrace, "expected '{' after pack name"); err != nil {
		return ir.Pack{}, err
	}
	for !p.check(tokenRBrace) && !p.check(tokenEOF) {
		if p.match(tokenSemicolon) {
			continue
		}
		if p.matchIdent("module") {
			mod, err := p.parseModule(nil)
			if err != nil {
				return ir.Pack{}, err
			}
			pack.Modules = append(pack.Modules, mod)
			continue
		}
		if p.matchIdent("policy") {
			pol, err := p.parsePolicyDecl()
			if err != nil {
				return ir.Pack{}, err
			}
			pack.Policies = append(pack.Policies, pol)
			continue
		}
		if p.matchIdent("constraint") {
			c, err := p.parseConstraintDecl()
			if err != nil {
				return ir.Pack{}, err
			}
			pack.Constraints = append(pack.Constraints, c)
			continue
		}
		return ir.Pack{}, p.errorf(p.peek(), "unexpected token in pack")
	}
	if err := p.expect(tokenRBrace, "expected '}' to close pack"); err != nil {
		return ir.Pack{}, err
	}
	return pack, nil
}

func (p *parser) parseModule(gen *ir.GenMeta) (ir.Module, error) {
	name, err := p.parseQName()
	if err != nil {
		return ir.Module{}, err
	}
	mod := ir.Module{Name: name, Gen: gen}
	if p.matchIdent("as") {
		roleTok, err := p.expectIdent("expected role name")
		if err != nil {
			return ir.Module{}, err
		}
		mod.Role = roleTok.lexeme
	}
	if err := p.expect(tokenLBrace, "expected '{' after module header"); err != nil {
		return ir.Module{}, err
	}
	for !p.check(tokenRBrace) && !p.check(tokenEOF) {
		if p.match(tokenSemicolon) {
			continue
		}
		if p.matchIdent("type") {
			decl, err := p.parseTypeDecl()
			if err != nil {
				return ir.Module{}, err
			}
			mod.Types = append(mod.Types, decl)
			continue
		}
		if p.matchIdent("enum") {
			decl, err := p.parseEnumDecl()
			if err != nil {
				return ir.Module{}, err
			}
			mod.Enums = append(mod.Enums, decl)
			continue
		}
		if p.matchIdent("port") {
			decl, err := p.parsePortDecl()
			if err != nil {
				return ir.Module{}, err
			}
			mod.Ports = append(mod.Ports, decl)
			continue
		}
		if p.matchIdent("usecase") {
			decl, err := p.parseUsecaseDecl()
			if err != nil {
				return ir.Module{}, err
			}
			mod.Usecases = append(mod.Usecases, decl)
			continue
		}
		if p.matchIdent("adapter") {
			decl, err := p.parseAdapterDecl()
			if err != nil {
				return ir.Module{}, err
			}
			mod.Adapters = append(mod.Adapters, decl)
			continue
		}
		if p.matchIdent("wiring") {
			decl, err := p.parseWiringDecl()
			if err != nil {
				return ir.Module{}, err
			}
			mod.Wirings = append(mod.Wirings, decl)
			continue
		}
		if p.matchIdent("constraint") {
			decl, err := p.parseConstraintDecl()
			if err != nil {
				return ir.Module{}, err
			}
			mod.Constraints = append(mod.Constraints, decl)
			continue
		}
		if p.matchIdent("prefer") {
			decl, err := p.parsePreferDecl()
			if err != nil {
				return ir.Module{}, err
			}
			mod.Preferences = append(mod.Preferences, decl)
			continue
		}
		if p.matchIdent("hole") {
			decl, err := p.parseHoleDecl()
			if err != nil {
				return ir.Module{}, err
			}
			mod.Holes = append(mod.Holes, decl)
			continue
		}
		if p.matchIdent("candidate") {
			decl, err := p.parseCandidateDecl()
			if err != nil {
				return ir.Module{}, err
			}
			mod.Candidates = append(mod.Candidates, decl)
			continue
		}
		return ir.Module{}, p.errorf(p.peek(), "unexpected token in module")
	}
	if err := p.expect(tokenRBrace, "expected '}' to close module"); err != nil {
		return ir.Module{}, err
	}
	return mod, nil
}

func (p *parser) parseTypeDecl() (ir.TypeDecl, error) {
	nameTok, err := p.expectIdent("expected type name")
	if err != nil {
		return ir.TypeDecl{}, err
	}
	if err := p.expect(tokenAssign, "expected '=' after type name"); err != nil {
		return ir.TypeDecl{}, err
	}
	base, err := p.parseTypeRef()
	if err != nil {
		return ir.TypeDecl{}, err
	}
	decl := ir.TypeDecl{Name: nameTok.lexeme, Base: base}
	if p.matchIdent("where") {
		pred, err := p.collectUntil(tokenSemicolon)
		if err != nil {
			return ir.TypeDecl{}, err
		}
		decl.Predicate = pred
		return decl, nil
	}
	if err := p.expect(tokenSemicolon, "expected ';' after type declaration"); err != nil {
		return ir.TypeDecl{}, err
	}
	return decl, nil
}

func (p *parser) parseEnumDecl() (ir.EnumDecl, error) {
	nameTok, err := p.expectIdent("expected enum name")
	if err != nil {
		return ir.EnumDecl{}, err
	}
	if err := p.expect(tokenLBrace, "expected '{' after enum name"); err != nil {
		return ir.EnumDecl{}, err
	}
	values := []string{}
	for !p.check(tokenRBrace) && !p.check(tokenEOF) {
		if p.match(tokenComma) {
			continue
		}
		valTok, err := p.expectIdent("expected enum value")
		if err != nil {
			return ir.EnumDecl{}, err
		}
		values = append(values, valTok.lexeme)
		p.match(tokenComma)
	}
	if err := p.expect(tokenRBrace, "expected '}' after enum values"); err != nil {
		return ir.EnumDecl{}, err
	}
	p.match(tokenSemicolon)
	return ir.EnumDecl{Name: nameTok.lexeme, Values: values}, nil
}

func (p *parser) parsePortDecl() (ir.PortDecl, error) {
	nameTok, err := p.expectIdent("expected port name")
	if err != nil {
		return ir.PortDecl{}, err
	}
	if err := p.expect(tokenLBrace, "expected '{' after port name"); err != nil {
		return ir.PortDecl{}, err
	}
	port := ir.PortDecl{Name: nameTok.lexeme}
	for !p.check(tokenRBrace) && !p.check(tokenEOF) {
		if p.match(tokenSemicolon) {
			continue
		}
		if p.matchIdent("fn") || p.matchIdent("method") {
			meth, err := p.parseMethodDecl()
			if err != nil {
				return ir.PortDecl{}, err
			}
			port.Methods = append(port.Methods, meth)
			continue
		}
		return ir.PortDecl{}, p.errorf(p.peek(), "unexpected token in port")
	}
	if err := p.expect(tokenRBrace, "expected '}' after port body"); err != nil {
		return ir.PortDecl{}, err
	}
	return port, nil
}

func (p *parser) parseMethodDecl() (ir.FuncDecl, error) {
	nameTok, err := p.expectIdent("expected method name")
	if err != nil {
		return ir.FuncDecl{}, err
	}
	if err := p.expect(tokenLParen, "expected '(' after method name"); err != nil {
		return ir.FuncDecl{}, err
	}
	params, err := p.parseFieldList(tokenRParen)
	if err != nil {
		return ir.FuncDecl{}, err
	}
	if err := p.expect(tokenRParen, "expected ')' after params"); err != nil {
		return ir.FuncDecl{}, err
	}
	returns := []ir.Field{}
	if p.match(tokenArrow) {
		if err := p.expect(tokenLParen, "expected '(' after ->"); err != nil {
			return ir.FuncDecl{}, err
		}
		returns, err = p.parseFieldList(tokenRParen)
		if err != nil {
			return ir.FuncDecl{}, err
		}
		if err := p.expect(tokenRParen, "expected ')' after returns"); err != nil {
			return ir.FuncDecl{}, err
		}
	}
	if err := p.expect(tokenSemicolon, "expected ';' after method declaration"); err != nil {
		return ir.FuncDecl{}, err
	}
	return ir.FuncDecl{Name: nameTok.lexeme, Params: params, Returns: returns}, nil
}

func (p *parser) parseUsecaseDecl() (ir.UsecaseDecl, error) {
	nameTok, err := p.expectIdent("expected usecase name")
	if err != nil {
		return ir.UsecaseDecl{}, err
	}
	if err := p.expect(tokenLBrace, "expected '{' after usecase name"); err != nil {
		return ir.UsecaseDecl{}, err
	}
	uc := ir.UsecaseDecl{Name: nameTok.lexeme}
	for !p.check(tokenRBrace) && !p.check(tokenEOF) {
		if p.match(tokenSemicolon) {
			continue
		}
		if p.matchIdent("input") {
			fields, err := p.parseFieldBlock()
			if err != nil {
				return ir.UsecaseDecl{}, err
			}
			uc.Inputs = append(uc.Inputs, fields...)
			p.match(tokenSemicolon)
			continue
		}
		if p.matchIdent("output") {
			fields, err := p.parseFieldBlock()
			if err != nil {
				return ir.UsecaseDecl{}, err
			}
			uc.Outputs = append(uc.Outputs, fields...)
			p.match(tokenSemicolon)
			continue
		}
		if p.matchIdent("effects") {
			effects, err := p.parseEffects()
			if err != nil {
				return ir.UsecaseDecl{}, err
			}
			uc.Effects = append(uc.Effects, effects...)
			continue
		}
		stmt, err := p.parseStmt()
		if err != nil {
			return ir.UsecaseDecl{}, err
		}
		uc.Body = append(uc.Body, stmt)
	}
	if err := p.expect(tokenRBrace, "expected '}' after usecase body"); err != nil {
		return ir.UsecaseDecl{}, err
	}
	return uc, nil
}

func (p *parser) parseAdapterDecl() (ir.AdapterDecl, error) {
	nameTok, err := p.expectIdent("expected adapter name")
	if err != nil {
		return ir.AdapterDecl{}, err
	}
	ad := ir.AdapterDecl{Name: nameTok.lexeme}
	if p.matchIdent("implements") {
		sym, err := p.parseQName()
		if err != nil {
			return ir.AdapterDecl{}, err
		}
		ad.Implements = sym
	}
	if err := p.expect(tokenLBrace, "expected '{' after adapter name"); err != nil {
		return ir.AdapterDecl{}, err
	}
	for !p.check(tokenRBrace) && !p.check(tokenEOF) {
		if p.match(tokenSemicolon) {
			continue
		}
		if p.matchIdent("input") {
			fields, err := p.parseFieldBlock()
			if err != nil {
				return ir.AdapterDecl{}, err
			}
			ad.Inputs = append(ad.Inputs, fields...)
			p.match(tokenSemicolon)
			continue
		}
		if p.matchIdent("output") {
			fields, err := p.parseFieldBlock()
			if err != nil {
				return ir.AdapterDecl{}, err
			}
			ad.Outputs = append(ad.Outputs, fields...)
			p.match(tokenSemicolon)
			continue
		}
		if p.matchIdent("effects") {
			effects, err := p.parseEffects()
			if err != nil {
				return ir.AdapterDecl{}, err
			}
			ad.Effects = append(ad.Effects, effects...)
			continue
		}
		stmt, err := p.parseStmt()
		if err != nil {
			return ir.AdapterDecl{}, err
		}
		ad.Body = append(ad.Body, stmt)
	}
	if err := p.expect(tokenRBrace, "expected '}' after adapter body"); err != nil {
		return ir.AdapterDecl{}, err
	}
	return ad, nil
}

func (p *parser) parseWiringDecl() (ir.WiringDecl, error) {
	nameTok, err := p.expectIdent("expected wiring name")
	if err != nil {
		return ir.WiringDecl{}, err
	}
	if err := p.expect(tokenLBrace, "expected '{' after wiring name"); err != nil {
		return ir.WiringDecl{}, err
	}
	w := ir.WiringDecl{Name: nameTok.lexeme}
	for !p.check(tokenRBrace) && !p.check(tokenEOF) {
		if p.match(tokenSemicolon) {
			continue
		}
		if !p.matchIdent("bind") {
			return ir.WiringDecl{}, p.errorf(p.peek(), "expected bind in wiring")
		}
		left, err := p.parseQName()
		if err != nil {
			return ir.WiringDecl{}, err
		}
		if err := p.expect(tokenArrow, "expected '->' in bind"); err != nil {
			return ir.WiringDecl{}, err
		}
		right, err := p.parseQName()
		if err != nil {
			return ir.WiringDecl{}, err
		}
		if err := p.expect(tokenSemicolon, "expected ';' after bind"); err != nil {
			return ir.WiringDecl{}, err
		}
		w.Binds = append(w.Binds, fmt.Sprintf("%s -> %s", left, right))
	}
	if err := p.expect(tokenRBrace, "expected '}' after wiring"); err != nil {
		return ir.WiringDecl{}, err
	}
	return w, nil
}

func (p *parser) parseConstraintDecl() (ir.ConstraintDecl, error) {
	nameTok, err := p.expectIdent("expected constraint name")
	if err != nil {
		return ir.ConstraintDecl{}, err
	}
	if err := p.expect(tokenColon, "expected ':' after constraint name"); err != nil {
		return ir.ConstraintDecl{}, err
	}
	expr, err := p.collectUntil(tokenSemicolon)
	if err != nil {
		return ir.ConstraintDecl{}, err
	}
	return ir.ConstraintDecl{Name: nameTok.lexeme, Expr: expr}, nil
}

func (p *parser) parsePreferDecl() (ir.PreferDecl, error) {
	nameTok, err := p.expectIdent("expected prefer name")
	if err != nil {
		return ir.PreferDecl{}, err
	}
	if err := p.expect(tokenColon, "expected ':' after prefer name"); err != nil {
		return ir.PreferDecl{}, err
	}
	weight := 0.0
	weightSet := false
	var exprTokens []token
	for !p.check(tokenSemicolon) && !p.check(tokenEOF) {
		if p.matchIdent("weight") {
			valTok := p.peek()
			if !p.match(tokenNumber) && !p.match(tokenIdent) {
				return ir.PreferDecl{}, p.errorf(valTok, "expected weight number")
			}
			val, err := strconv.ParseFloat(valTok.lexeme, 64)
			if err != nil {
				return ir.PreferDecl{}, p.errorf(valTok, "invalid weight")
			}
			weight = val
			weightSet = true
			continue
		}
		exprTokens = append(exprTokens, p.advance())
	}
	if err := p.expect(tokenSemicolon, "expected ';' after prefer"); err != nil {
		return ir.PreferDecl{}, err
	}
	expr := tokensToString(exprTokens)
	decl := ir.PreferDecl{Name: nameTok.lexeme, Expr: expr}
	if weightSet {
		decl.Weight = weight
	}
	return decl, nil
}

func (p *parser) parseHoleDecl() (ir.HoleDecl, error) {
	nameTok, err := p.expectIdent("expected hole name")
	if err != nil {
		return ir.HoleDecl{}, err
	}
	if err := p.expect(tokenColon, "expected ':' after hole name"); err != nil {
		return ir.HoleDecl{}, err
	}
	contract, err := p.collectUntil(tokenSemicolon)
	if err != nil {
		return ir.HoleDecl{}, err
	}
	return ir.HoleDecl{Name: nameTok.lexeme, Contract: contract}, nil
}

func (p *parser) parseCandidateDecl() (ir.CandidateDecl, error) {
	sym, err := p.parseQName()
	if err != nil {
		return ir.CandidateDecl{}, err
	}
	cand := ir.CandidateDecl{Symbol: sym}
	if p.matchIdent("score") {
		valTok := p.peek()
		if !p.match(tokenNumber) && !p.match(tokenIdent) {
			return ir.CandidateDecl{}, p.errorf(valTok, "expected score number")
		}
		val, err := strconv.ParseFloat(valTok.lexeme, 64)
		if err != nil {
			return ir.CandidateDecl{}, p.errorf(valTok, "invalid score")
		}
		cand.Score = val
	}
	if p.match(tokenLBrace) {
		for !p.check(tokenRBrace) && !p.check(tokenEOF) {
			if p.match(tokenSemicolon) {
				continue
			}
			if p.matchIdent("score") {
				valTok := p.peek()
				if !p.match(tokenNumber) && !p.match(tokenIdent) {
					return ir.CandidateDecl{}, p.errorf(valTok, "expected score number")
				}
				val, err := strconv.ParseFloat(valTok.lexeme, 64)
				if err != nil {
					return ir.CandidateDecl{}, p.errorf(valTok, "invalid score")
				}
				cand.Score = val
				if err := p.expect(tokenSemicolon, "expected ';' after score"); err != nil {
					return ir.CandidateDecl{}, err
				}
				continue
			}
			if p.matchIdent("constraint") {
				c, err := p.parseConstraintDecl()
				if err != nil {
					return ir.CandidateDecl{}, err
				}
				cand.Constraints = append(cand.Constraints, c)
				continue
			}
			return ir.CandidateDecl{}, p.errorf(p.peek(), "unexpected token in candidate")
		}
		if err := p.expect(tokenRBrace, "expected '}' after candidate block"); err != nil {
			return ir.CandidateDecl{}, err
		}
		p.match(tokenSemicolon)
		return cand, nil
	}
	p.match(tokenSemicolon)
	return cand, nil
}

func (p *parser) parseEffects() ([]string, error) {
	if err := p.expect(tokenLBracket, "expected '[' after effects"); err != nil {
		return nil, err
	}
	var effects []string
	for !p.check(tokenRBracket) && !p.check(tokenEOF) {
		if p.match(tokenComma) {
			continue
		}
		tok := p.peek()
		if !p.match(tokenIdent) && !p.match(tokenString) {
			return nil, p.errorf(tok, "expected effect name")
		}
		value := tok.lexeme
		if tok.typ == tokenString {
			unq, err := unquoteString(tok.lexeme)
			if err != nil {
				return nil, err
			}
			value = unq
		}
		effects = append(effects, value)
		p.match(tokenComma)
	}
	if err := p.expect(tokenRBracket, "expected ']' after effects"); err != nil {
		return nil, err
	}
	p.match(tokenSemicolon)
	return effects, nil
}

func (p *parser) parseFieldBlock() ([]ir.Field, error) {
	if err := p.expect(tokenLBrace, "expected '{' for field block"); err != nil {
		return nil, err
	}
	fields, err := p.parseFieldList(tokenRBrace)
	if err != nil {
		return nil, err
	}
	if err := p.expect(tokenRBrace, "expected '}' after field block"); err != nil {
		return nil, err
	}
	return fields, nil
}

func (p *parser) parseFieldList(stop tokenType) ([]ir.Field, error) {
	var fields []ir.Field
	if p.check(stop) {
		return fields, nil
	}
	for !p.check(stop) && !p.check(tokenEOF) {
		nameTok, err := p.expectIdent("expected field name")
		if err != nil {
			return nil, err
		}
		if err := p.expect(tokenColon, "expected ':' after field name"); err != nil {
			return nil, err
		}
		typeRef, err := p.parseTypeRef()
		if err != nil {
			return nil, err
		}
		fields = append(fields, ir.Field{Name: nameTok.lexeme, Type: typeRef})
		if p.match(tokenComma) {
			continue
		}
		if p.match(tokenSemicolon) {
			continue
		}
		if p.check(stop) {
			break
		}
		return nil, p.errorf(p.peek(), "expected ',' or '%s'", tokenName(stop))
	}
	return fields, nil
}

func (p *parser) parseStmt() (ir.Stmt, error) {
	if p.matchIdent("let") {
		nameTok, err := p.expectIdent("expected identifier after let")
		if err != nil {
			return ir.Stmt{}, err
		}
		if err := p.expect(tokenAssign, "expected '=' after let name"); err != nil {
			return ir.Stmt{}, err
		}
		expr, err := p.parseExpr()
		if err != nil {
			return ir.Stmt{}, err
		}
		if err := p.expect(tokenSemicolon, "expected ';' after let"); err != nil {
			return ir.Stmt{}, err
		}
		return ir.Stmt{Kind: "let", Let: &ir.LetStmt{Name: nameTok.lexeme, Value: expr}}, nil
	}
	if p.matchIdent("return") {
		if p.check(tokenSemicolon) {
			p.advance()
			return ir.Stmt{Kind: "return", Return: &ir.ReturnStmt{}}, nil
		}
		expr, err := p.parseExpr()
		if err != nil {
			return ir.Stmt{}, err
		}
		if err := p.expect(tokenSemicolon, "expected ';' after return"); err != nil {
			return ir.Stmt{}, err
		}
		return ir.Stmt{Kind: "return", Return: &ir.ReturnStmt{Value: &expr}}, nil
	}
	if p.matchIdent("if") {
		cond, err := p.parseExpr()
		if err != nil {
			return ir.Stmt{}, err
		}
		thenBlock, err := p.parseStmtBlock()
		if err != nil {
			return ir.Stmt{}, err
		}
		var elseBlock []ir.Stmt
		if p.matchIdent("else") {
			elseBlock, err = p.parseStmtBlock()
			if err != nil {
				return ir.Stmt{}, err
			}
		}
		return ir.Stmt{Kind: "if", If: &ir.IfStmt{Cond: cond, Then: thenBlock, Else: elseBlock}}, nil
	}
	if p.matchIdent("while") {
		cond, err := p.parseExpr()
		if err != nil {
			return ir.Stmt{}, err
		}
		body, err := p.parseStmtBlock()
		if err != nil {
			return ir.Stmt{}, err
		}
		return ir.Stmt{Kind: "while", While: &ir.WhileStmt{Cond: cond, Body: body}}, nil
	}
	if p.matchIdent("for") {
		nameTok, err := p.expectIdent("expected loop variable")
		if err != nil {
			return ir.Stmt{}, err
		}
		if !p.matchIdent("in") {
			return ir.Stmt{}, p.errorf(p.peek(), "expected 'in' after loop variable")
		}
		iter, err := p.parseExpr()
		if err != nil {
			return ir.Stmt{}, err
		}
		body, err := p.parseStmtBlock()
		if err != nil {
			return ir.Stmt{}, err
		}
		return ir.Stmt{Kind: "for", For: &ir.ForStmt{Var: nameTok.lexeme, Iter: iter, Body: body}}, nil
	}
	if p.matchIdent("break") {
		if err := p.expect(tokenSemicolon, "expected ';' after break"); err != nil {
			return ir.Stmt{}, err
		}
		return ir.Stmt{Kind: "break"}, nil
	}
	if p.matchIdent("continue") {
		if err := p.expect(tokenSemicolon, "expected ';' after continue"); err != nil {
			return ir.Stmt{}, err
		}
		return ir.Stmt{Kind: "continue"}, nil
	}

	if p.check(tokenIdent) && p.peekNext().typ == tokenAssign {
		nameTok := p.advance()
		p.advance()
		expr, err := p.parseExpr()
		if err != nil {
			return ir.Stmt{}, err
		}
		if err := p.expect(tokenSemicolon, "expected ';' after assignment"); err != nil {
			return ir.Stmt{}, err
		}
		return ir.Stmt{Kind: "assign", Assign: &ir.AssignStmt{Name: nameTok.lexeme, Value: expr}}, nil
	}

	expr, err := p.parseExpr()
	if err != nil {
		return ir.Stmt{}, err
	}
	if err := p.expect(tokenSemicolon, "expected ';' after expression"); err != nil {
		return ir.Stmt{}, err
	}
	return ir.Stmt{Kind: "expr", ExprStmt: &expr}, nil
}

func (p *parser) parseStmtBlock() ([]ir.Stmt, error) {
	if err := p.expect(tokenLBrace, "expected '{' to start block"); err != nil {
		return nil, err
	}
	var stmts []ir.Stmt
	for !p.check(tokenRBrace) && !p.check(tokenEOF) {
		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}
	if err := p.expect(tokenRBrace, "expected '}' to close block"); err != nil {
		return nil, err
	}
	return stmts, nil
}

func (p *parser) parseExpr() (ir.Expr, error) {
	return p.parseOrExpr()
}

func (p *parser) parseOrExpr() (ir.Expr, error) {
	left, err := p.parseAndExpr()
	if err != nil {
		return ir.Expr{}, err
	}
	for p.match(tokenOrOr) {
		op := "||"
		right, err := p.parseAndExpr()
		if err != nil {
			return ir.Expr{}, err
		}
		left = ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: op, Left: left, Right: right}}
	}
	return left, nil
}

func (p *parser) parseAndExpr() (ir.Expr, error) {
	left, err := p.parseEqualityExpr()
	if err != nil {
		return ir.Expr{}, err
	}
	for p.match(tokenAndAnd) {
		op := "&&"
		right, err := p.parseEqualityExpr()
		if err != nil {
			return ir.Expr{}, err
		}
		left = ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: op, Left: left, Right: right}}
	}
	return left, nil
}

func (p *parser) parseEqualityExpr() (ir.Expr, error) {
	left, err := p.parseComparisonExpr()
	if err != nil {
		return ir.Expr{}, err
	}
	for {
		if p.match(tokenEqEq) {
			right, err := p.parseComparisonExpr()
			if err != nil {
				return ir.Expr{}, err
			}
			left = ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: "==", Left: left, Right: right}}
			continue
		}
		if p.match(tokenNotEq) {
			right, err := p.parseComparisonExpr()
			if err != nil {
				return ir.Expr{}, err
			}
			left = ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: "!=", Left: left, Right: right}}
			continue
		}
		break
	}
	return left, nil
}

func (p *parser) parseComparisonExpr() (ir.Expr, error) {
	left, err := p.parseTermExpr()
	if err != nil {
		return ir.Expr{}, err
	}
	for {
		if p.match(tokenLT) {
			right, err := p.parseTermExpr()
			if err != nil {
				return ir.Expr{}, err
			}
			left = ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: "<", Left: left, Right: right}}
			continue
		}
		if p.match(tokenLTE) {
			right, err := p.parseTermExpr()
			if err != nil {
				return ir.Expr{}, err
			}
			left = ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: "<=", Left: left, Right: right}}
			continue
		}
		if p.match(tokenGT) {
			right, err := p.parseTermExpr()
			if err != nil {
				return ir.Expr{}, err
			}
			left = ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: ">", Left: left, Right: right}}
			continue
		}
		if p.match(tokenGTE) {
			right, err := p.parseTermExpr()
			if err != nil {
				return ir.Expr{}, err
			}
			left = ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: ">=", Left: left, Right: right}}
			continue
		}
		break
	}
	return left, nil
}

func (p *parser) parseTermExpr() (ir.Expr, error) {
	left, err := p.parseFactorExpr()
	if err != nil {
		return ir.Expr{}, err
	}
	for {
		if p.match(tokenPlus) {
			right, err := p.parseFactorExpr()
			if err != nil {
				return ir.Expr{}, err
			}
			left = ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: "+", Left: left, Right: right}}
			continue
		}
		if p.match(tokenMinus) {
			right, err := p.parseFactorExpr()
			if err != nil {
				return ir.Expr{}, err
			}
			left = ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: "-", Left: left, Right: right}}
			continue
		}
		break
	}
	return left, nil
}

func (p *parser) parseFactorExpr() (ir.Expr, error) {
	left, err := p.parseUnaryExpr()
	if err != nil {
		return ir.Expr{}, err
	}
	for {
		if p.match(tokenStar) {
			right, err := p.parseUnaryExpr()
			if err != nil {
				return ir.Expr{}, err
			}
			left = ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: "*", Left: left, Right: right}}
			continue
		}
		if p.match(tokenSlash) {
			right, err := p.parseUnaryExpr()
			if err != nil {
				return ir.Expr{}, err
			}
			left = ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: "/", Left: left, Right: right}}
			continue
		}
		if p.match(tokenPercent) {
			right, err := p.parseUnaryExpr()
			if err != nil {
				return ir.Expr{}, err
			}
			left = ir.Expr{Kind: "binary", Binary: &ir.BinaryExpr{Op: "%", Left: left, Right: right}}
			continue
		}
		break
	}
	return left, nil
}

func (p *parser) parseUnaryExpr() (ir.Expr, error) {
	if p.match(tokenBang) {
		expr, err := p.parseUnaryExpr()
		if err != nil {
			return ir.Expr{}, err
		}
		return ir.Expr{Kind: "unary", Unary: &ir.UnaryExpr{Op: "!", Expr: expr}}, nil
	}
	if p.match(tokenMinus) {
		expr, err := p.parseUnaryExpr()
		if err != nil {
			return ir.Expr{}, err
		}
		return ir.Expr{Kind: "unary", Unary: &ir.UnaryExpr{Op: "-", Expr: expr}}, nil
	}
	return p.parseCallExpr()
}

func (p *parser) parseCallExpr() (ir.Expr, error) {
	expr, err := p.parsePrimaryExpr()
	if err != nil {
		return ir.Expr{}, err
	}
	for {
		if p.match(tokenLParen) {
			args, err := p.parseExprList(tokenRParen)
			if err != nil {
				return ir.Expr{}, err
			}
			if err := p.expect(tokenRParen, "expected ')' after args"); err != nil {
				return ir.Expr{}, err
			}
			expr = ir.Expr{Kind: "call", Call: &ir.CallExpr{Callee: expr, Args: args}}
			continue
		}
		if p.match(tokenDot) {
			fieldTok, err := p.expectIdent("expected field name after '.'")
			if err != nil {
				return ir.Expr{}, err
			}
			expr = ir.Expr{Kind: "member", Member: &ir.MemberExpr{Object: expr, Field: fieldTok.lexeme}}
			continue
		}
		if p.match(tokenLBracket) {
			idx, err := p.parseExpr()
			if err != nil {
				return ir.Expr{}, err
			}
			if err := p.expect(tokenRBracket, "expected ']' after index"); err != nil {
				return ir.Expr{}, err
			}
			expr = ir.Expr{Kind: "index", Index: &ir.IndexExpr{Object: expr, Index: idx}}
			continue
		}
		break
	}
	return expr, nil
}

func (p *parser) parsePrimaryExpr() (ir.Expr, error) {
	if p.match(tokenNumber) {
		tok := p.previous()
		return ir.Expr{Kind: "literal", Literal: &ir.Literal{Kind: "number", Value: tok.lexeme}}, nil
	}
	if p.match(tokenString) {
		tok := p.previous()
		val, err := unquoteString(tok.lexeme)
		if err != nil {
			return ir.Expr{}, err
		}
		return ir.Expr{Kind: "literal", Literal: &ir.Literal{Kind: "string", Value: val}}, nil
	}
	if p.match(tokenIdent) {
		tok := p.previous()
		if tok.lexeme == "true" || tok.lexeme == "false" {
			return ir.Expr{Kind: "literal", Literal: &ir.Literal{Kind: "bool", Value: tok.lexeme}}, nil
		}
		return ir.Expr{Kind: "ident", Ident: tok.lexeme}, nil
	}
	if p.match(tokenLParen) {
		expr, err := p.parseExpr()
		if err != nil {
			return ir.Expr{}, err
		}
		if err := p.expect(tokenRParen, "expected ')' after expression"); err != nil {
			return ir.Expr{}, err
		}
		return expr, nil
	}
	if p.match(tokenLBracket) {
		elems, err := p.parseExprList(tokenRBracket)
		if err != nil {
			return ir.Expr{}, err
		}
		if err := p.expect(tokenRBracket, "expected ']' after list"); err != nil {
			return ir.Expr{}, err
		}
		return ir.Expr{Kind: "list", List: &ir.ListExpr{Elements: elems}}, nil
	}
	return ir.Expr{}, p.errorf(p.peek(), "expected expression")
}

func (p *parser) parseExprList(stop tokenType) ([]ir.Expr, error) {
	var exprs []ir.Expr
	if p.check(stop) {
		return exprs, nil
	}
	for !p.check(stop) && !p.check(tokenEOF) {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)
		if p.match(tokenComma) {
			continue
		}
		if p.check(stop) {
			break
		}
		return nil, p.errorf(p.peek(), "expected ',' or '%s'", tokenName(stop))
	}
	return exprs, nil
}

func (p *parser) parseGenMeta() (*ir.GenMeta, error) {
	if err := p.expect(tokenLBrace, "expected '{' after @gen"); err != nil {
		return nil, err
	}
	meta := &ir.GenMeta{}
	for !p.check(tokenRBrace) && !p.check(tokenEOF) {
		if p.match(tokenComma) || p.match(tokenSemicolon) {
			continue
		}
		keyTok, err := p.expectIdent("expected @gen key")
		if err != nil {
			return nil, err
		}
		if err := p.expect(tokenColon, "expected ':' after @gen key"); err != nil {
			return nil, err
		}
		switch keyTok.lexeme {
		case "prompt_ref":
			val, err := p.parseLiteralValue()
			if err != nil {
				return nil, err
			}
			meta.PromptRef = val
		case "prompt_hash":
			val, err := p.parseLiteralValue()
			if err != nil {
				return nil, err
			}
			meta.PromptHash = val
		case "model_id":
			val, err := p.parseLiteralValue()
			if err != nil {
				return nil, err
			}
			meta.ModelID = val
		case "generator_pass":
			val, err := p.parseLiteralValue()
			if err != nil {
				return nil, err
			}
			meta.GeneratorPass = val
		case "timestamp":
			val, err := p.parseLiteralValue()
			if err != nil {
				return nil, err
			}
			meta.Timestamp = val
		case "context_refs":
			list, err := p.parseStringList()
			if err != nil {
				return nil, err
			}
			meta.ContextRefs = append(meta.ContextRefs, list...)
		case "tools_trace_refs":
			list, err := p.parseStringList()
			if err != nil {
				return nil, err
			}
			meta.ToolsTraceRefs = append(meta.ToolsTraceRefs, list...)
		case "model_params":
			params, err := p.parseKVList()
			if err != nil {
				return nil, err
			}
			meta.ModelParams = append(meta.ModelParams, params...)
		default:
			// Skip unknown keys to keep parser forward-compatible.
			if _, err := p.parseLiteralValue(); err != nil {
				return nil, err
			}
		}
		p.match(tokenComma)
	}
	if err := p.expect(tokenRBrace, "expected '}' after @gen"); err != nil {
		return nil, err
	}
	return meta, nil
}

func (p *parser) parseKVList() ([]ir.KV, error) {
	var params []ir.KV
	if p.match(tokenLBrace) {
		for !p.check(tokenRBrace) && !p.check(tokenEOF) {
			if p.match(tokenComma) || p.match(tokenSemicolon) {
				continue
			}
			keyTok, err := p.expectIdent("expected model_params key")
			if err != nil {
				return nil, err
			}
			if err := p.expect(tokenColon, "expected ':' after model_params key"); err != nil {
				return nil, err
			}
			val, err := p.parseLiteralValue()
			if err != nil {
				return nil, err
			}
			params = append(params, ir.KV{Key: keyTok.lexeme, Value: val})
			p.match(tokenComma)
		}
		if err := p.expect(tokenRBrace, "expected '}' after model_params"); err != nil {
			return nil, err
		}
		return params, nil
	}
	val, err := p.parseLiteralValue()
	if err != nil {
		return nil, err
	}
	return []ir.KV{{Key: "raw", Value: val}}, nil
}

func (p *parser) parseStringList() ([]string, error) {
	if err := p.expect(tokenLBracket, "expected '[' for list"); err != nil {
		return nil, err
	}
	var out []string
	for !p.check(tokenRBracket) && !p.check(tokenEOF) {
		if p.match(tokenComma) {
			continue
		}
		val, err := p.parseLiteralValue()
		if err != nil {
			return nil, err
		}
		out = append(out, val)
		p.match(tokenComma)
	}
	if err := p.expect(tokenRBracket, "expected ']' after list"); err != nil {
		return nil, err
	}
	return out, nil
}

func (p *parser) parseLiteralValue() (string, error) {
	tok := p.peek()
	if p.match(tokenString) {
		return unquoteString(tok.lexeme)
	}
	if p.match(tokenNumber) || p.match(tokenIdent) {
		return tok.lexeme, nil
	}
	return "", p.errorf(tok, "expected literal value")
}

func (p *parser) parseTapePath() (string, error) {
	if p.check(tokenString) {
		tok := p.advance()
		return unquoteString(tok.lexeme)
	}
	if p.check(tokenIdent) {
		first := p.advance().lexeme
		if p.check(tokenString) {
			tok := p.advance()
			val, err := unquoteString(tok.lexeme)
			if err != nil {
				return "", err
			}
			return val, nil
		}
		return first, nil
	}
	return "", p.errorf(p.peek(), "expected tape path")
}

func (p *parser) parsePackRef() (ir.PackRef, error) {
	nameTok, err := p.expectIdent("expected pack name")
	if err != nil {
		return ir.PackRef{}, err
	}
	ref := ir.PackRef{Name: nameTok.lexeme}
	if p.match(tokenAt) {
		ver, err := p.parseVersion()
		if err != nil {
			return ir.PackRef{}, err
		}
		ref.Version = ver
	}
	return ref, nil
}

func (p *parser) parseVersion() (string, error) {
	valTok := p.peek()
	if !p.match(tokenIdent) && !p.match(tokenNumber) && !p.match(tokenString) {
		return "", p.errorf(valTok, "expected version")
	}
	if valTok.typ == tokenString {
		return unquoteString(valTok.lexeme)
	}
	return valTok.lexeme, nil
}

func (p *parser) parsePolicyDecl() (ir.PolicyDecl, error) {
	nameTok, err := p.expectIdent("expected policy name")
	if err != nil {
		return ir.PolicyDecl{}, err
	}
	if err := p.expect(tokenColon, "expected ':' after policy name"); err != nil {
		return ir.PolicyDecl{}, err
	}
	body, err := p.collectUntil(tokenSemicolon)
	if err != nil {
		return ir.PolicyDecl{}, err
	}
	return ir.PolicyDecl{Name: nameTok.lexeme, Body: body}, nil
}

func (p *parser) parseQName() (string, error) {
	first, err := p.expectIdent("expected identifier")
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString(first.lexeme)
	for {
		if p.match(tokenDot) {
			b.WriteString(".")
			part, err := p.expectIdent("expected identifier after '.'")
			if err != nil {
				return "", err
			}
			b.WriteString(part.lexeme)
			continue
		}
		if p.match(tokenDoubleColon) {
			b.WriteString("::")
			part, err := p.expectIdent("expected identifier after '::'")
			if err != nil {
				return "", err
			}
			b.WriteString(part.lexeme)
			continue
		}
		if p.match(tokenColon) {
			b.WriteString(":")
			part, err := p.expectIdent("expected identifier after ':'")
			if err != nil {
				return "", err
			}
			b.WriteString(part.lexeme)
			continue
		}
		break
	}
	return b.String(), nil
}

func (p *parser) parseTypeRef() (string, error) {
	var parts []string
	depthAngle := 0
	depthBracket := 0
	for {
		tok := p.peek()
		if tok.typ == tokenEOF {
			return "", p.errorf(tok, "unexpected end of type reference")
		}
		if depthAngle == 0 && depthBracket == 0 {
			if tok.typ == tokenComma || tok.typ == tokenSemicolon || tok.typ == tokenRBrace || tok.typ == tokenRParen {
				break
			}
		}
		if tok.typ == tokenLT {
			depthAngle++
		}
		if tok.typ == tokenGT && depthAngle > 0 {
			depthAngle--
		}
		if tok.typ == tokenLBracket {
			depthBracket++
		}
		if tok.typ == tokenRBracket && depthBracket > 0 {
			depthBracket--
		}
		parts = append(parts, tok.lexeme)
		p.advance()
	}
	if len(parts) == 0 {
		return "", p.errorf(p.peek(), "expected type reference")
	}
	return strings.Join(parts, ""), nil
}

func (p *parser) collectUntil(stop tokenType) (string, error) {
	var parts []token
	for !p.check(stop) && !p.check(tokenEOF) {
		parts = append(parts, p.advance())
	}
	if err := p.expect(stop, "expected terminator"); err != nil {
		return "", err
	}
	return tokensToString(parts), nil
}

func tokensToString(tokens []token) string {
	var b strings.Builder
	for i, tok := range tokens {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(tok.lexeme)
	}
	return strings.TrimSpace(b.String())
}

func (p *parser) match(typ tokenType) bool {
	if p.check(typ) {
		p.advance()
		return true
	}
	return false
}

func (p *parser) matchIdent(value string) bool {
	if p.check(tokenIdent) && p.peek().lexeme == value {
		p.advance()
		return true
	}
	return false
}

func (p *parser) check(typ tokenType) bool {
	return p.peek().typ == typ
}

func (p *parser) advance() token {
	if !p.check(tokenEOF) {
		p.pos++
	}
	return p.previous()
}

func (p *parser) previous() token {
	if p.pos == 0 {
		return p.tokens[0]
	}
	return p.tokens[p.pos-1]
}

func (p *parser) peek() token {
	return p.tokens[p.pos]
}

func (p *parser) peekNext() token {
	if p.pos+1 >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[p.pos+1]
}

func (p *parser) expect(typ tokenType, msg string) error {
	if p.check(typ) {
		p.advance()
		return nil
	}
	return p.errorf(p.peek(), msg)
}

func (p *parser) expectIdent(msg string) (token, error) {
	if p.check(tokenIdent) {
		return p.advance(), nil
	}
	return token{}, p.errorf(p.peek(), msg)
}

func (p *parser) errorf(tok token, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("parse error at %d:%d: %s", tok.line, tok.col, msg)
}

func tokenName(typ tokenType) string {
	switch typ {
	case tokenRParen:
		return ")"
	case tokenRBrace:
		return "}"
	case tokenRBracket:
		return "]"
	case tokenSemicolon:
		return ";"
	default:
		return "token"
	}
}

func unquoteString(raw string) (string, error) {
	val, err := strconv.Unquote(raw)
	if err != nil {
		return "", err
	}
	return val, nil
}
