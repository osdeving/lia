package parser

import (
	"regexp"
	"strings"
)

// NormalizeLLMOutput preprocesses raw LLM output before parsing.
// It fixes common LLM mistakes that would cause syntax errors:
//   - Strips markdown code fences (```lia ... ```)
//   - Normalizes fn/function/def → method-compatible syntax
//   - Fixes class/record/struct misuse
//   - Inserts missing semicolons before } and keywords
//   - Normalizes whitespace and line endings
//   - Strips trailing commas before closing braces
func NormalizeLLMOutput(input string) string {
	s := input
	s = normalizeLineEndings(s)
	s = stripMarkdownFences(s)
	s = fixKeywordAliases(s)
	s = insertMissingSemicolons(s)
	s = stripTrailingCommasBeforeBraces(s)
	s = normalizeWhitespace(s)
	return s
}

// normalizeLineEndings converts \r\n → \n and strips lone \r.
func normalizeLineEndings(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

// stripMarkdownFences removes ```lia, ```lia-lang, ```, etc.
var fencePattern = regexp.MustCompile("(?m)^\\s*```(?:lia(?:-lang)?)?\\s*$")

func stripMarkdownFences(s string) string {
	return fencePattern.ReplaceAllString(s, "")
}

// fixKeywordAliases replaces common LLM keyword mistakes with LIA equivalents.
var keywordAliasPatterns = []struct {
	pattern     *regexp.Regexp
	replacement string
}{
	// fn → method (only inside port blocks — detected by line context)
	// function → fn or leave as-is since participle accepts both fn|method
	{regexp.MustCompile(`\bfunction\b`), "fn"},
	{regexp.MustCompile(`\bdef\b`), "fn"},
	// class → module (common Python/Java LLM output)
	{regexp.MustCompile(`(?m)^(\s*)class\s+(\w+)`), "${1}module ${2}"},
	// struct → type (common Go/Rust LLM output)
	{regexp.MustCompile(`(?m)^(\s*)struct\s+(\w+)\s*\{`), "${1}record ${2} {"},
	// interface → port (common Java/Go output)
	{regexp.MustCompile(`(?m)^(\s*)interface\s+(\w+)\s*\{`), "${1}port ${2} {"},
}

func fixKeywordAliases(s string) string {
	for _, alias := range keywordAliasPatterns {
		s = alias.pattern.ReplaceAllString(s, alias.replacement)
	}
	return s
}

// insertMissingSemicolons adds missing semicolons before } and before keywords
// on new lines where the previous non-empty line doesn't end with ; or { or }.
var missingSemicolonPattern = regexp.MustCompile(`(?m)([^;{}\s])\s*\n(\s*})`)

func insertMissingSemicolons(s string) string {
	return missingSemicolonPattern.ReplaceAllString(s, "${1};\n${2}")
}

// stripTrailingCommasBeforeBraces removes ,} and ,] patterns.
var trailingCommaPattern = regexp.MustCompile(`(?m),(\s*[}\]])`)

func stripTrailingCommasBeforeBraces(s string) string {
	return trailingCommaPattern.ReplaceAllString(s, "${1}")
}

// normalizeWhitespace ensures single blank lines, trims trailing whitespace.
func normalizeWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	var result []string
	prevEmpty := false
	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		isEmpty := trimmed == ""
		if isEmpty && prevEmpty {
			continue // skip consecutive blank lines
		}
		result = append(result, trimmed)
		prevEmpty = isEmpty
	}
	return strings.TrimSpace(strings.Join(result, "\n")) + "\n"
}
