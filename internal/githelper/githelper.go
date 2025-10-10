// Package githelper provides utilities for analyzing git history and correlating
// it with Claude Code session activity.
package githelper

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Commit represents a git commit with metadata
type Commit struct {
	Hash         string
	Author       string
	Timestamp    time.Time
	Message      string
	FilesChanged int
	Insertions   int
	Deletions    int
}

// FileChange represents a file modification in a commit
type FileChange struct {
	FilePath   string
	Insertions int
	Deletions  int
	ChangeType string // "A" (added), "M" (modified), "D" (deleted), "R" (renamed)
}

// GetCommitsSince returns all commits since a given time
func GetCommitsSince(since time.Time) ([]Commit, error) {
	// git log --since="2024-01-01 00:00:00" --format="%H|%an|%at|%s" --numstat
	sinceStr := since.Format("2006-01-02 15:04:05")

	cmd := exec.Command("git", "log",
		fmt.Sprintf("--since=%s", sinceStr),
		"--format=%H|%an|%at|%s",
		"--numstat")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	return parseCommits(output)
}

// GetCommitsBetween returns commits in a time range
func GetCommitsBetween(start, end time.Time) ([]Commit, error) {
	startStr := start.Format("2006-01-02 15:04:05")
	endStr := end.Format("2006-01-02 15:04:05")

	cmd := exec.Command("git", "log",
		fmt.Sprintf("--since=%s", startStr),
		fmt.Sprintf("--until=%s", endStr),
		"--format=%H|%an|%at|%s",
		"--numstat")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	return parseCommits(output)
}

// GetChangedFiles returns files modified in a specific commit
func GetChangedFiles(commitHash string) ([]FileChange, error) {
	cmd := exec.Command("git", "show", "--format=", "--numstat", commitHash)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git show failed: %w", err)
	}

	return parseFileChanges(output)
}

// GetCommitMessage returns the full commit message for a commit
func GetCommitMessage(commitHash string) (string, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%B", commitHash)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetCommitAt returns the commit at or nearest to a specific timestamp
func GetCommitAt(timestamp time.Time) (*Commit, error) {
	// Get commit closest to timestamp (within ±5 minutes)
	before := timestamp.Add(-5 * time.Minute).Format("2006-01-02 15:04:05")
	after := timestamp.Add(5 * time.Minute).Format("2006-01-02 15:04:05")

	cmd := exec.Command("git", "log",
		fmt.Sprintf("--since=%s", before),
		fmt.Sprintf("--until=%s", after),
		"--format=%H|%an|%at|%s",
		"--numstat",
		"-1")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	commits, err := parseCommits(output)
	if err != nil {
		return nil, err
	}

	if len(commits) == 0 {
		return nil, nil
	}

	return &commits[0], nil
}

// IsGitRepository checks if current directory is a git repository
func IsGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// GetRecentCommitCount returns number of commits in the last N days
func GetRecentCommitCount(days int) (int, error) {
	since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	cmd := exec.Command("git", "rev-list", "--count", "--since="+since, "HEAD")

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("git rev-list failed: %w", err)
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, fmt.Errorf("failed to parse commit count: %w", err)
	}

	return count, nil
}

// GetFileChurnSince returns files sorted by total changes since a time
func GetFileChurnSince(since time.Time) (map[string]int, error) {
	sinceStr := since.Format("2006-01-02 15:04:05")

	cmd := exec.Command("git", "log",
		fmt.Sprintf("--since=%s", sinceStr),
		"--format=",
		"--numstat")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	churn := make(map[string]int)
	scanner := bufio.NewScanner(bytes.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		insertions, _ := strconv.Atoi(parts[0])
		deletions, _ := strconv.Atoi(parts[1])
		filepath := parts[2]

		churn[filepath] += insertions + deletions
	}

	return churn, nil
}

// parseCommits parses git log output with --numstat
func parseCommits(output []byte) ([]Commit, error) {
	var commits []Commit
	scanner := bufio.NewScanner(bytes.NewReader(output))

	var currentCommit *Commit

	for scanner.Scan() {
		line := scanner.Text()

		// Commit header line: hash|author|timestamp|message
		if strings.Contains(line, "|") {
			if currentCommit != nil {
				commits = append(commits, *currentCommit)
			}

			parts := strings.SplitN(line, "|", 4)
			if len(parts) < 4 {
				continue
			}

			timestamp, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				continue
			}

			currentCommit = &Commit{
				Hash:      parts[0],
				Author:    parts[1],
				Timestamp: time.Unix(timestamp, 0),
				Message:   parts[3],
			}
			continue
		}

		// Numstat line: insertions deletions filepath
		if currentCommit != nil && line != "" {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				insertions, _ := strconv.Atoi(parts[0])
				deletions, _ := strconv.Atoi(parts[1])

				currentCommit.FilesChanged++
				currentCommit.Insertions += insertions
				currentCommit.Deletions += deletions
			}
		}
	}

	// Add last commit
	if currentCommit != nil {
		commits = append(commits, *currentCommit)
	}

	return commits, nil
}

// parseFileChanges parses git show --numstat output
func parseFileChanges(output []byte) ([]FileChange, error) {
	var changes []FileChange
	scanner := bufio.NewScanner(bytes.NewReader(output))

	// Pattern: insertions deletions filepath
	numstatPattern := regexp.MustCompile(`^(\d+|-)\s+(\d+|-)\s+(.+)$`)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		matches := numstatPattern.FindStringSubmatch(line)
		if len(matches) != 4 {
			continue
		}

		insertions := 0
		deletions := 0

		if matches[1] != "-" {
			insertions, _ = strconv.Atoi(matches[1])
		}
		if matches[2] != "-" {
			deletions, _ = strconv.Atoi(matches[2])
		}

		filepath := matches[3]
		changeType := "M" // Default to modified

		// Detect change type from filepath patterns
		if strings.Contains(filepath, " => ") {
			changeType = "R" // Renamed
		}

		changes = append(changes, FileChange{
			FilePath:   filepath,
			Insertions: insertions,
			Deletions:  deletions,
			ChangeType: changeType,
		})
	}

	return changes, nil
}
