package cli

import (
	"context"
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

		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return err
		}
		tapePath := filepath.Join(outDir, spec.TapeFile)
		if err := os.MkdirAll(filepath.Dir(tapePath), 0o755); err != nil {
			return err
		}

		generator := llmgen.NewGenerator(provider, tapePath)
		assets, err := generateProjectModules(context.Background(), generator, spec, cmd)
		if err != nil {
			return err
		}

		projectPath := filepath.Join(outDir, "project.lia")
		projectSource := renderProjectSource(spec, assets)
		if err := os.WriteFile(projectPath, []byte(projectSource), 0o644); err != nil {
			return err
		}

		prog, err := parser.ParseFile(projectPath)
		if err != nil {
			return fmt.Errorf("parse generated project: %w", err)
		}

		liaoPath := filepath.Join(outDir, "project.liao")
		if err := codec.WriteProgramFile(liaoPath, prog); err != nil {
			return err
		}

		packs, packDiags := loadPacksForProgram(cmd, prog)
		diags := append(packDiags, check.CheckProgram(prog, packs)...)
		printDiags(cmd, diags)
		if check.HasErrors(diags) {
			return fmt.Errorf("generated project failed validation")
		}

		linked, log, linkDiags, err := linker.Link([]*ir.Program{prog}, packs)
		if err != nil {
			return err
		}
		printDiags(cmd, linkDiags)
		if check.HasErrors(linkDiags) {
			return fmt.Errorf("generated project failed linking")
		}

		lialPath := filepath.Join(outDir, "project.lial")
		if err := codec.WriteProgramFile(lialPath, linked); err != nil {
			return err
		}
		logPath := filepath.Join(outDir, "project.lial.decision-log.json")
		if err := linker.WriteDecisionLog(logPath, log); err != nil {
			return err
		}

		tape, err := repro.LoadTape(tapePath)
		if err != nil {
			return err
		}
		replayDiags, summary := repro.ValidateProgramAgainstTape(linked, tape)
		printDiags(cmd, replayDiags)
		if check.HasErrors(replayDiags) {
			return fmt.Errorf("generated project failed replay validation")
		}

		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", projectPath)
		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", liaoPath)
		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", lialPath)
		fmt.Fprintf(cmd.OutOrStdout(), "decision log %s\n", logPath)
		fmt.Fprintf(cmd.OutOrStdout(), "prompt tape %s\n", tapePath)
		fmt.Fprintf(cmd.OutOrStdout(), "replay ok: modules %d, generated %d, validated %d\n", summary.Modules, summary.GeneratedModules, summary.ValidatedModules)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(genCmd)
	genCmd.AddCommand(genProjectCmd)

	genProjectCmd.Flags().String("spec", "", "path to project-spec.json")
	genProjectCmd.Flags().String("out-dir", "", "directory for generated project artifacts")
	genProjectCmd.Flags().String("provider", "ollama", "provider: ollama|openai-compatible")
	genProjectCmd.Flags().String("base-url", "", "provider base URL")
	genProjectCmd.Flags().String("api-key-env", "OPENAI_API_KEY", "env var for openai-compatible API key")
	genProjectCmd.Flags().String("model", "", "default model for modules without an explicit model")
	genProjectCmd.Flags().Float64("temperature", -1, "default temperature for modules without an explicit temperature")
	addPackDirFlag(genProjectCmd)
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

func generateProjectModules(ctx context.Context, generator *llmgen.Generator, spec *llmgen.ProjectSpec, cmd *cobra.Command) ([]*llmgen.GeneratedModule, error) {
	defaultModel, _ := cmd.Flags().GetString("model")
	defaultTemp, _ := cmd.Flags().GetFloat64("temperature")

	var assets []*llmgen.GeneratedModule
	for _, mod := range spec.Modules {
		model := mod.Model
		if model == "" {
			model = defaultModel
		}
		if model == "" {
			return nil, fmt.Errorf("no model configured for module %s", mod.Name)
		}

		temp := mod.Temperature
		if temp == 0 && defaultTemp >= 0 {
			temp = defaultTemp
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
		fmt.Fprintf(cmd.OutOrStdout(), "generated %s (%s)\n", mod.Name, mod.Role)
	}
	return assets, nil
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
