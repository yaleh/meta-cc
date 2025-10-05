package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/analyzer"
	"github.com/yale/meta-cc/internal/locator"
	"github.com/yale/meta-cc/internal/parser"
	"github.com/yale/meta-cc/pkg/output"
)

var (
	analyzeWindow int // 分析窗口大小（最近 N 个 Turn）
)

// analyzeCmd 表示 analyze 子命令
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze Claude Code session patterns",
	Long: `Analyze Claude Code session data to detect patterns and insights.

Supports error pattern detection, tool usage analysis, and more.

Examples:
  meta-cc analyze errors
  meta-cc analyze errors --window 20
  meta-cc analyze errors --output md`,
}

// analyzeErrorsCmd 表示 analyze errors 子子命令
var analyzeErrorsCmd = &cobra.Command{
	Use:   "errors",
	Short: "Detect error patterns in session",
	Long: `Detect repeated error patterns in Claude Code session data.

Identifies errors that occur 3 or more times within the analysis window.
Provides context including turn sequences, timestamps, and time spans.

Examples:
  meta-cc analyze errors
  meta-cc analyze errors --window 20 --output md
  meta-cc analyze errors --output json`,
	RunE: runAnalyzeErrors,
}

func init() {
	// 将 analyze 子命令添加到 root
	rootCmd.AddCommand(analyzeCmd)

	// 将 errors 子子命令添加到 analyze
	analyzeCmd.AddCommand(analyzeErrorsCmd)

	// errors 子命令的参数
	analyzeErrorsCmd.Flags().IntVarP(&analyzeWindow, "window", "w", 0, "Analyze last N turns (0 = analyze entire session)")

	// --output 参数已在 root.go 中定义为全局参数
}

func runAnalyzeErrors(cmd *cobra.Command, args []string) error {
	// Step 1: 定位会话文件（使用 Phase 1 的 locator）
	loc := locator.NewSessionLocator()
	sessionPath, err := loc.Locate(locator.LocateOptions{
		SessionID:   sessionID,   // 来自全局参数
		ProjectPath: projectPath, // 来自全局参数
	})
	if err != nil {
		return fmt.Errorf("failed to locate session file: %w", err)
	}

	// Step 2: 解析会话文件（使用 Phase 2 的 parser）
	sessionParser := parser.NewSessionParser(sessionPath)
	entries, err := sessionParser.ParseEntries()
	if err != nil {
		return fmt.Errorf("failed to parse session file: %w", err)
	}

	// Step 3: 应用窗口过滤（如果指定）
	if analyzeWindow > 0 && len(entries) > analyzeWindow {
		entries = entries[len(entries)-analyzeWindow:]
	}

	// Step 4: 提取工具调用
	toolCalls := parser.ExtractToolCalls(entries)

	// Step 5: 检测错误模式
	patterns := analyzer.DetectErrorPatterns(entries, toolCalls)

	// Step 6: 格式化输出
	var outputStr string
	var formatErr error

	if len(patterns) == 0 {
		// 无错误模式
		switch outputFormat {
		case "jsonl":
			outputStr = "[]"
		case "tsv":
			outputStr = ""
		default:
			return fmt.Errorf("unsupported output format: %s (supported: jsonl, tsv)", outputFormat)
		}
	} else {
		// 有错误模式
		switch outputFormat {
		case "jsonl":
			outputStr, formatErr = output.FormatJSONL(patterns)
		case "tsv":
			outputStr, formatErr = output.FormatTSV(patterns)
		default:
			return fmt.Errorf("unsupported output format: %s (supported: jsonl, tsv)", outputFormat)
		}

		if formatErr != nil {
			return fmt.Errorf("failed to format output: %w", formatErr)
		}
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)

	return nil
}

// formatErrorPatternsMarkdown 格式化错误模式为 Markdown 报告
func formatErrorPatternsMarkdown(patterns []analyzer.ErrorPattern) (string, error) {
	var sb strings.Builder

	sb.WriteString("# Error Pattern Analysis\n\n")
	sb.WriteString(fmt.Sprintf("Found %d error pattern(s):\n\n", len(patterns)))

	for i, pattern := range patterns {
		sb.WriteString(fmt.Sprintf("## Pattern %d: %s\n\n", i+1, pattern.ToolName))

		sb.WriteString(fmt.Sprintf("- **Type**: %s\n", pattern.Type))
		sb.WriteString(fmt.Sprintf("- **Occurrences**: %d times\n", pattern.Occurrences))
		sb.WriteString(fmt.Sprintf("- **Signature**: `%s`\n", pattern.Signature))
		sb.WriteString(fmt.Sprintf("- **Error**: %s\n", pattern.ErrorText))
		sb.WriteString("\n")

		// Context section
		sb.WriteString("### Context\n\n")
		sb.WriteString(fmt.Sprintf("- **First Occurrence**: %s\n", pattern.FirstSeen))
		sb.WriteString(fmt.Sprintf("- **Last Occurrence**: %s\n", pattern.LastSeen))
		sb.WriteString(fmt.Sprintf("- **Time Span**: %d seconds (%.1f minutes)\n",
			pattern.TimeSpanSeconds, float64(pattern.TimeSpanSeconds)/60))
		sb.WriteString(fmt.Sprintf("- **Affected Turns**: %d\n", len(pattern.Context.TurnUUIDs)))
		sb.WriteString("\n")

		// Turn sequence (limited to first 5)
		if len(pattern.Context.TurnUUIDs) > 0 {
			sb.WriteString("**Turn Sequence** (first 5):\n")
			limit := len(pattern.Context.TurnUUIDs)
			if limit > 5 {
				limit = 5
			}
			for j := 0; j < limit; j++ {
				sb.WriteString(fmt.Sprintf("- `%s`\n", pattern.Context.TurnUUIDs[j]))
			}
			if len(pattern.Context.TurnUUIDs) > 5 {
				sb.WriteString(fmt.Sprintf("- ... and %d more\n", len(pattern.Context.TurnUUIDs)-5))
			}
			sb.WriteString("\n")
		}

		sb.WriteString("---\n\n")
	}

	return sb.String(), nil
}
