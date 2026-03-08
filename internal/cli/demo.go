package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/cobra"

	"github.com/willams/lia/internal/check"
	"github.com/willams/lia/internal/codec"
	"github.com/willams/lia/internal/ir"
	"github.com/willams/lia/internal/linker"
	"github.com/willams/lia/internal/llmgen"
	lowerjava "github.com/willams/lia/internal/lower/java"
	"github.com/willams/lia/internal/parser"
	"github.com/willams/lia/internal/repro"
)

const maxDirectJavaAttempts = 3

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Run presentation-friendly demos",
}

var demoCompareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Generate and compare direct Java vs LIA -> Java",
	RunE: func(cmd *cobra.Command, args []string) error {
		specPath, _ := cmd.Flags().GetString("spec")
		if strings.TrimSpace(specPath) == "" {
			return fmt.Errorf("--spec is required")
		}
		outDir, _ := cmd.Flags().GetString("out-dir")
		if strings.TrimSpace(outDir) == "" {
			return fmt.Errorf("--out-dir is required")
		}
		referenceDir, _ := cmd.Flags().GetString("reference-dir")

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

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		briefPath := filepath.Join(outDir, "common-brief.md")
		if err := os.WriteFile(briefPath, []byte(renderCommonBrief(spec)), 0o644); err != nil {
			return err
		}

		directDir := filepath.Join(outDir, "direct-java")
		directModel, directTemp, err := resolveCompareModelConfig(cmd, "direct")
		if err != nil {
			return err
		}
		directResult, err := generateDirectJavaProject(ctx, provider, spec, directDir, directModel, directTemp)
		if err != nil {
			return err
		}

		liaDir := filepath.Join(outDir, "lia-artifacts")
		liaModel, liaTemp, err := resolveCompareModelConfig(cmd, "lia")
		if err != nil {
			return err
		}
		liaResult, err := generateLIAProjectForCompare(ctx, cmd, provider, spec, liaDir, liaModel, liaTemp)
		if err != nil {
			return err
		}

		liaJavaDir := filepath.Join(outDir, "lia-java")
		loweredProject, err := lowerjava.LowerProject(liaResult.Linked)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(liaJavaDir, 0o755); err != nil {
			return err
		}
		if err := lowerjava.WriteProject(liaJavaDir, loweredProject); err != nil {
			return err
		}

		liaJavaCompile := compileJavaProject(liaJavaDir)
		directMetrics := analyzeJavaProject(directDir)
		liaJavaMetrics := analyzeJavaProject(liaJavaDir)

		var reference *comparisonVariant
		if strings.TrimSpace(referenceDir) != "" {
			if info, err := os.Stat(referenceDir); err == nil && info.IsDir() {
				refCompile := compileJavaProject(referenceDir)
				refMetrics := analyzeJavaProject(referenceDir)
				if err := os.WriteFile(filepath.Join(outDir, "reference-java.compile.txt"), []byte(renderCompileResult(refCompile)), 0o644); err != nil {
					return err
				}
				reference = &comparisonVariant{
					Name:         "Reference Java",
					Dir:          referenceDir,
					Compile:      refCompile,
					Metrics:      refMetrics,
					HasLIAInputs: false,
				}
			}
		}

		directVariant := comparisonVariant{
			Name:           "Direct Java",
			Dir:            directDir,
			Compile:        directResult.Compile,
			Metrics:        directMetrics,
			PromptPath:     directResult.PromptPath,
			RawResponse:    directResult.RawResponsePath,
			RepairAttempts: directResult.Repairs,
			HasLIAInputs:   false,
		}
		if err := os.WriteFile(filepath.Join(outDir, "direct-java.compile.txt"), []byte(renderCompileResult(directResult.Compile)), 0o644); err != nil {
			return err
		}
		liaVariant := comparisonVariant{
			Name:           "LIA -> Java",
			Dir:            liaJavaDir,
			Compile:        liaJavaCompile,
			Metrics:        liaJavaMetrics,
			PromptPath:     filepath.Join(liaDir, "project.lia"),
			RawResponse:    filepath.Join(liaDir, "prompt-tape.json"),
			RepairAttempts: 0,
			HasLIAInputs:   true,
		}
		if err := os.WriteFile(filepath.Join(outDir, "lia-java.compile.txt"), []byte(renderCompileResult(liaJavaCompile)), 0o644); err != nil {
			return err
		}

		reportPath := filepath.Join(outDir, "COMPARISON.md")
		report := renderComparisonReport(spec, briefPath, reference, directVariant, liaVariant, liaResult)
		if err := os.WriteFile(reportPath, []byte(report), 0o644); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "direct java: %s (%s)\n", directDir, compileLabel(directResult.Compile.Success))
		fmt.Fprintf(cmd.OutOrStdout(), "lia artifacts: %s\n", liaDir)
		fmt.Fprintf(cmd.OutOrStdout(), "lia java: %s (%s)\n", liaJavaDir, compileLabel(liaJavaCompile.Success))
		fmt.Fprintf(cmd.OutOrStdout(), "comparison report: %s\n", reportPath)
		return nil
	},
}

