package policy

import (
	"strconv"
	"strings"

	"github.com/willams/lia/internal/ir"
)

// Engine applies a minimal policy/constraint DSL.
type Engine struct {
	RoleAllowlist   map[string]bool
	MaxModules      int
	MaxDepsPerMod   int
	EffectRules     []EffectRule
	DependencyRules []DependencyRule
	Warnings        []ir.Diagnostic
}

// EffectRule restricts effects by role.
type EffectRule struct {
	Role   string
	Effect string
	Kind   string // forbid|require
}

// DependencyRule restricts dependencies by role.
type DependencyRule struct {
	Role       string
	TargetRole string
	Kind       string // forbid|require
}

// NewEngine parses packs into an executable engine.
func NewEngine(packs []ir.Pack) (*Engine, []ir.Diagnostic) {
	eng := &Engine{
		RoleAllowlist: map[string]bool{},
	}
	var diags []ir.Diagnostic

	for _, p := range packs {
		for _, pol := range p.Policies {
			polLines := splitPolicyLines(pol.Body)
			for _, line := range polLines {
				if line == "" {
					continue
				}
				applied := eng.applyPolicyLine(pol.Name, line, &diags)
				if !applied {
					diags = append(diags, ir.Diagnostic{
						Severity: "warning",
						Message:  "policy not enforced (v0.1): " + pol.Name + ": " + line,
					})
				}
			}
		}

		for _, c := range p.Constraints {
			if rule, ok := parseEffectConstraint(c.Expr); ok {
				eng.EffectRules = append(eng.EffectRules, rule)
				continue
			}
			if rule, ok := parseDependencyConstraint(c.Expr); ok {
				eng.DependencyRules = append(eng.DependencyRules, rule)
				continue
			}
			diags = append(diags, ir.Diagnostic{
				Severity: "warning",
				Message:  "constraint not enforced (v0.1): " + c.Name + ": " + c.Expr,
			})
		}
	}

	return eng, diags
}

// ApplyLocal enforces local (single-module) rules.
func (e *Engine) ApplyLocal(p *ir.Program) []ir.Diagnostic {
	var diags []ir.Diagnostic
	if p == nil {
		return diags
	}

	for i := range p.Modules {
		m := &p.Modules[i]
		if len(e.RoleAllowlist) > 0 && m.Role != "" {
			if !e.RoleAllowlist[m.Role] {
				diags = append(diags, ir.Diagnostic{
					Severity: "error",
					Message:  "role not allowed by policy: " + m.Role,
					Path:     m.Name,
				})
			}
		}

		effects := moduleEffects(m)
		for _, rule := range e.EffectRules {
			if rule.Kind != "forbid" {
				continue
			}
			if (rule.Role == "" || rule.Role == m.Role) && effects[rule.Effect] {
				diags = append(diags, ir.Diagnostic{
					Severity: "error",
					Message:  "effect forbidden by policy: " + rule.Effect,
					Path:     m.Name,
				})
			}
		}
	}

	return diags
}

// ApplyGlobal enforces cross-module rules.
func (e *Engine) ApplyGlobal(p *ir.Program, depRoles map[string]map[string]bool) []ir.Diagnostic {
	var diags []ir.Diagnostic
	if p == nil {
		return diags
	}

	if e.MaxModules > 0 && len(p.Modules) > e.MaxModules {
		diags = append(diags, ir.Diagnostic{
			Severity: "error",
			Message:  "max_modules exceeded",
		})
	}

	for i := range p.Modules {
		m := &p.Modules[i]
		if e.MaxDepsPerMod > 0 && len(m.Requires) > e.MaxDepsPerMod {
			diags = append(diags, ir.Diagnostic{
				Severity: "error",
				Message:  "max_deps_per_module exceeded",
				Path:     m.Name,
			})
		}

		for _, rule := range e.DependencyRules {
			if rule.Kind != "forbid" {
				continue
			}
			if rule.Role != "" && rule.Role != m.Role {
				continue
			}
			deps := depRoles[m.Name]
			if deps != nil && deps[rule.TargetRole] {
				diags = append(diags, ir.Diagnostic{
					Severity: "error",
					Message:  "dependency forbidden by policy: " + rule.TargetRole,
					Path:     m.Name,
				})
			}
		}
	}

	return diags
}

