package appserver

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
)

// MapThread converts an app-server Thread (as returned by thread/list or
// thread/read) into the canonical conversation.Session, reusing the DIR-028
// Thread->Turn->Item model directly rather than a parallel struct. Turns is
// only populated when the caller requested includeTurns (thread/read); a
// thread/list-sourced Thread maps to a Session with no Turns, matching the
// files backend's ListSessions/GetSession vs LoadTurns split.
func MapThread(t Thread) conversation.Session {
	title := t.Preview
	if t.Name != nil && *t.Name != "" {
		title = *t.Name
	}

	turns := make([]conversation.Turn, 0, len(t.Turns))
	for _, turn := range t.Turns {
		turns = append(turns, mapTurn(turn))
	}

	parentThreadID := ""
	if t.ParentThreadID != nil {
		parentThreadID = *t.ParentThreadID
	}
	sourceKind := parseSourceKind(t.Source)
	isSubagent := strings.HasPrefix(sourceKind, "subAgent")

	ext, _ := json.Marshal(map[string]interface{}{
		"backend":          "app_server",
		"session_id":       t.SessionID,
		"cli_version":      t.CLIVersion,
		"source":           json.RawMessage(t.Source),
		"parent_thread_id": t.ParentThreadID,
	})

	return conversation.Session{
		ID:             t.ID,
		Provider:       conversation.ProviderCodex,
		Title:          title,
		CWD:            t.CWD,
		Model:          t.ModelProvider,
		ModelProvider:  t.ModelProvider,
		SourceKind:     sourceKind,
		Status:         "active", // caller (listSessions) overwrites with Archived/Status when it knows which archived-pass this thread came from
		ParentThreadID: parentThreadID,
		Lineage:        mapLineage(parentThreadID, isSubagent),
		IsSubagent:     isSubagent,
		CreatedAt:      time.Unix(t.CreatedAt, 0).UTC(),
		UpdatedAt:      time.Unix(t.UpdatedAt, 0).UTC(),
		Turns:          turns,
		Extensions:     ext,
	}
}

// mapLineage classifies what MapThread actually knows about a thread's
// spawn relationship (DIR-032). A non-empty parentThreadID is a confirmed
// child. An empty parentThreadID from a subagent source kind is treated as
// unknown rather than root: subagent threads are spawned by construction,
// so a missing parent edge there means the app-server suppressed/omitted
// the spawn metadata, not that the thread has no parent. Every other empty
// case (an ordinary cli/vscode/exec/appServer thread) is a confirmed root.
func mapLineage(parentThreadID string, isSubagent bool) conversation.LineageStatus {
	if parentThreadID != "" {
		return conversation.LineageStatusChild
	}
	if isSubagent {
		return conversation.LineageStatusUnknown
	}
	return conversation.LineageStatusRoot
}

