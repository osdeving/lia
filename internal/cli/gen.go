package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/willams/lia/internal/check"
	"github.com/willams/lia/internal/codec"
	"github.com/willams/lia/internal/ir"
	"github.com/willams/lia/internal/linker"
	"github.com/willams/lia/internal/llmgen"
	"github.com/willams/lia/internal/lower/java"
	"github.com/willams/lia/internal/packs"
	"github.com/willams/lia/internal/parser"
	"github.com/willams/lia/internal/repro"
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate LIA artifacts with an AI provider",
}

var genProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Generate a LIA project from a JSON spec",
	RunE: func(cmd *cobra.Command, args []string) error {
		specPath, _ := cmd.Flags().GetString("spec")
		if strings.TrimSpace(specPath) == "" {
			return fmt.Errorf("--spec is required")
		}
		outDir, _ := cmd.Flags().GetString("out-dir")
		if strings.TrimSpace(outDir) == "" {
			return fmt.Errorf("--out-dir is required")
		}

		spec, err := llmgen.LoadProjectSpec(specPath)
		if err != nil {
			return err
		}
		provider, err := buildGenerationProvider(cmd)
		if err != nil {
			return err
		}
		defaults, err := generationDefaultsFromFlags(cmd, nil)
		if err != nil {
			return err
		}

		result, err := generateProjectArtifacts(context.Background(), cmd, spec, provider, outDir, defaults)
		if err != nil {
			return err
		}
		printGeneratedProjectSummary(cmd, result)
		return nil
	},
}

var genAppCmd = &cobra.Command{
	Use:   "app",
	Short: "Generate an app from a free-form prompt",
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt, _ := cmd.Flags().GetString("prompt")
		if strings.TrimSpace(prompt) == "" {
			return fmt.Errorf("--prompt is required")
		}
		outDir, _ := cmd.Flags().GetString("out-dir")
		if strings.TrimSpace(outDir) == "" {
			return fmt.Errorf("--out-dir is required")
		}
		manifestPath, _ := cmd.Flags().GetString("manifest")
		target, _ := cmd.Flags().GetString("target")
		javaProfileName, _ := cmd.Flags().GetString("java-profile")
		javaProfilesValue, _ := cmd.Flags().GetString("java-profiles")
		javaStrict, _ := cmd.Flags().GetBool("java-strict")

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		workspace, err := llmgen.LoadWorkspaceContext(cwd, manifestPath, resolvePackDirs(cmd))
		if err != nil {
			return err
		}

		provider, err := buildGenerationProvider(cmd)
		if err != nil {
			return err
		}
		defaults, err := generationDefaultsFromFlags(cmd, workspace)
		if err != nil {
			return err
		}
		if strings.TrimSpace(target) == "" {
			target = workspace.Target
		}
		if strings.TrimSpace(target) == "" {
			target = "java"
		}

		manifestDirs := dedupeDirList(append(resolvePackDirs(cmd), workspace.PackDirs...))
		manifests, err := packs.LoadManifests(manifestDirs)
		if err != nil {
			return err
		}

		plan, err := llmgen.PlanProject(context.Background(), provider, llmgen.PlanRequest{
			Prompt:      prompt,
			Target:      target,
			Model:       defaults.Model,
			Temperature: defaults.Temperature,
			Workspace:   workspace,
			Manifests:   manifests,
		})
		if err != nil {
			return err
		}
		if err := writePlannerArtifacts(outDir, workspace, plan); err != nil {
			return err
		}

		result, err := generateProjectArtifacts(context.Background(), cmd, plan.Spec, provider, outDir, defaults)
		if err != nil {
			return err
		}
		printGeneratedProjectSummary(cmd, result)

		switch target {
		case "", "none":
			return nil
		case "java":
			if strings.TrimSpace(javaProfilesValue) == "" && strings.TrimSpace(javaProfileName) == "plain" && workspace != nil {
				switch strings.TrimSpace(strings.ToLower(workspace.Framework)) {
				case "spring", "spring-boot", "springboot":
					javaProfileName = "spring-boot"
				case "quarkus":
					javaProfileName = "quarkus"
				}
			}
			profiles, err := resolveJavaProfiles(javaProfileName, javaProfilesValue)
			if err != nil {
				return err
			}
			multi := len(profiles) > 1
			for _, profile := range profiles {
				javaOut := javaProfileOutputDir(outDir, profile, multi)
				project, err := java.LowerProjectWithOptions(result.Linked, java.Options{Profile: profile, Strict: javaStrict})
				if err != nil {
					return err
				}
				if err := os.MkdirAll(javaOut, 0o755); err != nil {
					return err
				}
				if err := java.WriteProject(javaOut, project); err != nil {
					return err
				}
				compile := compileJavaProject(javaOut)
				if err := os.WriteFile(javaProfileCompilePath(outDir, profile, multi), []byte(renderCompileResult(compile)), 0o644); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "java output %s (%s)\n", javaOut, profile)
				fmt.Fprintf(cmd.OutOrStdout(), "java compile %s (%s)\n", compileLabel(compile.Success), profile)
			}
			return nil
		default:
			return fmt.Errorf("unsupported target: %s", target)
		}
	},
}

