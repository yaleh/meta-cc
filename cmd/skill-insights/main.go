package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yaleh/meta-cc/internal/skillinsights"
)

func main() {
	if code := run(os.Args[1:], os.Stdout, os.Stderr); code != 0 {
		os.Exit(code)
	}
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("skill-insights", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var (
		days       = flags.Int("days", 30, "analyze session files modified within the last N days; set 0 for all")
		rootList   = flags.String("roots", "", "comma-separated roots or JSONL files; defaults to Claude and Codex session roots")
		maxFiles   = flags.Int("max-files", 0, "maximum JSONL files to scan; 0 means no limit")
		maxFileMB  = flags.Int64("max-file-mb", 25, "skip JSONL files larger than this size in MiB; set 0 for no size limit")
		maxSamples = flags.Int("max-samples", 12, "maximum rows in sample sections")
		format     = flags.String("format", "markdown", "output format: markdown or json")
		outPath    = flags.String("out", "", "write report to this file instead of stdout")
	)
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	var since time.Time
	if *days > 0 {
		since = time.Now().AddDate(0, 0, -*days)
	}

	opts := skillinsights.Options{
		Roots:      splitRoots(*rootList),
		Since:      since,
		MaxFiles:   *maxFiles,
		MaxFileMB:  *maxFileMB,
		MaxSamples: *maxSamples,
	}
	report, err := skillinsights.Analyze(opts)
	if err != nil {
		fmt.Fprintf(stderr, "skill-insights: %v\n", err)
		return 1
	}

	var output []byte
	switch strings.ToLower(*format) {
	case "json":
		output, err = json.MarshalIndent(report, "", "  ")
		if err == nil {
			output = append(output, '\n')
		}
	case "markdown", "md":
		var b strings.Builder
		skillinsights.WriteMarkdown(&b, report, *maxSamples)
		output = []byte(b.String())
	default:
		err = fmt.Errorf("unsupported format %q", *format)
	}
	if err != nil {
		fmt.Fprintf(stderr, "skill-insights: %v\n", err)
		return 1
	}

	if *outPath == "" {
		fmt.Fprint(stdout, string(output))
		return 0
	}
	if err := os.MkdirAll(filepath.Dir(*outPath), 0755); err != nil {
		fmt.Fprintf(stderr, "skill-insights: %v\n", err)
		return 1
	}
	if err := os.WriteFile(*outPath, output, 0644); err != nil {
		fmt.Fprintf(stderr, "skill-insights: %v\n", err)
		return 1
	}
	return 0
}

func splitRoots(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	roots := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			roots = append(roots, part)
		}
	}
	return roots
}
