package response

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DIR-080 / ADR-009: bounded, redacted structured samples attached to
// file_ref metadata. A sample illustrates the declared shape — it is NOT a
// complete or authoritative subset of the result.

const (
	// maxSampleRecords bounds how many records a sample contains.
	maxSampleRecords = 2
	// maxSampleArrayElements bounds array elements retained per array.
	maxSampleArrayElements = 2
	// maxSampleDepth bounds object nesting in the sample.
	maxSampleDepth = 8
	// maxSampleStringLen bounds individual string values.
	maxSampleStringLen = 240
	// maxSampleTotalBytes bounds the final JSON serialization of the sample;
	// when exceeded, string values are re-truncated more aggressively.
	maxSampleTotalBytes = 2048
	// minSampleStringCap is the smallest string budget the shrink loop uses.
	minSampleStringCap = 16

	// redactedMarker replaces values of sensitive-looking fields.
	redactedMarker = "[REDACTED]"
)

// sensitiveKeyPatterns mark field names whose values are replaced by
// redactedMarker. Matched case-insensitively as substrings of the field
// name, so nested keys (client_secret, X-Api-Key, sessionToken) are covered.
var sensitiveKeyPatterns = []string{
	"api_key", "apikey", "token", "secret", "password", "passwd",
	"credential", "authorization", "private_key", "cookie",
}

// BuildBoundedSample selects up to maxSampleRecords representative records
// and returns sanitized copies: sensitive fields redacted, oversized strings
// truncated, arrays and depth capped, and the total serialization kept under
// maxSampleTotalBytes. Deterministic: record 0 plus the record with the most
// distinct paths (ties resolved by earliest index).
func BuildBoundedSample(records []interface{}) []interface{} {
	if len(records) == 0 {
		return []interface{}{}
	}

	selected := make([]interface{}, 0, maxSampleRecords)
	seen := map[int]bool{}
	for _, idx := range selectSampleIndexes(records) {
		if seen[idx] {
			continue
		}
		seen[idx] = true
		selected = append(selected, records[idx])
	}

	stringCap := maxSampleStringLen
	for {
		sample := make([]interface{}, 0, len(selected))
		for _, record := range selected {
			sanitized, _ := redactAndTruncate(record, 0, stringCap)
			sample = append(sample, sanitized)
		}
		serialized, err := json.Marshal(sample)
		if err != nil {
			return []interface{}{}
		}
		if len(serialized) <= maxSampleTotalBytes || stringCap <= minSampleStringCap {
			return sample
		}
		stringCap /= 2
		if stringCap < minSampleStringCap {
			stringCap = minSampleStringCap
		}
	}
}

// selectSampleIndexes returns the deterministic sample selection: always
// record 0, plus the record covering the most distinct paths (so optional
// fields are illustrated), tie-broken by earliest index.
func selectSampleIndexes(records []interface{}) []int {
	indexes := []int{0}
	bestIdx, bestCount := -1, -1
	for i, record := range records {
		count := countDistinctPaths(record, 0, 128)
		if count > bestCount {
			bestIdx, bestCount = i, count
		}
	}
	if bestIdx > 0 {
		indexes = append(indexes, bestIdx)
	}
	return indexes
}

// countDistinctPaths approximates shape coverage of one record for sample
// selection, bounded by cap.
func countDistinctPaths(v interface{}, depth, cap int) int {
	if depth > maxSampleDepth || cap <= 0 {
		return 0
	}
	switch val := v.(type) {
	case map[string]interface{}:
		total := 0
		for _, child := range val {
			total += 1 + countDistinctPaths(child, depth+1, cap-total)
			if total >= cap {
				return cap
			}
		}
		return total
	case []interface{}:
		total := 0
		for _, elem := range val {
			total += 1 + countDistinctPaths(elem, depth+1, cap-total)
			if total >= cap {
				return cap
			}
		}
		return total
	default:
		return 0
	}
}

// redactAndTruncate returns a sanitized deep copy of v and the number of
// sensitive fields redacted. Inputs are never mutated.
func redactAndTruncate(v interface{}, depth, stringCap int) (interface{}, int) {
	if depth > maxSampleDepth {
		return "[depth limit]", 0
	}
	switch val := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(val))
		redacted := 0
		for k, child := range val {
			if isSensitiveKey(k) {
				if child != redactedMarker {
					redacted++
				}
				out[k] = redactedMarker
				continue
			}
			sanitized, n := redactAndTruncate(child, depth+1, stringCap)
			out[k] = sanitized
			redacted += n
		}
		return out, redacted
	case []interface{}:
		n := len(val)
		kept := n
		if kept > maxSampleArrayElements {
			kept = maxSampleArrayElements
		}
		out := make([]interface{}, 0, kept+1)
		redacted := 0
		for _, elem := range val[:kept] {
			sanitized, c := redactAndTruncate(elem, depth+1, stringCap)
			out = append(out, sanitized)
			redacted += c
		}
		if n > kept {
			out = append(out, fmt.Sprintf("[+%d more]", n-kept))
		}
		return out, redacted
	case string:
		if len(val) > stringCap {
			return val[:stringCap] + fmt.Sprintf("...[truncated: %d chars]", len(val)), 0
		}
		return val, 0
	default:
		return v, 0
	}
}

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	for _, pattern := range sensitiveKeyPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}