type generatedJavaProject struct {
	PromptPath      string
	RawResponsePath string
	Compile         javaCompileResult
	Repairs         int
}

type liaCompareResult struct {
	Linked       *ir.Program
	Replay       repro.ValidationSummary
	ProjectPath  string
	TapePath     string
	DecisionPath string
}

type javaCompileResult struct {
	Success bool
	Errors  string
	Files   int
}

type javaProjectMetrics struct {
	JavaFiles       int
	Records         int
	Interfaces      int
	Enums           int
	UsecaseClasses  int
	AdapterClasses  int
	WiringClasses   int
	DomainFiles     int
	PortFiles       int
	ApplicationDirs int
	BootstrapFiles  int
}

type comparisonVariant struct {
	Name           string
	Dir            string
	Compile        javaCompileResult
	Metrics        javaProjectMetrics
	PromptPath     string
	RawResponse    string
	RepairAttempts int
	HasLIAInputs   bool
}

type textFile struct {
	Path    string
	Content string
}

func init() {
	rootCmd.AddCommand(demoCmd)
	demoCmd.AddCommand(demoCompareCmd)

	demoCompareCmd.Flags().String("spec", "", "path to project-spec.json")
	demoCompareCmd.Flags().String("out-dir", "", "directory for comparison outputs")
	demoCompareCmd.Flags().String("reference-dir", filepath.Join("examples", "java-reference", "orders-service"), "optional handcrafted Java reference project")
	demoCompareCmd.Flags().String("provider", "ollama", "provider: ollama|openai-compatible")
	demoCompareCmd.Flags().String("base-url", "", "provider base URL")
	demoCompareCmd.Flags().String("api-key-env", "OPENAI_API_KEY", "env var for openai-compatible API key")
	demoCompareCmd.Flags().String("model", "", "default model for direct Java and LIA generation")
	demoCompareCmd.Flags().String("direct-model", "", "model override for the direct Java branch")
	demoCompareCmd.Flags().String("lia-model", "", "model override for the LIA branch")
	demoCompareCmd.Flags().Float64("temperature", 0.1, "default temperature for both branches")
	demoCompareCmd.Flags().Float64("direct-temperature", -1, "temperature override for the direct Java branch")
	demoCompareCmd.Flags().Float64("lia-temperature", -1, "temperature override for the LIA branch")
	addPackDirFlag(demoCompareCmd)
}

func resolveCompareModelConfig(cmd *cobra.Command, branch string) (string, float64, error) {
	defaultModel, _ := cmd.Flags().GetString("model")
	defaultTemp, _ := cmd.Flags().GetFloat64("temperature")

	var model string
	var temp float64
	switch branch {
	case "direct":
		model, _ = cmd.Flags().GetString("direct-model")
		temp, _ = cmd.Flags().GetFloat64("direct-temperature")
	case "lia":
		model, _ = cmd.Flags().GetString("lia-model")
		temp, _ = cmd.Flags().GetFloat64("lia-temperature")
	default:
		return "", 0, fmt.Errorf("unknown branch: %s", branch)
	}
	if strings.TrimSpace(model) == "" {
		model = defaultModel
	}
	if strings.TrimSpace(model) == "" {
		return "", 0, fmt.Errorf("no model configured for %s branch", branch)
	}
	if temp < 0 {
		temp = defaultTemp
	}
	return model, temp, nil
}

