package query

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/itchyny/gojq"

	"github.com/yaleh/meta-cc/internal/locator"
	"github.com/yaleh/meta-cc/internal/parser"
	querycommon "github.com/yaleh/meta-cc/internal/query/common"
	"github.com/yaleh/meta-cc/internal/session"
)

// ParsedTimeRange specifies optional lower and upper bounds for timestamp filtering.
type ParsedTimeRange struct {
	Since *time.Time
	Until *time.Time
}

// ParseTimeRange parses since/until strings (RFC3339) into a ParsedTimeRange.
func ParseTimeRange(sinceStr, untilStr string) (ParsedTimeRange, error) {
	var tr ParsedTimeRange
	if sinceStr != "" {
		t, err := time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			return ParsedTimeRange{}, fmt.Errorf("invalid since value %q: must be RFC3339 (e.g. 2026-03-07T00:00:00Z)", sinceStr)
		}
		tr.Since = &t
	}
	if untilStr != "" {
		t, err := time.Parse(time.RFC3339, untilStr)
		if err != nil {
			return ParsedTimeRange{}, fmt.Errorf("invalid until value %q: must be RFC3339 (e.g. 2026-03-09T00:00:00Z)", untilStr)
		}
		tr.Until = &t
	}
	return tr, nil
}

// QueryRequest represents a query request
type QueryRequest struct {
	JQFilter    string
	JQTransform string
	Scope       string
	Limit       int
	SortBy      string
}

// QueryResponse represents query results
type QueryResponse struct {
	Entries []interface{}
}

// QueryResult represents query results with optional warnings about skipped files
type QueryResult struct {
	Entries     []interface{}
	Warnings    []string
	BypassStats bool // skip StatsOnly/StatsFirst processing; used by handlers that inject synthetic diagnostic entries
}

// ExpressionCache provides LRU caching for compiled jq expressions
type ExpressionCache struct {
	mu      sync.RWMutex
	Entries map[string]interface{} // stores *gojq.Code
	Keys    []string               // LRU tracking
	MaxSize int
}

// Get retrieves a cached expression
func (c *ExpressionCache) Get(expr string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Entries[expr]
}

// Put stores a compiled expression in cache with LRU eviction
func (c *ExpressionCache) Put(expr string, code interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.Entries[expr]; exists {
		c.removeKey(expr)
		c.Keys = append(c.Keys, expr)
		c.Entries[expr] = code
		return
	}

	if len(c.Entries) >= c.MaxSize {
		oldest := c.Keys[0]
		delete(c.Entries, oldest)
		c.Keys = c.Keys[1:]
	}

	c.Entries[expr] = code
	c.Keys = append(c.Keys, expr)
}

func (c *ExpressionCache) removeKey(key string) {
	for i, k := range c.Keys {
		if k == key {
			c.Keys = append(c.Keys[:i], c.Keys[i+1:]...)
			return
		}
	}
}

// QueryExecutor executes jq queries on JSONL session data with expression caching
type QueryExecutor struct {
	baseDir string
	Cache   *ExpressionCache
}

// NewQueryExecutor creates a new query executor
func NewQueryExecutor(baseDir string) *QueryExecutor {
	return &QueryExecutor{
		baseDir: baseDir,
		Cache: &ExpressionCache{
			Entries: make(map[string]interface{}),
			Keys:    []string{},
			MaxSize: 100,
		},
	}
}

// BuildExpression combines filter and transform into a single jq expression
func (e *QueryExecutor) BuildExpression(filter, transform string) string {
	if filter == "" {
		filter = "."
	}
	if transform != "" {
		return fmt.Sprintf("%s | %s", filter, transform)
	}
	return filter
}

// CompileExpression compiles a jq expression with caching
func (e *QueryExecutor) CompileExpression(expr string) (*gojq.Code, error) {
	if expr == "" {
		expr = "."
	}

	if cached := e.Cache.Get(expr); cached != nil {
		if code, ok := cached.(*gojq.Code); ok {
			return code, nil
		}
	}

	query, err := gojq.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid jq expression '%s': %w", expr, err)
	}

	code, err := gojq.Compile(query)
	if err != nil {
		return nil, fmt.Errorf("failed to compile jq expression '%s': %w", expr, err)
	}

	e.Cache.Put(expr, code)
	return code, nil
}

