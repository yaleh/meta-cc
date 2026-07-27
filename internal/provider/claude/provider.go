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
		turns = append(turns, conversation.Turn{
			ID:            fmt.Sprintf("%s-%d", sessionID, idx+1),
			UserText:      entryText(pair.user),
			AssistantText: entryText(pair.assistant),
			ToolCalls:     joinToolCalls(pair),
			Timestamp:     ts.UTC(),
		})
	}
	return turns, nil
}

func (p *Provider) findSessionFile(sessionID string) (string, error) {
	file, err := p.locator.FromSessionID(sessionID)
	if err == nil {
		return file, nil
	}

	projectPath := p.workingDir
	if projectPath == "" {
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			return "", err
		}
		projectPath = cwd
	}
	files, listErr := p.locator.AllSessionsFromProject(projectPath)
	if listErr != nil {
		return "", err
	}
	for _, candidate := range files {
		session, sessionErr := p.sessionFromFile(candidate)
		if sessionErr == nil && session.ID == sessionID {
			return candidate, nil
		}
	}
	return "", err
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
