package analyzer

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/yaleh/meta-cc/internal/types"
)

var markerPattern = regexp.MustCompile(`\b(TODO|FIXME|HACK|XXX)\b`)

// MarkerCount holds the count for a single debt marker label.
type MarkerCount struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// FileDebt holds per-file marker count for hotspot ranking.
type FileDebt struct {
	File        string `json:"file"`
	MarkerCount int    `json:"marker_count"`
}

// TechDebtResult is the output of GetTechDebt.
// DataSource is "measured": Markers and HotspotFiles are scanned directly
// from tool output text. NOTE: OpenIssues uses a heuristic (error call with
// no subsequent success for the same tool) and is technically "estimated";
// the top-level DataSource reflects the dominant measured provenance.
type TechDebtResult struct {
	Markers      []MarkerCount `json:"markers"`
	HotspotFiles []FileDebt    `json:"hotspot_files"`
	OpenIssues   int           `json:"open_issues"`
	DataSource   DataSource    `json:"data_source"`
}

// scannerToolNames is the set of tool names whose Output we scan for markers.
var scannerToolNames = map[string]bool{
	"Read":  true,
	"Edit":  true,
	"Write": true,
	"Bash":  true,
}

// knownCodeExtensions is the set of file extensions scanned during source-dir walk.
var knownCodeExtensions = map[string]bool{
	".go": true, ".py": true, ".ts": true, ".js": true, ".tsx": true, ".jsx": true,
	".java": true, ".rs": true, ".c": true, ".h": true, ".cpp": true, ".hpp": true,
	".sh": true, ".yaml": true, ".yml": true, ".md": true, ".json": true, ".toml": true,
}

// hiddenDirNames is the set of directory names skipped during source-dir walk.
var hiddenDirNames = map[string]bool{
	".git": true, "node_modules": true, "vendor": true, ".venv": true, "__pycache__": true,
}

// GetTechDebt scans toolCalls for TODO/FIXME/HACK/XXX markers in outputs and
// detects unresolved errors (tool calls with status "error" that have no
// subsequent success call with the same tool name).
func GetTechDebt(entries []types.SessionEntry, toolCalls []types.ToolCall) (*TechDebtResult, error) {
	labelCounts := make(map[string]int)
	fileCounts := make(map[string]int)

	for _, tc := range toolCalls {
		if !scannerToolNames[tc.ToolName] {
			continue
		}
		matches := markerPattern.FindAllString(tc.Output, -1)
		if len(matches) == 0 {
			continue
		}
		for _, m := range matches {
			labelCounts[m]++
		}
		fp := getFilePath(tc.Input)
		if fp != "" {
			fileCounts[fp] += len(matches)
		}
	}

	// Build Markers slice
	var markers []MarkerCount
	for label, count := range labelCounts {
		markers = append(markers, MarkerCount{Label: label, Count: count})
	}
	sort.Slice(markers, func(i, j int) bool {
		return markers[i].Count > markers[j].Count
	})

	// Build HotspotFiles slice sorted descending by MarkerCount
	var hotspots []FileDebt
	for file, count := range fileCounts {
		hotspots = append(hotspots, FileDebt{File: file, MarkerCount: count})
	}
	sort.Slice(hotspots, func(i, j int) bool {
		if hotspots[i].MarkerCount != hotspots[j].MarkerCount {
			return hotspots[i].MarkerCount > hotspots[j].MarkerCount
		}
		return hotspots[i].File < hotspots[j].File
	})

	// Detect open issues: error calls with no subsequent success for same tool
	openIssues := 0
	for i, tc := range toolCalls {
		if tc.Status != "error" {
			continue
		}
		fixed := false
		for j := i + 1; j < len(toolCalls); j++ {
			if toolCalls[j].ToolName == tc.ToolName && toolCalls[j].Status == "success" {
				fixed = true
				break
			}
		}
		if !fixed {
			openIssues++
		}
	}

	return &TechDebtResult{
		Markers:      markers,
		HotspotFiles: hotspots,
		OpenIssues:   openIssues,
		DataSource:   DataSourceMeasured,
	}, nil
}

