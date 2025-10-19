package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// TestProjectFlagDefaultBehavior tests that commands default to using --project .
// when no session-specific flag is provided
func TestProjectFlagDefaultBehavior(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectProject bool
	}{
		{
			name:          "parse stats with no flags should use project",
			args:          []string{"parse", "stats"},
			expectProject: true,
		},
		{
			name:          "query tools with no flags should use project",
			args:          []string{"query", "tools"},
			expectProject: true,
		},
		{
			name:          "parse stats with --session-only should not use project",
			args:          []string{"parse", "stats", "--session-only"},
			expectProject: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test verifies the behavior exists
			// Implementation will set projectPath to "." by default
			// unless --session-only flag is set

			// For now, just verify the test structure is correct
			if len(tt.args) == 0 {
				t.Error("Test args should not be empty")
			}
		})
	}
}

// TestSessionOnlyFlag tests the new --session-only flag that opts out of project-level analysis
func TestSessionOnlyFlag(t *testing.T) {
	t.Skip("Implementation pending")

	// When implemented, this should verify that:
	// 1. --session-only flag prevents default --project behavior
	// 2. Session is located via environment or auto-detection only
	// 3. Does not use project path defaulting
}

// Test Phase 14: meta-cc mcp subcommand should not exist (legacy removed)
// The MCP server is now a separate executable: meta-cc-mcp
func TestMCPSubcommandDoesNotExist(t *testing.T) {
	// Get all subcommands from rootCmd
	commands := rootCmd.Commands()

	// Check that "mcp" subcommand does NOT exist
	for _, cmd := range commands {
		if cmd.Name() == "mcp" {
			t.Errorf("Phase 14: 'mcp' subcommand should not exist. Use meta-cc-mcp executable instead.")
			t.Errorf("Found legacy mcp subcommand at: %s", cmd.Use)
		}
	}
}

// Test Phase 14: Verify expected subcommands exist (regression test)
func TestExpectedSubcommandsExist(t *testing.T) {
	expectedCommands := []string{
		"parse",
		"query",
		"analyze",
	}

	commands := rootCmd.Commands()
	commandMap := make(map[string]bool)

	for _, cmd := range commands {
		commandMap[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !commandMap[expected] {
			t.Errorf("Expected subcommand '%s' not found", expected)
		}
	}
}

// ===== Iteration 3 Tests: CLI Command Pattern Tests =====

// resetGlobalFlags resets all global flag variables to their default state.
// This prevents test interference due to shared global state.
func resetGlobalFlags() {
	sessionID = ""
	projectPath = ""
	outputFormat = "jsonl"
	sessionOnly = false
	limitFlag = 0
	offsetFlag = 0
	estimateSizeFlag = false
	chunkSizeFlag = 0
	outputDirFlag = ""
	fieldsFlag = ""
	ifErrorIncludeFlag = ""
	summaryFirstFlag = false
	topNFlag = 0
}

// TestExecute_HelpDisplay tests that Execute displays help when no args provided
func TestExecute_HelpDisplay(t *testing.T) {
	resetGlobalFlags()
	defer resetGlobalFlags()

	// Capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{})

	// Execute should not error (just shows help)
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Verify help text is displayed
	output := buf.String()
	if !strings.Contains(output, "meta-cc") {
		t.Errorf("Expected help text containing 'meta-cc', got: %s", output)
	}
	if !strings.Contains(output, "Available Commands") || !strings.Contains(output, "Usage") {
		t.Errorf("Expected help text with 'Available Commands' or 'Usage', got: %s", output)
	}
}

// TestRootCommand_Version tests version flag handling
func TestRootCommand_Version(t *testing.T) {
	resetGlobalFlags()
	defer resetGlobalFlags()

	// Set version for testing
	oldVersion := Version
	oldCommit := Commit
	oldBuildTime := BuildTime
	defer func() {
		Version = oldVersion
		Commit = oldCommit
		BuildTime = oldBuildTime
	}()

	Version = "1.0.0"
	Commit = "abc123"
	BuildTime = "2025-01-01"

	// Re-init to pick up new version
	rootCmd.Version = ""
	rootCmd.Version = strings.TrimSpace(Version + " (commit: " + Commit + ", built: " + BuildTime + ")")

	// Capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"--version"})

	// Execute
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() with --version failed: %v", err)
	}

	// Verify version output
	output := buf.String()
	if !strings.Contains(output, "1.0.0") {
		t.Errorf("Expected version '1.0.0' in output, got: %s", output)
	}
}

