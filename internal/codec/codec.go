package codec

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/willams/lia/internal/ir"
)

// CanonicalizeProgram returns canonical JSON for hashing/storage.
// Note: This sorts slices in-place for determinism.
func CanonicalizeProgram(p *ir.Program) ([]byte, error) {
	if p == nil {
		return nil, errors.New("nil program")
	}
	SortAll(p)

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(p); err != nil {
		return nil, err
	}
	b := bytes.TrimRight(buf.Bytes(), "\n")
	return b, nil
}

// HashBytes returns a sha256 hex digest.
func HashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// HashProgram hashes the canonical JSON representation of a program.
func HashProgram(p *ir.Program) (string, error) {
	b, err := CanonicalizeProgram(p)
	if err != nil {
		return "", err
	}
	return HashBytes(b), nil
}

// DecodeProgram decodes canonical JSON into a Program.
func DecodeProgram(r io.Reader) (*ir.Program, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var p ir.Program
	if err := dec.Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

// LoadProgramFile loads a program from a JSON file.
func LoadProgramFile(path string) (*ir.Program, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return DecodeProgram(f)
}

// WriteProgramFile writes a canonical JSON file.
func WriteProgramFile(path string, p *ir.Program) error {
	b, err := CanonicalizeProgram(p)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// NormalizeQName normalizes qualified names in a minimal way.
func NormalizeQName(q string) string {
	return strings.TrimSpace(q)
}

// SortAll sorts all collections in place for deterministic output.
func SortAll(p *ir.Program) {
	if p == nil {
		return
	}

	sort.Slice(p.Projects, func(i, j int) bool { return p.Projects[i].Name < p.Projects[j].Name })
	for i := range p.Projects {
		proj := &p.Projects[i]
		sort.Slice(proj.Uses, func(a, b int) bool {
			if proj.Uses[a].Name == proj.Uses[b].Name {
				return proj.Uses[a].Version < proj.Uses[b].Version
			}
			return proj.Uses[a].Name < proj.Uses[b].Name
		})
		sort.Slice(proj.Policies, func(a, b int) bool { return proj.Policies[a].Name < proj.Policies[b].Name })
		sort.Slice(proj.Constraints, func(a, b int) bool { return proj.Constraints[a].Name < proj.Constraints[b].Name })
	}

	sort.Slice(p.Packs, func(i, j int) bool { return p.Packs[i].Name < p.Packs[j].Name })
	for i := range p.Packs {
		pack := &p.Packs[i]
		sort.Slice(pack.Modules, func(a, b int) bool { return pack.Modules[a].Name < pack.Modules[b].Name })
		sort.Slice(pack.Policies, func(a, b int) bool { return pack.Policies[a].Name < pack.Policies[b].Name })
		sort.Slice(pack.Constraints, func(a, b int) bool { return pack.Constraints[a].Name < pack.Constraints[b].Name })
	}

	sort.Slice(p.Modules, func(i, j int) bool { return p.Modules[i].Name < p.Modules[j].Name })
	for i := range p.Modules {
		m := &p.Modules[i]
		m.Name = NormalizeQName(m.Name)
		sort.Slice(m.Provides, func(a, b int) bool { return m.Provides[a].QName < m.Provides[b].QName })
		sort.Slice(m.Requires, func(a, b int) bool { return m.Requires[a].QName < m.Requires[b].QName })
		sort.Slice(m.Types, func(a, b int) bool { return m.Types[a].Name < m.Types[b].Name })
		sort.Slice(m.Enums, func(a, b int) bool { return m.Enums[a].Name < m.Enums[b].Name })
		sort.Slice(m.Ports, func(a, b int) bool { return m.Ports[a].Name < m.Ports[b].Name })
		sort.Slice(m.Usecases, func(a, b int) bool { return m.Usecases[a].Name < m.Usecases[b].Name })
		sort.Slice(m.Adapters, func(a, b int) bool { return m.Adapters[a].Name < m.Adapters[b].Name })
		sort.Slice(m.Wirings, func(a, b int) bool { return m.Wirings[a].Name < m.Wirings[b].Name })
		sort.Slice(m.Constraints, func(a, b int) bool { return m.Constraints[a].Name < m.Constraints[b].Name })
		sort.Slice(m.Preferences, func(a, b int) bool { return m.Preferences[a].Name < m.Preferences[b].Name })
		sort.Slice(m.Holes, func(a, b int) bool { return m.Holes[a].Name < m.Holes[b].Name })
		sort.Slice(m.Candidates, func(a, b int) bool { return m.Candidates[a].Symbol < m.Candidates[b].Symbol })

		if m.Gen != nil {
			sort.Slice(m.Gen.ModelParams, func(a, b int) bool { return m.Gen.ModelParams[a].Key < m.Gen.ModelParams[b].Key })
			sort.Strings(m.Gen.ContextRefs)
			sort.Strings(m.Gen.ToolsTraceRefs)
		}

		for j := range m.Ports {
			port := &m.Ports[j]
			sort.Slice(port.Methods, func(a, b int) bool { return port.Methods[a].Name < port.Methods[b].Name })
			for k := range port.Methods {
				meth := &port.Methods[k]
				sort.Slice(meth.Params, func(a, b int) bool { return meth.Params[a].Name < meth.Params[b].Name })
				sort.Slice(meth.Returns, func(a, b int) bool { return meth.Returns[a].Name < meth.Returns[b].Name })
			}
		}

		for j := range m.Usecases {
			uc := &m.Usecases[j]
			sort.Slice(uc.Inputs, func(a, b int) bool { return uc.Inputs[a].Name < uc.Inputs[b].Name })
			sort.Slice(uc.Outputs, func(a, b int) bool { return uc.Outputs[a].Name < uc.Outputs[b].Name })
			sort.Strings(uc.Effects)
		}

		for j := range m.Adapters {
			ad := &m.Adapters[j]
			sort.Slice(ad.Inputs, func(a, b int) bool { return ad.Inputs[a].Name < ad.Inputs[b].Name })
			sort.Slice(ad.Outputs, func(a, b int) bool { return ad.Outputs[a].Name < ad.Outputs[b].Name })
			sort.Strings(ad.Effects)
		}

		for j := range m.Wirings {
			w := &m.Wirings[j]
			sort.Strings(w.Binds)
		}

		if m.Doc != nil && m.Doc.RTF != nil {
			sort.Strings(m.Doc.RTF.Inputs)
			sort.Strings(m.Doc.RTF.Must)
			sort.Strings(m.Doc.RTF.Avoid)
			sort.Strings(m.Doc.RTF.Tests)
			sort.Strings(m.Doc.RTF.Repair)
		}
	}
}
