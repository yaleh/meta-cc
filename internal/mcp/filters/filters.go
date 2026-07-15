package filters

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/yaleh/meta-cc/internal/parser"
)

const (
	// DefaultMaxMessageLength is the default for max_message_length parameter.
	// Set to 0 to disable truncation (use hybrid output mode for large results).
	DefaultMaxMessageLength = 0

	DefaultPreviewLength = 100
)

// TruncateMessageContent truncates the 'content' field in user messages
// to prevent context overflow from large session summaries.
//
// Use hybrid output mode instead of truncation for complete data preservation.
//
// Parameters:
//   - messages: Array of message objects (maps)
//   - maxLen: Maximum length for content field (0 or negative = no truncation)
//
// Returns:
//   - New array with truncated messages (originals unchanged)
//
// Truncated messages include:
//   - content_truncated: true
//   - original_length: original content length
func TruncateMessageContent(messages []interface{}, maxLen int) []interface{} {
	if maxLen <= 0 {
		return messages
	}

	truncated := make([]interface{}, len(messages))

	for i, msg := range messages {
		// Convert to map for manipulation
		msgMap, ok := msg.(map[string]interface{})
		if !ok {
			// Not a map, return as-is
			truncated[i] = msg
			continue
		}

		// Create shallow copy to avoid mutating original
		newMap := make(map[string]interface{})
		for k, v := range msgMap {
			newMap[k] = v
		}

		// Try nested message.content first
		if msgObj, ok := newMap["message"].(map[string]interface{}); ok {
			if content, ok := msgObj["content"].(string); ok && len(content) > maxLen {
				// Deep copy the message map to avoid mutating original
				newMsg := make(map[string]interface{})
				for k, v := range msgObj {
					newMsg[k] = v
				}
				newMsg["content"] = content[:maxLen] + "... [TRUNCATED]"
				newMap["message"] = newMsg
				newMap["content_truncated"] = true
				newMap["original_length"] = len(content)
			}
		} else if content, ok := newMap["content"].(string); ok {
			// Fallback: flat content
			if len(content) > maxLen {
				newMap["content"] = content[:maxLen] + "... [TRUNCATED]"
				newMap["content_truncated"] = true
				newMap["original_length"] = len(content)
			}
		}

		truncated[i] = newMap
	}

	return truncated
}

// ApplyContentSummary returns only message metadata (no full content).
// Useful for pattern matching without needing full message text.
//
// Use hybrid output mode instead for complete data preservation.
//
// Parameters:
//   - messages: Array of message objects (maps)
//   - previewLength: Max runes for content_preview (0 or negative = DefaultPreviewLength)
//
// Returns:
//   - New array with summary objects containing:
//   - turn_sequence: message turn number
//   - timestamp: message timestamp
//   - content_preview: first previewLength characters of content
//
// All other fields are removed to reduce output size.
func ApplyContentSummary(messages []interface{}, previewLength int) []interface{} {
	if previewLength <= 0 {
		previewLength = DefaultPreviewLength
	}

	summary := make([]interface{}, len(messages))

	for i, msg := range messages {
		msgMap, ok := msg.(map[string]interface{})
		if !ok {
			// Not a map, return as-is
			summary[i] = msg
			continue
		}

		// Extract preview from nested or flat content using rune-safe truncation
		preview := ""
		content := extractContentString(msgMap)
		if content != "" {
			runes := []rune(content)
			if len(runes) > previewLength {
				preview = string(runes[:previewLength]) + "..."
			} else {
				preview = content
			}
		}

		// Create summary object
		summary[i] = map[string]interface{}{
			"session_id":      msgMap["sessionId"],
			"turn_sequence":   i,
			"uuid":            msgMap["uuid"],
			"timestamp":       msgMap["timestamp"],
			"content_preview": preview,
		}
	}

	return summary
}

// extractContentString extracts content string from a message map.
// Handles both nested (message.content) and flat (content) structures.
func extractContentString(msgMap map[string]interface{}) string {
	// Try nested: message.content
	if msg, ok := msgMap["message"].(map[string]interface{}); ok {
		if content, ok := msg["content"].(string); ok {
			return content
		}
	}
	// Fallback: flat content
	if content, ok := msgMap["content"].(string); ok {
		return content
	}
	return ""
}

// ApplyMessageFiltersToData applies content truncation or summary mode to user messages (data array)
func ApplyMessageFiltersToData(messages []interface{}, maxMessageLength int, contentSummary bool, previewLength int) []interface{} {
	if contentSummary {
		return ApplyContentSummary(messages, previewLength)
	}
	return TruncateMessageContent(messages, maxMessageLength)
}

