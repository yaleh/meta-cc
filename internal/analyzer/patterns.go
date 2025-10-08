package analyzer

import (
	"sort"
	"time"

	"github.com/yaleh/meta-cc/internal/parser"
)

// ErrorPattern 表示检测到的错误模式
type ErrorPattern struct {
	PatternID       string         `json:"pattern_id"`        // 模式唯一标识符（基于签名）
	Type            string         `json:"type"`              // 模式类型（目前为 "repeated_error"）
	Occurrences     int            `json:"occurrences"`       // 出现次数
	Signature       string         `json:"signature"`         // 错误签名
	ToolName        string         `json:"tool_name"`         // 工具名称
	ErrorText       string         `json:"error_text"`        // 错误文本示例（第一次出现的错误文本）
	FirstSeen       string         `json:"first_seen"`        // 首次出现的时间戳
	LastSeen        string         `json:"last_seen"`         // 最后一次出现的时间戳
	TimeSpanSeconds int            `json:"time_span_seconds"` // 时间跨度（秒）
	Context         PatternContext `json:"context"`           // 模式上下文信息
}

// PatternContext 表示错误模式的上下文信息
type PatternContext struct {
	TurnUUIDs   []string `json:"turn_uuids"`   // 包含此错误的 Turn UUID 序列
	TurnIndices []int    `json:"turn_indices"` // Turn 在 entries 中的索引
}

// DetectErrorPatterns 检测错误模式
// 返回在会话中重复出现的错误（出现次数 >= 3）
func DetectErrorPatterns(entries []parser.SessionEntry, toolCalls []parser.ToolCall) []ErrorPattern {
	// 构建 UUID -> 索引映射
	uuidToIndex := make(map[string]int)
	for i, entry := range entries {
		uuidToIndex[entry.UUID] = i
	}

	// 构建 UUID -> 时间戳映射
	uuidToTimestamp := make(map[string]string)
	for _, entry := range entries {
		uuidToTimestamp[entry.UUID] = entry.Timestamp
	}

	// 按签名分组错误
	errorGroups := make(map[string][]parser.ToolCall)

	for _, tc := range toolCalls {
		// 仅处理错误状态的工具调用
		if tc.Status != "error" && tc.Error == "" {
			continue
		}

		// 计算错误签名
		signature := CalculateErrorSignature(tc.ToolName, tc.Error)

		// 按签名分组
		errorGroups[signature] = append(errorGroups[signature], tc)
	}

	// 检测模式（出现次数 >= 3）
	var patterns []ErrorPattern

	for signature, group := range errorGroups {
		if len(group) < 3 {
			continue // 少于 3 次不形成模式
		}

		// 构建模式
		pattern := ErrorPattern{
			PatternID:   signature,
			Type:        "repeated_error",
			ToolName:    group[0].ToolName,
			Occurrences: len(group),
			Signature:   signature,
			ErrorText:   group[0].Error,
			Context:     buildPatternContext(group, uuidToIndex, uuidToTimestamp),
		}

		// 设置时间信息
		timestamps := extractTimestamps(group, uuidToTimestamp)
		if len(timestamps) > 0 {
			pattern.FirstSeen = timestamps[0]
			pattern.LastSeen = timestamps[len(timestamps)-1]
			pattern.TimeSpanSeconds = calculateTimeSpan(timestamps[0], timestamps[len(timestamps)-1])
		}

		patterns = append(patterns, pattern)
	}

	// 按出现次数降序排序
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Occurrences > patterns[j].Occurrences
	})

	return patterns
}

// buildPatternContext 构建模式上下文
func buildPatternContext(toolCalls []parser.ToolCall, uuidToIndex map[string]int, uuidToTimestamp map[string]string) PatternContext {
	context := PatternContext{
		TurnUUIDs:   make([]string, 0, len(toolCalls)),
		TurnIndices: make([]int, 0, len(toolCalls)),
	}

	for _, tc := range toolCalls {
		context.TurnUUIDs = append(context.TurnUUIDs, tc.UUID)
		if index, found := uuidToIndex[tc.UUID]; found {
			context.TurnIndices = append(context.TurnIndices, index)
		}
	}

	return context
}

// extractTimestamps 提取时间戳并排序
func extractTimestamps(toolCalls []parser.ToolCall, uuidToTimestamp map[string]string) []string {
	timestamps := make([]string, 0, len(toolCalls))

	for _, tc := range toolCalls {
		if ts, found := uuidToTimestamp[tc.UUID]; found && ts != "" {
			timestamps = append(timestamps, ts)
		}
	}

	// 按时间排序
	sort.Strings(timestamps)

	return timestamps
}

// calculateTimeSpan 计算时间跨度（秒）
func calculateTimeSpan(first, last string) int {
	if first == "" || last == "" {
		return 0
	}

	firstTime, err1 := time.Parse(time.RFC3339, first)
	lastTime, err2 := time.Parse(time.RFC3339, last)

	if err1 != nil || err2 != nil {
		return 0
	}

	return int(lastTime.Sub(firstTime).Seconds())
}
