package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	lowerjava "github.com/willams/lia/internal/lower/java"
)

func addJavaProfileFlags(cmd flagBinder) {
	cmd.String("java-profile", "plain", "java lowering profile: plain|spring-boot|quarkus")
	cmd.String("java-profiles", "", "comma-separated Java lowering profiles; overrides --java-profile")
}

type flagBinder interface {
	String(name string, value string, usage string) *string
}

func resolveJavaProfiles(single, many string) ([]lowerjava.Profile, error) {
	raw := many
	if strings.TrimSpace(raw) == "" {
		raw = single
	}
	if strings.TrimSpace(raw) == "" {
		raw = "plain"
	}
	seen := map[lowerjava.Profile]bool{}
	var profiles []lowerjava.Profile
	for _, part := range strings.Split(raw, ",") {
		profile, err := lowerjava.ParseProfile(part)
		if err != nil {
			return nil, err
		}
		if seen[profile] {
			continue
		}
		seen[profile] = true
		profiles = append(profiles, profile)
	}
	if len(profiles) == 0 {
		return nil, fmt.Errorf("no java profiles resolved")
	}
	return profiles, nil
}

func javaProfileSlug(profile lowerjava.Profile) string {
	return strings.ReplaceAll(string(profile), "_", "-")
}

func javaProfileOutputDir(root string, profile lowerjava.Profile, multi bool) string {
	if !multi {
		return filepath.Join(root, "java")
	}
	return filepath.Join(root, "java-"+javaProfileSlug(profile))
}

func javaProfileCompilePath(root string, profile lowerjava.Profile, multi bool) string {
	if !multi {
		return filepath.Join(root, "java.compile.txt")
	}
	return filepath.Join(root, "java-"+javaProfileSlug(profile)+".compile.txt")
}
