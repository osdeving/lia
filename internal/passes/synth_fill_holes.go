package passes

import (
	"github.com/willams/lia/internal/ir"
)

type SynthFillHolesPass struct{}

func NewSynthFillHolesPass() *SynthFillHolesPass {
	return &SynthFillHolesPass{}
}

func (s *SynthFillHolesPass) Name() string { return "SynthFillHoles" }
func (s *SynthFillHolesPass) Version() string { return "0.1" }
func (s *SynthFillHolesPass) Kind() PassKind { return SynthesizePass }

func (s *SynthFillHolesPass) Run(ctx *Context, p *ir.Program) (*Result, error) {
	var diags []ir.Diagnostic
	for i := range p.Modules {
		mod := &p.Modules[i]
		if len(mod.Holes) == 0 {
			continue
		}

		var remainingHoles []ir.HoleDecl
		for _, hole := range mod.Holes {
			resolved := false
			for _, cand := range mod.Candidates {
				if cand.Symbol == hole.Name {
					resolved = true
					diags = append(diags, ir.Diagnostic{
						Severity: "info",
						Message:  "resolved hole " + hole.Name + " with candidate " + cand.Symbol,
						Path:     mod.Name,
					})
					break
				}
			}
			if !resolved {
				remainingHoles = append(remainingHoles, hole)
			}
		}
		mod.Holes = remainingHoles
	}
	return &Result{Diags: diags}, nil
}
