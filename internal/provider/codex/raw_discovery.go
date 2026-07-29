package codex

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
)

// discoverRolloutSessions is the metadata-light fallback used when no
// compatible state_N.sqlite exists. It reads only the first session_meta line
// from each rollout and never infers a cwd, preventing cross-project leakage.
func discoverRolloutSessions(roots []rolloutRoot, filter conversation.SessionFilter) ([]conversation.Session, []string, error) {
	var sessions []conversation.Session
	var warnings []string
	for _, root := range roots {
		err := filepath.WalkDir(root.path, func(path string, entry os.DirEntry, err error) error {
			if os.IsNotExist(err) {
				return nil
			}
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("codex rollout fallback: cannot inspect %s: %v", path, err))
				return nil
			}
			if entry.IsDir() || filepath.Ext(path) != ".jsonl" {
				return nil
			}
			session, err := rolloutMetadata(path, root.archived)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("codex rollout fallback: skipped %s: %v", path, err))
				return nil
			}
			if filter.CWD != "" && session.CWD != filter.CWD {
				return nil
			}
			sessions = append(sessions, session)
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			return nil, warnings, err
		}
	}
	return conversation.ApplyFilter(sessions, filter), warnings, nil
}

type rolloutRoot struct {
	path     string
	archived bool
}

func rolloutMetadata(path string, archived bool) (conversation.Session, error) {
	f, err := os.Open(path)
	if err != nil {
		return conversation.Session{}, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var event struct {
			Timestamp string `json:"timestamp"`
			Type      string `json:"type"`
			Payload   struct {
				ID            string `json:"id"`
				CWD           string `json:"cwd"`
				Model         string `json:"model"`
				ModelProvider string `json:"model_provider"`
				Source        string `json:"source"`
			} `json:"payload"`
		}
		if json.Unmarshal(scanner.Bytes(), &event) != nil || event.Type != "session_meta" {
			continue
		}
		if event.Payload.ID == "" || event.Payload.CWD == "" {
			return conversation.Session{}, fmt.Errorf("session_meta lacks id or cwd")
		}
		stamp, _ := time.Parse(time.RFC3339, event.Timestamp)
		ext, _ := json.Marshal(map[string]string{"rollout_path": path, "source": firstNonEmpty(event.Payload.Source, "cli"), "model_provider": event.Payload.ModelProvider})
		status := "active"
		if archived {
			status = "archived"
		}
		return conversation.Session{
			ID: event.Payload.ID, Provider: conversation.ProviderCodex,
			Title: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)), CWD: event.Payload.CWD,
			Model: firstNonEmpty(event.Payload.Model, event.Payload.ModelProvider), ModelProvider: event.Payload.ModelProvider,
			SourceKind: firstNonEmpty(event.Payload.Source, "cli"), Status: status, Archived: archived,
			CreatedAt: stamp.UTC(), UpdatedAt: stamp.UTC(), Lineage: conversation.LineageStatusUnknown, Extensions: ext,
		}, nil
	}
	if err := scanner.Err(); err != nil {
		return conversation.Session{}, err
	}
	return conversation.Session{}, fmt.Errorf("no usable session_meta")
}
