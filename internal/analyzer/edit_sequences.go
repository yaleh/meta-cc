package analyzer

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yaleh/meta-cc/internal/types"
)

// EditContent holds optional full content for an edit event.
type EditContent struct {
	OldString string `json:"oldString,omitempty"`
	NewString string `json:"newString,omitempty"`
}

// EditEvent represents a single file-access event (Read, Edit, or Write).
type EditEvent struct {
	Timestamp   string       `json:"timestamp"`
	SessionID   string       `json:"sessionId"`
	Tool        string       `json:"tool"`
	ContentHint string       `json:"contentHint"`
	FileType    string       `json:"fileType"`
	DocRole     string       `json:"docRole,omitempty"`
	Content     *EditContent `json:"content,omitempty"`
}

// CoAccessedDoc represents a doc file co-accessed with a source file.
type CoAccessedDoc struct {
	FilePath      string `json:"filePath"`
	DocRole       string `json:"docRole"`
	CoAccessCount int    `json:"coAccessCount"`
	TotalDocReads int    `json:"totalDocReads"`
}

// FileEditSequence holds the edit/read history for a single file.
type FileEditSequence struct {
	SessionCount     int             `json:"sessionCount"`
	TotalReads       int             `json:"totalReads"`
	TotalEdits       int             `json:"totalEdits"`
	ReadEditRatio    float64         `json:"readEditRatio"`
	PatternHint      string          `json:"patternHint"`
	Events           []EditEvent     `json:"events"`
	CoAccessedDocs   []CoAccessedDoc `json:"coAccessedDocs,omitempty"`
	DocVoid          bool            `json:"docVoid"`
	SpecPrecisionGap bool            `json:"specPrecisionGap"`
}

// EditSequenceSummary holds aggregate stats across all files.
type EditSequenceSummary struct {
	TotalFiles          int            `json:"totalFiles"`
	PatternDistribution map[string]int `json:"patternDistribution"`
}

// EditSequencesResult is the top-level result type for BuildEditSequences.
type EditSequencesResult struct {
	Files   map[string]FileEditSequence `json:"files"`
	Summary EditSequenceSummary         `json:"summary"`
}

// classifyFileType returns "doc", "source", "config", or "other" based on extension.
func classifyFileType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".md", ".rst", ".txt":
		return "doc"
	case ".go", ".ts", ".py", ".java", ".cpp", ".rs", ".kt":
		return "source"
	case ".json", ".yaml", ".toml", ".env", ".lock":
		return "config"
	default:
		return "other"
	}
}

// buildContentHint generates a human-readable hint string for an event.
func buildContentHint(tool string, input map[string]interface{}, _ bool) string {
	switch tool {
	case "Read":
		fp, _ := input["file_path"].(string)
		return fmt.Sprintf("file_path=%s", fp)
	case "Edit":
		old, _ := input["old_string"].(string)
		newStr, _ := input["new_string"].(string)
		if len(old) > 40 {
			old = old[:40]
		}
		if len(newStr) > 40 {
			newStr = newStr[:40]
		}
		return fmt.Sprintf("old: '%s' → new: '%s'", old, newStr)
	case "Write":
		fp, _ := input["file_path"].(string)
		content, _ := input["content"].(string)
		return fmt.Sprintf("write %d bytes to %s", len(content), fp)
	default:
		return ""
	}
}

// computeDocRole returns the docRole for a doc file given its read/edit stats.
func computeDocRole(reads, edits int) string {
	if edits == 0 {
		// Read-only doc: classify based on reads
		if reads > 0 {
			return "spec"
		}
		return ""
	}
	ratio := float64(reads) / float64(edits)
	if ratio >= 3.0 {
		return "spec"
	}
	if ratio <= 0.5 && edits >= 3 {
		return "output"
	}
	return "mixed"
}

// computePatternHint returns "A", "B", or "C" based on read/edit ratio.
func computePatternHint(reads, edits int) string {
	var ratio float64
	if edits > 0 {
		ratio = float64(reads) / float64(edits)
	} else {
		ratio = float64(reads)
	}
	if ratio >= 3.0 {
		return "A"
	}
	if ratio <= 0.8 && edits >= 5 {
		return "B"
	}
	return "C"
}

