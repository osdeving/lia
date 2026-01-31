package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/willams/lia/internal/check"
	"github.com/willams/lia/internal/codec"
	"github.com/willams/lia/internal/ir"
	"github.com/willams/lia/internal/linker"
	"github.com/willams/lia/internal/lower/java"
	"github.com/willams/lia/internal/lower/python"
	"github.com/willams/lia/internal/packs"
	"github.com/willams/lia/internal/parser"
)

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string     `json:"jsonrpc"`
	ID      any        `json:"id,omitempty"`
	Result  any        `json:"result,omitempty"`
	Error   *respError `json:"error,omitempty"`
}

type respError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"input_schema,omitempty"`
}

type toolCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type parseArgs struct {
	Input  string `json:"input"`
	Output string `json:"output,omitempty"`
}

type checkArgs struct {
	Input    string   `json:"input"`
	PackDirs []string `json:"pack_dirs,omitempty"`
}

type linkArgs struct {
	Inputs      []string `json:"inputs"`
	Output      string   `json:"output,omitempty"`
	DecisionLog string   `json:"decision_log,omitempty"`
	PackDirs    []string `json:"pack_dirs,omitempty"`
}

type lowerArgs struct {
	Input  string `json:"input"`
	Target string `json:"target,omitempty"`
	Output string `json:"output,omitempty"`
}

type explainArgs struct {
	Input       string `json:"input"`
	DecisionLog string `json:"decision_log,omitempty"`
}

type buildArgs struct {
	Input    string   `json:"input"`
	OutDir   string   `json:"out_dir,omitempty"`
	Target   string   `json:"target,omitempty"`
	PackDirs []string `json:"pack_dirs,omitempty"`
}

func main() {
	dec := json.NewDecoder(os.Stdin)
	enc := json.NewEncoder(os.Stdout)
	for {
		var req request
		if err := dec.Decode(&req); err != nil {
			if err == io.EOF {
				return
			}
			_ = enc.Encode(response{JSONRPC: "2.0", Error: &respError{Code: -32700, Message: err.Error()}})
			return
		}

		result, errObj := handle(req)
		resp := response{JSONRPC: "2.0", ID: req.ID, Result: result, Error: errObj}
		if req.ID == nil {
			continue
		}
		_ = enc.Encode(resp)
	}
}

func handle(req request) (any, *respError) {
	switch req.Method {
	case "initialize":
		return map[string]any{
			"name":    "lia-mcp",
			"version": "0.1-playground",
			"capabilities": map[string]any{
				"tools": true,
			},
		}, nil
	case "tools/list":
		return map[string]any{"tools": listTools()}, nil
	case "tools/call":
		var call toolCall
		if err := json.Unmarshal(req.Params, &call); err != nil {
			return nil, &respError{Code: -32602, Message: "invalid params", Data: err.Error()}
		}
		return callTool(call)
	default:
		return nil, &respError{Code: -32601, Message: "method not found"}
	}
}

func listTools() []tool {
	return []tool{
		{
			Name:        "lia.parse",
			Description: "Parse .lia source into .liao",
			InputSchema: objectSchema([]string{"input"}, map[string]any{
				"input":  stringSchema(),
				"output": stringSchema(),
			}),
		},
		{
			Name:        "lia.check",
			Description: "Validate a .liao or .lial file",
			InputSchema: objectSchema([]string{"input"}, map[string]any{
				"input":     stringSchema(),
				"pack_dirs": arraySchema(stringSchema()),
			}),
		},
		{
			Name:        "lia.link",
			Description: "Link .liao inputs into .lial + decision log",
			InputSchema: objectSchema([]string{"inputs"}, map[string]any{
				"inputs":       arraySchema(stringSchema()),
				"output":       stringSchema(),
				"decision_log": stringSchema(),
				"pack_dirs":    arraySchema(stringSchema()),
			}),
		},
		{
			Name:        "lia.lower",
			Description: "Lower .lial to target language (python/java)",
			InputSchema: objectSchema([]string{"input"}, map[string]any{
				"input":  stringSchema(),
				"target": stringSchema(),
				"output": stringSchema(),
			}),
		},
		{
			Name:        "lia.explain",
			Description: "Explain hashes and counts for a .liao/.lial",
			InputSchema: objectSchema([]string{"input"}, map[string]any{
				"input":        stringSchema(),
				"decision_log": stringSchema(),
			}),
		},
		{
			Name:        "lia.build",
			Description: "Parse + check + link + optional lower (playground)",
			InputSchema: objectSchema([]string{"input"}, map[string]any{
				"input":     stringSchema(),
				"out_dir":   stringSchema(),
				"target":    stringSchema(),
				"pack_dirs": arraySchema(stringSchema()),
			}),
		},
	}
}

func callTool(call toolCall) (any, *respError) {
	switch call.Name {
	case "lia.parse":
		var args parseArgs
		if err := json.Unmarshal(call.Arguments, &args); err != nil {
			return nil, invalidParams(err)
		}
		if args.Output == "" {
			args.Output = replaceExt(args.Input, ".liao")
		}
		prog, err := parser.ParseFile(args.Input)
		if err != nil {
			return nil, toolError(err)
		}
		if err := codec.WriteProgramFile(args.Output, prog); err != nil {
			return nil, toolError(err)
		}
		return map[string]any{"output": args.Output}, nil
	case "lia.check":
		var args checkArgs
		if err := json.Unmarshal(call.Arguments, &args); err != nil {
			return nil, invalidParams(err)
		}
		prog, err := codec.LoadProgramFile(args.Input)
		if err != nil {
			return nil, toolError(err)
		}
		packsLoaded, packDiags := loadPacksForProgram(prog, args.PackDirs)
		diags := append(packDiags, check.CheckProgram(prog, packsLoaded)...)
		return map[string]any{"ok": !check.HasErrors(diags), "diagnostics": diags}, nil
	case "lia.link":
		var args linkArgs
		if err := json.Unmarshal(call.Arguments, &args); err != nil {
			return nil, invalidParams(err)
		}
		if len(args.Inputs) == 0 {
			return nil, invalidParams(fmt.Errorf("inputs required"))
		}
		if args.Output == "" {
			args.Output = "out.lial"
		}
		if args.DecisionLog == "" {
			args.DecisionLog = args.Output + ".decision-log.json"
		}
		var inputs []*ir.Program
		var allPacks []ir.Pack
		var diags []ir.Diagnostic
		for _, path := range args.Inputs {
			p, err := codec.LoadProgramFile(path)
			if err != nil {
				return nil, toolError(err)
			}
			inputs = append(inputs, p)
			packsLoaded, packDiags := loadPacksForProgram(p, args.PackDirs)
			diags = append(diags, packDiags...)
			allPacks = append(allPacks, packsLoaded...)
		}
		linked, log, linkDiags, err := linker.Link(inputs, allPacks)
		if err != nil {
			return nil, toolError(err)
		}
		diags = append(diags, linkDiags...)
		if err := codec.WriteProgramFile(args.Output, linked); err != nil {
			return nil, toolError(err)
		}
		if err := linker.WriteDecisionLog(args.DecisionLog, log); err != nil {
			return nil, toolError(err)
		}
		return map[string]any{
			"output":       args.Output,
			"decision_log": args.DecisionLog,
			"diagnostics":  diags,
			"ok":           !check.HasErrors(diags),
		}, nil
	case "lia.lower":
		var args lowerArgs
		if err := json.Unmarshal(call.Arguments, &args); err != nil {
			return nil, invalidParams(err)
		}
		if args.Target == "" {
			args.Target = "python"
		}
		if args.Output == "" {
			if args.Target == "java" {
				args.Output = replaceExt(args.Input, ".java")
			} else {
				args.Output = replaceExt(args.Input, ".py")
			}
		}
		prog, err := codec.LoadProgramFile(args.Input)
		if err != nil {
			return nil, toolError(err)
		}
		var data []byte
		switch strings.ToLower(args.Target) {
		case "java":
			data, err = java.Lower(prog)
		case "python", "py":
			data, err = python.Lower(prog)
		default:
			return nil, invalidParams(fmt.Errorf("unknown target: %s", args.Target))
		}
		if err != nil {
			return nil, toolError(err)
		}
		if err := os.WriteFile(args.Output, data, 0o644); err != nil {
			return nil, toolError(err)
		}
		return map[string]any{"output": args.Output}, nil
	case "lia.explain":
		var args explainArgs
		if err := json.Unmarshal(call.Arguments, &args); err != nil {
			return nil, invalidParams(err)
		}
		prog, err := codec.LoadProgramFile(args.Input)
		if err != nil {
			return nil, toolError(err)
		}
		hash, err := codec.HashProgram(prog)
		if err != nil {
			return nil, toolError(err)
		}
		res := map[string]any{
			"hash":     hash,
			"projects": len(prog.Projects),
			"packs":    len(prog.Packs),
			"modules":  len(prog.Modules),
		}
		if args.DecisionLog != "" {
			log, err := loadDecisionLog(args.DecisionLog)
			if err != nil {
				return nil, toolError(err)
			}
			res["decision_log_entries"] = len(log.Entries)
			res["decision_log_hash"] = log.Hash
		}
		return res, nil
	case "lia.build":
		var args buildArgs
		if err := json.Unmarshal(call.Arguments, &args); err != nil {
			return nil, invalidParams(err)
		}
		if args.OutDir == "" {
			args.OutDir = "out"
		}
		if err := os.MkdirAll(args.OutDir, 0o755); err != nil {
			return nil, toolError(err)
		}
		base := strings.TrimSuffix(filepath.Base(args.Input), filepath.Ext(args.Input))
		outLiao := filepath.Join(args.OutDir, base+".liao")
		outLial := filepath.Join(args.OutDir, base+".lial")
		outDecision := outLial + ".decision-log.json"

		prog, err := parser.ParseFile(args.Input)
		if err != nil {
			return nil, toolError(err)
		}
		if err := codec.WriteProgramFile(outLiao, prog); err != nil {
			return nil, toolError(err)
		}
		packsLoaded, packDiags := loadPacksForProgram(prog, args.PackDirs)
		diags := append(packDiags, check.CheckProgram(prog, packsLoaded)...)

		linked, log, linkDiags, err := linker.Link([]*ir.Program{prog}, packsLoaded)
		if err != nil {
			return nil, toolError(err)
		}
		diags = append(diags, linkDiags...)
		if err := codec.WriteProgramFile(outLial, linked); err != nil {
			return nil, toolError(err)
		}
		if err := linker.WriteDecisionLog(outDecision, log); err != nil {
			return nil, toolError(err)
		}

		output := map[string]any{
			"liao":         outLiao,
			"lial":         outLial,
			"decision_log": outDecision,
			"diagnostics":  diags,
			"ok":           !check.HasErrors(diags),
		}

		if args.Target != "" {
			lowerOut := filepath.Join(args.OutDir, base+"."+targetExt(args.Target))
			data, err := lowerToTarget(linked, args.Target)
			if err != nil {
				return nil, toolError(err)
			}
			if err := os.WriteFile(lowerOut, data, 0o644); err != nil {
				return nil, toolError(err)
			}
			output["lowered"] = lowerOut
		}

		return output, nil
	default:
		return nil, &respError{Code: -32601, Message: "tool not found"}
	}
}

func loadDecisionLog(path string) (*linker.DecisionLog, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	var log linker.DecisionLog
	if err := dec.Decode(&log); err != nil {
		return nil, err
	}
	return &log, nil
}

func replaceExt(path, ext string) string {
	base := strings.TrimSuffix(path, filepath.Ext(path))
	return base + ext
}

func targetExt(target string) string {
	switch strings.ToLower(target) {
	case "java":
		return "java"
	default:
		return "py"
	}
}

func lowerToTarget(p *ir.Program, target string) ([]byte, error) {
	switch strings.ToLower(target) {
	case "java":
		return java.Lower(p)
	case "python", "py":
		return python.Lower(p)
	default:
		return nil, fmt.Errorf("unknown target: %s", target)
	}
}

func loadPacksForProgram(p *ir.Program, packDirs []string) ([]ir.Pack, []ir.Diagnostic) {
	refs := gatherPackRefs(p)
	if len(refs) == 0 {
		return nil, nil
	}
	loader := packs.Loader{SearchDirs: resolvePackDirs(packDirs)}
	return loader.LoadAll(refs)
}

func gatherPackRefs(p *ir.Program) []ir.PackRef {
	var refs []ir.PackRef
	seen := map[string]bool{}
	for _, proj := range p.Projects {
		for _, ref := range proj.Uses {
			key := ref.Name + "@" + ref.Version
			if seen[key] {
				continue
			}
			seen[key] = true
			refs = append(refs, ref)
		}
	}
	return refs
}

func resolvePackDirs(custom []string) []string {
	if len(custom) == 0 {
		return packs.DefaultSearchDirs()
	}
	defaults := packs.DefaultSearchDirs()
	seen := map[string]bool{}
	var dirs []string
	for _, d := range append(custom, defaults...) {
		if d == "" || seen[d] {
			continue
		}
		if info, err := os.Stat(d); err == nil && info.IsDir() {
			seen[d] = true
			dirs = append(dirs, d)
		}
	}
	return dirs
}

func stringSchema() map[string]any {
	return map[string]any{"type": "string"}
}

func arraySchema(item map[string]any) map[string]any {
	return map[string]any{"type": "array", "items": item}
}

func objectSchema(required []string, props map[string]any) map[string]any {
	obj := map[string]any{"type": "object", "properties": props}
	if len(required) > 0 {
		obj["required"] = required
	}
	return obj
}

func invalidParams(err error) *respError {
	return &respError{Code: -32602, Message: err.Error()}
}

func toolError(err error) *respError {
	return &respError{Code: -32000, Message: err.Error()}
}
