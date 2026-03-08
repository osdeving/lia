package llmgen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// WorkspaceManifest is an optional root manifest for prompt-driven generation.
type WorkspaceManifest struct {
	Name        string                 `json:"name,omitempty"`
	Target      string                 `json:"target,omitempty"`
	Repro       string                 `json:"repro,omitempty"`
	Model       string                 `json:"model,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	Workspace   WorkspaceMetadata      `json:"workspace,omitempty"`
	Packs       WorkspacePackSelection `json:"packs,omitempty"`
}

// WorkspaceMetadata describes the current repository/runtime shape.
type WorkspaceMetadata struct {
	Build       string `json:"build,omitempty"`
	Framework   string `json:"framework,omitempty"`
	BasePackage string `json:"base_package,omitempty"`
}

// WorkspacePackSelection lets a workspace pin or extend pack discovery.
type WorkspacePackSelection struct {
	Builtin   []string `json:"builtin,omitempty"`
	ExtraDirs []string `json:"extra_dirs,omitempty"`
}

// WorkspaceContext is the resolved context used by the planner.
type WorkspaceContext struct {
	RootDir      string   `json:"root_dir"`
	ManifestPath string   `json:"manifest_path,omitempty"`
	Name         string   `json:"name,omitempty"`
	Target       string   `json:"target,omitempty"`
	Repro        string   `json:"repro,omitempty"`
	Model        string   `json:"model,omitempty"`
	Temperature  float64  `json:"temperature,omitempty"`
	Build        string   `json:"build,omitempty"`
	Framework    string   `json:"framework,omitempty"`
	BasePackage  string   `json:"base_package,omitempty"`
	Signals      []string `json:"signals,omitempty"`
	PackDirs     []string `json:"pack_dirs,omitempty"`
	BuiltinPacks []string `json:"builtin_packs,omitempty"`
}

// LoadWorkspaceContext loads an optional lia.json and combines it with local repo signals.
func LoadWorkspaceContext(rootDir, manifestPath string, extraPackDirs []string) (*WorkspaceContext, error) {
	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		rootDir = "."
	}
	absRoot, err := filepath.Abs(rootDir)
	if err == nil {
		rootDir = absRoot
	}

	ctx := &WorkspaceContext{
		RootDir: rootDir,
	}
	manifest, found, err := loadWorkspaceManifest(manifestPath)
	if err != nil {
		return nil, err
	}
	if found {
		ctx.ManifestPath = manifestPath
		ctx.Name = strings.TrimSpace(manifest.Name)
		ctx.Target = strings.TrimSpace(manifest.Target)
		ctx.Repro = strings.TrimSpace(manifest.Repro)
		ctx.Model = strings.TrimSpace(manifest.Model)
		ctx.Temperature = manifest.Temperature
		ctx.Build = strings.TrimSpace(manifest.Workspace.Build)
		ctx.Framework = strings.TrimSpace(manifest.Workspace.Framework)
		ctx.BasePackage = strings.TrimSpace(manifest.Workspace.BasePackage)
		ctx.BuiltinPacks = append(ctx.BuiltinPacks, manifest.Packs.Builtin...)
		ctx.PackDirs = append(ctx.PackDirs, manifest.Packs.ExtraDirs...)
	}

	signals := detectWorkspaceSignals(rootDir)
	ctx.Signals = append(ctx.Signals, signals...)
	if ctx.Name == "" {
		ctx.Name = sanitizeWorkspaceName(filepath.Base(rootDir))
	}
	if ctx.Target == "" {
		ctx.Target = inferTargetFromSignals(signals)
	}
	if ctx.Build == "" {
		ctx.Build = inferBuildFromSignals(signals)
	}
	ctx.PackDirs = dedupeExistingDirs(rootDir, append(ctx.PackDirs, extraPackDirs...))
	ctx.BuiltinPacks = dedupeStrings(ctx.BuiltinPacks)
	ctx.Signals = dedupeStrings(ctx.Signals)
	return ctx, nil
}

func loadWorkspaceManifest(path string) (*WorkspaceManifest, bool, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, false, nil
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, false, err
	}
	var manifest WorkspaceManifest
	if err := json.Unmarshal(b, &manifest); err != nil {
		return nil, false, err
	}
	return &manifest, true, nil
}

func detectWorkspaceSignals(rootDir string) []string {
	var signals []string
	add := func(signal string) {
		signal = strings.TrimSpace(signal)
		if signal != "" {
			signals = append(signals, signal)
		}
	}

	if fileExists(filepath.Join(rootDir, "pom.xml")) {
		add("pom.xml")
		add("maven")
		add("java")
	}
	if fileExists(filepath.Join(rootDir, "build.gradle")) || fileExists(filepath.Join(rootDir, "build.gradle.kts")) {
		add("gradle")
		add("java")
	}
	if dirExists(filepath.Join(rootDir, "src", "main", "java")) {
		add("src/main/java")
		add("java")
	}
	if fileExists(filepath.Join(rootDir, "package.json")) {
		add("package.json")
	}
	return dedupeStrings(signals)
}

func inferTargetFromSignals(signals []string) string {
	for _, signal := range signals {
		if signal == "java" || signal == "pom.xml" || signal == "gradle" || signal == "src/main/java" {
			return "java"
		}
	}
	return ""
}

func inferBuildFromSignals(signals []string) string {
	for _, signal := range signals {
		switch signal {
		case "maven", "pom.xml":
			return "maven"
		case "gradle":
			return "gradle"
		}
	}
	return ""
}

func sanitizeWorkspaceName(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "." {
		return "lia-app"
	}
	raw = strings.TrimSuffix(raw, filepath.Ext(raw))
	raw = strings.ReplaceAll(raw, "_", "-")
	return raw
}

func dedupeExistingDirs(rootDir string, dirs []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, dir := range dirs {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(rootDir, dir)
		}
		if seen[dir] {
			continue
		}
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			seen[dir] = true
			out = append(out, dir)
		}
	}
	sort.Strings(out)
	return out
}

func dedupeStrings(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