// BuildEditSequences analyzes session entries and returns file edit sequences.
func BuildEditSequences(entries []types.SessionEntry, files []string, includeContent bool, limitPerFile int) EditSequencesResult {
	if limitPerFile <= 0 {
		limitPerFile = 50
	}

	// Build file filter set
	fileSet := make(map[string]bool, len(files))
	for _, f := range files {
		fileSet[f] = true
	}

	// Build uuid→sessionID lookup
	uuidToSession := make(map[string]string)
	for _, e := range entries {
		uuidToSession[e.UUID] = e.SessionID
	}

	// Extract tool calls
	toolCalls := types.ExtractToolCalls(entries)

	// Group raw events by file path
	type rawEvent struct {
		timestamp string
		sessionID string
		tool      string
		input     map[string]interface{}
	}
	fileEvents := make(map[string][]rawEvent)

	for _, tc := range toolCalls {
		action := types.FileActionType(tc.ToolName)
		if action == "" {
			continue
		}
		fp, _ := tc.Input["file_path"].(string)
		if fp == "" {
			continue
		}
		if len(fileSet) > 0 && !fileSet[fp] {
			continue
		}
		sid := uuidToSession[tc.UUID]
		fileEvents[fp] = append(fileEvents[fp], rawEvent{
			timestamp: tc.Timestamp,
			sessionID: sid,
			tool:      action,
			input:     tc.Input,
		})
	}

	result := EditSequencesResult{
		Files: make(map[string]FileEditSequence),
		Summary: EditSequenceSummary{
			PatternDistribution: map[string]int{"A": 0, "B": 0, "C": 0},
		},
	}

	// Build per-file sequences
	for fp, evs := range fileEvents {
		sort.Slice(evs, func(i, j int) bool { return evs[i].timestamp < evs[j].timestamp })

		ft := classifyFileType(fp)
		sessions := make(map[string]bool)
		reads, edits := 0, 0

		var events []EditEvent
		for _, ev := range evs {
			sessions[ev.sessionID] = true
			hint := buildContentHint(ev.tool, ev.input, includeContent)
			event := EditEvent{
				Timestamp:   ev.timestamp,
				SessionID:   ev.sessionID,
				Tool:        ev.tool,
				ContentHint: hint,
				FileType:    ft,
			}
			switch ev.tool {
			case "Read":
				reads++
			case "Edit", "Write":
				edits++
			}
			events = append(events, event)
		}

		pattern := computePatternHint(reads, edits)
		docRole := ""
		if ft == "doc" {
			docRole = computeDocRole(reads, edits)
		}

		// Denormalize fileType and docRole onto all events
		for i := range events {
			events[i].FileType = ft
			if ft == "doc" {
				events[i].DocRole = docRole
			}
		}

		// Apply limitPerFile
		if len(events) > limitPerFile {
			events = events[:limitPerFile]
		}

		seq := FileEditSequence{
			SessionCount: len(sessions),
			TotalReads:   reads,
			TotalEdits:   edits,
			ReadEditRatio: func() float64 {
				if edits > 0 {
					return float64(reads) / float64(edits)
				}
				return float64(reads)
			}(),
			PatternHint: pattern,
			Events:      events,
		}
		result.Files[fp] = seq
		result.Summary.PatternDistribution[pattern]++
	}

	// ─── Phase C: CoAccessedDocs + DocVoid + SpecPrecisionGap ────────────────

	// Build globalDocStats: doc file path → {reads, edits} across ALL tool calls
	// (not just the filtered files). This allows DocRole to be computed for doc
	// files that were co-accessed but NOT in the input files list.
	type docStats struct{ reads, edits int }
	globalDocStats := make(map[string]*docStats)
	for _, tc := range toolCalls {
		action := types.FileActionType(tc.ToolName)
		if action == "" {
			continue
		}
		fp, _ := tc.Input["file_path"].(string)
		if fp == "" || classifyFileType(fp) != "doc" {
			continue
		}
		if globalDocStats[fp] == nil {
			globalDocStats[fp] = &docStats{}
		}
		switch action {
		case "Read":
			globalDocStats[fp].reads++
		case "Edit", "Write":
			globalDocStats[fp].edits++
		}
	}

	// Build sessionToFiles: sessionID → set of file paths touched
	sessionToFiles := make(map[string]map[string]bool)
	for _, tc := range toolCalls {
		if types.FileActionType(tc.ToolName) == "" {
			continue
		}
		fp, _ := tc.Input["file_path"].(string)
		if fp == "" {
			continue
		}
		sid := uuidToSession[tc.UUID]
		if sessionToFiles[sid] == nil {
			sessionToFiles[sid] = make(map[string]bool)
		}
		sessionToFiles[sid][fp] = true
	}

	// Build sessionToDocReads: sessionID → docFilePath → read count
	sessionToDocReads := make(map[string]map[string]int)
	for _, tc := range toolCalls {
		if tc.ToolName != "Read" {
			continue
		}
		fp, _ := tc.Input["file_path"].(string)
		if fp == "" || classifyFileType(fp) != "doc" {
			continue
		}
		sid := uuidToSession[tc.UUID]
		if sessionToDocReads[sid] == nil {
			sessionToDocReads[sid] = make(map[string]int)
		}
		sessionToDocReads[sid][fp]++
	}

	// For each file sequence, compute CoAccessedDocs, DocVoid, SpecPrecisionGap
	for fp, seq := range result.Files {
		// Collect sessions where this file was touched
		fileSessions := make(map[string]bool)
		for _, ev := range seq.Events {
			fileSessions[ev.SessionID] = true
		}
		// Also include sessions beyond the limitPerFile cut
		for _, ev := range fileEvents[fp] {
			fileSessions[ev.sessionID] = true
		}

		// Aggregate co-accessed docs
		docCoAccess := make(map[string]*CoAccessedDoc)
		for sid := range fileSessions {
			for docFP := range sessionToFiles[sid] {
				if classifyFileType(docFP) != "doc" {
					continue
				}
				if docFP == fp {
					continue
				}
				if docCoAccess[docFP] == nil {
					// Determine docRole: prefer the doc's own FileEditSequence (if it was
					// in the queried files list), otherwise fall back to globalDocStats so
					// that non-queried doc files still get a computed role.
					docDocRole := ""
					if docSeq, ok := result.Files[docFP]; ok {
						docDocRole = computeDocRole(docSeq.TotalReads, docSeq.TotalEdits)
					} else if gs, ok := globalDocStats[docFP]; ok {
						docDocRole = computeDocRole(gs.reads, gs.edits)
					}
					docCoAccess[docFP] = &CoAccessedDoc{FilePath: docFP, DocRole: docDocRole}
				}
				docCoAccess[docFP].CoAccessCount++
				docCoAccess[docFP].TotalDocReads += sessionToDocReads[sid][docFP]
			}
		}

		// Build sorted slice
		var coAccessed []CoAccessedDoc
		for _, d := range docCoAccess {
			coAccessed = append(coAccessed, *d)
		}
		sort.Slice(coAccessed, func(i, j int) bool {
			if coAccessed[i].CoAccessCount != coAccessed[j].CoAccessCount {
				return coAccessed[i].CoAccessCount > coAccessed[j].CoAccessCount
			}
			return coAccessed[i].FilePath < coAccessed[j].FilePath
		})

		// Compute DocVoid: pattern B + no co-accessed docs + reads < edits*0.8
		docVoid := seq.PatternHint == "B" &&
			len(coAccessed) == 0 &&
			float64(seq.TotalReads) < float64(seq.TotalEdits)*0.8

		// Compute SpecPrecisionGap: pattern B + co-accessed spec doc with totalDocReads ≥ 3
		specPrecisionGap := false
		if seq.PatternHint == "B" {
			for _, d := range coAccessed {
				if d.DocRole == "spec" && d.TotalDocReads >= 3 {
					specPrecisionGap = true
					break
				}
			}
		}

		seq.CoAccessedDocs = coAccessed
		seq.DocVoid = docVoid
		seq.SpecPrecisionGap = specPrecisionGap
		result.Files[fp] = seq
	}

	result.Summary.TotalFiles = len(result.Files)
	return result
}
