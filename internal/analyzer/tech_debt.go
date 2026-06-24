package analyzer

import (
	"regexp"
	"sort"

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
