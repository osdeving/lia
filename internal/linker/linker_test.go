package linker

import (
	"testing"

	"github.com/willams/lia/internal/ir"
)

func TestLink_EmptyInputs(t *testing.T) {
	_, _, _, err := Link(nil, nil)
	if err == nil {
		t.Fatal("expected error for empty inputs, got nil")
	}
}

func TestLink_ResolvesRequires(t *testing.T) {
	provider := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name:  "orders.port",
				Role:  "port",
				Ports: []ir.PortDecl{{Name: "OrderRepository"}},
			},
		},
	}
	consumer := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name: "orders.app",
				Role: "usecase",
				Adapters: []ir.AdapterDecl{{
					Name:       "OrdersDb",
					Implements: "orders.port::port:OrderRepository",
				}},
			},
		},
	}

	linked, log, diags, err := Link([]*ir.Program{provider, consumer}, nil)
	if err != nil {
		t.Fatalf("Link failed: %v", err)
	}
	if linked == nil {
		t.Fatal("expected linked program")
	}
	if len(log.Entries) != 1 {
		t.Fatalf("expected 1 decision entry, got %d", len(log.Entries))
	}
	entry := log.Entries[0]
	if entry.Requester != "orders.app" {
		t.Errorf("expected requester orders.app, got %s", entry.Requester)
	}
	if entry.Symbol != "orders.port::port:OrderRepository" {
		t.Errorf("unexpected symbol %s", entry.Symbol)
	}
	if entry.Chosen != "orders.port" {
		t.Errorf("unexpected chosen %s", entry.Chosen)
	}
	if entry.Score != 0 {
		t.Errorf("expected score 0, got %f", entry.Score)
	}
	if len(diags) != 0 {
		for _, d := range diags {
			if d.Severity == "error" {
				t.Fatalf("unexpected error diagnostic: %s", d.Message)
			}
		}
	}
}

func TestLink_MissingSymbol(t *testing.T) {
	consumer := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name: "orders.app",
				Role: "usecase",
				Adapters: []ir.AdapterDecl{{
					Name:       "OrdersDb",
					Implements: "orders.port::port:OrderRepository",
				}},
			},
		},
	}

	_, log, diags, err := Link([]*ir.Program{consumer}, nil)
	if err != nil {
		t.Fatalf("Link failed: %v", err)
	}
	if log == nil {
		t.Fatalf("expected decision log")
	}
	found := false
	for _, d := range diags {
		if d.Severity == "error" && d.Message == "missing symbol: orders.port::port:OrderRepository" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected missing symbol diagnostic")
	}
}

func TestLink_CandidateScore(t *testing.T) {
	providerLow := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name:  "payments.port",
				Role:  "port",
				Ports: []ir.PortDecl{{Name: "Billing"}},
				Candidates: []ir.CandidateDecl{{
					Symbol: "payments.port::port:Billing",
					Score:  0.2,
				}},
			},
		},
	}
	providerHigh := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name:  "payments.port",
				Role:  "port",
				Ports: []ir.PortDecl{{Name: "Billing"}},
				Candidates: []ir.CandidateDecl{{
					Symbol: "payments.port::port:Billing",
					Score:  2.5,
				}},
			},
		},
	}
	consumer := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name: "payments.app",
				Role: "usecase",
				Adapters: []ir.AdapterDecl{{
					Name:       "BillingAdapter",
					Implements: "payments.port::port:Billing",
				}},
			},
		},
	}

	_, log, _, err := Link([]*ir.Program{providerLow, providerHigh, consumer}, nil)
	if err != nil {
		t.Fatalf("Link failed: %v", err)
	}
	if len(log.Entries) != 1 {
		t.Fatalf("expected 1 decision entry, got %d", len(log.Entries))
	}
	if log.Entries[0].Score != 2.5 {
		t.Errorf("expected score 2.5, got %f", log.Entries[0].Score)
	}
}

func TestWriteDecisionLog(t *testing.T) {
	log := &DecisionLog{
		Entries: []DecisionEntry{
			{Requester: "mod", Symbol: "sym", Chosen: "impl", Reason: "test"},
		},
	}

	tmpfile := t.TempDir() + "/decision.json"
	if err := WriteDecisionLog(tmpfile, log); err != nil {
		t.Fatalf("WriteDecisionLog failed: %v", err)
	}
}

func TestWriteDecisionLog_Nil(t *testing.T) {
	err := WriteDecisionLog("/tmp/test.json", nil)
	if err == nil {
		t.Fatal("expected error for nil log, got nil")
	}
}

func TestHashDecisionLog_Determinism(t *testing.T) {
	log := &DecisionLog{
		Entries: []DecisionEntry{
			{Requester: "a", Symbol: "a", Chosen: "impl_a"},
			{Requester: "b", Symbol: "b", Chosen: "impl_b"},
		},
	}

	hash1 := hashDecisionLog(log)
	hash2 := hashDecisionLog(log)

	if hash1 != hash2 {
		t.Errorf("hash not deterministic: %s != %s", hash1, hash2)
	}

	if hash1 == "" {
		t.Error("expected non-empty hash")
	}
}

func TestPreferenceScore(t *testing.T) {
	req := &ir.Module{
		Name: "req",
		Preferences: []ir.PreferDecl{
			{Expr: "candidate.mod", Weight: 2.5},
			{Expr: "Sym", Weight: 0},
		},
	}
	score := preferenceScore(req, "candidate.mod", "candidate.mod::port:Sym")
	if score < 3.4 || score > 3.6 {
		t.Fatalf("unexpected preference score: %f", score)
	}
}

func TestMatchesSymbol(t *testing.T) {
	if !matchesSymbol("m::port:Repo", "m::port:Repo", "m") {
		t.Fatalf("expected exact match")
	}
	if matchesSymbol("other::port:Repo", "m::port:Repo", "m") {
		t.Fatalf("expected non-match for different module")
	}
	if !matchesSymbol("port:Repo", "m::port:Repo", "m") {
		t.Fatalf("expected kind-prefixed match")
	}
	if !matchesSymbol("Repo", "m::port:Repo", "m") {
		t.Fatalf("expected suffix match")
	}
	if matchesSymbol("", "m::port:Repo", "m") {
		t.Fatalf("expected empty candidate to fail")
	}
}

func TestSortDecisionEntries(t *testing.T) {
	log := &DecisionLog{
		Entries: []DecisionEntry{
			{Requester: "b", Symbol: "b", Chosen: "x"},
			{Requester: "a", Symbol: "b", Chosen: "x"},
			{Requester: "a", Symbol: "a", Chosen: "x"},
		},
	}
	sortDecisionEntries(log)
	if log.Entries[0].Requester != "a" || log.Entries[0].Symbol != "a" {
		t.Fatalf("unexpected sort order")
	}
}

func TestBuildDepRoleMap_Nil(t *testing.T) {
	deps := buildDepRoleMap(nil, map[string]string{})
	if len(deps) != 0 {
		t.Fatalf("expected empty dep map")
	}
}
