package packs

import "testing"

func TestPackFileCandidates_UppercaseUnderscore(t *testing.T) {
	candidates := packFileCandidates("ArchBaseline", "")
	if !contains(candidates, "ARCH_BASELINE.lia") {
		t.Fatalf("expected ARCH_BASELINE.lia in candidates: %v", candidates)
	}
	if !contains(candidates, "arch-baseline.lia") {
		t.Fatalf("expected arch-baseline.lia in candidates: %v", candidates)
	}

	withVersion := packFileCandidates("ArchBaseline", "0.1.0")
	if !contains(withVersion, "ARCH_BASELINE_0.1.0.lia") {
		t.Fatalf("expected versioned underscore candidate, got %v", withVersion)
	}
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
