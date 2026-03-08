package linker

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"sort"
	"strings"

	"github.com/willams/lia/internal/codec"
	"github.com/willams/lia/internal/ir"
	"github.com/willams/lia/internal/policy"
	"github.com/willams/lia/internal/symbols"
)

// DecisionLog records linker decisions for replay/audit.
type DecisionLog struct {
	Entries []DecisionEntry `json:"entries"`
	Hash    string          `json:"hash,omitempty"`
}

// DecisionEntry represents a single selection decision.
type DecisionEntry struct {
	Requester string  `json:"requester,omitempty"`
	Symbol    string  `json:"symbol"`
	Chosen    string  `json:"chosen"`
	Score     float64 `json:"score,omitempty"`
	Reason    string  `json:"reason,omitempty"`
	TieBreak  string  `json:"tie_break,omitempty"`
}

// Link merges multiple programs into a linked unit (v0.1 stub).
// Returns diagnostics for policy/constraint checks.
func Link(inputs []*ir.Program, packs []ir.Pack) (*ir.Program, *DecisionLog, []ir.Diagnostic, error) {
	if len(inputs) == 0 {
		return nil, nil, nil, errors.New("no inputs to link")
	}

	out := &ir.Program{Version: "0.1"}
	log := &DecisionLog{}
	var diags []ir.Diagnostic

	for _, p := range inputs {
		if p == nil {
			continue
		}
		diags = append(diags, symbols.DeriveProgramSymbols(p)...)
		out.Projects = append(out.Projects, p.Projects...)
		out.Packs = append(out.Packs, p.Packs...)
		out.Modules = append(out.Modules, p.Modules...)
	}

	providers := map[string][]*ir.Module{}
	roleByModule := map[string]string{}
	for i := range out.Modules {
		m := &out.Modules[i]
		roleByModule[m.Name] = m.Role
		for _, sym := range m.Provides {
			providers[sym.QName] = append(providers[sym.QName], m)
		}
	}

	for i := range out.Modules {
		m := &out.Modules[i]
		for _, req := range m.Requires {
			cands := providers[req.QName]
			if len(cands) == 0 {
				diags = append(diags, ir.Diagnostic{
					Severity: "error",
					Message:  "missing symbol: " + req.QName,
					Path:     m.Name,
				})
				continue
			}
			chosen, score, reason, tie := selectCandidate(m, req, cands)
			log.Entries = append(log.Entries, DecisionEntry{
				Requester: m.Name,
				Symbol:    req.QName,
				Chosen:    chosen.Name,
				Score:     score,
				Reason:    reason,
				TieBreak:  tie,
			})
		}
	}

	sortDecisionEntries(log)
	log.Hash = hashDecisionLog(log)

	depRoles := buildDepRoleMap(log, roleByModule)
	eng, pdiags := policy.NewEngine(packs)
	diags = append(diags, pdiags...)
	diags = append(diags, eng.ApplyGlobal(out, depRoles)...)
	diags = append(diags, validateLinkedProgram(out)...)
	return out, log, diags, nil
}

// WriteDecisionLog writes the decision log in canonical JSON.
func WriteDecisionLog(path string, log *DecisionLog) error {
	if log == nil {
		return errors.New("nil decision log")
	}
	b, err := encodeDecisionLog(log)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func encodeDecisionLog(log *DecisionLog) ([]byte, error) {
	sortDecisionEntries(log)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(log); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

func hashDecisionLog(log *DecisionLog) string {
	b, err := encodeDecisionLog(log)
	if err != nil {
		return ""
	}
	return codec.HashBytes(b)
}

func buildDepRoleMap(log *DecisionLog, roleByModule map[string]string) map[string]map[string]bool {
	depRoles := map[string]map[string]bool{}
	if log == nil {
		return depRoles
	}
	for _, entry := range log.Entries {
		role := roleByModule[entry.Chosen]
		if role == "" {
			continue
		}
		if depRoles[entry.Requester] == nil {
			depRoles[entry.Requester] = map[string]bool{}
		}
		depRoles[entry.Requester][role] = true
	}
	return depRoles
}

func sortDecisionEntries(log *DecisionLog) {
	if log == nil {
		return
	}
	sort.Slice(log.Entries, func(i, j int) bool {
		if log.Entries[i].Requester == log.Entries[j].Requester {
			return log.Entries[i].Symbol < log.Entries[j].Symbol
		}
		return log.Entries[i].Requester < log.Entries[j].Requester
	})
}

type scoredCandidate struct {
	mod   *ir.Module
	score float64
	deps  int
}

func selectCandidate(requester *ir.Module, req ir.SymbolRef, cands []*ir.Module) (*ir.Module, float64, string, string) {
	scored := make([]scoredCandidate, 0, len(cands))
	for _, cand := range cands {
		score := candidateScore(cand, req.QName)
		score += preferenceScore(requester, cand.Name, req.QName)
		scored = append(scored, scoredCandidate{
			mod:   cand,
			score: score,
			deps:  len(cand.Requires),
		})
	}
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		if scored[i].deps != scored[j].deps {
			return scored[i].deps < scored[j].deps
		}
		return scored[i].mod.Name < scored[j].mod.Name
	})
	chosen := scored[0]
	tieBreak := "score desc, deps asc, lexical"
	reason := "selected candidate"
	return chosen.mod, chosen.score, reason, tieBreak
}

func candidateScore(mod *ir.Module, required string) float64 {
	score := 0.0
	for _, cand := range mod.Candidates {
		if matchesSymbol(cand.Symbol, required, mod.Name) {
			if cand.Score > score {
				score = cand.Score
			}
		}
	}
	return score
}

func preferenceScore(requester *ir.Module, candidateName, required string) float64 {
	if requester == nil {
		return 0
	}
	score := 0.0
	for _, pref := range requester.Preferences {
		expr := strings.TrimSpace(pref.Expr)
		if expr == "" {
			continue
		}
		if expr == candidateName || expr == required || strings.HasSuffix(required, ":"+expr) {
			if pref.Weight != 0 {
				score += pref.Weight
			} else {
				score += 1
			}
		}
	}
	return score
}

func matchesSymbol(candidate, required, moduleName string) bool {
	candidate = strings.TrimSpace(candidate)
	required = strings.TrimSpace(required)
	if candidate == "" || required == "" {
		return false
	}
	if candidate == required {
		return true
	}
	if strings.Contains(candidate, "::") {
		return false
	}
	if strings.Contains(candidate, ":") {
		return moduleName+"::"+candidate == required
	}
	return strings.HasSuffix(required, ":"+candidate)
}

func validateLinkedProgram(p *ir.Program) []ir.Diagnostic {
	if p == nil {
		return nil
	}

	strict := false
	for _, proj := range p.Projects {
		if proj.Repro == ir.ReproStrict {
			strict = true
			break
		}
	}
	if !strict {
		return nil
	}

	var diags []ir.Diagnostic
	for _, mod := range p.Modules {
		for _, hole := range mod.Holes {
			diags = append(diags, ir.Diagnostic{
				Severity: "error",
				Message:  "unresolved hole is not allowed in repro=strict: " + hole.Name,
				Path:     mod.Name,
			})
		}
	}
	return diags
}
