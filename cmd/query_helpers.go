package cmd

import (
	"strings"

	pipelinepkg "github.com/yaleh/meta-cc/pkg/pipeline"
)

func toPipelineOptions(opts GlobalOptions) pipelinepkg.GlobalOptions {
	return pipelinepkg.GlobalOptions{
		SessionID:   opts.SessionID,
		ProjectPath: opts.ProjectPath,
		SessionOnly: opts.SessionOnly,
	}
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
