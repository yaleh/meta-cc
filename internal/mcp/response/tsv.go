package response

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// SerializeResponseTSV serializes a hybrid-mode response (as built by
// AdaptResponse) into tab-separated values (DIR-044). It only handles the
// inline-mode envelope shape ({"mode":"inline","data":[...],...}) — for any
// other shape (notably file_ref mode, where the actual records already live
// in a separate JSONL temp file rather than the response body) it returns
// ok=false so the caller falls back to the standard JSON serialization.
//
// Output shape: a header row of the sorted union of all record field names,
// followed by one tab-separated row per record. Non-"data" envelope fields
// (mode, pagination, warnings) are not silently dropped — they're rendered
// as leading "#"-prefixed comment lines before the header, so a TSV consumer
// that only wants the data can skip lines starting with "#".
func SerializeResponseTSV(response interface{}) (string, bool) {
	m, ok := response.(map[string]interface{})
	if !ok {
		return "", false
	}
	mode, _ := m["mode"].(string)
	if mode != OutputModeInline {
		// file_ref mode: the records live in a JSONL temp file already;
		// converting the file_ref envelope itself to TSV isn't meaningful.
		// Fall back to the normal JSON response for this mode.
		return "", false
	}

	data, _ := m["data"].([]interface{})

	var sb strings.Builder
	sb.WriteString(tsvMetaComment(m))
	sb.WriteString(RecordsToTSV(data))
	return sb.String(), true
}

// tsvMetaComment renders envelope metadata fields (mode, pagination,
// warnings) as "#"-prefixed comment lines so they aren't silently lost when
// the body switches from a JSON object to TSV.
func tsvMetaComment(m map[string]interface{}) string {
	var lines []string
	for _, key := range []string{"mode", "pagination", "warnings"} {
		v, ok := m[key]
		if !ok {
			continue
		}
		b, err := json.Marshal(v)
		if err != nil {
			continue
		}
		lines = append(lines, fmt.Sprintf("# %s: %s", key, string(b)))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

// RecordsToTSV converts a slice of records (each typically
// map[string]interface{}-shaped) into TSV: a header row (the sorted union of
// keys across all records) followed by one row per record. A record that
// isn't a JSON object is rendered as a single JSON-encoded cell on its own
// line. Non-scalar field values (nested objects/arrays) are JSON-encoded
// inline within their cell rather than flattened. Tab/newline/backslash
// characters inside scalar string values are backslash-escaped so row
// structure (one record per line, fields split on tab) is preserved.
func RecordsToTSV(data []interface{}) string {
	if len(data) == 0 {
		return ""
	}

	keySet := make(map[string]bool)
	var keys []string
	for _, rec := range data {
		obj, ok := rec.(map[string]interface{})
		if !ok {
			continue
		}
		for k := range obj {
			if !keySet[k] {
				keySet[k] = true
				keys = append(keys, k)
			}
		}
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString(strings.Join(keys, "\t"))
	sb.WriteString("\n")

	for _, rec := range data {
		obj, ok := rec.(map[string]interface{})
		if !ok {
			b, err := json.Marshal(rec)
			if err != nil {
				b = []byte(fmt.Sprintf("%v", rec))
			}
			sb.WriteString(tsvEscape(string(b)))
			sb.WriteString("\n")
			continue
		}
		cells := make([]string, len(keys))
		for i, k := range keys {
			cells[i] = tsvCell(obj[k])
		}
		sb.WriteString(strings.Join(cells, "\t"))
		sb.WriteString("\n")
	}

	return sb.String()
}

// tsvCell renders a single field value as a TSV cell.
func tsvCell(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return tsvEscape(val)
	case float64, bool:
		return fmt.Sprintf("%v", val)
	default:
		// Nested object/array: JSON-encode inline rather than flatten.
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return tsvEscape(string(b))
	}
}

// tsvEscape backslash-escapes characters that would otherwise break TSV's
// tab-separated/one-record-per-line structure.
func tsvEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\t", "\\t")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	return s
}