// StreamFiles processes multiple JSONL files with streaming
func (e *QueryExecutor) StreamFiles(ctx context.Context, files []string, code *gojq.Code, limit int) QueryResult {
	var result QueryResult

	for _, file := range files {
		select {
		case <-ctx.Done():
			return result
		default:
		}

		fileResults, err := e.ProcessFile(ctx, file, code)
		if err != nil {
			slog.Warn("skipping file due to read error", "file", file, "error", err)
			result.Warnings = append(result.Warnings, fmt.Sprintf("skipped %s: %s", file, err.Error()))
			continue
		}

		result.Entries = append(result.Entries, fileResults...)

		if limit > 0 && len(result.Entries) >= limit {
			result.Entries = result.Entries[:limit]
			return result
		}
	}

	return result
}

// StreamFilesWithTimeRange is like StreamFiles but applies TimeRange filtering before jq execution.
func (e *QueryExecutor) StreamFilesWithTimeRange(ctx context.Context, files []string, code *gojq.Code, limit int, tr ParsedTimeRange) QueryResult {
	var result QueryResult

	for _, file := range files {
		select {
		case <-ctx.Done():
			return result
		default:
		}

		fileResults, err := e.ProcessFileWithTimeRange(ctx, file, code, tr)
		if err != nil {
			slog.Warn("skipping file due to read error", "file", file, "error", err)
			result.Warnings = append(result.Warnings, fmt.Sprintf("skipped %s: %s", file, err.Error()))
			continue
		}

		result.Entries = append(result.Entries, fileResults...)

		if limit > 0 && len(result.Entries) >= limit {
			result.Entries = result.Entries[:limit]
			return result
		}
	}

	return result
}

// ProcessFile processes a single JSONL file
func (e *QueryExecutor) ProcessFile(ctx context.Context, filepath string, code *gojq.Code) ([]interface{}, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filepath, err)
	}
	defer file.Close()

	var results []interface{}
	r := bufio.NewReader(file)
	normalizer := session.NewNormalizer()

	for {
		select {
		case <-ctx.Done():
			return results, nil
		default:
		}

		rawLine, _, readErr := parser.ReadLineFiltered(r, parser.StrategyDefault)
		trimmed := bytes.TrimSpace(rawLine)
		if len(trimmed) == 0 {
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				return results, fmt.Errorf("error reading file %s: %w", filepath, readErr)
			}
			continue
		}

		var rawEntry map[string]interface{}
		if err := json.Unmarshal(trimmed, &rawEntry); err != nil {
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				return results, fmt.Errorf("error reading file %s: %w", filepath, readErr)
			}
			continue
		}

		for _, entry := range normalizer.NormalizeRecord(rawEntry) {
			iter := code.Run(entry)
			for {
				value, ok := iter.Next()
				if !ok {
					break
				}
				if err, ok := value.(error); ok {
					_ = err
					continue
				}
				results = append(results, value)
			}
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return results, fmt.Errorf("error reading file %s: %w", filepath, readErr)
		}
	}

	return results, nil
}

// ProcessFileWithTimeRange is like ProcessFile but filters by timestamp.
func (e *QueryExecutor) ProcessFileWithTimeRange(ctx context.Context, filepath string, code *gojq.Code, tr ParsedTimeRange) ([]interface{}, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filepath, err)
	}
	defer file.Close()

	var results []interface{}
	r := bufio.NewReader(file)
	normalizer := session.NewNormalizer()

	for {
		select {
		case <-ctx.Done():
			return results, nil
		default:
		}

		rawLine, _, readErr := parser.ReadLineFiltered(r, parser.StrategyDefault)
		trimmed := bytes.TrimSpace(rawLine)
		if len(trimmed) == 0 {
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				return results, fmt.Errorf("error reading file %s: %w", filepath, readErr)
			}
			continue
		}

		var rawEntry map[string]interface{}
		if err := json.Unmarshal(trimmed, &rawEntry); err != nil {
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				return results, fmt.Errorf("error reading file %s: %w", filepath, readErr)
			}
			continue
		}

		for _, entry := range normalizer.NormalizeRecord(rawEntry) {
			if tr.Since != nil || tr.Until != nil {
				if ts, ok := entry["timestamp"].(string); ok {
					if t, err := time.Parse(time.RFC3339, ts); err == nil {
						if tr.Since != nil && t.Before(*tr.Since) {
							continue
						}
						if tr.Until != nil && !t.Before(*tr.Until) {
							continue
						}
					}
				}
			}

			iter := code.Run(entry)
			for {
				value, ok := iter.Next()
				if !ok {
					break
				}
				if err, ok := value.(error); ok {
					_ = err
					continue
				}
				results = append(results, value)
			}
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return results, fmt.Errorf("error reading file %s: %w", filepath, readErr)
		}
	}

	return results, nil
}

