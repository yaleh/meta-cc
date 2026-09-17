package locator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SubagentDirName is the directory Claude Code writes a session's subagent
// transcripts into: <projectDir>/<session-uuid>/subagents/. It is named here
// rather than spelled out at each call site because "is this file a subagent
// transcript?" (IsSubagentTranscript), "what is this transcript's id?"
// (SubagentIDFromPath) and "where do a project's subagents live?"
// (SubagentTranscriptsUnder) must all agree on the spelling — the
// gap-claude-listsessions-misses-subagents defect was exactly a disagreement
// about what the corpus is.
const SubagentDirName = "subagents"

// FromSessionID 通过会话 ID 查找会话文件
// 遍历支持的 transcript roots，查找匹配的 {session-id}.jsonl
// 如果找到多个（跨项目同名会话），返回最新的
//
// A subagent transcript is addressed by its AGENT id, which Claude Code
// encodes in the filename (agent-<agentId>.jsonl) rather than in the file's
// sessionId field — that field carries the PARENT session's uuid, so looking
// up an id the listing just produced would miss without the second search
// below. It reuses SubagentTranscriptsUnder/SubagentIDFromPath, the same pair
// the listing uses, so "which files are subagent transcripts, and what ids do
// they answer to" has exactly one definition.
func (l *SessionLocator) FromSessionID(sessionID string) (string, error) {
	var candidates []string
	sessionFilename := sessionID + ".jsonl"
	var checked []string

	for _, root := range l.TranscriptRoots() {
		checked = append(checked, formatRoot(root))
		if _, err := os.Stat(root.Path); os.IsNotExist(err) {
			continue
		}

		if root.ProjectHashed {
			projectDirs, err := os.ReadDir(root.Path)
			if err != nil {
				continue
			}
			for _, projectDir := range projectDirs {
				if !projectDir.IsDir() {
					continue
				}

				projectDirPath := filepath.Join(root.Path, projectDir.Name())
				sessionPath := filepath.Join(projectDirPath, sessionFilename)
				if _, err := os.Stat(sessionPath); err == nil {
					candidates = append(candidates, sessionPath)
				}

				for _, subagentFile := range SubagentTranscriptsUnder(projectDirPath) {
					if SubagentIDFromPath(subagentFile) == sessionID {
						candidates = append(candidates, subagentFile)
					}
				}
			}
			continue
		}

		sessionPath := filepath.Join(root.Path, sessionFilename)
		if _, err := os.Stat(sessionPath); err == nil {
			candidates = append(candidates, sessionPath)
			continue
		}

		matches, err := findSessionFilesRecursive(root.Path, sessionFilename)
		if err == nil {
			candidates = append(candidates, matches...)
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("session file not found for ID %q; checked transcript roots: %s",
			sessionID, strings.Join(checked, ", "))
	}

	// 如果找到多个，返回最新的
	return findNewestFile(candidates)
}

// FromSessionIDScoped resolves sessionID to a file path exactly like
// FromSessionID, but additionally enforces the caller's working_dir/cwd
// boundary before returning it.
//
// FromSessionID (above) is a GLOBAL search: it walks every project-hash
// directory on disk looking for a matching {session_id}.jsonl and returns
// whatever it finds, without ever comparing the result against the
// caller's working directory. Used directly, that shape is a cross-project
// leak: any caller who learns a session_id can read that session's content
// regardless of which project it claims to be scoped to.
//
// This exact defect was independently introduced and independently caught
// three separate times in this repo's history — internal/mcp/executor's
// ExecuteQueryForSession (DIR-030), internal/provider/claude's
// findSessionFile (found during the DIR-032 build), and
// internal/analysis/service.go's loadData (found by a DIR-032 adversarial
// audit AFTER the bug class was believed closed). Each fix hand-wrote the
// same boundary comparison: resolve workingDir to the project-hash
// directory name Claude Code itself uses (PathToHash) and reject the match
// if the resolved session file does not live under it. DIR-033
// crystallizes that one comparison here so no future caller has to
// remember to reimplement it — FromSessionID itself stays unscoped and is
// only ever called from within this package; every external caller must
// go through FromSessionIDScoped instead.
//
// An empty workingDir is a no-op (matches the pre-existing per-callsite
// behavior of skipping the boundary check when no project scope was
// requested/available).
func (l *SessionLocator) FromSessionIDScoped(sessionID, workingDir string) (string, error) {
	file, err := l.FromSessionID(sessionID)
	if err != nil {
		return "", fmt.Errorf("session_id %q not found: %w", sessionID, err)
	}

	boundaryDir := workingDir
	if abs, absErr := filepath.Abs(boundaryDir); absErr == nil {
		boundaryDir = abs
	}
	if expectedHash := PathToHash(boundaryDir); expectedHash != "" {
		actualHash := projectHashDirName(file)
		if actualHash != expectedHash {
			return "", fmt.Errorf("session_id %q not found for project %q", sessionID, boundaryDir)
		}
	}

	return file, nil
}

// projectHashDirName returns the name of the project-hash directory a resolved
// transcript file lives in. For a top-level session that is the file's own
// parent; for a subagent transcript (<hash>/<uuid>/subagents/agent-*.jsonl) it
// is three levels up.
//
// DIR-033's boundary check compares this against PathToHash(workingDir). The
// subagent depth has to be handled here rather than left to the caller: a
// directory literally named "subagents" never equals a project hash, so the
// naive parent-only comparison would reject exactly the subagent transcripts
// FromSessionID had just resolved — the boundary check would silently undo the
// enumeration fix for every session_id-taking tool.
func projectHashDirName(file string) string {
	dir := filepath.Dir(file)
	if filepath.Base(dir) == SubagentDirName {
		dir = filepath.Dir(filepath.Dir(dir))
	}
	return filepath.Base(dir)
}

// FromProjectPath 通过项目路径查找最新会话
// 1. 将项目路径转换为哈希（/ → -）
// 2. 定位 ~/.claude/projects/{hash}/
// 3. 返回该目录下最新的 .jsonl 文件
func (l *SessionLocator) FromProjectPath(projectPath string) (string, error) {
	// 解析相对路径为绝对路径（如 "." -> "/home/yale/work/meta-cc"）
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve project path: %w", err)
	}

	projectHash := PathToHash(absPath)

	sessions, err := l.sessionsFromProject(absPath, projectHash)
	if err != nil {
		return "", fmt.Errorf("no sessions found for project %q (hash: %s): %w",
			projectPath, projectHash, err)
	}

	return findNewestFile(sessions)
}