type generationDefaults struct {
	Model       string
	Temperature float64
}

type generatedProjectResult struct {
	ProjectPath string
	LiaoPath    string
	LialPath    string
	LogPath     string
	TapePath    string
	Linked      *ir.Program
	Replay      repro.ValidationSummary
}

func init() {
	rootCmd.AddCommand(genCmd)
	genCmd.AddCommand(genProjectCmd)
	genCmd.AddCommand(genAppCmd)

	genProjectCmd.Flags().String("spec", "", "path to project-spec.json")
	genProjectCmd.Flags().String("out-dir", "", "directory for generated project artifacts")
	addPackDirFlag(genProjectCmd)
	addGenerationFlags(genProjectCmd)

	genAppCmd.Flags().String("prompt", "", "free-form app prompt")
	genAppCmd.Flags().String("out-dir", "", "directory for generated app artifacts")
	genAppCmd.Flags().String("manifest", "lia.json", "optional workspace manifest")
	genAppCmd.Flags().String("target", "java", "output target: java|none")
	addJavaProfileFlags(genAppCmd.Flags())
	genAppCmd.Flags().Bool("java-strict", false, "fail instead of generating placeholders or adapter stubs in java lowering")
	addPackDirFlag(genAppCmd)
	addGenerationFlags(genAppCmd)
}

func addGenerationFlags(cmd *cobra.Command) {
	cmd.Flags().String("provider", "ollama", "provider: ollama|openai-compatible")
	cmd.Flags().String("base-url", "", "provider base URL")
	cmd.Flags().String("api-key-env", "OPENAI_API_KEY", "env var for openai-compatible API key")
	cmd.Flags().String("model", "", "default model for planning/modules without an explicit model")
	cmd.Flags().Float64("temperature", -1, "default temperature for planning/modules without an explicit temperature")
}

func buildGenerationProvider(cmd *cobra.Command) (llmgen.Provider, error) {
	providerName, _ := cmd.Flags().GetString("provider")
	baseURL, _ := cmd.Flags().GetString("base-url")
	apiKeyEnv, _ := cmd.Flags().GetString("api-key-env")

	switch providerName {
	case "ollama":
		return llmgen.NewOllamaProvider(baseURL), nil
	case "openai-compatible":
		var apiKey string
		if strings.TrimSpace(apiKeyEnv) != "" {
			apiKey = os.Getenv(apiKeyEnv)
		}
		if strings.TrimSpace(baseURL) == "" {
			return nil, fmt.Errorf("--base-url is required for provider openai-compatible")
		}
		return llmgen.NewOpenAICompatibleProvider(baseURL, apiKey), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}
}

