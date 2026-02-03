package symbols

import (
	"testing"

	"github.com/willams/lia/internal/ir"
)

func TestSymbolQName(t *testing.T) {
	q := SymbolQName("orders.app", "port", "Repo")
	if q != "orders.app::port:Repo" {
		t.Fatalf("unexpected qname: %s", q)
	}
}

func TestQualifySymbol(t *testing.T) {
	if got := QualifySymbol("m", "port", "m::port:Repo"); got != "m::port:Repo" {
		t.Fatalf("unexpected qualified: %s", got)
	}
	if got := QualifySymbol("m", "port", "port:Repo"); got != "m::port:Repo" {
		t.Fatalf("unexpected qualified: %s", got)
	}
	if got := QualifySymbol("m", "port", "Repo"); got != "m::port:Repo" {
		t.Fatalf("unexpected qualified: %s", got)
	}
}

func TestDeriveModuleSymbols(t *testing.T) {
	m := &ir.Module{
		Name:  "orders.app",
		Role:  "usecase",
		Types: []ir.TypeDecl{{Name: "OrderId"}},
		Enums: []ir.EnumDecl{{Name: "Status"}},
		Ports: []ir.PortDecl{{Name: "OrderPort"}},
		Usecases: []ir.UsecaseDecl{{
			Name: "CreateOrder",
		}},
		Adapters: []ir.AdapterDecl{{
			Name:       "OrdersDb",
			Implements: "orders.port::port:OrderPort",
		}},
		Wirings: []ir.WiringDecl{{
			Name:  "Bindings",
			Binds: []string{"orders.port::port:OrderPort -> orders.app::adapter:OrdersDb"},
		}},
	}

	provides, requires, diags := DeriveModuleSymbols(m)
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if len(provides) < 5 {
		t.Fatalf("expected provides, got %d", len(provides))
	}
	foundReq := false
	for _, req := range requires {
		if req.QName == "orders.port::port:OrderPort" {
			foundReq = true
			break
		}
	}
	if !foundReq {
		t.Fatalf("expected requires to include orders.port::port:OrderPort")
	}
}

func TestDeriveModuleSymbols_Duplicate(t *testing.T) {
	m := &ir.Module{
		Name:  "orders.app",
		Types: []ir.TypeDecl{{Name: "OrderId"}, {Name: "OrderId"}},
	}
	_, _, diags := DeriveModuleSymbols(m)
	if len(diags) == 0 {
		t.Fatalf("expected duplicate diagnostic")
	}
}

func TestDeriveProgramSymbols(t *testing.T) {
	p := &ir.Program{
		Version: "0.1",
		Modules: []ir.Module{
			{
				Name:  "orders.app",
				Types: []ir.TypeDecl{{Name: "OrderId"}},
			},
		},
	}
	diags := DeriveProgramSymbols(p)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(p.Modules[0].Provides) != 1 {
		t.Fatalf("expected provides to be derived")
	}
}

func TestParseBind_Invalid(t *testing.T) {
	if _, _, ok := parseBind("invalid"); ok {
		t.Fatalf("expected parseBind to fail")
	}
}
