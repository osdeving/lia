package repro

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"sort"

	"github.com/willams/lia/internal/ir"
)

// PromptTape is a versioned registry of prompts used in generation.
type PromptTape struct {
	Version string        `json:"version,omitempty"`
	Prompts []PromptEntry `json:"prompts"`
}

// PromptEntry maps a prompt ref to its body and metadata.
type PromptEntry struct {
	Ref         string   `json:"ref"`
	Body        string   `json:"body"`
	Hash        string   `json:"hash,omitempty"`
	ContextRefs []string `json:"context_refs,omitempty"`
	Params      []ir.KV  `json:"params,omitempty"`
}

// LoadTape reads a prompt tape from disk.
func LoadTape(path string) (*PromptTape, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var t PromptTape
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&t); err != nil {
		return nil, err
	}
	NormalizeTape(&t)
	return &t, nil
}

// SaveTape writes a prompt tape to disk in canonical ordering.
func SaveTape(path string, t *PromptTape) error {
	NormalizeTape(t)
	b, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// ComputePromptHash returns sha256 hash of the prompt body.
func ComputePromptHash(body string) string {
	sum := sha256.Sum256([]byte(body))
	return hex.EncodeToString(sum[:])
}

// AttachGenMeta attaches provenance to a module.
func AttachGenMeta(m *ir.Module, meta ir.GenMeta) {
	m.Gen = &meta
}

// NormalizeTape sorts prompts deterministically and fills missing hashes.
func NormalizeTape(t *PromptTape) {
	if t == nil {
		return
	}
	for i := range t.Prompts {
		if t.Prompts[i].Hash == "" {
			t.Prompts[i].Hash = ComputePromptHash(t.Prompts[i].Body)
		}
		sort.Strings(t.Prompts[i].ContextRefs)
		sort.Slice(t.Prompts[i].Params, func(a, b int) bool {
			return t.Prompts[i].Params[a].Key < t.Prompts[i].Params[b].Key
		})
	}
	sort.Slice(t.Prompts, func(i, j int) bool { return t.Prompts[i].Ref < t.Prompts[j].Ref })
}

// (no helpers)
