package policy

import (
	"testing"

	"github.com/willams/lia/internal/ir"
)

func TestNewEngine_AllowRoles(t *testing.T) {
	packs := []ir.Pack{
		{
			Name: "TestPack",
			Policies: []ir.PolicyDecl{
				{Name: "roles", Body: `allow_roles ["domain", "usecase", "port"]`},
			},
		},
	}
	eng, diags := NewEngine(packs)
	if len(diags) > 0 {
		for _, d := range diags {
			if d.Severity == "error" {
				t.Errorf("unexpected error: %s", d.Message)
			}
		}
	}
	if !eng.RoleAllowlist["domain"] {
		t.Error("expected domain in allowlist")
	}
	if !eng.RoleAllowlist["usecase"] {
		t.Error("expected usecase in allowlist")
	}
}

func TestNewEngine_MaxModules(t *testing.T) {
	packs := []ir.Pack{
		{
			Name: "TestPack",
			Policies: []ir.PolicyDecl{
				{Name: "limits", Body: "max_modules 5"},
			},
		},
	}
	eng, _ := NewEngine(packs)
	if eng.MaxModules != 5 {
		t.Errorf("expected max_modules=5, got %d", eng.MaxModules)
	}
}

func TestNewEngine_MaxDepsPerModule(t *testing.T) {
	packs := []ir.Pack{
		{
			Name: "TestPack",
			Policies: []ir.PolicyDecl{
				{Name: "limits", Body: "max_deps_per_module 3"},
			},
		},
	}
	eng, _ := NewEngine(packs)
	if eng.MaxDepsPerMod != 3 {
		t.Errorf("expected max_deps=3, got %d", eng.MaxDepsPerMod)
	}
}

func TestNewEngine_InvalidMaxModules(t *testing.T) {
	packs := []ir.Pack{
		{
			Name: "TestPack",
			Policies: []ir.PolicyDecl{
				{Name: "limits", Body: "max_modules abc"},
			},
		},
	}
	_, diags := NewEngine(packs)
	found := false
	for _, d := range diags {
		if d.Message == "invalid max_modules value: abc" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected invalid max_modules warning")
	}
}

func TestNewEngine_EffectConstraint(t *testing.T) {
	packs := []ir.Pack{
		{
			Name: "TestPack",
			Constraints: []ir.ConstraintDecl{
				{Name: "no_io", Expr: `forbid(effect io)`},
			},
		},
	}
	eng, _ := NewEngine(packs)
	if len(eng.EffectRules) != 1 {
		t.Fatalf("expected 1 effect rule, got %d", len(eng.EffectRules))
	}
	if eng.EffectRules[0].Effect != "io" || eng.EffectRules[0].Kind != "forbid" {
		t.Errorf("unexpected rule: %+v", eng.EffectRules[0])
	}
}

func TestNewEngine_DependencyConstraint(t *testing.T) {
	packs := []ir.Pack{
		{
			Name: "TestPack",
			Constraints: []ir.ConstraintDecl{
				{Name: "no_infra_dep", Expr: `forbid(depends_on role == "infra")`},
			},
		},
	}
	eng, _ := NewEngine(packs)
	if len(eng.DependencyRules) != 1 {
		t.Fatalf("expected 1 dep rule, got %d", len(eng.DependencyRules))
	}
	if eng.DependencyRules[0].TargetRole != "infra" {
		t.Errorf("unexpected target role: %s", eng.DependencyRules[0].TargetRole)
	}
}

func TestNewEngine_UnknownConstraint(t *testing.T) {
	packs := []ir.Pack{
		{
			Name: "TestPack",
			Constraints: []ir.ConstraintDecl{
				{Name: "custom", Expr: "custom_thing"},
			},
		},
	}
	_, diags := NewEngine(packs)
	found := false
	for _, d := range diags {
		if d.Severity == "warning" && d.Message == "constraint not enforced (v0.1): custom: custom_thing" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected unknown constraint warning")
	}
}

func TestApplyLocal_RoleNotAllowed(t *testing.T) {
	eng := &Engine{
		RoleAllowlist: map[string]bool{"domain": true},
	}
	prog := &ir.Program{
		Modules: []ir.Module{{Name: "test", Role: "infra"}},
	}
	diags := eng.ApplyLocal(prog)
	if len(diags) == 0 {
		t.Fatal("expected role not allowed")
	}
	if diags[0].Message != "role not allowed by policy: infra" {
		t.Errorf("unexpected: %s", diags[0].Message)
	}
}

func TestApplyLocal_EffectForbidden(t *testing.T) {
	eng := &Engine{
		RoleAllowlist: map[string]bool{},
		EffectRules: []EffectRule{
			{Kind: "forbid", Effect: "io"},
		},
	}
	prog := &ir.Program{
		Modules: []ir.Module{
			{
				Name: "test",
				Usecases: []ir.UsecaseDecl{
					{Name: "Test", Effects: []string{"io"}},
				},
			},
		},
	}
	diags := eng.ApplyLocal(prog)
	found := false
	for _, d := range diags {
		if d.Message == "effect forbidden by policy: io" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected effect forbidden diagnostic")
	}
}

func TestApplyLocal_NilProgram(t *testing.T) {
	eng := &Engine{RoleAllowlist: map[string]bool{}}
	diags := eng.ApplyLocal(nil)
	if len(diags) != 0 {
		t.Error("expected no diags for nil program")
	}
}

func TestApplyGlobal_MaxModulesExceeded(t *testing.T) {
	eng := &Engine{
		RoleAllowlist: map[string]bool{},
		MaxModules:    1,
	}
	prog := &ir.Program{
		Modules: []ir.Module{{Name: "a"}, {Name: "b"}},
	}
	diags := eng.ApplyGlobal(prog, nil)
	found := false
	for _, d := range diags {
		if d.Message == "max_modules exceeded" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected max_modules exceeded")
	}
}

