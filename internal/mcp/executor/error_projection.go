package executor

import (
	"strings"

	"github.com/yaleh/meta-cc/internal/analyzer"
)

// error_projection.go implements the DIR-097 stable projection for
// query_session_signals(type="errors").
//
// Before DIR-097 the errors signal returned raw user-role JSONL records
// verbatim. Their shape is irregular and undocumented:
//
//   - the top-level `toolUseResult` field is a plain string in some records
//     and an object (stdout/stderr, filePath/structuredPatch, ...) in others;
//   - the actual error text is nested inside `message.content[].content`,
//     which is itself a string in most records and an array of text blocks in
//     a few;
//   - the tool name is not on the record at all — it lives on the *assistant*
//     record that issued the call, correlated by `tool_use_id`.
//
// A consumer that built a jq extraction from the docs' role=tool "block
// fields" guidance therefore extracted nothing and had to hand-inspect raw
// JSONL. The projection below normalizes every error record to the same five
// fields regardless of that variance, so the docs can promise one shape.
// Collection and filtering are untouched — only the output projection is.
const (
	// fieldTimestamp / ... name the five projected fields. Every projected
	// record carries all of them, even when a value cannot be resolved (the
	// field is then ""), so a consumer's jq never has to guard for absence.
	fieldTimestamp = "timestamp"
	fieldSessionID = "session_id"
	fieldToolName  = "tool_name"
	fieldErrorText = "error_text"
	fieldCategory  = "category"
)

// toolUseJoinFilter selects the assistant records that carry `tool_use`
// blocks. It is built from the assistant records' `message.content[].id` ->
// `.name` pairs so a tool_result block's `tool_use_id` can be resolved to a
// tool name. The same filter shape is already used by handleQueryTools'
// status join, and it works for both record families the errors signal can
// return: raw Claude JSONL (tool_use blocks live on assistant records) and
// the normalized cross-provider shape (providerrecords.Normalize emits the
// same tool_use blocks, but NO `uuid`/`sourceToolAssistantUUID`, which is why
// `tool_use_id` — not the source-assistant uuid — is the join key here).
const toolUseJoinFilter = `select(.type == "assistant") | select(.message.content[] | .type == "tool_use")`

// projectedErrorFields lists the projected shape's keys in a stable order.
// Used by tests and documentation to assert the contract.
var projectedErrorFields = []string{fieldTimestamp, fieldSessionID, fieldToolName, fieldErrorText, fieldCategory}

// projectErrorEntries projects every raw error entry to the stable shape.
// Non-object entries (which the errors filter cannot produce, but which a
// provider is free to return) are passed through untouched rather than
// fabricating an empty projection around them.
func projectErrorEntries(entries []interface{}, toolNames map[string]string) []interface{} {
	out := make([]interface{}, 0, len(entries))
	for _, entry := range entries {
		rec, ok := entry.(map[string]interface{})
		if !ok {
			out = append(out, entry)
			continue
		}
		out = append(out, projectErrorRecord(rec, toolNames))
	}
	return out
}

// projectErrorRecord projects one raw tool-error record to
// {timestamp, session_id, tool_name, error_text, category}.
//
// error_text mirrors the extraction `analyze_errors` performs on a tool
// result (internal/types.ToolResult.UnmarshalJSON): a string `content` is
// used as-is, an array `content` joins its blocks' non-empty `.text` with a
// newline. When several `tool_result` blocks on the one record are errors,
// their texts are joined the same way. Only if no block yields any text — a
// record the pair-block extraction cannot describe — does the projection fall
// back to the record-level `toolUseResult` (string, or object's
// error/stderr/stdout/content), which is exactly the string/object variance
// DIR-097 was filed about.
//
// category is produced by analyzer.ClassifyErrorType — the same function
// analyze_errors uses to label its ByType groups — applied to the same
// (tool_name, error_text) pair, so the two tools speak one label vocabulary.
func projectErrorRecord(rec map[string]interface{}, toolNames map[string]string) map[string]interface{} {
	timestamp := stringValue(rec[fieldTimestamp])
	sessionID := firstNonEmptyString(stringValue(rec["sessionId"]), stringValue(rec[fieldSessionID]))

	var toolName, errorText string
	for _, block := range errorBlocks(rec) {
		if toolName == "" {
			toolName = toolNames[toolUseKey(sessionID, stringValue(block["tool_use_id"]))]
		}
		text := toolResultText(block)
		if text == "" {
			text = toolUseResultText(rec)
		}
		if text == "" {
			continue
		}
		if errorText == "" {
			errorText = text
			continue
		}
		errorText += "\n" + text
	}

	return map[string]interface{}{
		fieldTimestamp: timestamp,
		fieldSessionID: sessionID,
		fieldToolName:  toolName,
		fieldErrorText: errorText,
		fieldCategory:  analyzer.ClassifyErrorType(toolName, errorText),
	}
}

