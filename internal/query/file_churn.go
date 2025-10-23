package query

import (
	"github.com/yaleh/meta-cc/internal/analyzer"
	"github.com/yaleh/meta-cc/internal/parser"
)

type FileChurnOptions struct {
	Threshold int
}

func DetectFileChurn(entries []parser.SessionEntry, opts FileChurnOptions) []analyzer.FileChurnDetail {
	result := analyzer.DetectFileChurn(entries, opts.Threshold)
	return result.HighChurnFiles
}