func generationDefaultsFromFlags(cmd *cobra.Command, workspace *llmgen.WorkspaceContext) (generationDefaults, error) {
	model, _ := cmd.Flags().GetString("model")
	temp, _ := cmd.Flags().GetFloat64("temperature")
	if strings.TrimSpace(model) == "" && workspace != nil {
		model = workspace.Model
	}
	if temp < 0 && workspace != nil && workspace.Temperature > 0 {
		temp = workspace.Temperature
	}
	if strings.TrimSpace(model) == "" {
		return generationDefaults{}, fmt.Errorf("no model configured")
	}
	if temp < 0 {
		temp = 0.1
	}
	return generationDefaults{
		Model:       model,
		Temperature: temp,
	}, nil
}

func generateProjectArtifacts(ctx context.Context, cmd *cobra.Command, spec *llmgen.ProjectSpec, provider llmgen.Provider, outDir string, defaults generationDefaults) (*generatedProjectResult, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	tapePath := filepath.Join(outDir, spec.TapeFile)
	if err := os.MkdirAll(filepath.Dir(tapePath), 0o755); err != nil {
		return nil, err
	}

	generator := llmgen.NewGenerator(provider, tapePath)
	assets, err := generateProjectModulesWithDefaults(ctx, generator, spec, defaults, cmd)
	if err != nil {
		return nil, err
	}

	projectPath := filepath.Join(outDir, "project.lia")
	projectSource := renderProjectSource(spec, assets)
	if err := os.WriteFile(projectPath, []byte(projectSource), 0o644); err != nil {
		return nil, err
	}

	prog, err := parser.ParseFile(projectPath)
	if err != nil {
		return nil, fmt.Errorf("parse generated project: %w", err)
	}

	liaoPath := filepath.Join(outDir, "project.liao")
	if err := codec.WriteProgramFile(liaoPath, prog); err != nil {
		return nil, err
	}

	loadedPacks, packDiags := loadPacksForProgram(cmd, prog)
	diags := append(packDiags, check.CheckProgram(prog, loadedPacks)...)
	printDiags(cmd, diags)
	if check.HasErrors(diags) {
		return nil, fmt.Errorf("generated project failed validation")
	}

	linked, log, linkDiags, err := linker.Link([]*ir.Program{prog}, loadedPacks)
	if err != nil {
		return nil, err
	}
	printDiags(cmd, linkDiags)
	if check.HasErrors(linkDiags) {
		return nil, fmt.Errorf("generated project failed linking")
	}

	lialPath := filepath.Join(outDir, "project.lial")
	if err := codec.WriteProgramFile(lialPath, linked); err != nil {
		return nil, err
	}
	logPath := filepath.Join(outDir, "project.lial.decision-log.json")
	if err := linker.WriteDecisionLog(logPath, log); err != nil {
		return nil, err
	}

	tape, err := repro.LoadTape(tapePath)
	if err != nil {
		return nil, err
	}
	replayDiags, summary := repro.ValidateProgramAgainstTape(linked, tape)
	printDiags(cmd, replayDiags)
	if check.HasErrors(replayDiags) {
		return nil, fmt.Errorf("generated project failed replay validation")
	}

	return &generatedProjectResult{
		ProjectPath: projectPath,
		LiaoPath:    liaoPath,
		LialPath:    lialPath,
		LogPath:     logPath,
		TapePath:    tapePath,
		Linked:      linked,
		Replay:      summary,
	}, nil
}

func generateProjectModules(ctx context.Context, generator *llmgen.Generator, spec *llmgen.ProjectSpec, cmd *cobra.Command) ([]*llmgen.GeneratedModule, error) {
	defaults, err := generationDefaultsFromFlags(cmd, nil)
	if err != nil {
		return nil, err
	}
	return generateProjectModulesWithDefaults(ctx, generator, spec, defaults, cmd)
}

