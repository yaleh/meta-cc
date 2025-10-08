package output

import (
	"fmt"

	pkgOutput "github.com/yaleh/meta-cc/pkg/output"
)

// FormatOutput formats data based on the output format
// Supports all data types via generic TSV formatting
func FormatOutput(data interface{}, format string) (string, error) {
	switch format {
	case "jsonl":
		return pkgOutput.FormatJSONL(data)
	case "tsv":
		return pkgOutput.FormatTSV(data)
	default:
		return "", fmt.Errorf("unsupported output format: %s (supported: jsonl, tsv)", format)
	}
}