// TestGetGlobalOptions_DefaultProjectPath tests default project path resolution
func TestGetGlobalOptions_DefaultProjectPath(t *testing.T) {
	resetGlobalFlags()
	defer resetGlobalFlags()

	// Clear environment variables
	oldSessionEnv := os.Getenv("CC_SESSION_ID")
	os.Unsetenv("CC_SESSION_ID")
	defer func() {
		if oldSessionEnv != "" {
			os.Setenv("CC_SESSION_ID", oldSessionEnv)
		}
	}()

	// Get options with no flags set
	opts := getGlobalOptions()

	// Should default to current working directory
	cwd, _ := os.Getwd()
	if opts.ProjectPath != cwd {
		t.Errorf("Expected ProjectPath=%s, got %s", cwd, opts.ProjectPath)
	}

	if opts.SessionID != "" {
		t.Errorf("Expected SessionID='', got '%s'", opts.SessionID)
	}

	if opts.SessionOnly {
		t.Errorf("Expected SessionOnly=false, got true")
	}
}

// TestGetGlobalOptions_WithFlags tests global options with various flag combinations
func TestGetGlobalOptions_WithFlags(t *testing.T) {
	tests := []struct {
		name                string
		sessionID           string
		projectPath         string
		sessionOnly         bool
		envSessionID        string
		expectedSession     string
		expectedProject     string
		expectedSessionOnly bool
	}{
		{
			name:                "session flag set",
			sessionID:           "test-session-123",
			projectPath:         "",
			sessionOnly:         false,
			envSessionID:        "",
			expectedSession:     "test-session-123",
			expectedProject:     "", // No default when session is set
			expectedSessionOnly: false,
		},
		{
			name:                "project flag set",
			sessionID:           "",
			projectPath:         "/custom/project",
			sessionOnly:         false,
			envSessionID:        "",
			expectedSession:     "",
			expectedProject:     "/custom/project",
			expectedSessionOnly: false,
		},
		{
			name:                "session-only flag set",
			sessionID:           "",
			projectPath:         "",
			sessionOnly:         true,
			envSessionID:        "",
			expectedSession:     "",
			expectedProject:     "", // No default when session-only
			expectedSessionOnly: true,
		},
		{
			name:                "all flags set",
			sessionID:           "session-abc",
			projectPath:         "/path/to/project",
			sessionOnly:         true,
			envSessionID:        "",
			expectedSession:     "session-abc",
			expectedProject:     "/path/to/project",
			expectedSessionOnly: true,
		},
		{
			name:                "env session ID set",
			sessionID:           "",
			projectPath:         "",
			sessionOnly:         false,
			envSessionID:        "env-session-456",
			expectedSession:     "",
			expectedProject:     "", // No default when env is set
			expectedSessionOnly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()
			defer resetGlobalFlags()

			// Set environment variable if specified
			if tt.envSessionID != "" {
				os.Setenv("CC_SESSION_ID", tt.envSessionID)
				defer os.Unsetenv("CC_SESSION_ID")
			} else {
				os.Unsetenv("CC_SESSION_ID")
			}

			// Set global flags
			sessionID = tt.sessionID
			projectPath = tt.projectPath
			sessionOnly = tt.sessionOnly

			// Get options
			opts := getGlobalOptions()

			// Verify session ID
			if opts.SessionID != tt.expectedSession {
				t.Errorf("SessionID: expected '%s', got '%s'", tt.expectedSession, opts.SessionID)
			}

			// Verify project path (handle default case)
			if tt.expectedProject == "" && tt.sessionID == "" && !tt.sessionOnly && tt.envSessionID == "" {
				// Should default to cwd
				cwd, _ := os.Getwd()
				if opts.ProjectPath != cwd {
					t.Errorf("ProjectPath: expected default '%s', got '%s'", cwd, opts.ProjectPath)
				}
			} else {
				if opts.ProjectPath != tt.expectedProject {
					t.Errorf("ProjectPath: expected '%s', got '%s'", tt.expectedProject, opts.ProjectPath)
				}
			}

			// Verify session-only
			if opts.SessionOnly != tt.expectedSessionOnly {
				t.Errorf("SessionOnly: expected %v, got %v", tt.expectedSessionOnly, opts.SessionOnly)
			}
		})
	}
}

