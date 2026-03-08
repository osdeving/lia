package policy

import "testing"

func TestParseList_TrimsWhitespaceAndQuotes(t *testing.T) {
	got := parseList(`["domain", "usecase", "port"]`)
	if len(got) != 3 {
		t.Fatalf("expected 3 roles, got %d", len(got))
	}
	if got[0] != "domain" || got[1] != "usecase" || got[2] != "port" {
		t.Fatalf("unexpected parsed roles: %#v", got)
	}
}
