package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/itchyny/gojq"
)

// QueryExecutor executes jq queries on JSONL session data with expression caching
type QueryExecutor struct {
	baseDir string
	cache   *ExpressionCache
}

// ExpressionCache provides LRU caching for compiled jq expressions
type ExpressionCache struct {
	mu      sync.RWMutex
	entries map[string]interface{} // stores *gojq.Code
	keys    []string               // LRU tracking
	maxSize int
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

// NewQueryExecutor creates a new query executor
func NewQueryExecutor(baseDir string) *QueryExecutor {
	return &QueryExecutor{
		baseDir: baseDir,
		cache: &ExpressionCache{
			entries: make(map[string]interface{}),
			keys:    []string{},
			maxSize: 100,
		},
	}
}

// buildExpression combines filter and transform into a single jq expression
func (e *QueryExecutor) buildExpression(filter, transform string) string {
	// Default to identity filter
	if filter == "" {
		filter = "."
	}

	// If transform is provided, pipe it
	if transform != "" {
		return fmt.Sprintf("%s | %s", filter, transform)
	}

	return filter
}

// compileExpression compiles a jq expression with caching
func (e *QueryExecutor) compileExpression(expr string) (*gojq.Code, error) {
	// Normalize empty expression to identity
	if expr == "" {
		expr = "."
	}

	// Check cache first
	if cached := e.cache.Get(expr); cached != nil {
		if code, ok := cached.(*gojq.Code); ok {
			return code, nil
		}
	}

	// Parse and compile expression
	query, err := gojq.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid jq expression '%s': %w", expr, err)
	}

	// Compile to bytecode
	code, err := gojq.Compile(query)
	if err != nil {
		return nil, fmt.Errorf("failed to compile jq expression '%s': %w", expr, err)
	}

	// Cache the compiled code
	e.cache.Put(expr, code)

	return code, nil
}

// streamFiles processes multiple JSONL files with streaming
func (e *QueryExecutor) streamFiles(ctx context.Context, files []string, code *gojq.Code, limit int) []interface{} {
	var results []interface{}

	for _, file := range files {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return results
		default:
		}

		fileResults, err := e.processFile(ctx, file, code)
		if err != nil {
			// Log error but continue processing other files
			continue
		}

		results = append(results, fileResults...)

		// Check limit
		if limit > 0 && len(results) >= limit {
			return results[:limit]
		}
	}

	return results
}

// processFile processes a single JSONL file
func (e *QueryExecutor) processFile(ctx context.Context, filepath string, code *gojq.Code) ([]interface{}, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filepath, err)
	}
	defer file.Close()

	var results []interface{}
	scanner := bufio.NewScanner(file)

	// Increase buffer size for large lines
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	lineNum := 0
	for scanner.Scan() {
		lineNum++

		// Check context cancellation
		select {
		case <-ctx.Done():
			return results, nil
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse JSON line
		var entry interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			// Skip invalid JSON lines (don't fail entire file)
			continue
		}

		// Execute jq query on this entry
		iter := code.Run(entry)
		for {
			value, ok := iter.Next()
			if !ok {
				break
			}

			// Check for errors
			if err, ok := value.(error); ok {
				// Skip entries that cause jq errors
				_ = err
				continue
			}

			results = append(results, value)
		}
	}

	if err := scanner.Err(); err != nil {
		return results, fmt.Errorf("error reading file %s: %w", filepath, err)
	}

	return results, nil
}

// Get retrieves a cached expression
func (c *ExpressionCache) Get(expr string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.entries[expr]
}

// Put stores a compiled expression in cache with LRU eviction
func (c *ExpressionCache) Put(expr string, code interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if entry already exists
	if _, exists := c.entries[expr]; exists {
		// Update existing entry (move to end for LRU)
		c.removeKey(expr)
		c.keys = append(c.keys, expr)
		c.entries[expr] = code
		return
	}

	// LRU eviction if cache is full
	if len(c.entries) >= c.maxSize {
		oldest := c.keys[0]
		delete(c.entries, oldest)
		c.keys = c.keys[1:]
	}

	// Add new entry
	c.entries[expr] = code
	c.keys = append(c.keys, expr)
}

// removeKey removes a key from the keys slice (helper for LRU)
func (c *ExpressionCache) removeKey(key string) {
	for i, k := range c.keys {
		if k == key {
			c.keys = append(c.keys[:i], c.keys[i+1:]...)
			return
		}
	}
}
