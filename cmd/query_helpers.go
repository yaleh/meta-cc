package cmd

import (
	"strings"

	"github.com/yaleh/meta-cc/internal/parser"
	pipelinepkg "github.com/yaleh/meta-cc/pkg/pipeline"
)

func toPipelineOptions(opts GlobalOptions) pipelinepkg.GlobalOptions {
	return pipelinepkg.GlobalOptions{
		SessionID:   opts.SessionID,
		ProjectPath: opts.ProjectPath,
		SessionOnly: opts.SessionOnly,
	}
}

func buildTurnIndex(entries []parser.SessionEntry) map[string]int {
	turnIndex := make(map[string]int)
	turn := 0

	for _, entry := range entries {
		if entry.IsMessage() {
			turnIndex[entry.UUID] = turn
			turn++
		}
	}

	return turnIndex
}

func isSystemMessage(content string) bool {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return false
	}

	systemPrefixes := []string{
		"<command-message>",
		"<command-name>",
		"<command-args>",
		"<local-command",
		"Caveat:",
		"# meta-",
	}

	for _, prefix := range systemPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}

	return false
}
