package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yaleh/meta-cc/internal/validation"
)

func main() {
	// Parse flags
	filePath := flag.String("file", "cmd/mcp-server/tools.go", "Path to tools.go")
	fast := flag.Bool("fast", true, "Run fast checks only (MVP mode)")
	quiet := flag.Bool("quiet", false, "Suppress output except errors")
	jsonOutput := flag.Bool("json", false, "Output as JSON")
	flag.Parse()

	// Parse tools.go
	tools, err := validation.ParseTools(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing tools.go: %v\n", err)
		os.Exit(2)
	}

	// Run validations
	validator := validation.NewValidator(*fast)
	report := validator.Validate(tools)

	// Output results
	reporter := validation.NewReporter(*quiet, *jsonOutput)
	reporter.Print(report)

	// Exit with appropriate code
	if report.Failed > 0 {
		os.Exit(1)
	}

	os.Exit(0)
}
