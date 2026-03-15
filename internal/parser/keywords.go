package parser

// LIAKeywords is the single source of truth for all LIA language keywords.
// It is referenced by TypeRef.Parse, RawExpr.Parse, the lexer Keyword pattern,
// and any future component that needs to distinguish keywords from identifiers.
var LIAKeywords = map[string]bool{
	// top-level constructs
	"project": true,
	"module":  true,
	"pack":    true,

	// module roles
	"as": true,

	// declarations
	"type":    true,
	"enum":    true,
	"port":    true,
	"usecase": true,
	"adapter": true,
	"wiring":  true,
	"record":  true,

	// blocks
	"input":   true,
	"output":  true,
	"effects": true,

	// modifiers & references
	"implements": true,
	"where":      true,
	"use":        true,

	// project-level
	"repro":      true,
	"tape":       true,
	"policy":     true,
	"constraint": true,
	"prefer":     true,
	"hole":       true,
	"candidate":  true,

	// statements
	"let":      true,
	"return":   true,
	"if":       true,
	"else":     true,
	"while":    true,
	"for":      true,
	"in":       true,
	"break":    true,
	"continue": true,

	// method declaration
	"fn":     true,
	"method": true,

	// wiring
	"bind": true,

	// score (candidate)
	"score":  true,
	"weight": true,
}

// IsLIAKeyword checks if a token value is a LIA keyword.
func IsLIAKeyword(s string) bool {
	return LIAKeywords[s]
}

// LIATypeTerminators returns the set of keywords that terminate a type reference
// when encountered at depth 0 (no angle/bracket nesting). This is a subset of
// LIAKeywords that is meaningful as a type-ref boundary.
func LIATypeTerminators() map[string]bool {
	return map[string]bool{
		"as":         true,
		"usecase":    true,
		"adapter":    true,
		"port":       true,
		"wiring":     true,
		"use":        true,
		"project":    true,
		"module":     true,
		"pack":       true,
		"repro":      true,
		"tape":       true,
		"policy":     true,
		"constraint": true,
		"prefer":     true,
		"hole":       true,
		"candidate":  true,
		"input":      true,
		"output":     true,
		"effects":    true,
		"implements": true,
		"type":       true,
		"enum":       true,
		"record":     true,
		"where":      true,
		"let":        true,
		"return":     true,
		"if":         true,
		"else":       true,
		"while":      true,
		"for":        true,
		"break":      true,
		"continue":   true,
		"fn":         true,
		"method":     true,
		"bind":       true,
	}
}
