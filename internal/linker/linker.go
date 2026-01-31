package linker

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"sort"

	"github.com/willams/lia/internal/codec"
	"github.com/willams/lia/internal/ir"
	"github.com/willams/lia/internal/policy"
)

// DecisionLog records linker decisions for replay/audit.
type DecisionLog struct {
	Entries []DecisionEntry `json:"entries"`
	Hash    string          `json:"hash,omitempty"`
}

// DecisionEntry represents a single selection decision.
type DecisionEntry struct {
	Symbol   string  `json:"symbol"`
	Chosen   string  `json:"chosen"`
	Score    float64 `json:"score,omitempty"`
	Reason   string  `json:"reason,omitempty"`
	TieBreak string  `json:"tie_break,omitempty"`
}

// Link merges multiple programs into a linked unit (v0.1 stub).
// Returns diagnostics for policy/constraint checks.
func Link(inputs []*ir.Program, packs []ir.Pack) (*ir.Program, *DecisionLog, []ir.Diagnostic, error) {
	if len(inputs) == 0 {
		return nil, nil, nil, errors.New("no inputs to link")
	}

	out := &ir.Program{Version: "0.1"}
	log := &DecisionLog{}

	for _, p := range inputs {
		if p == nil {
			continue
		}
		out.Projects = append(out.Projects, p.Projects...)
		out.Packs = append(out.Packs, p.Packs...)
		for _, m := range p.Modules {
			out.Modules = append(out.Modules, m)
			log.Entries = append(log.Entries, DecisionEntry{
				Symbol: m.Name,
				Chosen: m.Name,
				Reason: "v0.1 stub linker: choose input module",
			})
		}
	}

	sort.Slice(log.Entries, func(i, j int) bool { return log.Entries[i].Symbol < log.Entries[j].Symbol })
	log.Hash = hashDecisionLog(log)

	depRoles := buildDepRoleMap(out)
	eng, pdiags := policy.NewEngine(packs)
	diags := append(pdiags, eng.ApplyGlobal(out, depRoles)...)
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
	sort.Slice(log.Entries, func(i, j int) bool { return log.Entries[i].Symbol < log.Entries[j].Symbol })
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

func buildDepRoleMap(p *ir.Program) map[string]map[string]bool {
	roleBySymbol := map[string]string{}
	for _, m := range p.Modules {
		for _, sym := range m.Provides {
			roleBySymbol[sym.QName] = m.Role
		}
	}

	depRoles := map[string]map[string]bool{}
	for _, m := range p.Modules {
		for _, req := range m.Requires {
			role := roleBySymbol[req.QName]
			if role == "" {
				continue
			}
			if depRoles[m.Name] == nil {
				depRoles[m.Name] = map[string]bool{}
			}
			depRoles[m.Name][role] = true
		}
	}
	return depRoles
}
