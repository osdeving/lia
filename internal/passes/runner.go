package passes

import "github.com/willams/lia/internal/ir"

// Runner executes a sequence of passes.
type Runner struct {
	passes []Pass
}

// NewRunner creates a new pass runner.
func NewRunner() *Runner {
	return &Runner{}
}

// Add appends a pass to the runner.
func (r *Runner) Add(pass Pass) {
	r.passes = append(r.passes, pass)
}

// Run executes all registered passes on the program context.
func (r *Runner) Run(ctx *Context, p *ir.Program) (*Result, error) {
	var allDiags []ir.Diagnostic
	res := &Result{}
	if p == nil {
		return res, nil
	}
	for _, pass := range r.passes {
		passRes, err := pass.Run(ctx, p)
		if err != nil {
			return nil, err
		}
		if passRes != nil {
			allDiags = append(allDiags, passRes.Diags...)
		}
	}
	res.Diags = allDiags
	return res, nil
}