// errorBlocks returns the `tool_result` blocks on rec whose `is_error` is
// true. The errors filter guarantees at least one, but a projected record is
// still built defensively so a provider that widens the filter cannot make
// the projection panic or emit a missing field.
func errorBlocks(rec map[string]interface{}) []map[string]interface{} {
	message, _ := rec["message"].(map[string]interface{})
	content, _ := message["content"].([]interface{})
	var out []map[string]interface{}
	for _, block := range content {
		bm, ok := block.(map[string]interface{})
		if !ok || bm["type"] != "tool_result" {
			continue
		}
		if isErr, _ := bm["is_error"].(bool); !isErr {
			continue
		}
		out = append(out, bm)
	}
	return out
}

// toolResultText extracts a tool_result block's text, mirroring
// internal/types.ToolResult.UnmarshalJSON's string-or-array-of-text-blocks
// handling so error_text matches what analyze_errors sees on the same record.
func toolResultText(block map[string]interface{}) string {
	switch content := block["content"].(type) {
	case string:
		return content
	case []interface{}:
		texts := make([]string, 0, len(content))
		for _, entry := range content {
			bm, ok := entry.(map[string]interface{})
			if !ok {
				continue
			}
			if text := stringValue(bm["text"]); text != "" {
				texts = append(texts, text)
			}
		}
		return strings.Join(texts, "\n")
	default:
		return ""
	}
}

// toolUseResultText reads the record-level `toolUseResult` fallback, handling
// both variants DIR-097 was filed about: the string form is used as-is, and
// the object form contributes its first non-empty text-bearing field.
func toolUseResultText(rec map[string]interface{}) string {
	switch result := rec["toolUseResult"].(type) {
	case string:
		return result
	case map[string]interface{}:
		return firstNonEmptyString(
			stringValue(result["error"]),
			stringValue(result["stderr"]),
			stringValue(result["stdout"]),
			stringValue(result["content"]),
		)
	default:
		return ""
	}
}

// collectToolNames indexes assistant `tool_use` blocks by
// (session, tool_use_id) -> tool name. Session-scoping keeps the key unique
// across the multi-session record streams a project-scope query returns.
func collectToolNames(entries []interface{}) map[string]string {
	names := make(map[string]string)
	for _, entry := range entries {
		rec, ok := entry.(map[string]interface{})
		if !ok || rec["type"] != "assistant" {
			continue
		}
		sessionID := firstNonEmptyString(stringValue(rec["sessionId"]), stringValue(rec[fieldSessionID]))
		message, _ := rec["message"].(map[string]interface{})
		content, _ := message["content"].([]interface{})
		for _, block := range content {
			bm, ok := block.(map[string]interface{})
			if !ok || bm["type"] != "tool_use" {
				continue
			}
			id := stringValue(bm["id"])
			name := stringValue(bm["name"])
			if id == "" || name == "" {
				continue
			}
			names[toolUseKey(sessionID, id)] = name
		}
	}
	return names
}

func toolUseKey(sessionID, toolUseID string) string {
	return sessionID + "|" + toolUseID
}

// stringValue returns v when it is a non-empty string, else "".
func stringValue(v interface{}) string {
	s, _ := v.(string)
	return s
}

func firstNonEmptyString(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