func generateDirectJavaProject(ctx context.Context, provider llmgen.Provider, spec *llmgen.ProjectSpec, outDir, model string, temperature float64) (*generatedJavaProject, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}

	basePackage := "demo.compare." + sanitizeCompareSegment(spec.Name)
	initialPrompt := buildDirectJavaPrompt(spec, basePackage)
	promptPath := filepath.Join(outDir, "prompt.txt")
	if err := os.WriteFile(promptPath, []byte(initialPrompt), 0o644); err != nil {
		return nil, err
	}

	prompt := initialPrompt
	rawResponsePath := filepath.Join(outDir, "response.txt")
	result := &generatedJavaProject{PromptPath: promptPath, RawResponsePath: rawResponsePath}
	var lastCompile javaCompileResult
	var lastAttemptDir string

	for attempt := 1; attempt <= maxDirectJavaAttempts; attempt++ {
		resp, err := provider.Generate(ctx, llmgen.GenerateRequest{
			Prompt:      prompt,
			Model:       model,
			Temperature: temperature,
			MaxTokens:   6000,
		})
		if err != nil {
			return nil, fmt.Errorf("generate direct java: %w", err)
		}
		if err := os.WriteFile(rawResponsePath, []byte(resp.Content), 0o644); err != nil {
			return nil, err
		}

		files, err := parseTextFiles(resp.Content)
		if err != nil {
			prompt = buildDirectJavaRepairPrompt(initialPrompt, "The previous answer did not respect the FILE marker format: "+err.Error())
			result.Repairs = attempt
			continue
		}

		attemptDir := filepath.Join(outDir, fmt.Sprintf("attempt-%d", attempt))
		if err := writeTextProject(attemptDir, files); err != nil {
			return nil, err
		}
		lastAttemptDir = attemptDir
		lastCompile = compileJavaProject(attemptDir)
		if lastCompile.Success {
			if err := copyDir(attemptDir, outDir); err != nil {
				return nil, err
			}
			result.Compile = compileJavaProject(outDir)
			result.Repairs = attempt - 1
			return result, nil
		}

		prompt = buildDirectJavaRepairPrompt(initialPrompt, lastCompile.Errors)
		result.Repairs = attempt
	}

	result.Compile = lastCompile
	if lastAttemptDir != "" {
		if err := copyDir(lastAttemptDir, outDir); err != nil {
			return nil, err
		}
		result.Compile = compileJavaProject(outDir)
	}
	return result, nil
}

func generateLIAProjectForCompare(ctx context.Context, cmd *cobra.Command, provider llmgen.Provider, spec *llmgen.ProjectSpec, outDir, model string, temperature float64) (*liaCompareResult, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}

	tapePath := filepath.Join(outDir, spec.TapeFile)
	if err := os.MkdirAll(filepath.Dir(tapePath), 0o755); err != nil {
		return nil, err
	}

	generator := llmgen.NewGenerator(provider, tapePath)
	assets, err := generateProjectModulesForCompare(ctx, generator, spec, model, temperature)
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

	packs, packDiags := loadPacksForProgram(cmd, prog)
	diags := append(packDiags, check.CheckProgram(prog, packs)...)
	if check.HasErrors(diags) {
		return nil, fmt.Errorf("generated project failed validation")
	}

	linked, log, linkDiags, err := linker.Link([]*ir.Program{prog}, packs)
	if err != nil {
		return nil, err
	}
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
	if check.HasErrors(replayDiags) {
		return nil, fmt.Errorf("generated project failed replay validation")
	}

	return &liaCompareResult{
		Linked:       linked,
		Replay:       summary,
		ProjectPath:  projectPath,
		TapePath:     tapePath,
		DecisionPath: logPath,
	}, nil
}

