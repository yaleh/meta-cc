package validation

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ParseTools extracts tool definitions from tools.go using regex
func ParseTools(filePath string) ([]Tool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return parseToolsFromContent(string(content))
}

func parseToolsFromContent(content string) ([]Tool, error) {
	var tools []Tool

	// Extract the getToolDefinitions() function body
	funcPattern := regexp.MustCompile(`(?s)func getToolDefinitions\(\).*?\{(.*?)\n\}`)
	funcMatch := funcPattern.FindStringSubmatch(content)
	if len(funcMatch) < 2 {
		return nil, fmt.Errorf("could not find getToolDefinitions() function")
	}

	funcBody := funcMatch[1]

	// Extract individual tool definitions
	toolPattern := regexp.MustCompile(`(?s)\{[\s\n]*Name:\s*"([^"]+)",[\s\n]*Description:\s*"([^"]+)",[\s\n]*InputSchema:.*?\},`)
	toolMatches := toolPattern.FindAllStringSubmatch(funcBody, -1)

	for _, match := range toolMatches {
		if len(match) < 3 {
			continue
		}

		name := match[1]
		description := match[2]

		// Extract parameters and required fields for this tool
		namePos := strings.Index(funcBody, fmt.Sprintf(`Name:        "%s"`, name))
		if namePos == -1 {
			continue
		}

		// Find the opening brace before the Name field (struct literal starts with {)
		toolDefStart := strings.LastIndex(funcBody[:namePos], "{")
		if toolDefStart == -1 {
			continue
		}

		// Find the closing brace for this tool definition
		toolDefEnd := findClosingBrace(funcBody[toolDefStart:])
		if toolDefEnd == -1 {
			continue
		}

		toolDef := funcBody[toolDefStart : toolDefStart+toolDefEnd+1]

		properties := parseProperties(toolDef)
		required := parseRequired(toolDef)

		tool := Tool{
			Name:        name,
			Description: description,
			InputSchema: InputSchema{
				Type:       "object",
				Properties: properties,
				Required:   required,
			},
		}

		tools = append(tools, tool)
	}

	return tools, nil
}

func parseProperties(toolDef string) map[string]Property {
	properties := make(map[string]Property)

	// Match property definitions - support both struct field syntax and map key syntax
	// Struct field: "param_name": { Type: "type", Description: "desc" }
	// Map key: "param_name": map[string]interface{}{ "type": "type", "description": "desc" }

	// Try struct field syntax first
	propPattern := regexp.MustCompile(`"([^"]+)":\s*\{[\s\n]*Type:\s*"([^"]+)",[\s\n]*Description:\s*"([^"]+)",?[\s\n]*\}`)
	propMatches := propPattern.FindAllStringSubmatch(toolDef, -1)

	for _, match := range propMatches {
		if len(match) < 4 {
			continue
		}

		paramName := match[1]
		paramType := match[2]
		paramDesc := match[3]

		// Skip standard parameters (these are added by MergeParameters)
		if isStandardParameter(paramName) {
			continue
		}

		properties[paramName] = Property{
			Type:        paramType,
			Description: paramDesc,
		}
	}

	// Try map key syntax (lowercase field names in maps)
	mapPropPattern := regexp.MustCompile(`"([^"]+)":\s*map\[string\]interface\{\}\s*\{[\s\n]*"type":\s*"([^"]+)",[\s\n]*"description":\s*"([^"]+)",?[\s\n]*\}`)
	mapPropMatches := mapPropPattern.FindAllStringSubmatch(toolDef, -1)

	for _, match := range mapPropMatches {
		if len(match) < 4 {
			continue
		}

		paramName := match[1]
		paramType := match[2]
		paramDesc := match[3]

		// Skip standard parameters (these are added by MergeParameters)
		if isStandardParameter(paramName) {
			continue
		}

		properties[paramName] = Property{
			Type:        paramType,
			Description: paramDesc,
		}
	}

	return properties
}

func parseRequired(toolDef string) []string {
	var required []string

	// Match Required: []string{"param1", "param2"} (struct field syntax)
	requiredPattern := regexp.MustCompile(`Required:\s*\[\]string\{([^}]+)\}`)
	requiredMatch := requiredPattern.FindStringSubmatch(toolDef)

	if len(requiredMatch) > 1 {
		// Extract quoted strings
		quotedPattern := regexp.MustCompile(`"([^"]+)"`)
		quotedMatches := quotedPattern.FindAllStringSubmatch(requiredMatch[1], -1)

		for _, match := range quotedMatches {
			if len(match) > 1 {
				required = append(required, match[1])
			}
		}
	}

	// Match "required": []string{"param1", "param2"} (map key syntax)
	mapRequiredPattern := regexp.MustCompile(`"required":\s*\[\]string\{([^}]+)\}`)
	mapRequiredMatch := mapRequiredPattern.FindStringSubmatch(toolDef)

	if len(mapRequiredMatch) > 1 {
		// Extract quoted strings
		quotedPattern := regexp.MustCompile(`"([^"]+)"`)
		quotedMatches := quotedPattern.FindAllStringSubmatch(mapRequiredMatch[1], -1)

		for _, match := range quotedMatches {
			if len(match) > 1 {
				required = append(required, match[1])
			}
		}
	}

	return required
}

func isStandardParameter(name string) bool {
	standardParams := []string{
		"scope", "jq_filter", "stats_only", "stats_first",
		"inline_threshold_bytes", "output_format",
	}

	for _, param := range standardParams {
		if name == param {
			return true
		}
	}

	return false
}

func findClosingBrace(s string) int {
	depth := 0
	for i, char := range s {
		if char == '{' {
			depth++
		} else if char == '}' {
			if depth == 0 {
				// Found closing before opening
				return i
			}
			depth--
			if depth == 0 {
				// Found matching closing brace
				return i
			}
		}
	}
	return -1
}
