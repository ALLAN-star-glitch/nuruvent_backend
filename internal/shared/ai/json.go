// internal/shared/ai/json.go

package ai

import "strings"

// CleanJSONResponse strips markdown code fences and extracts the first
// top-level JSON object from an LLM response.
//
// LLMs frequently wrap their JSON in ```json ... ``` fences even when
// explicitly told not to. This function normalizes that output so the
// caller can json.Unmarshal directly.
//
// Example inputs and outputs:
//
//	"```json\n{\"a\":1}\n```"      → "{\"a\":1}"
//	"```\n{\"a\":1}\n```"          → "{\"a\":1}"
//	"Here you go:\n{\"a\":1}\n"    → "{\"a\":1}"
//	"{\"a\":1}"                     → "{\"a\":1}"
func CleanJSONResponse(raw string) string {
	s := strings.TrimSpace(raw)

	// Strip leading fence
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```JSON")
	s = strings.TrimPrefix(s, "```")

	// Strip trailing fence
	s = strings.TrimSuffix(s, "```")

	s = strings.TrimSpace(s)

	// Extract the outermost { ... } block
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start != -1 && end != -1 && end > start {
		s = s[start : end+1]
	}

	return s
}