func generateProjectModulesForCompare(ctx context.Context, generator *llmgen.Generator, spec *llmgen.ProjectSpec, model string, temperature float64) ([]*llmgen.GeneratedModule, error) {
	var assets []*llmgen.GeneratedModule
	for _, mod := range spec.Modules {
		contextText := combineGenerationContext(spec.Brief, mod.Context)
		asset, err := generator.GenerateLIAModuleAsset(ctx, llmgen.ModuleSpec{
			Name:        mod.Name,
			Role:        mod.Role,
			Model:       model,
			Temperature: temperature,
			Context:     contextText,
		})
		if err != nil {
			return nil, fmt.Errorf("generate module %s: %w", mod.Name, err)
		}
		assets = append(assets, asset)
	}
	return assets, nil
}

func compileJavaProject(dir string) javaCompileResult {
	absDir, err := filepath.Abs(dir)
	if err == nil {
		dir = absDir
	}
	files := collectJavaFiles(dir)
	if len(files) == 0 {
		return javaCompileResult{
			Success: false,
			Errors:  "no Java files were found",
			Files:   0,
		}
	}
	buildDir := filepath.Join(dir, ".build")
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		return javaCompileResult{Success: false, Errors: err.Error(), Files: len(files)}
	}
	args := append([]string{"-d", buildDir}, files...)
	cmd := exec.Command("javac", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return javaCompileResult{
			Success: false,
			Errors:  strings.TrimSpace(string(out)),
			Files:   len(files),
		}
	}
	return javaCompileResult{
		Success: true,
		Files:   len(files),
	}
}

func analyzeJavaProject(dir string) javaProjectMetrics {
	var metrics javaProjectMetrics
	for _, path := range collectJavaFiles(dir) {
		metrics.JavaFiles++
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		text := string(content)
		normPath := filepath.ToSlash(path)
		if strings.Contains(normPath, "/domain/") {
			metrics.DomainFiles++
		}
		if strings.Contains(normPath, "/port/") {
			metrics.PortFiles++
		}
		if strings.Contains(normPath, "/app/") || strings.Contains(normPath, "/application/") {
			metrics.ApplicationDirs++
		}
		if strings.Contains(normPath, "/bootstrap/") || strings.Contains(text, "Wiring") {
			metrics.BootstrapFiles++
		}
		if strings.Contains(text, " record ") || strings.Contains(text, "record ") {
			metrics.Records++
		}
		if strings.Contains(text, " interface ") || strings.Contains(text, "interface ") {
			metrics.Interfaces++
		}
		if strings.Contains(text, " enum ") || strings.Contains(text, "enum ") {
			metrics.Enums++
		}
		if strings.Contains(normPath, "UseCase.java") || strings.Contains(text, "UseCase") {
			metrics.UsecaseClasses++
		}
		if strings.Contains(normPath, "Adapter.java") || strings.Contains(text, "Adapter") {
			metrics.AdapterClasses++
		}
		if strings.Contains(normPath, "Wiring.java") || strings.Contains(text, "Wiring") {
			metrics.WiringClasses++
		}
	}
	return metrics
}

func collectJavaFiles(dir string) []string {
	var files []string
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".build" || strings.HasPrefix(name, "attempt-") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ".java") {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	return files
}

