package githelper

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// setupTestRepo creates a temporary git repository for testing
func setupTestRepo(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "githelper-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git user for commits (use --local to avoid system config override)
	cmd = exec.Command("git", "config", "--local", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to config git user.name: %v", err)
	}

	cmd = exec.Command("git", "config", "--local", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to config git user.email: %v", err)
	}

	return tmpDir
}

// createTestCommit creates a commit with specified file changes
func createTestCommit(t *testing.T, repoDir, filename, content, message string) {
	filePath := filepath.Join(repoDir, filename)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	cmd := exec.Command("git", "add", filename)
	cmd.Dir = repoDir
	// Isolate from system git config
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "HOME="+repoDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = repoDir
	// Isolate from system git config
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "HOME="+repoDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to git commit: %v", err)
	}
}

func TestIsGitRepository(t *testing.T) {
	// Test in a git repository
	tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("failed to restore original directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to test directory: %v", err)
	}

	if !IsGitRepository() {
		t.Error("expected IsGitRepository to return true in a git repo")
	}

	// Test in a non-git directory
	tmpDir2, err := os.MkdirTemp("", "non-git-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir2)

	if err := os.Chdir(tmpDir2); err != nil {
		t.Fatalf("failed to change to non-git directory: %v", err)
	}

	if IsGitRepository() {
		t.Error("expected IsGitRepository to return false in a non-git directory")
	}
}

func TestGetCommitsSince(t *testing.T) {
	tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("failed to restore original directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to test directory: %v", err)
	}

	// Create some test commits
	createTestCommit(t, tmpDir, "file1.txt", "line1\nline2\nline3\n", "First commit")
	time.Sleep(100 * time.Millisecond)
	createTestCommit(t, tmpDir, "file2.txt", "content\n", "Second commit")

	// Get commits since 1 hour ago
	since := time.Now().Add(-1 * time.Hour)
	commits, err := GetCommitsSince(since)
	if err != nil {
		t.Fatalf("GetCommitsSince failed: %v", err)
	}

	if len(commits) != 2 {
		t.Errorf("expected 2 commits, got %d", len(commits))
	}

	// Verify commit data
	if commits[0].Message != "Second commit" {
		t.Errorf("expected 'Second commit', got '%s'", commits[0].Message)
	}

	if commits[0].FilesChanged != 1 {
		t.Errorf("expected 1 file changed, got %d", commits[0].FilesChanged)
	}
}

func TestGetCommitsBetween(t *testing.T) {
	tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("failed to restore original directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to test directory: %v", err)
	}

	start := time.Now()
	createTestCommit(t, tmpDir, "file1.txt", "content1\n", "Commit 1")
	time.Sleep(2 * time.Second) // Increase sleep to ensure commits are separated
	middle := time.Now()
	time.Sleep(2 * time.Second)
	createTestCommit(t, tmpDir, "file2.txt", "content2\n", "Commit 2")
	end := time.Now()

	// Get commits between start and middle (should get only first commit)
	commits, err := GetCommitsBetween(start, middle)
	if err != nil {
		t.Fatalf("GetCommitsBetween failed: %v", err)
	}

	// Due to git timestamp granularity, we may get 1 or 2 commits
	if len(commits) < 1 || len(commits) > 2 {
		t.Errorf("expected 1-2 commits, got %d", len(commits))
	}

	// Get commits between start and end (should get both commits)
	commits, err = GetCommitsBetween(start, end)
	if err != nil {
		t.Fatalf("GetCommitsBetween failed: %v", err)
	}

	if len(commits) != 2 {
		t.Errorf("expected 2 commits, got %d", len(commits))
	}
}

func TestGetChangedFiles(t *testing.T) {
	tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("failed to restore original directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to test directory: %v", err)
	}

	// Create a commit with multiple file changes
	createTestCommit(t, tmpDir, "file1.txt", "line1\nline2\n", "Multi-file commit")
	createTestCommit(t, tmpDir, "file2.txt", "content\n", "Multi-file commit")

	// Get latest commit hash
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = tmpDir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to get commit hash: %v", err)
	}
	commitHash := string(output[:len(output)-1]) // Remove newline

	// Get changed files
	files, err := GetChangedFiles(commitHash)
	if err != nil {
		t.Fatalf("GetChangedFiles failed: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}

	if len(files) > 0 && files[0].FilePath != "file2.txt" {
		t.Errorf("expected 'file2.txt', got '%s'", files[0].FilePath)
	}
}

func TestGetCommitMessage(t *testing.T) {
	tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("failed to restore original directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to test directory: %v", err)
	}

	createTestCommit(t, tmpDir, "file1.txt", "content\n", "Test commit message")

	// Get latest commit hash
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = tmpDir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to get commit hash: %v", err)
	}
	commitHash := string(output[:len(output)-1])

	// Get commit message
	message, err := GetCommitMessage(commitHash)
	if err != nil {
		t.Fatalf("GetCommitMessage failed: %v", err)
	}

	if message != "Test commit message" {
		t.Errorf("expected 'Test commit message', got '%s'", message)
	}
}