func splitPolicyLines(body string) []string {
	lines := strings.Split(body, "\n")
	var out []string
	for _, line := range lines {
		trim := strings.TrimSpace(strings.TrimSuffix(line, ";"))
		trim = stripInlineComment(trim)
		trim = strings.TrimSpace(trim)
		if trim != "" {
			out = append(out, trim)
		}
	}
	return out
}

func (e *Engine) applyPolicyLine(policyName, line string, diags *[]ir.Diagnostic) bool {
	if strings.HasPrefix(line, "allow_roles ") {
		list := parseList(strings.TrimPrefix(line, "allow_roles "))
		for _, role := range list {
			e.RoleAllowlist[role] = true
		}
		return true
	}
	if strings.HasPrefix(line, "max_modules ") {
		val := strings.TrimSpace(strings.TrimPrefix(line, "max_modules "))
		if n, err := strconv.Atoi(val); err == nil {
			e.MaxModules = n
			return true
		}
		*diags = append(*diags, ir.Diagnostic{
			Severity: "warning",
			Message:  "invalid max_modules value: " + val,
		})
		return false
	}
	if strings.HasPrefix(line, "max_deps_per_module ") {
		val := strings.TrimSpace(strings.TrimPrefix(line, "max_deps_per_module "))
		if n, err := strconv.Atoi(val); err == nil {
			e.MaxDepsPerMod = n
			return true
		}
		*diags = append(*diags, ir.Diagnostic{
			Severity: "warning",
			Message:  "invalid max_deps_per_module value: " + val,
		})
		return false
	}
	return false
}

func parseEffectConstraint(expr string) (EffectRule, bool) {
	whenRole := parseWhenRole(expr)
	if !strings.HasPrefix(expr, "forbid(") {
		return EffectRule{}, false
	}
	if !strings.Contains(expr, "effect ") {
		return EffectRule{}, false
	}
	start := strings.Index(expr, "effect ")
	if start == -1 {
		return EffectRule{}, false
	}
	rest := expr[start+len("effect "):]
	end := strings.Index(rest, ")")
	if end == -1 {
		return EffectRule{}, false
	}
	effect := strings.TrimSpace(rest[:end])
	return EffectRule{Role: whenRole, Effect: effect, Kind: "forbid"}, true
}

func parseDependencyConstraint(expr string) (DependencyRule, bool) {
	whenRole := parseWhenRole(expr)
	if !strings.HasPrefix(expr, "forbid(") {
		return DependencyRule{}, false
	}
	needle := "depends_on role == "
	start := strings.Index(expr, needle)
	if start == -1 {
		return DependencyRule{}, false
	}
	rest := expr[start+len(needle):]
	role := trimQuoted(rest)
	if role == "" {
		return DependencyRule{}, false
	}
	return DependencyRule{Role: whenRole, TargetRole: role, Kind: "forbid"}, true
}

func parseWhenRole(expr string) string {
	needle := "when role == "
	idx := strings.Index(expr, needle)
	if idx == -1 {
		return ""
	}
	rest := strings.TrimSpace(expr[idx+len(needle):])
	return trimQuoted(rest)
}

func trimQuoted(s string) string {
	trim := strings.TrimSpace(s)
	if strings.HasPrefix(trim, "\"") {
		trim = strings.TrimPrefix(trim, "\"")
		if end := strings.Index(trim, "\""); end != -1 {
			return trim[:end]
		}
	}
	return ""
}

func parseList(s string) []string {
	start := strings.Index(s, "[")
	end := strings.Index(s, "]")
	if start == -1 || end == -1 || end <= start {
		return nil
	}
	body := s[start+1 : end]
	parts := strings.Split(body, ",")
	var out []string
	for _, p := range parts {
		item := strings.Trim(strings.TrimSpace(p), "\"")
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func stripInlineComment(line string) string {
	var out strings.Builder
	inQuote := false
	for i := 0; i < len(line); i++ {
		if line[i] == '"' {
			inQuote = !inQuote
		}
		if !inQuote && i+1 < len(line) && line[i] == '/' && line[i+1] == '/' {
			break
		}
		out.WriteByte(line[i])
	}
	return out.String()
}

func moduleEffects(m *ir.Module) map[string]bool {
	effects := map[string]bool{}
	for _, uc := range m.Usecases {
		for _, ef := range uc.Effects {
			effects[ef] = true
		}
	}
	for _, ad := range m.Adapters {
		for _, ef := range ad.Effects {
			effects[ef] = true
		}
	}
	return effects
}
