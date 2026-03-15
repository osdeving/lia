package parser

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/willams/lia/internal/ir"
	"github.com/willams/lia/internal/symbols"
)

// ParseDiagnostic represents a parse error with position information.
type ParseDiagnostic struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
}

func (d ParseDiagnostic) String() string {
	return fmt.Sprintf("line %d:%d: %s", d.Line, d.Column, d.Message)
}

// RecoveryResult holds the partial parse result and any diagnostics.
type RecoveryResult struct {
	Program     *ir.Program       `json:"program"`
	Diagnostics []ParseDiagnostic `json:"diagnostics,omitempty"`
	TotalBlocks int               `json:"total_blocks"`
	Parsed      int               `json:"parsed"`
}

// HasErrors returns true if any diagnostics were recorded.
func (r *RecoveryResult) HasErrors() bool {
	return len(r.Diagnostics) > 0
}

// topLevelPattern matches the start of a top-level construct (not indented).
var topLevelPattern = regexp.MustCompile(`(?m)^(?:@gen\s*\{[^}]*\}\s*)?(?:project|module|pack)\b`)

// ParseWithRecovery parses a .lia source string in error-tolerant mode.
// Instead of aborting on the first error, it splits the input into top-level
// blocks and parses each independently. Successfully parsed blocks are
// collected into a partial ir.Program; failed blocks produce diagnostics.
func ParseWithRecovery(input string) *RecoveryResult {
	blocks := splitTopLevelBlocks(input)
	result := &RecoveryResult{
		Program:     &ir.Program{Version: "0.1"},
		TotalBlocks: len(blocks),
	}

	if len(blocks) == 0 {
		// Try parsing the entire input as-is
		prog, err := ParseString(input)
		if err != nil {
			result.Diagnostics = append(result.Diagnostics, ParseDiagnostic{
				Line:    1,
				Column:  1,
				Message: err.Error(),
			})
			return result
		}
		result.Program = prog
		result.Parsed = 1
		result.TotalBlocks = 1
		return result
	}

	for _, block := range blocks {
		prog, err := ParseString(block.Content)
		if err != nil {
			result.Diagnostics = append(result.Diagnostics, ParseDiagnostic{
				Line:    block.StartLine,
				Column:  1,
				Message: err.Error(),
			})
			continue
		}
		result.Parsed++
		mergeProgram(result.Program, prog)
	}

	// Derive symbols on the merged result
	if len(result.Program.Modules) > 0 || len(result.Program.Projects) > 0 {
		_ = symbols.DeriveProgramSymbols(result.Program)
	}

	return result
}

// ParseFileWithRecovery is the file-based variant of ParseWithRecovery.
func ParseFileWithRecovery(path string) *RecoveryResult {
	b, err := os.ReadFile(path)
	if err != nil {
		return &RecoveryResult{
			Program: &ir.Program{Version: "0.1"},
			Diagnostics: []ParseDiagnostic{{
				Line:    1,
				Column:  1,
				Message: err.Error(),
			}},
		}
	}
	return ParseWithRecovery(string(b))
}

// sourceBlock represents a contiguous top-level block of source code.
type sourceBlock struct {
	Content   string
	StartLine int
}

// splitTopLevelBlocks splits the input into independent top-level blocks,
// each starting with project/module/pack (optionally preceded by @gen).
func splitTopLevelBlocks(input string) []sourceBlock {
	locs := topLevelPattern.FindAllStringIndex(input, -1)
	if len(locs) == 0 {
		return nil
	}

	var blocks []sourceBlock
	for i, loc := range locs {
		start := loc[0]
		var end int
		if i+1 < len(locs) {
			end = locs[i+1][0]
		} else {
			end = len(input)
		}
		content := input[start:end]
		startLine := countLines(input[:start]) + 1
		blocks = append(blocks, sourceBlock{
			Content:   strings.TrimRight(content, " \t\n\r"),
			StartLine: startLine,
		})
	}
	return blocks
}

func countLines(s string) int {
	return strings.Count(s, "\n")
}

// mergeProgram merges src into dst.
func mergeProgram(dst, src *ir.Program) {
	dst.Projects = append(dst.Projects, src.Projects...)
	dst.Packs = append(dst.Packs, src.Packs...)
	dst.Modules = append(dst.Modules, src.Modules...)
}