func TestApplyGlobal_MaxDepsExceeded(t *testing.T) {
	eng := &Engine{
		RoleAllowlist: map[string]bool{},
		MaxDepsPerMod: 1,
	}
	prog := &ir.Program{
		Modules: []ir.Module{
			{
				Name: "test",
				Requires: []ir.SymbolRef{
					{QName: "a::type:A"},
					{QName: "b::type:B"},
				},
			},
		},
	}
	diags := eng.ApplyGlobal(prog, nil)
	found := false
	for _, d := range diags {
		if d.Message == "max_deps_per_module exceeded" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected max_deps exceeded")
	}
}

func TestApplyGlobal_DependencyForbidden(t *testing.T) {
	eng := &Engine{
		RoleAllowlist: map[string]bool{},
		DependencyRules: []DependencyRule{
			{Kind: "forbid", TargetRole: "infra"},
		},
	}
	prog := &ir.Program{
		Modules: []ir.Module{{Name: "test", Role: "domain"}},
	}
	depRoles := map[string]map[string]bool{
		"test": {"infra": true},
	}
	diags := eng.ApplyGlobal(prog, depRoles)
	found := false
	for _, d := range diags {
		if d.Message == "dependency forbidden by policy: infra" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected dependency forbidden")
	}
}

func TestApplyGlobal_NilProgram(t *testing.T) {
	eng := &Engine{RoleAllowlist: map[string]bool{}}
	diags := eng.ApplyGlobal(nil, nil)
	if len(diags) != 0 {
		t.Error("expected no diags for nil program")
	}
}

func TestSplitPolicyLines(t *testing.T) {
	body := `allow_roles ["domain"];
max_modules 10;
// comment
`
	lines := splitPolicyLines(body)
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %v", len(lines), lines)
	}
}

func TestStripInlineComment(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`hello // world`, "hello "},
		{`"url://test" value`, `"url://test" value`},
		{`no comment`, "no comment"},
	}
	for _, tc := range tests {
		got := stripInlineComment(tc.input)
		if got != tc.want {
			t.Errorf("stripInlineComment(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestParseList_Empty(t *testing.T) {
	got := parseList("no brackets here")
	if len(got) != 0 {
		t.Errorf("expected empty list, got %v", got)
	}
}

func TestParseList_EmptyBrackets(t *testing.T) {
	got := parseList("[]")
	if len(got) != 0 {
		t.Errorf("expected empty list, got %v", got)
	}
}

func TestTrimQuoted(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"hello"`, "hello"},
		{`"test" extra`, "test"},
		{"unquoted", ""},
		{`""`, ""},
	}
	for _, tc := range tests {
		got := trimQuoted(tc.input)
		if got != tc.want {
			t.Errorf("trimQuoted(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestModuleEffects(t *testing.T) {
	m := &ir.Module{
		Usecases: []ir.UsecaseDecl{
			{Name: "A", Effects: []string{"io", "tx"}},
		},
		Adapters: []ir.AdapterDecl{
			{Name: "B", Effects: []string{"emit"}},
		},
	}
	effects := moduleEffects(m)
	if !effects["io"] || !effects["tx"] || !effects["emit"] {
		t.Errorf("expected io, tx, emit, got %v", effects)
	}
}

func TestParseEffectConstraint_Invalid(t *testing.T) {
	_, ok := parseEffectConstraint("require(something)")
	if ok {
		t.Error("expected false for non-forbid")
	}
	_, ok = parseEffectConstraint("forbid(depends_on role)")
	if ok {
		t.Error("expected false for non-effect constraint")
	}
}

func TestParseDependencyConstraint_Invalid(t *testing.T) {
	_, ok := parseDependencyConstraint("require(something)")
	if ok {
		t.Error("expected false for non-forbid")
	}
	_, ok = parseDependencyConstraint("forbid(effect io)")
	if ok {
		t.Error("expected false for effect constraint")
	}
}

func TestParseWhenRole(t *testing.T) {
	got := parseWhenRole(`when role == "domain" forbid(effect io)`)
	if got != "domain" {
		t.Errorf("expected domain, got %s", got)
	}
	got = parseWhenRole("no when here")
	if got != "" {
		t.Errorf("expected empty, got %s", got)
	}
}

func TestNewEngine_UnknownPolicyLine(t *testing.T) {
	packs := []ir.Pack{
		{
			Name: "TestPack",
			Policies: []ir.PolicyDecl{
				{Name: "custom", Body: "unknown_directive value"},
			},
		},
	}
	_, diags := NewEngine(packs)
	found := false
	for _, d := range diags {
		if d.Severity == "warning" && d.Message == "policy not enforced (v0.1): custom: unknown_directive value" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected unknown policy warning")
	}
}

func TestEffectConstraint_WithRole(t *testing.T) {
	rule, ok := parseEffectConstraint(`forbid(effect io) when role == "domain"`)
	if !ok {
		t.Fatal("expected valid constraint")
	}
	if rule.Role != "domain" {
		t.Errorf("expected domain role, got %s", rule.Role)
	}
	if rule.Effect != "io" {
		t.Errorf("expected io effect, got %s", rule.Effect)
	}
}

func TestNewEngine_EmptyPacks(t *testing.T) {
	eng, diags := NewEngine(nil)
	if eng == nil {
		t.Fatal("expected non-nil engine")
	}
	if len(diags) != 0 {
		t.Errorf("unexpected diags: %v", diags)
	}
}