// JQRunner executes jq queries against JSONL session files.
type JQRunner interface {
	RunQuery(ctx context.Context, files []string, filter, transform string, limit int) (QueryResult, error)
	RunQueryWithTimeRange(ctx context.Context, files []string, filter, transform string, limit int, tr ParsedTimeRange) (QueryResult, error)
}

// Ensure QueryExecutor implements JQRunner at compile time.
var _ JQRunner = (*QueryExecutor)(nil)

const transformAllNullWarning = "transform produced all-null fields: check your field paths (e.g. .timestamp not .Timestamp). Use inspect_session_files(include_samples=true) to see actual field names."

// RunQuery executes a jq query against the given JSONL files.
func (e *QueryExecutor) RunQuery(ctx context.Context, files []string, filter, transform string, limit int) (QueryResult, error) {
	expr := e.BuildExpression(filter, transform)
	code, err := e.CompileExpression(expr)
	if err != nil {
		return QueryResult{}, err
	}
	result := e.StreamFiles(ctx, files, code, limit)
	if transform != "" && querycommon.AllNullOrEmpty(result.Entries) {
		result.Warnings = append(result.Warnings, transformAllNullWarning)
	}
	return result, nil
}

// RunQueryWithTimeRange is like RunQuery but applies time range filtering before jq execution.
func (e *QueryExecutor) RunQueryWithTimeRange(ctx context.Context, files []string, filter, transform string, limit int, tr ParsedTimeRange) (QueryResult, error) {
	expr := e.BuildExpression(filter, transform)
	code, err := e.CompileExpression(expr)
	if err != nil {
		return QueryResult{}, err
	}
	result := e.StreamFilesWithTimeRange(ctx, files, code, limit, tr)
	if transform != "" && querycommon.AllNullOrEmpty(result.Entries) {
		result.Warnings = append(result.Warnings, transformAllNullWarning)
	}
	return result, nil
}

// GetQueryFiles returns the resolved list of JSONL file paths for the given scope
// and includeSubagents flag.
//
//   - scope=session, includeSubagents=false → [<current_session>.jsonl]
//   - scope=session, includeSubagents=true  → [<current_session>.jsonl] + <uuid>/subagents/*.jsonl
//   - scope=project, includeSubagents=false → all top-level *.jsonl (existing GetJSONLFiles behaviour)
//   - scope=project, includeSubagents=true  → top-level + all */subagents/*.jsonl
//
// Subagent scanning is fixed at two levels deep (<projectDir>/<uuid>/subagents/) to avoid
// picking up tool-results/ or other subdirectories.
func GetQueryFiles(scope, workingDir string, includeSubagents bool) ([]string, error) {
	projectPath := workingDir
	if projectPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "."
		}
		projectPath = cwd
	}

	loc := locator.NewSessionLocator()

	if scope == "session" {
		sessionFile, err := loc.FromProjectPath(projectPath)
		if err != nil {
			return nil, fmt.Errorf("failed to locate current session: %w", err)
		}
		files := []string{sessionFile}
		if includeSubagents {
			// Derive the subagents directory: <projectDir>/<uuid>/subagents/
			// sessionFile is <projectDir>/<uuid>.jsonl
			sessionDir := filepath.Dir(sessionFile)
			uuid := strings.TrimSuffix(filepath.Base(sessionFile), ".jsonl")
			subagentDir := filepath.Join(sessionDir, uuid, "subagents")
			subFiles, err := getSubagentJSONLFiles(subagentDir)
			if err == nil {
				files = append(files, subFiles...)
			}
		}
		return files, nil
	}

	// project scope
	baseDir, err := GetQueryBaseDir(scope, projectPath)
	if err != nil {
		return nil, err
	}

	topLevel, err := GetJSONLFiles(baseDir)
	if err != nil {
		return nil, err
	}

	if !includeSubagents {
		return topLevel, nil
	}

	// Scan <baseDir>/<entry>/subagents/*.jsonl for each directory entry
	all := make([]string, len(topLevel))
	copy(all, topLevel)

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return all, nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		subagentDir := filepath.Join(baseDir, entry.Name(), "subagents")
		subFiles, err := getSubagentJSONLFiles(subagentDir)
		if err != nil {
			continue
		}
		all = append(all, subFiles...)
	}

	return all, nil
}