// AllSessionsFromProject 通过项目路径查找所有会话文件
// 1. 将项目路径转换为哈希（/ → -）
// 2. 定位 ~/.claude/projects/{hash}/
// 3. 返回该目录下所有 .jsonl 文件的路径
func (l *SessionLocator) AllSessionsFromProject(projectPath string) ([]string, error) {
	// 解析相对路径为绝对路径（如 "." -> "/home/yale/work/meta-cc"）
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project path: %w", err)
	}

	projectHash := PathToHash(absPath)

	sessions, err := l.sessionsFromProject(absPath, projectHash)
	if err != nil {
		return nil, fmt.Errorf("no sessions found for project %q (hash: %s): %w",
			projectPath, projectHash, err)
	}

	// 返回所有会话文件
	return sessions, nil
}

func (l *SessionLocator) sessionsFromProject(projectPath, projectHash string) ([]string, error) {
	var sessions []string
	var checked []string

	for _, root := range l.TranscriptRoots() {
		checked = append(checked, formatRoot(root))
		if !root.ProjectHashed {
			continue
		}
		if _, err := os.Stat(root.Path); os.IsNotExist(err) {
			continue
		}

		sessionDir := filepath.Join(root.Path, projectHash)
		rootSessions, err := filepath.Glob(filepath.Join(sessionDir, "*.jsonl"))
		if err != nil {
			return nil, fmt.Errorf("failed to search session files in %s: %w", sessionDir, err)
		}
		sessions = append(sessions, rootSessions...)
	}
	if len(sessions) > 0 {
		return sessions, nil
	}

	return nil, fmt.Errorf("checked transcript roots: %s", strings.Join(checked, ", "))
}

// AllTranscriptsFromProject returns a project's whole transcript corpus: the
// top-level {session-id}.jsonl files (what AllSessionsFromProject returns)
// PLUS every <uuid>/subagents/agent-*.jsonl subagent transcript beside them.
//
// This — not AllSessionsFromProject — is the corpus a session LISTING must
// enumerate. Subagent transcripts ARE sessions from every consumer's point of
// view: a real project on this host held 22 top-level transcripts and 2
// subagent ones, and a listing built on AllSessionsFromProject reported 22,
// making a subagent's own history unreachable through the documented
// "list, then query by session_id" route (gap-claude-listsessions-misses-subagents).
//
// AllSessionsFromProject deliberately keeps its narrower top-level-only meaning
// because callers use it to derive the project's session DIRECTORY
// (GetQueryBaseDir, stage.go) by taking filepath.Dir of the first result: a
// subagent file's directory is <uuid>/subagents, so widening that function
// would silently relocate those callers' base directory.
func (l *SessionLocator) AllTranscriptsFromProject(projectPath string) ([]string, error) {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project path: %w", err)
	}

	projectHash := PathToHash(absPath)

	// sessionsFromProject returns either a non-empty slice or an error, so the
	// pair below cannot both be empty with a nil error.
	topLevel, topErr := l.sessionsFromProject(absPath, projectHash)
	subagents := l.subagentTranscriptsForProject(projectHash)
	if len(topLevel)+len(subagents) == 0 {
		if topErr != nil {
			return nil, topErr
		}
		return nil, fmt.Errorf("no transcripts found for project %q (hash: %s)", projectPath, projectHash)
	}

	return append(topLevel, subagents...), nil
}