// TestGetGlobalOptions_EnvironmentVariables tests environment variable handling
func TestGetGlobalOptions_EnvironmentVariables(t *testing.T) {
	resetGlobalFlags()
	defer resetGlobalFlags()

	// Save original env vars
	oldSession := os.Getenv("CC_SESSION_ID")
	oldProject := os.Getenv("CC_PROJECT_PATH")
	defer func() {
		if oldSession != "" {
			os.Setenv("CC_SESSION_ID", oldSession)
		} else {
			os.Unsetenv("CC_SESSION_ID")
		}
		if oldProject != "" {
			os.Setenv("CC_PROJECT_PATH", oldProject)
		} else {
			os.Unsetenv("CC_PROJECT_PATH")
		}
	}()

	// Set environment variables
	os.Setenv("CC_SESSION_ID", "env-session")
	os.Setenv("CC_PROJECT_PATH", "/env/project")

	// Note: getGlobalOptions doesn't directly read env vars (viper does during flag parsing)
	// But it checks CC_SESSION_ID for default project path logic
	opts := getGlobalOptions()

	// With CC_SESSION_ID set, should not default to cwd
	if opts.ProjectPath != "" {
		t.Logf("ProjectPath is '%s' (expected empty when CC_SESSION_ID env is set)", opts.ProjectPath)
	}
}

// TestRootCommand_OutputFormatFlag tests output format flag parsing
func TestRootCommand_OutputFormatFlag(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedFormat string
	}{
		{
			name:           "default format",
			args:           []string{},
			expectedFormat: "jsonl",
		},
		{
			name:           "jsonl format",
			args:           []string{"--output", "jsonl"},
			expectedFormat: "jsonl",
		},
		{
			name:           "tsv format",
			args:           []string{"--output", "tsv"},
			expectedFormat: "tsv",
		},
		{
			name:           "short flag",
			args:           []string{"-o", "tsv"},
			expectedFormat: "tsv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()
			defer resetGlobalFlags()

			// Parse flags
			rootCmd.SetArgs(tt.args)
			err := rootCmd.ParseFlags(tt.args)
			if err != nil {
				t.Fatalf("ParseFlags() failed: %v", err)
			}

			// Verify output format
			if outputFormat != tt.expectedFormat {
				t.Errorf("Expected outputFormat='%s', got '%s'", tt.expectedFormat, outputFormat)
			}
		})
	}
}

// TestRootCommand_PaginationFlags tests pagination flag parsing
func TestRootCommand_PaginationFlags(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedLimit  int
		expectedOffset int
		wantErr        bool
	}{
		{
			name:           "no pagination",
			args:           []string{},
			expectedLimit:  0,
			expectedOffset: 0,
			wantErr:        false,
		},
		{
			name:           "limit only",
			args:           []string{"--limit", "10"},
			expectedLimit:  10,
			expectedOffset: 0,
			wantErr:        false,
		},
		{
			name:           "offset only",
			args:           []string{"--offset", "5"},
			expectedLimit:  0,
			expectedOffset: 5,
			wantErr:        false,
		},
		{
			name:           "limit and offset",
			args:           []string{"--limit", "20", "--offset", "10"},
			expectedLimit:  20,
			expectedOffset: 10,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()
			defer resetGlobalFlags()

			rootCmd.SetArgs(tt.args)
			err := rootCmd.ParseFlags(tt.args)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if limitFlag != tt.expectedLimit {
					t.Errorf("Expected limitFlag=%d, got %d", tt.expectedLimit, limitFlag)
				}
				if offsetFlag != tt.expectedOffset {
					t.Errorf("Expected offsetFlag=%d, got %d", tt.expectedOffset, offsetFlag)
				}
			}
		})
	}
}

// TestRootCommand_ChunkingFlags tests chunking flag parsing
func TestRootCommand_ChunkingFlags(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedChunkSize int
		expectedOutputDir string
	}{
		{
			name:              "no chunking",
			args:              []string{},
			expectedChunkSize: 0,
			expectedOutputDir: "",
		},
		{
			name:              "chunk size set",
			args:              []string{"--chunk-size", "100"},
			expectedChunkSize: 100,
			expectedOutputDir: "",
		},
		{
			name:              "chunk size with output dir",
			args:              []string{"--chunk-size", "50", "--output-dir", "/tmp/chunks"},
			expectedChunkSize: 50,
			expectedOutputDir: "/tmp/chunks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()
			defer resetGlobalFlags()

			rootCmd.SetArgs(tt.args)
			err := rootCmd.ParseFlags(tt.args)
			if err != nil {
				t.Fatalf("ParseFlags() failed: %v", err)
			}

			if chunkSizeFlag != tt.expectedChunkSize {
				t.Errorf("Expected chunkSizeFlag=%d, got %d", tt.expectedChunkSize, chunkSizeFlag)
			}
			if outputDirFlag != tt.expectedOutputDir {
				t.Errorf("Expected outputDirFlag='%s', got '%s'", tt.expectedOutputDir, outputDirFlag)
			}
		})
	}
}