// parseSourceKind extracts the "type" discriminant from a Thread's raw
// Source payload (a oneOf tagged by "type", matching the ThreadSourceKind
// enum documented in docs/reference/codex-app-server.md — "cli", "vscode",
// "exec", "subAgent", ...). An unparseable/empty Source yields "" rather
// than an error: source kind is a filter convenience, not something a
// listing should fail over.
func parseSourceKind(raw json.RawMessage) string {
	var payload struct {
		Type string `json:"type"`
	}
	if len(raw) == 0 {
		return ""
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	return payload.Type
}

func mapTurn(t Turn) conversation.Turn {
	out := conversation.Turn{
		ID:        t.ID,
		Status:    mapTurnStatus(t.Status),
		Timestamp: unixPtrOrZero(t.StartedAt, t.CompletedAt),
		// thread/read(includeTurns) is the only app-server turn-content
		// surface DIR-029/032 currently confirm; it always returns full
		// item content (no summary/unloaded projection is known to exist
		// yet — see docs/reference/codex-history-model.md's "Experimental
		// pagination" section), so every turn mapped here is Full. A future
		// confirmed partial-content surface should set a different
		// HistoryCompleteness value explicitly rather than defaulting here.
		Completeness: conversation.HistoryCompletenessFull,
	}
	for _, item := range t.Items {
		out.Items = append(out.Items, mapItems(item, out.Timestamp)...)
	}
	projectLegacyFields(&out)
	return out
}

func unixPtrOrZero(preferred, fallback *int64) time.Time {
	if preferred != nil {
		return time.Unix(*preferred, 0).UTC()
	}
	if fallback != nil {
		return time.Unix(*fallback, 0).UTC()
	}
	return time.Time{}
}

func mapTurnStatus(status string) conversation.TurnStatus {
	switch status {
	case "completed":
		return conversation.TurnStatusCompleted
	case "inProgress":
		return conversation.TurnStatusInProgress
	case "failed", "interrupted":
		return conversation.TurnStatusFailed
	default:
		return conversation.TurnStatusUnspecified
	}
}

// mapItems decodes one app-server ThreadItem into zero or more canonical
// Items. Most item types map 1:1; "commandExecution" and "mcpToolCall" map
// to a call+result pair (matching the Codex rollout adapter's ToolCall/
// ToolResult pairing convention) since the app-server reports them as a
// single already-resolved snapshot rather than separate open/close events.
//
// Coverage is intentionally scoped to the item types most useful for query
// (see docs/reference/codex-app-server.md): userMessage, agentMessage,
// commandExecution, fileChange, mcpToolCall, webSearch, reasoning, plan,
// and contextCompaction get dedicated mapping. Everything else (hookPrompt,
// dynamicToolCall, collabAgentToolCall, subAgentActivity, imageView, sleep,
// imageGeneration, enteredReviewMode, exitedReviewMode) is preserved
// losslessly via conversation.NewRawItem rather than dropped.
func mapItems(item ThreadItem, ts time.Time) []conversation.Item {
	switch item.Type {
	case "userMessage":
		var payload struct {
			ID      string `json:"id"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		}
		_ = json.Unmarshal(item.Raw, &payload)
		return []conversation.Item{{
			ID: payload.ID, Kind: conversation.ItemKindUserMessage, Role: "user",
			Text: joinTextParts(payload.Content), Timestamp: ts, Source: "app_server",
		}}
	case "agentMessage":
		var payload struct {
			ID    string  `json:"id"`
			Text  string  `json:"text"`
			Phase *string `json:"phase"`
		}
		_ = json.Unmarshal(item.Raw, &payload)
		return []conversation.Item{{
			ID: payload.ID, Kind: conversation.ItemKindAgentMessage, Role: "assistant",
			Text: payload.Text, Phase: mapPhase(payload.Phase), Timestamp: ts, Source: "app_server",
		}}
	case "commandExecution":
		var payload struct {
			ID               string `json:"id"`
			Command          string `json:"command"`
			Status           string `json:"status"`
			ExitCode         *int   `json:"exitCode"`
			AggregatedOutput string `json:"aggregatedOutput"`
		}
		_ = json.Unmarshal(item.Raw, &payload)
		return []conversation.Item{{
			ID: payload.ID, Kind: conversation.ItemKindCommandExecution,
			Command: payload.Command, ExitCode: payload.ExitCode,
			Status: mapItemStatus(payload.Status), Output: payload.AggregatedOutput,
			IsError: payload.Status == "failed", Timestamp: ts, Source: "app_server",
		}}
	case "fileChange":
		var payload struct {
			ID      string `json:"id"`
			Status  string `json:"status"`
			Changes []struct {
				Path string `json:"path"`
			} `json:"changes"`
		}
		_ = json.Unmarshal(item.Raw, &payload)
		paths := make([]string, 0, len(payload.Changes))
		for _, c := range payload.Changes {
			paths = append(paths, c.Path)
		}
		return []conversation.Item{{
			ID: payload.ID, Kind: conversation.ItemKindFileChange, Paths: paths,
			Status: mapItemStatus(payload.Status), Timestamp: ts, Source: "app_server",
		}}
	case "mcpToolCall":
		var payload struct {
			ID        string          `json:"id"`
			Server    string          `json:"server"`
			Tool      string          `json:"tool"`
			Arguments json.RawMessage `json:"arguments"`
			Status    string          `json:"status"`
			Result    json.RawMessage `json:"result"`
			Error     *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.Unmarshal(item.Raw, &payload)
		output := string(payload.Result)
		isError := payload.Status == "failed" || payload.Error != nil
		if payload.Error != nil {
			output = payload.Error.Message
		}
		name := payload.Tool
		if payload.Server != "" {
			name = fmt.Sprintf("%s.%s", payload.Server, payload.Tool)
		}
		return []conversation.Item{
			{
				ID: payload.ID, Kind: conversation.ItemKindToolCall, ToolCallID: payload.ID,
				ToolName: name, Input: payload.Arguments, Timestamp: ts, Source: "app_server",
			},
			{
				Kind: conversation.ItemKindToolResult, ToolCallID: payload.ID,
				Output: output, IsError: isError, Timestamp: ts, Source: "app_server",
			},
		}
	case "webSearch":
		var payload struct {
			ID    string `json:"id"`
			Query string `json:"query"`
		}
		_ = json.Unmarshal(item.Raw, &payload)
		return []conversation.Item{{
			ID: payload.ID, Kind: conversation.ItemKindWebSearch, Query: payload.Query,
			Timestamp: ts, Source: "app_server",
		}}
	case "reasoning":
		var payload struct {
			ID      string   `json:"id"`
			Content []string `json:"content"`
			Summary []string `json:"summary"`
		}
		_ = json.Unmarshal(item.Raw, &payload)
		text := strings.Join(append(append([]string{}, payload.Content...), payload.Summary...), "\n")
		return []conversation.Item{{
			ID: payload.ID, Kind: conversation.ItemKindReasoning, Text: text,
			Timestamp: ts, Source: "app_server",
		}}
	case "plan":
		var payload struct {
			ID   string `json:"id"`
			Text string `json:"text"`
		}
		_ = json.Unmarshal(item.Raw, &payload)
		return []conversation.Item{{
			ID: payload.ID, Kind: conversation.ItemKindPlanUpdate, PlanSteps: []string{payload.Text},
			Timestamp: ts, Source: "app_server",
		}}
	case "contextCompaction":
		// DIR-032: best-effort typed CompactionBoundary. reason/summary are
		// not yet confirmed against a live payload (see
		// docs/reference/codex-history-model.md), so both are optional and
		// this defensively decodes them if present rather than assuming the
		// shape — an absent field just yields a boundary with empty
		// Reason/Summary, still distinct from ItemKindUnknown.
		var payload struct {
			ID      string `json:"id"`
			Reason  string `json:"reason"`
			Summary string `json:"summary"`
		}
		_ = json.Unmarshal(item.Raw, &payload)
		return []conversation.Item{{
			ID: payload.ID, Kind: conversation.ItemKindCompaction, Timestamp: ts, Source: "app_server",
			Compaction: &conversation.CompactionBoundary{Reason: payload.Reason, Summary: payload.Summary},
		}}
	default:
		return []conversation.Item{conversation.NewRawItem(conversation.ItemKindUnknown, ts, item.Raw)}
	}
}

func joinTextParts(content []struct {
	Type string `json:"type"`
	Text string `json:"text"`
}) string {
	var parts []string
	for _, c := range content {
		if c.Text != "" {
			parts = append(parts, c.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func mapPhase(phase *string) conversation.AgentPhase {
	if phase == nil {
		return conversation.PhaseUnspecified
	}
	switch *phase {
	case "commentary":
		return conversation.PhaseCommentary
	case "final_answer":
		return conversation.PhaseFinal
	default:
		return conversation.PhaseUnspecified
	}
}

func mapItemStatus(status string) conversation.ItemStatus {
	switch status {
	case "completed":
		return conversation.StatusCompleted
	case "failed", "declined":
		return conversation.StatusFailed
	case "inProgress":
		return conversation.StatusInProgress
	default:
		return conversation.StatusUnspecified
	}
}

// projectLegacyFields derives Turn.UserText/AssistantText/ToolCalls from
// Items, mirroring internal/provider/codex/rollout.go's projection so both
// Codex backends (files and app_server) populate the same compatibility
// fields the same way. Duplicated rather than shared because appserver must
// not import the codex package (codex imports appserver, not vice versa).
func projectLegacyFields(t *conversation.Turn) {
	var userParts, assistantParts []string
	var calls []conversation.ToolCall
	callIndex := make(map[string]int)
	for _, item := range t.Items {
		switch item.Kind {
		case conversation.ItemKindUserMessage:
			userParts = append(userParts, item.Text)
		case conversation.ItemKindAgentMessage:
			assistantParts = append(assistantParts, item.Text)
		case conversation.ItemKindToolCall:
			callIndex[item.ToolCallID] = len(calls)
			calls = append(calls, conversation.ToolCall{
				ID: item.ToolCallID, Name: item.ToolName, Input: item.Input, Timestamp: item.Timestamp,
			})
		case conversation.ItemKindToolResult:
			if idx, ok := callIndex[item.ToolCallID]; ok {
				calls[idx].Output = item.Output
				calls[idx].IsError = item.IsError
			}
		}
	}
	if len(userParts) > 0 {
		t.UserText = strings.Join(userParts, "\n")
	}
	if len(assistantParts) > 0 {
		t.AssistantText = strings.Join(assistantParts, "\n")
	}
	t.ToolCalls = calls
}