// subagentTranscriptsForProject collects the subagent transcripts of every
// project-hashed transcript root, for the project identified by projectHash.
func (l *SessionLocator) subagentTranscriptsForProject(projectHash string) []string {
	var subagents []string
	for _, root := range l.TranscriptRoots() {
		if !root.ProjectHashed {
			continue
		}
		if _, err := os.Stat(root.Path); os.IsNotExist(err) {
			continue
		}
		subagents = append(subagents,
			SubagentTranscriptsUnder(filepath.Join(root.Path, projectHash))...)
	}
	return subagents
}

// SubagentTranscriptsUnder returns every subagent transcript belonging to any
// session in the transcript directory baseDir (a project's hash directory):
// <baseDir>/<entry>/subagents/*.jsonl for each directory entry of baseDir.
//
// The walk is fixed at exactly that shape — one directory level, then a
// directory literally named "subagents" — which is what keeps sibling
// directories such as tool-results/ out of the corpus: a path is only ever
// accepted by descending through that exact name. This is the single
// definition of "where a project's subagent transcripts are"; the session
// listing and the query-file resolution both call it rather than each
// re-deriving the rule (two copies of "what is the corpus" is the defect class
// this task was filed against).
func SubagentTranscriptsUnder(baseDir string) []string {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		files = append(files,
			SubagentTranscripts(filepath.Join(baseDir, entry.Name(), SubagentDirName))...)
	}
	return files
}

// SubagentTranscripts returns the .jsonl transcripts in one subagents/
// directory. A directory that does not exist is not an error: most sessions
// never spawn a subagent.
func SubagentTranscripts(subagentsDir string) []string {
	entries, err := os.ReadDir(subagentsDir)
	if err != nil {
		return nil
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" {
			continue
		}
		files = append(files, filepath.Join(subagentsDir, entry.Name()))
	}
	return files
}

// IsSubagentTranscript reports whether path is a subagent transcript, i.e.
// whether it sits directly inside a subagents/ directory.
func IsSubagentTranscript(path string) bool {
	return filepath.Base(filepath.Dir(path)) == SubagentDirName
}

// SubagentIDFromPath returns the id a subagent transcript is addressed by —
// Claude Code names those files agent-<agentId>.jsonl, and that agent id (not
// the parent session uuid the file's entries carry in sessionId) is what
// distinguishes one subagent from another and what a caller can pass back as a
// session_id. Returns "" for a path that is not a subagent transcript.
func SubagentIDFromPath(path string) string {
	if !IsSubagentTranscript(path) {
		return ""
	}
	name := strings.TrimSuffix(filepath.Base(path), ".jsonl")
	return strings.TrimPrefix(name, "agent-")
}

func findSessionFilesRecursive(rootPath, filename string) ([]string, error) {
	var matches []string
	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == filename {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, errors.New("no matching session files")
	}
	return matches, nil
}

func formatRoot(root SessionRoot) string {
	if root.ProjectHashed {
		return fmt.Sprintf("%s=%s (project-hash)", root.Host, root.Path)
	}
	return fmt.Sprintf("%s=%s", root.Host, root.Path)
}

// PathToHash converts a project path to the hashed directory name used by Claude Code.
// Example: /home/yale/work/myproject → -home-yale-work-myproject
// Windows: C:/Users/yale/work/myproject → C--Users-yale-work-myproject
//
// Resolves symlinks for consistent hashing across platforms
// (e.g., /var → /private/var on macOS).
func PathToHash(path string) string {
	// Handle empty path edge case
	if path == "" {
		return ""
	}

	// Resolve symlinks for consistent hashing (e.g., /var -> /private/var on macOS)
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		// If resolution fails (e.g., path doesn't exist), use original path
		resolved = path
	}

	// Normalize path separators (both forward slash and backslash) to -
	// First replace backslashes (Windows paths)
	hash := strings.ReplaceAll(resolved, "\\", "-")
	// Then replace forward slashes (Unix paths and normalized Windows paths)
	hash = strings.ReplaceAll(hash, "/", "-")
	// Finally replace colons (Windows drive letters like C:)
	hash = strings.ReplaceAll(hash, ":", "-")
	return hash
}
