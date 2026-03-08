package llmgen

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/willams/lia/internal/ir"
	"github.com/willams/lia/internal/packs"
)

const maxPlanningAttempts = 3

// PlanRequest describes a prompt-driven application planning request.
type PlanRequest struct {
	Prompt      string
	Target      string
	Model       string
	Temperature float64
	Workspace   *WorkspaceContext
	Manifests   []packs.Manifest
}

// PlanResult is the planner output plus traceable prompt/response artifacts.
type PlanResult struct {
	Spec           *ProjectSpec
	SelectedPacks  []packs.Manifest
	Workspace      *WorkspaceContext
	PlanningPrompt string
	RawResponse    string
}

// PlanProject generates an internal project plan from a free-form prompt.
func PlanProject(ctx context.Context, provider Provider, req PlanRequest) (*PlanResult, error) {
	prompt := buildPlanningPrompt(req)
	currentPrompt := prompt
	var lastResponse string
	var lastErr error

	for attempt := 1; attempt <= maxPlanningAttempts; attempt++ {
		resp, err := provider.Generate(ctx, GenerateRequest{
			Prompt:      currentPrompt,
			Model:       req.Model,
			Temperature: req.Temperature,
			MaxTokens:   4000,
			Metadata: map[string]interface{}{
				"kind":   "app-plan",
				"target": req.Target,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("plan project: %w", err)
		}
		lastResponse = resp.Content

		spec, parseErr := parsePlanResponse(resp.Content)
		if parseErr == nil {
			if strings.TrimSpace(spec.Brief) == "" {
				spec.Brief = strings.TrimSpace(req.Prompt)
			}
			if strings.TrimSpace(spec.Repro) == "" {
				spec.Repro = string(ir.ReproPinned)
			}
			applyWorkspaceDefaults(spec, req.Workspace)
			if err := applyPackDefaults(spec, req.Workspace, req.Manifests); err != nil {
				return nil, err
			}
			if err := ensureModuleTemplates(spec, req.Manifests); err != nil {
				return nil, err
			}
			if err := spec.Normalize(); err != nil {
				lastErr = err
			} else {
				return &PlanResult{
					Spec:           spec,
					SelectedPacks:  selectManifestDetails(spec.Packs, req.Manifests),
					Workspace:      req.Workspace,
					PlanningPrompt: prompt,
					RawResponse:    resp.Content,
				}, nil
			}
		} else {
			lastErr = parseErr
		}

		if attempt < maxPlanningAttempts {
			currentPrompt = buildPlanningRepairPrompt(prompt, lastResponse, lastErr)
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("planning failed")
	}
	return nil, fmt.Errorf("%w; last response: %s", lastErr, compactJSONText(lastResponse, 600))
}

func buildPlanningPrompt(req PlanRequest) string {
	var b strings.Builder
	b.WriteString("You are planning a LIA application from a free-form prompt.\n\n")
	b.WriteString("Return only JSON and nothing else.\n")
	b.WriteString("Do not use markdown fences.\n")
	b.WriteString("The JSON must match this schema:\n")
	b.WriteString("{\n")
	b.WriteString("  \"name\": \"ProjectName\",\n")
	b.WriteString("  \"brief\": \"normalized brief\",\n")
	b.WriteString("  \"repro\": \"pinned\",\n")
	b.WriteString("  \"packs\": [{\"name\":\"ArchBaseline\",\"version\":\"0.1.0\"}],\n")
	b.WriteString("  \"modules\": [\n")
	b.WriteString("    {\"name\":\"orders.domain\",\"role\":\"domain\",\"context\":\"...\"},\n")
	b.WriteString("    {\"name\":\"orders.port\",\"role\":\"port\",\"context\":\"...\"},\n")
	b.WriteString("    {\"name\":\"orders.app\",\"role\":\"usecase\",\"context\":\"...\"}\n")
	b.WriteString("  ]\n")
	b.WriteString("}\n\n")
	b.WriteString("Rules:\n")
	b.WriteString("- choose packs only from the available pack catalog below\n")
	b.WriteString("- prefer 3 to 5 modules\n")
	b.WriteString("- module names must be qualified names like orders.domain\n")
	b.WriteString("- roles allowed: domain, port, usecase\n")
	b.WriteString("- if ArchBaseline is available and the prompt describes a service/application, include it\n")
	b.WriteString("- adapt the plan to the workspace context and available packs\n")
	b.WriteString("- do not invent packs not present in the catalog\n")
	b.WriteString("- contexts should be short, concrete generation instructions for each module\n\n")
	b.WriteString("User prompt:\n")
	b.WriteString(strings.TrimSpace(req.Prompt))
	b.WriteString("\n\n")
	b.WriteString("Workspace context:\n")
	b.WriteString(renderWorkspaceContext(req.Workspace, req.Target))
	b.WriteString("\n\n")
	b.WriteString("Available pack catalog:\n")
	b.WriteString(renderPackCatalog(req.Manifests))
	b.WriteString("\n")
	return b.String()
}

func buildPlanningRepairPrompt(originalPrompt, previousOutput string, planErr error) string {
	return originalPrompt + "\n\nThe previous JSON was invalid.\nError:\n" + strings.TrimSpace(errorString(planErr)) + "\n\nPrevious output:\n" + compactJSONText(previousOutput, 1200) + "\n\nRewrite the full JSON from scratch."
}

func parsePlanResponse(content string) (*ProjectSpec, error) {
	raw := extractJSONObject(content)
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("planner did not return a JSON object")
	}
	var spec ProjectSpec
	if err := json.Unmarshal([]byte(raw), &spec); err != nil {
		return nil, err
	}
	if err := spec.Normalize(); err != nil && !strings.Contains(err.Error(), "requires at least one module") {
		return nil, err
	}
	return &spec, nil
}

func renderWorkspaceContext(workspace *WorkspaceContext, target string) string {
	if workspace == nil {
		return "(none)"
	}
	var lines []string
	if workspace.Name != "" {
		lines = append(lines, "- name: "+workspace.Name)
	}
	if target == "" {
		target = workspace.Target
	}
	if target != "" {
		lines = append(lines, "- target: "+target)
	}
	if workspace.Build != "" {
		lines = append(lines, "- build: "+workspace.Build)
	}
	if workspace.Framework != "" {
		lines = append(lines, "- framework: "+workspace.Framework)
	}
	if workspace.BasePackage != "" {
		lines = append(lines, "- base_package: "+workspace.BasePackage)
	}
	if len(workspace.Signals) > 0 {
		lines = append(lines, "- signals: "+strings.Join(workspace.Signals, ", "))
	}
	if len(workspace.BuiltinPacks) > 0 {
		lines = append(lines, "- builtin_packs: "+strings.Join(workspace.BuiltinPacks, ", "))
	}
	if len(lines) == 0 {
		return "(none)"
	}
	return strings.Join(lines, "\n")
}

func renderPackCatalog(manifests []packs.Manifest) string {
	if len(manifests) == 0 {
		return "(no pack manifests available)"
	}
	var lines []string
	for _, manifest := range manifests {
		line := "- " + manifest.Name
		if manifest.Version != "" {
			line += "@" + manifest.Version
		}
		if manifest.Summary != "" {
			line += ": " + manifest.Summary
		}
		if len(manifest.Keywords) > 0 {
			line += " | keywords=" + strings.Join(manifest.Keywords, ", ")
		}
		if len(manifest.ModuleTemplates) > 0 {
			var parts []string
			for _, tmpl := range manifest.ModuleTemplates {
				parts = append(parts, tmpl.Role+"/"+tmpl.Suffix)
			}
			line += " | modules=" + strings.Join(parts, ", ")
		}
		if manifest.DefaultSelected {
			line += " | default_selected=true"
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func applyWorkspaceDefaults(spec *ProjectSpec, workspace *WorkspaceContext) {
	if spec == nil || workspace == nil {
		return
	}
	if strings.TrimSpace(spec.Name) == "" && workspace.Name != "" {
		spec.Name = workspace.Name
	}
	if strings.TrimSpace(spec.Repro) == "" && workspace.Repro != "" {
		spec.Repro = workspace.Repro
	}
}

func applyPackDefaults(spec *ProjectSpec, workspace *WorkspaceContext, manifests []packs.Manifest) error {
	if spec == nil {
		return fmt.Errorf("nil project spec")
	}
	selected := map[string]bool{}
	for _, ref := range spec.Packs {
		selected[ref.Name+"@"+ref.Version] = true
	}

	for _, builtin := range builtinPackRefs(workspace, manifests) {
		key := builtin.Name + "@" + builtin.Version
		if selected[key] {
			continue
		}
		spec.Packs = append(spec.Packs, builtin)
		selected[key] = true
	}
	return nil
}

func builtinPackRefs(workspace *WorkspaceContext, manifests []packs.Manifest) []ir.PackRef {
	var refs []ir.PackRef
	for _, manifest := range manifests {
		if manifest.DefaultSelected {
			refs = append(refs, ir.PackRef{Name: manifest.Name, Version: manifest.Version})
		}
	}
	if workspace == nil {
		return dedupePackRefs(refs)
	}
	for _, name := range workspace.BuiltinPacks {
		for _, manifest := range manifests {
			if manifest.Name == name {
				refs = append(refs, ir.PackRef{Name: manifest.Name, Version: manifest.Version})
			}
		}
	}
	return dedupePackRefs(refs)
}

func ensureModuleTemplates(spec *ProjectSpec, manifests []packs.Manifest) error {
	if spec == nil {
		return fmt.Errorf("nil project spec")
	}
	if len(spec.Modules) > 0 {
		return nil
	}

	root := inferModuleRoot(spec)
	if root == "" {
		root = "app"
	}
	for _, manifest := range selectManifestDetails(spec.Packs, manifests) {
		for _, tmpl := range manifest.ModuleTemplates {
			if tmpl.Role == "" || tmpl.Suffix == "" {
				continue
			}
			spec.Modules = append(spec.Modules, ProjectModule{
				Name:    root + "." + tmpl.Suffix,
				Role:    tmpl.Role,
				Context: tmpl.Context,
			})
		}
	}
	if len(spec.Modules) == 0 {
		spec.Modules = []ProjectModule{
			{Name: root + ".domain", Role: "domain", Context: "Define the core domain vocabulary."},
			{Name: root + ".port", Role: "port", Context: "Define contracts and repositories needed by the application."},
			{Name: root + ".app", Role: "usecase", Context: "Define use cases, adapters, and wiring."},
		}
	}
	return nil
}

func inferModuleRoot(spec *ProjectSpec) string {
	if spec == nil {
		return ""
	}
	for _, mod := range spec.Modules {
		if idx := strings.Index(strings.TrimSpace(mod.Name), "."); idx > 0 {
			return strings.TrimSpace(mod.Name[:idx])
		}
	}
	name := strings.ToLower(strings.TrimSpace(spec.Name))
	if name == "" {
		return ""
	}
	name = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(name, ".")
	name = strings.Trim(name, ".")
	if name == "" {
		return ""
	}
	if idx := strings.Index(name, "."); idx > 0 {
		return name[:idx]
	}
	return name
}

func selectManifestDetails(refs []ir.PackRef, manifests []packs.Manifest) []packs.Manifest {
	selected := map[string]bool{}
	var out []packs.Manifest
	for _, ref := range refs {
		for _, manifest := range manifests {
			if manifest.Name != ref.Name {
				continue
			}
			if ref.Version != "" && manifest.Version != "" && manifest.Version != ref.Version {
				continue
			}
			key := manifest.Name + "@" + manifest.Version
			if selected[key] {
				continue
			}
			selected[key] = true
			out = append(out, manifest)
		}
	}
	return out
}

func dedupePackRefs(refs []ir.PackRef) []ir.PackRef {
	seen := map[string]bool{}
	var out []ir.PackRef
	for _, ref := range refs {
		key := ref.Name + "@" + ref.Version
		if ref.Name == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, ref)
	}
	return out
}

func extractJSONObject(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}
	input = strings.TrimPrefix(input, "```json")
	input = strings.TrimPrefix(input, "```")
	input = strings.TrimSuffix(input, "```")
	input = strings.TrimSpace(input)
	start := strings.IndexByte(input, '{')
	end := strings.LastIndexByte(input, '}')
	if start == -1 || end == -1 || end < start {
		return ""
	}
	return strings.TrimSpace(input[start : end+1])
}

func compactJSONText(input string, limit int) string {
	input = strings.TrimSpace(input)
	if len(input) <= limit {
		return input
	}
	return strings.TrimSpace(input[:limit]) + "..."
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