func generateProjectModulesWithDefaults(ctx context.Context, generator *llmgen.Generator, spec *llmgen.ProjectSpec, defaults generationDefaults, cmd *cobra.Command) ([]*llmgen.GeneratedModule, error) {
	var assets []*llmgen.GeneratedModule
	for _, mod := range spec.Modules {
		model := mod.Model
		if model == "" {
			model = defaults.Model
		}
		if model == "" {
			return nil, fmt.Errorf("no model configured for module %s", mod.Name)
		}

		temp := mod.Temperature
		if temp == 0 {
			temp = defaults.Temperature
		}

		contextText := combineGenerationContext(spec.Brief, mod.Context)
		asset, err := generator.GenerateLIAModuleAsset(ctx, llmgen.ModuleSpec{
			Name:        mod.Name,
			Role:        mod.Role,
			Model:       model,
			Temperature: temp,
			Context:     contextText,
		})
		if err != nil {
			return nil, fmt.Errorf("generate module %s: %w", mod.Name, err)
		}
		assets = append(assets, asset)
		if cmd != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "generated %s (%s)\n", mod.Name, mod.Role)
		}
	}
	return assets, nil
}

func printGeneratedProjectSummary(cmd *cobra.Command, result *generatedProjectResult) {
	if result == nil {
		return
	}
	fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", result.ProjectPath)
	fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", result.LiaoPath)
	fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", result.LialPath)
	fmt.Fprintf(cmd.OutOrStdout(), "decision log %s\n", result.LogPath)
	fmt.Fprintf(cmd.OutOrStdout(), "prompt tape %s\n", result.TapePath)
	fmt.Fprintf(cmd.OutOrStdout(), "replay ok: modules %d, generated %d, validated %d\n", result.Replay.Modules, result.Replay.GeneratedModules, result.Replay.ValidatedModules)
}

func writePlannerArtifacts(outDir string, workspace *llmgen.WorkspaceContext, plan *llmgen.PlanResult) error {
	if plan == nil || plan.Spec == nil {
		return fmt.Errorf("nil plan")
	}
	planDir := filepath.Join(outDir, ".lia")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		return err
	}
	if workspace != nil {
		b, err := json.MarshalIndent(workspace, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(planDir, "workspace.json"), append(b, '\n'), 0o644); err != nil {
			return err
		}
	}
	if err := os.WriteFile(filepath.Join(planDir, "plan.prompt.txt"), []byte(plan.PlanningPrompt), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(planDir, "plan.raw.txt"), []byte(plan.RawResponse), 0o644); err != nil {
		return err
	}
	b, err := json.MarshalIndent(plan.Spec, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(planDir, "plan.json"), append(b, '\n'), 0o644)
}

func dedupeDirList(dirs []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, dir := range dirs {
		dir = strings.TrimSpace(dir)
		if dir == "" || seen[dir] {
			continue
		}
		seen[dir] = true
		out = append(out, dir)
	}
	return out
}

func combineGenerationContext(projectBrief, moduleContext string) string {
	projectBrief = strings.TrimSpace(projectBrief)
	moduleContext = strings.TrimSpace(moduleContext)
	switch {
	case projectBrief != "" && moduleContext != "":
		return "Project brief:\n" + projectBrief + "\n\nModule brief:\n" + moduleContext
	case projectBrief != "":
		return projectBrief
	default:
		return moduleContext
	}
}

