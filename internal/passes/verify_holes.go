package passes

import (
	"fmt"

	"github.com/willams/lia/internal/ir"
)

type VerifyHolesPass struct{}

func NewVerifyHolesPass() *VerifyHolesPass {
	return &VerifyHolesPass{}
}

func (v *VerifyHolesPass) Name() string { return "VerifyHoles" }
func (v *VerifyHolesPass) Version() string { return "0.1" }
func (v *VerifyHolesPass) Kind() PassKind { return VerifyPass }

func (v *VerifyHolesPass) Run(ctx *Context, p *ir.Program) (*Result, error) {
	var diags []ir.Diagnostic
	for _, mod := range p.Modules {
		for _, hole := range mod.Holes {
			diags = append(diags, ir.Diagnostic{
				Severity: "error",
				Message:  fmt.Sprintf("unresolved hole: %s (contract: %s)", hole.Name, hole.Contract),
				Path:     mod.Name,
			})
		}
	}
	return &Result{Diags: diags}, nil
}
