package passes

import "github.com/willams/lia/internal/ir"

// PassKind describes the behavior of a pass.
type PassKind int

const (
	VerifyPass PassKind = iota
	TransformPass
	SynthesizePass
)

// Context provides execution context to passes.
type Context struct {
	ReproProfile ir.ReproProfile
	TapePath     string
}

// Pass defines the interface for pipeline passes.
type Pass interface {
	Name() string
	Version() string
	Kind() PassKind
	Run(ctx *Context, unit *ir.Program) (*Result, error)
}

// Result is the standardized output of a pass.
type Result struct {
	Patch   *ir.Patch
	Diags   []ir.Diagnostic
	GenMeta *ir.GenMeta
}