// ExpandContextTurns takes rawData (matched entries) and expands each matched turn
// by including up to N turns before and after it (within the same session).
// Matched turns are marked with "context":false; surrounding context turns with "context":true.
// Overlapping windows are merged (no duplicates). Order is chronological within each session.
// When excludeCompactSummaries is true, entries with isCompactSummary=true are filtered out
// from both the loaded session turns and the final results.
func ExpandContextTurns(rawData []interface{}, N int, baseDir string, excludeCompactSummaries bool) ([]interface{}, error) {
	if N <= 0 || len(rawData) == 0 {
		return rawData, nil
	}

	// 1. Build set of matched UUIDs and collect distinct sessionIds (preserving order)
	matchedUUIDs := make(map[string]bool)
	var sessionOrder []string
	sessionSeen := make(map[string]bool)

	for _, entry := range rawData {
		obj, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		uuid, _ := obj["uuid"].(string)
		if uuid != "" {
			matchedUUIDs[uuid] = true
		}
		// Raw data uses camelCase sessionId
		sessionID, _ := obj["sessionId"].(string)
		if sessionID == "" {
			// Fallback to snake_case
			sessionID, _ = obj["session_id"].(string)
		}
		if sessionID != "" && !sessionSeen[sessionID] {
			sessionOrder = append(sessionOrder, sessionID)
			sessionSeen[sessionID] = true
		}
	}

	// 2. For each distinct sessionId, load all turns for that session
	sessionTurns := make(map[string][]interface{})
	for _, sessionID := range sessionOrder {
		turns, err := loadTurnsForSession(baseDir, sessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to load turns for session %s: %w", sessionID, err)
		}
		// Filter out compact summaries before building the index
		if excludeCompactSummaries {
			filtered := turns[:0:0]
			for _, turn := range turns {
				obj, ok := turn.(map[string]interface{})
				if !ok {
					filtered = append(filtered, turn)
					continue
				}
				isCompact, _ := obj["isCompactSummary"].(bool)
				if !isCompact {
					filtered = append(filtered, turn)
				}
			}
			turns = filtered
		}
		sessionTurns[sessionID] = turns
	}

	// 3. For each session, find windows around matched turns and mark context field
	// Use seen-uuid map to deduplicate across overlapping windows
	seenUUIDs := make(map[string]bool)
	var result []interface{}

	for _, sessionID := range sessionOrder {
		turns := sessionTurns[sessionID]
		if len(turns) == 0 {
			continue
		}

		// Build UUID→index map for this session
		uuidToIndex := make(map[string]int, len(turns))
		for i, turn := range turns {
			obj, ok := turn.(map[string]interface{})
			if !ok {
				continue
			}
			uuid, _ := obj["uuid"].(string)
			if uuid != "" {
				uuidToIndex[uuid] = i
			}
		}

		// Collect all window indices for matched turns in this session
		// Process in index order for chronological output
		windowSet := make(map[int]bool)
		for _, entry := range rawData {
			obj, ok := entry.(map[string]interface{})
			if !ok {
				continue
			}
			entrySessionID, _ := obj["sessionId"].(string)
			if entrySessionID == "" {
				entrySessionID, _ = obj["session_id"].(string)
			}
			if entrySessionID != sessionID {
				continue
			}
			uuid, _ := obj["uuid"].(string)
			idx, exists := uuidToIndex[uuid]
			if !exists {
				continue
			}
			lo := idx - N
			if lo < 0 {
				lo = 0
			}
			hi := idx + N
			if hi >= len(turns) {
				hi = len(turns) - 1
			}
			for i := lo; i <= hi; i++ {
				windowSet[i] = true
			}
		}

		// Emit turns in index order, skipping duplicates
		for i := 0; i < len(turns); i++ {
			if !windowSet[i] {
				continue
			}
			turnObj, ok := turns[i].(map[string]interface{})
			if !ok {
				continue
			}
			uuid, _ := turnObj["uuid"].(string)
			if uuid != "" && seenUUIDs[uuid] {
				continue
			}
			if uuid != "" {
				seenUUIDs[uuid] = true
			}

			// Copy the object and add the "context" field
			newObj := make(map[string]interface{}, len(turnObj)+1)
			for k, v := range turnObj {
				newObj[k] = v
			}
			if matchedUUIDs[uuid] {
				newObj["context"] = false
			} else {
				newObj["context"] = true
			}
			result = append(result, newObj)
		}
	}

	return result, nil
}

// loadTurnsForSession reads all JSONL files in baseDir and returns the turns
// (entries) that belong to sessionID. Each JSONL file is scanned for entries
// where obj["sessionId"] == sessionID. Returns nil, nil if no entries are found.
func loadTurnsForSession(baseDir, sessionID string) ([]interface{}, error) {
	files, err := getJSONLFiles(baseDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		var turns []interface{}

		f, err := os.Open(file)
		if err != nil {
			continue
		}

		r := bufio.NewReader(f)
		rawMessages, readErr := parser.ReadAllFiltered(r, parser.StrategyDefault)
		f.Close()
		if readErr != nil {
			continue
		}

		for _, raw := range rawMessages {
			if len(raw) == 0 {
				continue
			}
			var obj map[string]interface{}
			if err := json.Unmarshal(raw, &obj); err != nil {
				continue
			}
			sid, _ := obj["sessionId"].(string)
			if sid == sessionID {
				turns = append(turns, obj)
			}
		}

		if len(turns) > 0 {
			return turns, nil
		}
	}

	return nil, nil
}

// getJSONLFiles returns all .jsonl files in a directory (non-recursive)
// Files are sorted by modification time (newest first) to prioritize recent sessions
func getJSONLFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Collect file info with modification times
	type fileInfo struct {
		path    string
		modTime int64 // Unix timestamp for easier sorting
	}
	var fileInfos []fileInfo

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".jsonl" {
			fullPath := filepath.Join(dir, entry.Name())

			// Get file stat for modification time
			info, err := entry.Info()
			if err != nil {
				// Skip files we can't stat
				continue
			}

			fileInfos = append(fileInfos, fileInfo{
				path:    fullPath,
				modTime: info.ModTime().Unix(),
			})
		}
	}

	// Sort by modification time (newest first = descending order)
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].modTime > fileInfos[j].modTime
	})

	// Extract paths
	var files []string
	for _, fi := range fileInfos {
		files = append(files, fi.path)
	}

	return files, nil
}
