package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidateDescription checks if tool descriptions follow template format
func ValidateDescription(tool Tool) Result {
	desc := tool.Description

	// Check for "Default scope:" presence
	if !strings.Contains(desc, "Default scope:") {
		return NewFailResult(
			tool.Name,
			"description_scope",
			"Description must include 'Default scope:' suffix",
			map[string]interface{}{
				"actual":    desc,
				"expected":  "<Action> <object>. Default scope: <project|session|none>.",
				"reference": "api-consistency-methodology.md (Section 4)",
			},
		)
	}

	// Check template pattern: starts with capital, ends with "Default scope: <X>."
	pattern := regexp.MustCompile(`^[A-Z].*\. Default scope: (project|session|none)\.$`)
	if !pattern.MatchString(desc) {
		return NewFailResult(
			tool.Name,
			"description_format",
			"Description must match template format",
			map[string]interface{}{
				"template": "<Action> <object>. Default scope: <project|session|none>.",
				"actual":   desc,
				"reference": "api-consistency-methodology.md (Section 4)",
			},
		)
	}

	// Check length (warning only)
	if len(desc) > 100 {
		return NewWarnResult(
			tool.Name,
			"description_length",
			fmt.Sprintf("Description exceeds 100 characters (%d chars)", len(desc)),
		)
	}

	return NewPassResult(tool.Name, "description")
}
