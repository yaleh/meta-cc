package claude

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
	"github.com/yaleh/meta-cc/internal/locator"
	"github.com/yaleh/meta-cc/internal/provider"
	"github.com/yaleh/meta-cc/internal/types"
)

var _ provider.Provider = (*Provider)(nil)

type Provider struct {
	locator    *locator.SessionLocator
	workingDir string
}

func NewProvider(loc *locator.SessionLocator, workingDir string) *Provider {
	return &Provider{locator: loc, workingDir: workingDir}
}

func (p *Provider) ID() conversation.ProviderID {
	return conversation.ProviderClaude
}

func (p *Provider) IsAvailable(context.Context) bool {
	if p.locator == nil {
		return false
	}
	root := os.Getenv("META_CC_PROJECTS_ROOT")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		root = filepath.Join(home, ".claude", "projects")
	}
	_, err := os.Stat(root)
	return err == nil
}

func (p *Provider) ListSessions(ctx context.Context) ([]conversation.Session, error) {
	_ = ctx
	projectPath := p.workingDir
	if projectPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		projectPath = cwd
	}

	files, err := p.locator.AllSessionsFromProject(projectPath)
	if err != nil {
		return nil, err
	}

	sessions := make([]conversation.Session, 0, len(files))
	for _, file := range files {
		session, err := p.sessionFromFile(file)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (p *Provider) GetSession(ctx context.Context, sessionID string) (conversation.Session, error) {
	_ = ctx
	file, err := p.findSessionFile(sessionID)
	if err != nil {
		return conversation.Session{}, err
	}
	return p.sessionFromFile(file)
}

func (p *Provider) LoadTurns(ctx context.Context, sessionID string) ([]conversation.Turn, error) {
	_ = ctx
	file, err := p.findSessionFile(sessionID)
	if err != nil {
		return nil, err
	}
	entries, err := parseClaudeEntries(file)
	if err != nil {
		return nil, err
	}

	pairs := buildTurns(entries)
	turns := make([]conversation.Turn, 0, len(pairs))
	for idx, pair := range pairs {
		ts := time.Time{}
		if pair.user != nil {
			ts, _ = time.Parse(time.RFC3339, pair.user.Timestamp)
		} else if pair.assistant != nil {
			ts, _ = time.Parse(time.RFC3339, pair.assistant.Timestamp)
		}
		ts = ts.UTC()
		id := fmt.Sprintf("%s-%d", sessionID, idx+1)
		userText := entryText(pair.user)
		assistantText := entryText(pair.assistant)
		calls := joinToolCalls(pair)
		status := conversation.TurnStatusInProgress
		if pair.assistant != nil {
			status = conversation.TurnStatusCompleted
		}
		turns = append(turns, conversation.Turn{
			ID:            id,
			Status:        status,
			UserText:      userText,
			AssistantText: assistantText,
			ToolCalls:     calls,
			Items:         itemsFromPair(id, userText, assistantText, calls, ts),
			Timestamp:     ts,
		})
	}
	return turns, nil
}

// findSessionFile resolves sessionID to its on-disk path. DIR-032:
// p.locator.FromSessionID is a GLOBAL search across every project-hash
// directory on disk, with no awareness of p.workingDir — used naively, it
// would let a working_dir scoped to project A resolve a session_id that
// actually belongs to project B (the same class of cross-project leak
// DIR-030's adversarial audit found and fixed on the
// provider_query.go/ExecuteQueryForSession path; this constructor-level
// GetSession/LoadTurns path had the identical gap, just not yet exercised
// by that audit — see query_sessions_handler.go and
// provider_query.go for the sibling fix this mirrors). A match from
// FromSessionID is only accepted when it falls under p.workingDir's own
// project-hash directory (via the same locator.PathToHash comparison
// ExecuteQueryForSession uses); otherwise this falls through to the
// already-scoped AllSessionsFromProject search, which naturally reports
// "not found" for a session belonging to a different project.
func (p *Provider) findSessionFile(sessionID string) (string, error) {
	if file, err := p.locator.FromSessionID(sessionID); err == nil && p.withinWorkingDirBoundary(file) {
		return file, nil
	}

	projectPath := p.workingDir
	if projectPath == "" {
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			return "", fmt.Errorf("session %q not found", sessionID)
		}
		projectPath = cwd
	}
	files, listErr := p.locator.AllSessionsFromProject(projectPath)
	if listErr != nil {
		return "", fmt.Errorf("session %q not found: %w", sessionID, listErr)
	}
	for _, candidate := range files {
		session, sessionErr := p.sessionFromFile(candidate)
		if sessionErr == nil && session.ID == sessionID {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("session %q not found", sessionID)
}

// withinWorkingDirBoundary reports whether file lives under the
// project-hash directory p.workingDir resolves to. An unset p.workingDir
// preserves the pre-DIR-032 unscoped-global-search behavior (no boundary
// configured, nothing to enforce), matching every other cwd-boundary check
// in this codebase (see ExecuteQueryForSession).
func (p *Provider) withinWorkingDirBoundary(file string) bool {
	if p.workingDir == "" {
		return true
	}
	projectPath := p.workingDir
	if abs, err := filepath.Abs(projectPath); err == nil {
		projectPath = abs
	}
	expectedHash := locator.PathToHash(projectPath)
	if expectedHash == "" {
		return true
	}
	return filepath.Base(filepath.Dir(file)) == expectedHash
}

// FilePath extracts the on-disk JSONL path recorded for a Claude session
// (stored in session.Extensions by sessionFromFile). Callers that need the
// raw file backing a Claude session — e.g. Stage 1 discovery tools — should
// use this instead of re-deriving the path themselves.
func FilePath(session conversation.Session) (string, error) {
	var ext struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(session.Extensions, &ext); err != nil {
		return "", err
	}
	if ext.Path == "" {
		return "", fmt.Errorf("missing path for session %s", session.ID)
	}
	return ext.Path, nil
}

func (p *Provider) sessionFromFile(file string) (conversation.Session, error) {
	entries, err := parseClaudeEntries(file)
	if err != nil {
		return conversation.Session{}, err
	}
	if len(entries) == 0 {
		return conversation.Session{}, fmt.Errorf("no Claude entries in %s", file)
	}

	first := entries[0]
	last := entries[len(entries)-1]
	createdAt, _ := time.Parse(time.RFC3339, first.Timestamp)
	tokenUsage := conversation.TokenUsage{}
	if last.Message != nil {
		tokenUsage = extractClaudeUsage(last.Message.Usage)
	}

	ext, _ := json.Marshal(map[string]string{"path": file})
	return conversation.Session{
		ID:         first.SessionID,
		Provider:   conversation.ProviderClaude,
		Title:      entryText(&first),
		CWD:        first.CWD,
		Model:      messageModel(last.Message),
		CreatedAt:  createdAt.UTC(),
		TokenUsage: tokenUsage,
		Extensions: ext,
	}, nil
}

func parseClaudeEntries(file string) ([]types.SessionEntry, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []types.SessionEntry
	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			var entry types.SessionEntry
			if jsonErr := json.Unmarshal(line, &entry); jsonErr == nil && entry.IsMessage() {
				entries = append(entries, entry)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return entries, nil
}

func extractClaudeUsage(usage map[string]interface{}) conversation.TokenUsage {
	return conversation.TokenUsage{
		InputTokens:  intFromMap(usage, "input_tokens"),
		OutputTokens: intFromMap(usage, "output_tokens"),
		CacheTokens:  intFromMap(usage, "cache_creation_input_tokens") + intFromMap(usage, "cache_read_input_tokens"),
	}
}

func intFromMap(values map[string]interface{}, key string) int {
	if values == nil {
		return 0
	}
	switch v := values[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return 0
	}
}

func messageModel(msg *types.Message) string {
	if msg == nil {
		return ""
	}
	return msg.Model
}