func renderProjectSource(spec *llmgen.ProjectSpec, assets []*llmgen.GeneratedModule) string {
	var b strings.Builder
	b.WriteString("project ")
	b.WriteString(spec.Name)
	b.WriteString(" {\n")
	if spec.Repro != "" {
		b.WriteString("  repro ")
		b.WriteString(spec.Repro)
		b.WriteString(";\n")
	}
	if spec.TapeFile != "" {
		b.WriteString("  tape prompt_tape ")
		b.WriteString(strconv.Quote(spec.TapeFile))
		b.WriteString(";\n")
	}
	if len(spec.Packs) > 0 {
		b.WriteByte('\n')
		for _, pack := range spec.Packs {
			b.WriteString("  use pack ")
			b.WriteString(pack.Name)
			if pack.Version != "" {
				b.WriteString("@")
				b.WriteString(pack.Version)
			}
			b.WriteString(";\n")
		}
	}
	if len(spec.Policies) > 0 {
		b.WriteByte('\n')
		for _, policy := range spec.Policies {
			b.WriteString("  policy ")
			b.WriteString(policy.Name)
			b.WriteString(": ")
			b.WriteString(policy.Expr)
			b.WriteString(";\n")
		}
	}
	if len(spec.Constraints) > 0 {
		b.WriteByte('\n')
		for _, constraint := range spec.Constraints {
			b.WriteString("  constraint ")
			b.WriteString(constraint.Name)
			b.WriteString(": ")
			b.WriteString(constraint.Expr)
			b.WriteString(";\n")
		}
	}
	if len(assets) > 0 {
		b.WriteByte('\n')
		for i, asset := range assets {
			b.WriteString(indentBlock(renderGeneratedModule(asset), "  "))
			b.WriteByte('\n')
			if i < len(assets)-1 {
				b.WriteByte('\n')
			}
		}
	}
	b.WriteString("}\n")
	return b.String()
}

func renderGeneratedModule(asset *llmgen.GeneratedModule) string {
	if asset == nil {
		return ""
	}
	source := asset.Source
	if source == "" && asset.Module != nil {
		source = "module " + asset.Module.Name + " as " + asset.Module.Role + " {\n}"
	}
	var b strings.Builder
	if asset.Module != nil && asset.Module.Gen != nil {
		b.WriteString(renderGenBlock(asset.Module.Gen))
		b.WriteByte('\n')
	}
	b.WriteString(strings.TrimSpace(source))
	return b.String()
}

func renderGenBlock(meta *ir.GenMeta) string {
	if meta == nil {
		return ""
	}
	var parts []string
	appendStringField := func(key, value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		parts = append(parts, key+": "+strconv.Quote(value))
	}

	appendStringField("prompt_ref", meta.PromptRef)
	appendStringField("prompt_hash", meta.PromptHash)
	appendStringField("model_id", meta.ModelID)
	if len(meta.ContextRefs) > 0 {
		parts = append(parts, "context_refs: "+renderStringList(meta.ContextRefs))
	}
	if len(meta.ToolsTraceRefs) > 0 {
		parts = append(parts, "tools_trace_refs: "+renderStringList(meta.ToolsTraceRefs))
	}
	appendStringField("generator_pass", meta.GeneratorPass)
	appendStringField("timestamp", meta.Timestamp)
	if len(meta.ModelParams) > 0 {
		parts = append(parts, "model_params: "+renderKVList(meta.ModelParams))
	}

	var b strings.Builder
	b.WriteString("@gen {\n")
	for i, part := range parts {
		b.WriteString("  ")
		b.WriteString(part)
		if i < len(parts)-1 {
			b.WriteString(",")
		}
		b.WriteByte('\n')
	}
	b.WriteString("}")
	return b.String()
}

func renderStringList(values []string) string {
	if len(values) == 0 {
		return "[]"
	}
	var quoted []string
	for _, value := range values {
		quoted = append(quoted, strconv.Quote(value))
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func renderKVList(values []ir.KV) string {
	if len(values) == 0 {
		return "{}"
	}
	var parts []string
	for _, value := range values {
		if value.Key == "" {
			continue
		}
		parts = append(parts, value.Key+": "+strconv.Quote(value.Value))
	}
	if len(parts) == 0 {
		return "{}"
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

func indentBlock(input, prefix string) string {
	lines := strings.Split(strings.TrimSpace(input), "\n")
	for i, line := range lines {
		if line == "" {
			lines[i] = prefix
			continue
		}
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}
