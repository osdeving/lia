package passes

import (
	"github.com/willams/lia/internal/ir"
)

type TransformDeduplicatePass struct{}

func NewTransformDeduplicatePass() *TransformDeduplicatePass {
	return &TransformDeduplicatePass{}
}

func (t *TransformDeduplicatePass) Name() string { return "TransformDeduplicate" }
func (t *TransformDeduplicatePass) Version() string { return "0.1" }
func (t *TransformDeduplicatePass) Kind() PassKind { return TransformPass }

func (t *TransformDeduplicatePass) Run(ctx *Context, p *ir.Program) (*Result, error) {
	var diags []ir.Diagnostic
	for i := range p.Modules {
		mod := &p.Modules[i]
		
		typeSet := make(map[string]bool)
		var uniqueTypes []ir.TypeDecl
		for _, decl := range mod.Types {
			key := decl.Name + ":" + decl.Base
			if !typeSet[key] {
				typeSet[key] = true
				uniqueTypes = append(uniqueTypes, decl)
			} else {
				diags = append(diags, ir.Diagnostic{
					Severity: "warning",
					Message:  "removed duplicate type declaration: " + decl.Name,
					Path:     mod.Name,
				})
			}
		}
		mod.Types = uniqueTypes

		enumSet := make(map[string]bool)
		var uniqueEnums []ir.EnumDecl
		for _, decl := range mod.Enums {
			if !enumSet[decl.Name] {
				enumSet[decl.Name] = true
				uniqueEnums = append(uniqueEnums, decl)
			} else {
				diags = append(diags, ir.Diagnostic{
					Severity: "warning",
					Message:  "removed duplicate enum declaration: " + decl.Name,
					Path:     mod.Name,
				})
			}
		}
		mod.Enums = uniqueEnums
	}
	return &Result{Diags: diags}, nil
}
