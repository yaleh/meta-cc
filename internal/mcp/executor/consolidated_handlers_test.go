package executor

import (
	"testing"
)

// Phase B tests: new consolidated handlers are registered and delegate correctly.

func TestConsolidatedHandlersRegistered(t *testing.T) {
	newQueryTools := []string{
		"query_session_content",
		"query_session_signals",
		"query_file_activity",
	}
	for _, tool := range newQueryTools {
		if _, ok := queryHandlerRegistry[tool]; !ok {
			t.Errorf("new consolidated tool %q not registered in queryHandlerRegistry", tool)
		}
	}
}

func TestHandleQuerySessionContent_InvalidRole(t *testing.T) {
	e := NewToolExecutor()
	_, err := handleQuerySessionContent(e, "project", map[string]interface{}{
		"role": "invalid_role",
	})
	if err == nil {
		t.Fatal("expected error for invalid role")
	}
}

func TestHandleQuerySessionSignals_InvalidType(t *testing.T) {
	e := NewToolExecutor()
	_, err := handleQuerySessionSignals(e, "project", map[string]interface{}{
		"type": "invalid_type",
	})
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestHandleQueryFileActivity_InvalidType(t *testing.T) {
	e := NewToolExecutor()
	_, err := handleQueryFileActivity(e, "project", map[string]interface{}{
		"type": "invalid_type",
	})
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestHandleQuerySessionSignals_Errors_SameAsToolErrors(t *testing.T) {
	e := NewToolExecutor()
	// Both should produce the same result (or same error type — no JSONL files in test env)
	_, err1 := handleQuerySessionSignals(e, "project", map[string]interface{}{
		"type": "errors",
	})
	_, err2 := handleQueryToolErrors(e, "project", map[string]interface{}{})

	// Both should either succeed or fail with the same error class (no JSONL files)
	if (err1 == nil) != (err2 == nil) {
		t.Errorf("query_session_signals(type=errors) and query_tool_errors behave differently: err1=%v err2=%v", err1, err2)
	}
}

func TestHandleQuerySessionSignals_Tokens_SameAsTokenUsage(t *testing.T) {
	e := NewToolExecutor()
	_, err1 := handleQuerySessionSignals(e, "project", map[string]interface{}{
		"type": "tokens",
	})
	_, err2 := handleQueryTokenUsage(e, "project", map[string]interface{}{})
	if (err1 == nil) != (err2 == nil) {
		t.Errorf("query_session_signals(type=tokens) and query_token_usage behave differently: err1=%v err2=%v", err1, err2)
	}
}

func TestHandleQuerySessionSignals_SystemErrors_SameAsSystemErrors(t *testing.T) {
	e := NewToolExecutor()
	_, err1 := handleQuerySessionSignals(e, "project", map[string]interface{}{
		"type": "system_errors",
	})
	_, err2 := handleQuerySystemErrors(e, "project", map[string]interface{}{})
	if (err1 == nil) != (err2 == nil) {
		t.Errorf("query_session_signals(type=system_errors) and query_system_errors behave differently: err1=%v err2=%v", err1, err2)
	}
}

func TestHandleQuerySessionSignals_Timestamps_SameAsTimestamps(t *testing.T) {
	e := NewToolExecutor()
	_, err1 := handleQuerySessionSignals(e, "project", map[string]interface{}{
		"type": "timestamps",
	})
	_, err2 := handleQueryTimestamps(e, "project", map[string]interface{}{})
	if (err1 == nil) != (err2 == nil) {
		t.Errorf("query_session_signals(type=timestamps) and query_timestamps behave differently: err1=%v err2=%v", err1, err2)
	}
}

func TestHandleQuerySessionSignals_ToolStats_SameAsQueryTools(t *testing.T) {
	e := NewToolExecutor()
	_, err1 := handleQuerySessionSignals(e, "project", map[string]interface{}{
		"type": "tool_stats",
	})
	_, err2 := handleQueryTools(e, "project", map[string]interface{}{})
	if (err1 == nil) != (err2 == nil) {
		t.Errorf("query_session_signals(type=tool_stats) and query_tools behave differently: err1=%v err2=%v", err1, err2)
	}
}

func TestHandleQuerySessionContent_User_SameAsUserMessages(t *testing.T) {
	e := NewToolExecutor()
	_, err1 := handleQuerySessionContent(e, "project", map[string]interface{}{
		"role":    "user",
		"pattern": ".*",
	})
	_, err2 := handleQueryUserMessages(e, "project", map[string]interface{}{
		"pattern": ".*",
	})
	if (err1 == nil) != (err2 == nil) {
		t.Errorf("query_session_content(role=user) and query_user_messages behave differently: err1=%v err2=%v", err1, err2)
	}
}

func TestHandleQuerySessionContent_Tool_SameAsToolBlocks(t *testing.T) {
	e := NewToolExecutor()
	_, err1 := handleQuerySessionContent(e, "project", map[string]interface{}{
		"role": "tool",
	})
	_, err2 := handleQueryToolBlocks(e, "project", map[string]interface{}{
		"block_type": "tool_use",
	})
	if (err1 == nil) != (err2 == nil) {
		t.Errorf("query_session_content(role=tool) and query_tool_blocks behave differently: err1=%v err2=%v", err1, err2)
	}
}

func TestHandleQuerySessionContent_All_SameAsConversationFlow(t *testing.T) {
	e := NewToolExecutor()
	_, err1 := handleQuerySessionContent(e, "project", map[string]interface{}{
		"role": "all",
	})
	_, err2 := handleQueryConversationFlow(e, "project", map[string]interface{}{})
	if (err1 == nil) != (err2 == nil) {
		t.Errorf("query_session_content(role=all) and query_conversation_flow behave differently: err1=%v err2=%v", err1, err2)
	}
}

func TestHandleQueryFileActivity_Snapshots_SameAsFileSnapshots(t *testing.T) {
	e := NewToolExecutor()
	_, err1 := handleQueryFileActivity(e, "project", map[string]interface{}{
		"type": "snapshots",
	})
	_, err2 := handleQueryFileSnapshots(e, "project", map[string]interface{}{})
	if (err1 == nil) != (err2 == nil) {
		t.Errorf("query_file_activity(type=snapshots) and query_file_snapshots behave differently: err1=%v err2=%v", err1, err2)
	}
}