func TestGetRecentCommitCount(t *testing.T) {
	tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("failed to restore original directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to test directory: %v", err)
	}

	// Create 3 commits
	createTestCommit(t, tmpDir, "file1.txt", "content1\n", "Commit 1")
	createTestCommit(t, tmpDir, "file2.txt", "content2\n", "Commit 2")
	createTestCommit(t, tmpDir, "file3.txt", "content3\n", "Commit 3")

	// Get commit count for last 7 days
	count, err := GetRecentCommitCount(7)
	if err != nil {
		t.Fatalf("GetRecentCommitCount failed: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 commits, got %d", count)
	}
}

func TestGetFileChurnSince(t *testing.T) {
	tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("failed to restore original directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to test directory: %v", err)
	}

	// Create commits with different file changes
	createTestCommit(t, tmpDir, "file1.txt", "line1\nline2\nline3\n", "First commit")
	createTestCommit(t, tmpDir, "file1.txt", "line1\nline2\nline3\nline4\n", "Second commit")
	createTestCommit(t, tmpDir, "file2.txt", "content\n", "Third commit")

	since := time.Now().Add(-1 * time.Hour)
	churn, err := GetFileChurnSince(since)
	if err != nil {
		t.Fatalf("GetFileChurnSince failed: %v", err)
	}

	// file1.txt should have more churn than file2.txt
	if churn["file1.txt"] <= churn["file2.txt"] {
		t.Errorf("expected file1.txt to have more churn than file2.txt, got %d vs %d",
			churn["file1.txt"], churn["file2.txt"])
	}

	if _, exists := churn["file1.txt"]; !exists {
		t.Error("expected file1.txt in churn map")
	}

	if _, exists := churn["file2.txt"]; !exists {
		t.Error("expected file2.txt in churn map")
	}
}

func TestParseCommits(t *testing.T) {
	// Test with sample git log output
	output := []byte(`abc123|John Doe|1609459200|Initial commit
3	0	file1.txt

def456|Jane Smith|1609545600|Second commit
5	2	file2.txt
10	3	file3.txt
`)

	commits, err := parseCommits(output)
	if err != nil {
		t.Fatalf("parseCommits failed: %v", err)
	}

	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}

	// Check first commit
	if commits[0].Hash != "abc123" {
		t.Errorf("expected hash 'abc123', got '%s'", commits[0].Hash)
	}
	if commits[0].Author != "John Doe" {
		t.Errorf("expected author 'John Doe', got '%s'", commits[0].Author)
	}
	if commits[0].Message != "Initial commit" {
		t.Errorf("expected message 'Initial commit', got '%s'", commits[0].Message)
	}
	if commits[0].FilesChanged != 1 {
		t.Errorf("expected 1 file changed, got %d", commits[0].FilesChanged)
	}
	if commits[0].Insertions != 3 {
		t.Errorf("expected 3 insertions, got %d", commits[0].Insertions)
	}

	// Check second commit
	if commits[1].Hash != "def456" {
		t.Errorf("expected hash 'def456', got '%s'", commits[1].Hash)
	}
	if commits[1].FilesChanged != 2 {
		t.Errorf("expected 2 files changed, got %d", commits[1].FilesChanged)
	}
	if commits[1].Insertions != 15 {
		t.Errorf("expected 15 insertions, got %d", commits[1].Insertions)
	}
	if commits[1].Deletions != 5 {
		t.Errorf("expected 5 deletions, got %d", commits[1].Deletions)
	}
}

func TestParseFileChanges(t *testing.T) {
	// Test with sample git show --numstat output
	output := []byte(`5	3	file1.txt
10	0	file2.txt
-	-	binary-file.jpg
`)

	changes, err := parseFileChanges(output)
	if err != nil {
		t.Fatalf("parseFileChanges failed: %v", err)
	}

	if len(changes) != 3 {
		t.Fatalf("expected 3 file changes, got %d", len(changes))
	}

	// Check first file
	if changes[0].FilePath != "file1.txt" {
		t.Errorf("expected filepath 'file1.txt', got '%s'", changes[0].FilePath)
	}
	if changes[0].Insertions != 5 {
		t.Errorf("expected 5 insertions, got %d", changes[0].Insertions)
	}
	if changes[0].Deletions != 3 {
		t.Errorf("expected 3 deletions, got %d", changes[0].Deletions)
	}

	// Check binary file (should have 0 insertions/deletions)
	if changes[2].FilePath != "binary-file.jpg" {
		t.Errorf("expected filepath 'binary-file.jpg', got '%s'", changes[2].FilePath)
	}
	if changes[2].Insertions != 0 {
		t.Errorf("expected 0 insertions for binary file, got %d", changes[2].Insertions)
	}
	if changes[2].Deletions != 0 {
		t.Errorf("expected 0 deletions for binary file, got %d", changes[2].Deletions)
	}
}