func renderComparisonReport(spec *llmgen.ProjectSpec, briefPath string, reference *comparisonVariant, direct, lia comparisonVariant, liaResult *liaCompareResult) string {
	var b strings.Builder
	b.WriteString("# Java Comparison Demo\n\n")
	b.WriteString("## Common Brief\n\n")
	b.WriteString("- Brief file: `")
	b.WriteString(filepath.ToSlash(briefPath))
	b.WriteString("`\n")
	b.WriteString("- Business brief: ")
	b.WriteString(spec.Brief)
	b.WriteString("\n\n")
	b.WriteString("## Variants\n\n")
	b.WriteString("| Variant | Compiles | Java files | Records | Interfaces | UseCases | Adapters | Wiring | Provenance |\n")
	b.WriteString("| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |\n")
	if reference != nil {
		b.WriteString(renderVariantRow(*reference))
	}
	b.WriteString(renderVariantRow(direct))
	b.WriteString(renderVariantRow(lia))
	b.WriteString("\n")
	b.WriteString("## Branch Artifacts\n\n")
	if reference != nil {
		b.WriteString("- Reference Java: `")
		b.WriteString(filepath.ToSlash(reference.Dir))
		b.WriteString("`\n")
		b.WriteString("- Reference Java compile report: `")
		b.WriteString(filepath.ToSlash(filepath.Join(filepath.Dir(briefPath), "reference-java.compile.txt")))
		b.WriteString("`\n")
	}
	b.WriteString("- Direct Java prompt: `")
	b.WriteString(filepath.ToSlash(direct.PromptPath))
	b.WriteString("`\n")
	b.WriteString("- Direct Java output: `")
	b.WriteString(filepath.ToSlash(direct.Dir))
	b.WriteString("`\n")
	b.WriteString("- Direct Java compile report: `")
	b.WriteString(filepath.ToSlash(filepath.Join(filepath.Dir(briefPath), "direct-java.compile.txt")))
	b.WriteString("`\n")
	b.WriteString("- Direct Java raw response: `")
	b.WriteString(filepath.ToSlash(direct.RawResponse))
	b.WriteString("`\n")
	b.WriteString("- LIA project: `")
	b.WriteString(filepath.ToSlash(liaResult.ProjectPath))
	b.WriteString("`\n")
	b.WriteString("- LIA prompt tape: `")
	b.WriteString(filepath.ToSlash(liaResult.TapePath))
	b.WriteString("`\n")
	b.WriteString("- LIA decision log: `")
	b.WriteString(filepath.ToSlash(liaResult.DecisionPath))
	b.WriteString("`\n")
	b.WriteString("- LIA -> Java output: `")
	b.WriteString(filepath.ToSlash(lia.Dir))
	b.WriteString("`\n")
	b.WriteString("- LIA -> Java compile report: `")
	b.WriteString(filepath.ToSlash(filepath.Join(filepath.Dir(briefPath), "lia-java.compile.txt")))
	b.WriteString("`\n\n")
	b.WriteString("## Notes\n\n")
	b.WriteString("- Direct Java repair attempts: ")
	b.WriteString(strconv.Itoa(direct.RepairAttempts))
	b.WriteString("\n")
	b.WriteString("- LIA replay summary: modules ")
	b.WriteString(strconv.Itoa(liaResult.Replay.Modules))
	b.WriteString(", generated ")
	b.WriteString(strconv.Itoa(liaResult.Replay.GeneratedModules))
	b.WriteString(", validated ")
	b.WriteString(strconv.Itoa(liaResult.Replay.ValidatedModules))
	b.WriteString("\n")
	b.WriteString("- The direct Java branch receives the same business brief but has no intermediate typed artifact, prompt tape, or decision log.\n")
	b.WriteString("- The LIA branch adds a structured intermediate (`project.lia`, `.lial`, tape, decision log) before lowering to Java.\n")
	return b.String()
}

func renderVariantRow(variant comparisonVariant) string {
	provenance := "none"
	if variant.HasLIAInputs {
		provenance = "project.lia + tape + decision log"
	}
	return fmt.Sprintf("| %s | %s | %d | %d | %d | %d | %d | %d | %s |\n",
		variant.Name,
		compileLabel(variant.Compile.Success),
		variant.Metrics.JavaFiles,
		variant.Metrics.Records,
		variant.Metrics.Interfaces,
		variant.Metrics.UsecaseClasses,
		variant.Metrics.AdapterClasses,
		variant.Metrics.WiringClasses,
		provenance,
	)
}

func compileLabel(ok bool) string {
	if ok {
		return "yes"
	}
	return "no"
}

func renderCompileResult(result javaCompileResult) string {
	var b strings.Builder
	b.WriteString("success: ")
	b.WriteString(compileLabel(result.Success))
	b.WriteByte('\n')
	b.WriteString("java_files: ")
	b.WriteString(strconv.Itoa(result.Files))
	b.WriteByte('\n')
	if strings.TrimSpace(result.Errors) != "" {
		b.WriteString("\n")
		b.WriteString(result.Errors)
		b.WriteByte('\n')
	}
	return b.String()
}

func renderCommonBrief(spec *llmgen.ProjectSpec) string {
	var b strings.Builder
	b.WriteString("# Common Brief\n\n")
	b.WriteString(spec.Brief)
	b.WriteString("\n\n## Structured Concerns\n\n")
	for _, mod := range spec.Modules {
		b.WriteString("- `")
		b.WriteString(mod.Name)
		b.WriteString("` (`")
		b.WriteString(mod.Role)
		b.WriteString("`): ")
		b.WriteString(mod.Context)
		b.WriteString("\n")
	}
	return b.String()
}