// ScanSourceDir walks sourceDir recursively, scanning known code files for
// TODO/FIXME/HACK/XXX markers. Hidden directories (.git, node_modules, etc.)
// are skipped. Scanning stops after maxFiles (safety cap).
// The returned result has DataSourceMeasured provenance.
func ScanSourceDir(sourceDir string) (*TechDebtResult, error) {
	labelCounts := make(map[string]int)
	fileCounts := make(map[string]int)
	filesScanned := 0
	const maxFiles = 10000

	err := filepath.Walk(sourceDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if info.IsDir() {
			base := info.Name()
			if hiddenDirNames[base] {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(p))
		if !knownCodeExtensions[ext] {
			return nil
		}
		filesScanned++
		if filesScanned > maxFiles {
			return nil // silently stop scanning
		}

		f, err := os.Open(p)
		if err != nil {
			return nil
		}

		// NOTE(DIR-038): this hand-rolled bufio.NewReader + ReadBytes('\n')
		// loop predates and duplicates the shape now crystallized in
		// parser.ReadLineBounded (internal/parser/bounded_reader.go).
		// Migrating it is tracked as optional/best-effort follow-up, not
		// done here, to keep this task's diff scoped to closing the
		// check-no-scanner violation in
		// internal/provider/codex/appserver/client.go.
		reader := bufio.NewReader(f)
		lineCount := 0
		for {
			line, readErr := reader.ReadBytes('\n')
			if len(line) > 0 {
				lineCount++
				// Cap lines per file to avoid pathological files
				if lineCount > 20000 {
					break
				}
				matches := markerPattern.FindAll(line, -1)
				if len(matches) > 0 {
					for _, m := range matches {
						labelCounts[string(m)]++
					}
					fileCounts[p] += len(matches)
				}
			}
			if readErr != nil {
				break
			}
		}
		f.Close()
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Build Markers slice
	var markers []MarkerCount
	for label, count := range labelCounts {
		markers = append(markers, MarkerCount{Label: label, Count: count})
	}
	sort.Slice(markers, func(i, j int) bool {
		return markers[i].Count > markers[j].Count
	})

	// Build HotspotFiles slice sorted descending by MarkerCount
	var hotspots []FileDebt
	for file, count := range fileCounts {
		hotspots = append(hotspots, FileDebt{File: file, MarkerCount: count})
	}
	sort.Slice(hotspots, func(i, j int) bool {
		if hotspots[i].MarkerCount != hotspots[j].MarkerCount {
			return hotspots[i].MarkerCount > hotspots[j].MarkerCount
		}
		return hotspots[i].File < hotspots[j].File
	})

	return &TechDebtResult{
		Markers:      markers,
		HotspotFiles: hotspots,
		OpenIssues:   0,
		DataSource:   DataSourceMeasured,
	}, nil
}

// MergeTechDebtResults merges two TechDebtResults into one by adding
// marker counts from both and deduplicating hotspot files by path
// (summing their counts). OpenIssues uses the max of both inputs.
// The DataSource is set to the provided combinedSource.
func MergeTechDebtResults(a, b *TechDebtResult, combinedSource DataSource) *TechDebtResult {
	labelCounts := make(map[string]int)
	fileCounts := make(map[string]int)

	for _, m := range a.Markers {
		labelCounts[m.Label] += m.Count
	}
	for _, m := range b.Markers {
		labelCounts[m.Label] += m.Count
	}

	for _, f := range a.HotspotFiles {
		fileCounts[f.File] += f.MarkerCount
	}
	for _, f := range b.HotspotFiles {
		fileCounts[f.File] += f.MarkerCount
	}

	var markers []MarkerCount
	for label, count := range labelCounts {
		markers = append(markers, MarkerCount{Label: label, Count: count})
	}
	sort.Slice(markers, func(i, j int) bool {
		return markers[i].Count > markers[j].Count
	})

	var hotspots []FileDebt
	for file, count := range fileCounts {
		hotspots = append(hotspots, FileDebt{File: file, MarkerCount: count})
	}
	sort.Slice(hotspots, func(i, j int) bool {
		if hotspots[i].MarkerCount != hotspots[j].MarkerCount {
			return hotspots[i].MarkerCount > hotspots[j].MarkerCount
		}
		return hotspots[i].File < hotspots[j].File
	})

	openIssues := a.OpenIssues
	if b.OpenIssues > openIssues {
		openIssues = b.OpenIssues
	}

	return &TechDebtResult{
		Markers:      markers,
		HotspotFiles: hotspots,
		OpenIssues:   openIssues,
		DataSource:   combinedSource,
	}
}

// getFilePath extracts file path from tool input map.
func getFilePath(input map[string]interface{}) string {
	for _, key := range []string{"file_path", "path"} {
		if v, ok := input[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}