// getSubagentJSONLFiles returns all .jsonl files in the given subagents directory.
// Returns nil without error if the directory does not exist.
func getSubagentJSONLFiles(subagentDir string) ([]string, error) {
	if _, err := os.Stat(subagentDir); os.IsNotExist(err) {
		return nil, nil
	}
	entries, err := os.ReadDir(subagentDir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".jsonl" {
			files = append(files, filepath.Join(subagentDir, entry.Name()))
		}
	}
	return files, nil
}

// GetQueryBaseDir returns the base directory for the given scope.
func GetQueryBaseDir(scope, workingDir string) (string, error) {
	projectPath := workingDir
	if projectPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "."
		}
		projectPath = cwd
	}

	loc := locator.NewSessionLocator()

	if scope == "session" {
		sessionFile, err := loc.FromProjectPath(projectPath)
		if err != nil {
			return "", fmt.Errorf("failed to locate current session: %w", err)
		}
		return filepath.Dir(sessionFile), nil
	}

	sessionFiles, err := loc.AllSessionsFromProject(projectPath)
	if err != nil {
		return "", fmt.Errorf("failed to locate project sessions: %w", err)
	}

	if len(sessionFiles) == 0 {
		return "", fmt.Errorf("no sessions found for project: %s", projectPath)
	}

	return filepath.Dir(sessionFiles[0]), nil
}

// GetJSONLFiles returns all .jsonl files in a directory sorted by modification time (newest first).
func GetJSONLFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	type fileInfo struct {
		path    string
		modTime int64
	}
	var fileInfos []fileInfo

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".jsonl" {
			fullPath := filepath.Join(dir, entry.Name())
			info, err := entry.Info()
			if err != nil {
				continue
			}
			fileInfos = append(fileInfos, fileInfo{path: fullPath, modTime: info.ModTime().Unix()})
		}
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].modTime > fileInfos[j].modTime
	})

	var files []string
	for _, fi := range fileInfos {
		files = append(files, fi.path)
	}
	return files, nil
}

// LoadTurnsForSession reads all JSONL files in baseDir and returns entries for sessionID.
func LoadTurnsForSession(baseDir, sessionID string) ([]interface{}, error) {
	files, err := GetJSONLFiles(baseDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		var turns []interface{}

		f, err := os.Open(file)
		if err != nil {
			continue
		}

		r := bufio.NewReader(f)
		normalizer := session.NewNormalizer()
		for {
			line, skipped, readErr := parser.ReadLineFiltered(r, parser.StrategyDefault)
			if skipped {
				if readErr == io.EOF {
					break
				}
				continue
			}
			if len(line) > 0 {
				lineStr := string(line)
				if len(lineStr) > 0 && lineStr[len(lineStr)-1] == '\n' {
					lineStr = lineStr[:len(lineStr)-1]
				}
				if lineStr != "" {
					var obj map[string]interface{}
					if err := json.Unmarshal([]byte(lineStr), &obj); err == nil {
						for _, normalized := range normalizer.NormalizeRecord(obj) {
							sid, _ := normalized["sessionId"].(string)
							if sid == sessionID {
								turns = append(turns, normalized)
							}
						}
					}
				}
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				break
			}
		}
		f.Close()

		if len(turns) > 0 {
			return turns, nil
		}
	}

	return nil, nil
}
