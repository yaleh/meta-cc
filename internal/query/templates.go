package query

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// QueryTemplate represents a query template definition
type QueryTemplate struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Category    string      `yaml:"category"`
	Filter      string      `yaml:"filter"`
	Examples    []Example   `yaml:"examples"`
	Parameters  []Parameter `yaml:"parameters"`
}

// Example represents a usage example for a query template
type Example struct {
	Description string `yaml:"description"`
	Command     string `yaml:"command"`
}

// Parameter represents a parameter for a query template
type Parameter struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
	Optional    bool   `yaml:"optional"`
}

// LoadTemplates loads all query templates from the templates directory
func LoadTemplates() (map[string]QueryTemplate, error) {
	templates := make(map[string]QueryTemplate)

	// Try to find the absolute path to templates directory
	// First check if we're running from the project root
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	// Try common paths
	possiblePaths := []string{
		filepath.Join(cwd, "internal", "query", "templates"),
		filepath.Join(cwd, "query", "templates"),
		"internal/query/templates",
		"templates",
	}

	var foundTemplatesDir string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			foundTemplatesDir = path
			break
		}
	}

	// If we didn't find the templates directory, return empty map
	if foundTemplatesDir == "" {
		return templates, nil
	}

	// Read all YAML files in templates directory
	files, err := os.ReadDir(foundTemplatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Only process .yaml files
		if filepath.Ext(file.Name()) != ".yaml" {
			continue
		}

		// Read template file
		templatePath := filepath.Join(foundTemplatesDir, file.Name())
		data, err := os.ReadFile(templatePath)
		if err != nil {
			// Skip files that can't be read
			continue
		}

		// Parse YAML
		var template QueryTemplate
		if err := yaml.Unmarshal(data, &template); err != nil {
			// Skip files that can't be parsed
			continue
		}

		// Add to templates map
		templates[template.Name] = template
	}

	return templates, nil
}