func buildDirectJavaPrompt(spec *llmgen.ProjectSpec, basePackage string) string {
	var b strings.Builder
	b.WriteString("Generate a plain Java 21 project directly, without using LIA.\n\n")
	b.WriteString("Business brief:\n")
	b.WriteString(strings.TrimSpace(spec.Brief))
	b.WriteString("\n\nStructured concerns to cover in the Java project:\n")
	for _, mod := range spec.Modules {
		b.WriteString("- ")
		b.WriteString(mod.Name)
		b.WriteString(" (")
		b.WriteString(mod.Role)
		b.WriteString("): ")
		b.WriteString(strings.TrimSpace(mod.Context))
		b.WriteByte('\n')
	}
	b.WriteString("\nRules:\n")
	b.WriteString("- Use plain Java 21.\n")
	b.WriteString("- No frameworks and no external dependencies.\n")
	b.WriteString("- Use Maven layout.\n")
	b.WriteString("- Keep the project compilable with javac.\n")
	b.WriteString("- Use base package ")
	b.WriteString(basePackage)
	b.WriteString(".\n")
	b.WriteString("- Separate domain, port, application, infrastructure, and bootstrap concerns when possible.\n")
	b.WriteString("- Output only file blocks in exactly this format:\n")
	b.WriteString("=== FILE: pom.xml ===\n")
	b.WriteString("<content>\n")
	b.WriteString("=== FILE: src/main/java/")
	b.WriteString(strings.ReplaceAll(basePackage, ".", "/"))
	b.WriteString("/Foo.java ===\n")
	b.WriteString("<content>\n")
	b.WriteString("- Do not wrap the answer in Markdown fences.\n")
	return b.String()
}

func buildDirectJavaRepairPrompt(initialPrompt, errors string) string {
	return initialPrompt + "\n\nThe previous answer was invalid or did not compile.\nCompiler/format feedback:\n" + strings.TrimSpace(errors) + "\n\nRewrite the full project from scratch and return all files again using the FILE markers."
}

func parseTextFiles(content string) ([]textFile, error) {
	lines := strings.Split(content, "\n")
	var files []textFile
	var current *textFile
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "=== FILE:") && strings.HasSuffix(trimmed, "===") {
			if current != nil {
				current.Content = strings.TrimLeft(strings.TrimRight(current.Content, "\n"), "\n")
				files = append(files, *current)
			}
			path := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "=== FILE:"), "==="))
			if err := validateGeneratedPath(path); err != nil {
				return nil, err
			}
			current = &textFile{Path: path}
			continue
		}
		if current == nil {
			continue
		}
		current.Content += line + "\n"
	}
	if current != nil {
		current.Content = strings.TrimLeft(strings.TrimRight(current.Content, "\n"), "\n")
		files = append(files, *current)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no FILE markers found")
	}
	return files, nil
}

func validateGeneratedPath(path string) error {
	path = filepath.ToSlash(strings.TrimSpace(path))
	switch {
	case path == "":
		return fmt.Errorf("empty file path")
	case strings.HasPrefix(path, "/"):
		return fmt.Errorf("absolute paths are not allowed: %s", path)
	case strings.Contains(path, ".."):
		return fmt.Errorf("relative parent paths are not allowed: %s", path)
	default:
		return nil
	}
}

func writeTextProject(dir string, files []textFile) error {
	for _, file := range files {
		target := filepath.Join(dir, filepath.FromSlash(file.Path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, []byte(file.Content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

func sanitizeCompareSegment(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "demo"
	}
	var out []rune
	for _, r := range raw {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			out = append(out, unicode.ToLower(r))
			continue
		}
		if len(out) == 0 || out[len(out)-1] != '_' {
			out = append(out, '_')
		}
	}
	clean := strings.Trim(string(out), "_")
	if clean == "" {
		return "demo"
	}
	if unicode.IsDigit(rune(clean[0])) {
		return "p" + clean
	}
	return clean
}
