package query

import (
	"strings"

	"github.com/yaleh/meta-cc/internal/analyzer"
)

// DIR-097: query_session_signals(type="errors") used to hand back the raw
// JSONL user-role record for every failing tool_result. That shape is
// irregular in ways a consumer cannot predict from the docs: `toolUseResult`
// is a bare string in some records and an object in others, `sessionId` is
// camelCase, the tool name lives on a *different* (assistant) record, and the
// error text is nested inside message.content[].content. A consumer's first
// jq extraction attempts therefore returned nothing, and the raw JSONL had to
// be inspected by hand — a projection gap, not an extraction bug.
//
// This file projects every error record onto five stable fields. Raw access is
// still available through the type=errors "raw" flag, which hands back the
// untouched record (see handleQueryToolErrors).

// ErrorSignalFields names, in order, the five fields every projected
// type=errors record exposes. Tests assert against this list so the documented
// contract and the implemented one cannot drift apart.
var ErrorSignalFields = []string{"timestamp", "session_id", "tool_name", "error_text", "category"}

// Join keys for the paired record stream ErrorSignalJoinJQ produces. They are
// internal to this projection: callers see only the five projected fields.
const (
	joinToolUseKey = "__tool_use"
	joinRecordKey  = "__error_record"
	joinBlockKey   = "__error_block"
)

// ErrorSignalJoinJQ is the jq program handleQueryToolErrors runs for the
// default (projected) type=errors path. It emits two kinds of record in one
// pass:
//
//   - one {"__tool_use": {id, name}} per assistant tool_use block, which is the
//     only place the tool name exists, reduced to the two fields the
//     projection needs;
//   - one {"__error_record", "__error_block"} pair per failing tool_result
//     block. Emitting the pair (rather than the bare record) keeps a record
//     with several failing blocks distinguishable, and keeps the successful
//     tool_result blocks of the same record from being projected at all.
//
// The caller fetches this unbounded and correlates it in Go: the jq pipeline
// runs one record at a time with no cross-record join, and the tool name sits
// on a different record from the error.
const ErrorSignalJoinJQ = `(select(.type == "assistant" and (.message.content | type == "array")) | ` +
	`.message.content[] | select(.type == "tool_use") | {"` + joinToolUseKey + `": {"id": .id, "name": .name}}), ` +
	`(select(.type == "user" and (.message.content | type == "array")) | . as $r | ` +
	`.message.content[] | select(.type == "tool_result" and .is_error == true) | ` +
	`{"` + joinRecordKey + `": $r, "` + joinBlockKey + `": .})`

// ProjectErrorSignals projects the paired records produced by
// ErrorSignalJoinJQ into the stable error-signal shape. Paired records that
// carry no error of their own (the __tool_use index entries) are consumed by
// the join and produce no output record.
func ProjectErrorSignals(entries []interface{}) []interface{} {
	toolNames := ToolNamesByToolUseID(entries)
	out := make([]interface{}, 0, len(entries))
	for _, entry := range entries {
		rec, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		record, ok := rec[joinRecordKey].(map[string]interface{})
		if !ok {
			continue
		}
		block, _ := rec[joinBlockKey].(map[string]interface{})
		out = append(out, ProjectErrorSignal(record, block, toolNames))
	}
	return out
}

// ToolNamesByToolUseID indexes every tool_use id in the paired stream to its
// tool name, so an error record can name the tool that failed even though the
// tool_use lives on an earlier assistant record.
func ToolNamesByToolUseID(entries []interface{}) map[string]string {
	names := map[string]string{}
	for _, entry := range entries {
		rec, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		use, ok := rec[joinToolUseKey].(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := use["id"].(string)
		name, _ := use["name"].(string)
		if id != "" && name != "" {
			names[id] = name
		}
	}
	return names
}

// ProjectErrorSignal projects one error record plus its failing tool_result
// block onto the five documented fields. A field whose source is absent is
// emitted as an empty string rather than omitted, so every record in the
// output has the same shape.
func ProjectErrorSignal(record, block map[string]interface{}, toolNames map[string]string) map[string]interface{} {
	toolUseID, _ := block["tool_use_id"].(string)
	toolName := toolNames[toolUseID]
	if toolName == "" {
		// The join is the authoritative source, but a hand-written or
		// normalized record may carry the name on the block itself; falling
		// back keeps a name that is present from being silently dropped.
		toolName, _ = block["name"].(string)
	}
	text := ErrorText(record, block)
	return map[string]interface{}{
		"timestamp":  stringValue(record["timestamp"]),
		"session_id": SessionIDOf(record),
		"tool_name":  toolName,
		"error_text": text,
		"category":   analyzer.ClassifyErrorType(toolName, text),
	}
}

// SessionIDOf returns the session identifier of a record, tolerating both the
// raw Claude JSONL spelling ("sessionId") and the normalized provider one
// ("session_id").
func SessionIDOf(record map[string]interface{}) string {
	if id := stringValue(record["session_id"]); id != "" {
		return id
	}
	return stringValue(record["sessionId"])
}

// ErrorText extracts the human-readable error message, handling every variant
// observed in this project's corpus.
//
// Precedence mirrors the parser that feeds analyze_errors: types.ToolResult
// takes an explicit block-level "error" field first and falls back to
// "content", so classifying the same text here is what makes the two tools'
// categories agree. Only when the block carries neither does the record-level
// "toolUseResult" get consulted.
func ErrorText(record, block map[string]interface{}) string {
	if text := stringValue(block["error"]); text != "" {
		return text
	}
	if text := contentText(block["content"]); text != "" {
		return text
	}
	return toolUseResultText(record["toolUseResult"])
}

// contentText renders a tool_result block's "content" field, which Claude Code
// writes either as a bare string or as an array of text blocks.
func contentText(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			switch block := item.(type) {
			case string:
				if block != "" {
					parts = append(parts, block)
				}
			case map[string]interface{}:
				if text := stringValue(block["text"]); text != "" {
					parts = append(parts, text)
				}
			}
		}
		return strings.Join(parts, "\n")
	}
	return ""
}

// toolUseResultText renders the record-level "toolUseResult" field, which this
// project's corpus carries as a bare string for failing calls but which Claude
// Code writes as an object for other tools. Only the object fields that hold
// the tool's own output are consulted — an unrecognized object deliberately
// yields "" rather than a serialization of the whole value, so the projection
// stays a bounded, documented shape instead of echoing arbitrarily large
// payloads (Edit results carry whole file bodies).
func toolUseResultText(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case map[string]interface{}:
		// stderr first: it is where the Bash harness puts a failed command's
		// diagnostics, with stdout as the fallback for commands that report
		// their failure on standard output.
		for _, key := range []string{"stderr", "error", "message", "stdout", "output"} {
			if text := stringValue(v[key]); strings.TrimSpace(text) != "" {
				return text
			}
		}
	}
	return ""
}

func stringValue(v interface{}) string {
	s, _ := v.(string)
	return s
}
