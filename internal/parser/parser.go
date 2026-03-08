package parser

import (
	"os"

	"github.com/alecthomas/participle/v2"

	"github.com/willams/lia/internal/ir"
	"github.com/willams/lia/internal/symbols"
)

var programParser = participle.MustBuild[Program](
	participle.Lexer(liaLexer),
	participle.Elide("Whitespace", "Comment"),
	participle.UseLookahead(4),
)

// ParseFile parses a .lia source file into an IR program.
func ParseFile(path string) (*ir.Program, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseBytes(path, b)
}

// ParseBytes parses raw .lia bytes into an IR program.
func ParseBytes(filename string, data []byte) (*ir.Program, error) {
	ast, err := programParser.ParseBytes(filename, data)
	if err != nil {
		return nil, err
	}
	prog, err := ast.ToIR()
	if err != nil {
		return nil, err
	}
	_ = symbols.DeriveProgramSymbols(prog)
	return prog, nil
}

// ParseString parses a .lia string into an IR program.
func ParseString(input string) (*ir.Program, error) {
	ast, err := programParser.ParseString("<string>", input)
	if err != nil {
		return nil, err
	}
	prog, err := ast.ToIR()
	if err != nil {
		return nil, err
	}
	_ = symbols.DeriveProgramSymbols(prog)
	return prog, nil
}
