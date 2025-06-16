package cli

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/testutil"
)

// testLogger is a minimal logger implementation for testing
type testLogger struct{}

func (tl *testLogger) Debug(format string, args ...interface{})                             {}
func (tl *testLogger) Info(format string, args ...interface{})                              {}
func (tl *testLogger) Warn(format string, args ...interface{})                              {}
func (tl *testLogger) Error(format string, args ...interface{})                             {}
func (tl *testLogger) Fatal(format string, args ...interface{})                             {}
func (tl *testLogger) Printf(format string, args ...interface{})                            {}
func (tl *testLogger) Println(args ...interface{})                                          {}
func (tl *testLogger) DebugContext(ctx context.Context, format string, args ...interface{}) {}
func (tl *testLogger) InfoContext(ctx context.Context, format string, args ...interface{})  {}
func (tl *testLogger) WarnContext(ctx context.Context, format string, args ...interface{})  {}
func (tl *testLogger) ErrorContext(ctx context.Context, format string, args ...interface{}) {}
func (tl *testLogger) FatalContext(ctx context.Context, format string, args ...interface{}) {}
func (tl *testLogger) WithContext(ctx context.Context) logutil.LoggerInterface              { return tl }

// cliTestRunner captures all outputs from a simulated CLI execution.
type cliTestRunner struct {
	stdout  string
	stderr  string
	logFile string
}

// runCliTest executes a simulated CLI run with specified configurations.
func runCliTest(t *testing.T, args []string, env map[string]string, isTTY bool) *cliTestRunner {
	t.Helper()
	// --- Capture stdout and stderr ---
	var outBuf, errBuf bytes.Buffer
	var wg sync.WaitGroup
	oldStdout, oldStderr := os.Stdout, os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout, os.Stderr = wOut, wErr

	// Use channels to safely communicate captured output
	stdoutChan := make(chan string, 1)
	stderrChan := make(chan string, 1)

	wg.Add(2)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				t.Logf("Error copying stdout: %v", err)
			}
			wg.Done()
		}()
		_, _ = io.Copy(&outBuf, rOut)
		stdoutChan <- outBuf.String()
	}()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				t.Logf("Error copying stderr: %v", err)
			}
			wg.Done()
		}()
		_, _ = io.Copy(&errBuf, rErr)
		stderrChan <- errBuf.String()
	}()

	// Cleanup function to properly close pipes and collect output
	cleanup := func() (string, string) {
		_ = wOut.Close()
		_ = wErr.Close()
		wg.Wait()
		stdout := <-stdoutChan
		stderr := <-stderrChan
		os.Stdout, os.Stderr = oldStdout, oldStderr
		return stdout, stderr
	}

	// --- Setup Test Environment ---
	tempDir := testutil.SetupTempDir(t, "clitest-")
	// Set required API key environment variable
	t.Setenv("OPENAI_API_KEY", "test-key-123")
	for key, val := range env {
		t.Setenv(key, val)
	}

	// --- Execute Mock Application Logic ---
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg, err := ParseFlagsWithEnv(flagSet, args, os.Getenv)
	if err != nil {
		// For early return, manually cleanup and capture output
		_, stderr := cleanup()
		return &cliTestRunner{stderr: stderr}
	}
	cfg.OutputDir = tempDir
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	logger := SetupLogging(cfg)
	consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return isTTY },
		// Note: CI detection is done automatically via environment variables
	})
	consoleWriter.SetQuiet(cfg.Quiet)
	consoleWriter.SetNoProgress(cfg.NoProgress)

	logger.Info("Simulating application start")
	consoleWriter.StartProcessing(2)
	consoleWriter.ModelCompleted("model-1", 1, 800*time.Millisecond, nil)
	consoleWriter.ModelCompleted("model-2", 2, 1200*time.Millisecond, errors.New("simulated error"))
	logger.Error("An error occurred", "err", "simulated error")

	// Force a flush to ensure all output is captured
	if f, ok := logger.(interface{ Sync() error }); ok {
		_ = f.Sync()
	}

	// Small delay to ensure all goroutines complete writing
	time.Sleep(10 * time.Millisecond)

	// --- Read log file if it was created ---
	var logFileContent string
	logFilePath := filepath.Join(tempDir, "thinktank.log")
	if content, err := os.ReadFile(logFilePath); err == nil {
		logFileContent = string(content)
	}

	// Capture final output
	stdout, stderr := cleanup()
	return &cliTestRunner{
		stdout:  stdout,
		stderr:  stderr,
		logFile: logFileContent,
	}
}

func TestCliLoggingCombinations(t *testing.T) {
	baseArgs := []string{"--instructions", "test.txt", "--model", "test-model", "test-path"}
	testCases := []struct {
		name              string
		flags             []string
		isTTY             bool
		env               map[string]string
		expectInStdout    []string
		expectNotInStdout []string
		expectInStderr    []string
		expectLogFile     bool
		expectInLogFile   []string
	}{
		{
			name:            "Default Interactive",
			flags:           []string{},
			isTTY:           true,
			expectInStdout:  []string{"🚀", "✓ completed", "✗ failed"},
			expectLogFile:   true,
			expectInLogFile: []string{`"level":"INFO"`, `"msg":"Simulating application start"`},
		},
		{
			name:              "Default CI",
			flags:             []string{},
			isTTY:             true,
			env:               map[string]string{"CI": "true"},
			expectInStdout:    []string{"Starting processing", "Completed model", "Failed model"},
			expectNotInStdout: []string{"🚀"},
			expectLogFile:     true,
		},
		{
			name:              "Quiet flag",
			flags:             []string{"--quiet"},
			isTTY:             true,
			expectNotInStdout: []string{"🚀"},
			expectLogFile:     true,
		},
		{
			name:           "JSON Logs flag",
			flags:          []string{"--json-logs"},
			isTTY:          true,
			expectInStdout: []string{"🚀", `"level":"INFO"`, `"msg":"Simulating application start"`},
			expectInStderr: []string{},
			expectLogFile:  false,
		},
		{
			name:              "No Progress flag",
			flags:             []string{"--no-progress"},
			isTTY:             true,
			expectInStdout:    []string{"🚀", "✗ failed"},
			expectNotInStdout: []string{"✓ completed"},
			expectLogFile:     true,
		},
		{
			name:           "Verbose flag (logs to stdout/stderr with stream separation)",
			flags:          []string{"--verbose"},
			isTTY:          true,
			expectInStdout: []string{"🚀", `"level":"INFO"`},
			expectInStderr: []string{`"level":"ERROR"`},
			expectLogFile:  false,
		},
		{
			name:              "Combined quiet and json-logs",
			flags:             []string{"--quiet", "--json-logs"},
			isTTY:             true,
			expectNotInStdout: []string{"🚀"},
			expectInStdout:    []string{`"level":"INFO"`},
			expectInStderr:    []string{},
			expectLogFile:     false,
		},
		{
			name:              "Combined no-progress and json-logs",
			flags:             []string{"--no-progress", "--json-logs"},
			isTTY:             true,
			expectInStdout:    []string{"🚀", "✗ failed", `"level":"INFO"`},
			expectNotInStdout: []string{"✓ completed"},
			expectInStderr:    []string{},
			expectLogFile:     false,
		},
		{
			name:              "CI environment with json-logs",
			flags:             []string{"--json-logs"},
			isTTY:             true,
			env:               map[string]string{"GITHUB_ACTIONS": "true"},
			expectInStdout:    []string{"Starting processing", "Failed model", `"level":"INFO"`},
			expectNotInStdout: []string{"🚀", "✓ completed"},
			expectInStderr:    []string{},
			expectLogFile:     false,
		},
		{
			name:              "Non-TTY environment",
			flags:             []string{},
			isTTY:             false,
			expectInStdout:    []string{"Starting processing", "Completed model", "Failed model"},
			expectNotInStdout: []string{"🚀"},
			expectLogFile:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := append(tc.flags, baseArgs...)
			// Set API key via environment variable instead of flag
			res := runCliTest(t, args, tc.env, tc.isTTY)

			// Assertions
			for _, s := range tc.expectInStdout {
				if !strings.Contains(res.stdout, s) {
					t.Errorf("stdout missing: %q", s)
				}
			}
			for _, s := range tc.expectNotInStdout {
				if strings.Contains(res.stdout, s) {
					t.Errorf("stdout has unexpected: %q", s)
				}
			}
			for _, s := range tc.expectInStderr {
				if !strings.Contains(res.stderr, s) {
					t.Errorf("stderr missing: %q", s)
				}
			}
			if (res.logFile != "") != tc.expectLogFile {
				t.Errorf("log file existence mismatch: got %v, want %v", (res.logFile != ""), tc.expectLogFile)
			}
			for _, s := range tc.expectInLogFile {
				if !strings.Contains(res.logFile, s) {
					t.Errorf("log file missing: %q", s)
				}
			}

			if t.Failed() {
				t.Logf("--- STDOUT ---\n%s\n--- STDERR ---\n%s\n--- LOGFILE ---\n%s", res.stdout, res.stderr, res.logFile)
			}
		})
	}
}

func TestCIEnvironmentDetection(t *testing.T) {
	baseArgs := []string{"--instructions", "test.txt", "--model", "test-model", "test-path"}

	ciEnvVars := map[string]string{
		"CI":                     "true",
		"GITHUB_ACTIONS":         "true",
		"GITLAB_CI":              "true",
		"TRAVIS":                 "true",
		"CIRCLECI":               "true",
		"JENKINS_URL":            "http://jenkins.example.com",
		"CONTINUOUS_INTEGRATION": "true",
	}

	for envVar, value := range ciEnvVars {
		t.Run("CI_Detection_"+envVar, func(t *testing.T) {
			env := map[string]string{envVar: value}
			res := runCliTest(t, baseArgs, env, true) // Even with TTY, CI should be detected

			// In CI mode, should not have interactive elements
			if strings.Contains(res.stdout, "🚀") {
				t.Errorf("CI mode should not contain emojis, but found them in: %s", res.stdout)
			}

			// Should have plain text output suitable for CI logs
			if !strings.Contains(res.stdout, "Starting processing") {
				t.Errorf("CI mode should contain plain text output, got: %s", res.stdout)
			}
		})
	}
}

func TestFlagValidationInIntegration(t *testing.T) {
	baseArgs := []string{"--instructions", "test.txt", "--model", "test-model", "test-path"}

	conflictTests := []struct {
		name        string
		flags       []string
		expectError bool
	}{
		{
			name:        "quiet and verbose conflict",
			flags:       []string{"--quiet", "--verbose"},
			expectError: true,
		},
		{
			name:        "quiet and json-logs coexist",
			flags:       []string{"--quiet", "--json-logs"},
			expectError: false,
		},
		{
			name:        "no-progress and verbose coexist",
			flags:       []string{"--no-progress", "--verbose"},
			expectError: false,
		},
	}

	for _, tc := range conflictTests {
		t.Run(tc.name, func(t *testing.T) {
			args := append(tc.flags, baseArgs...)

			getenv := func(key string) string {
				switch key {
				case "OPENAI_API_KEY":
					return "test-key-123"
				default:
					return ""
				}
			}

			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
			cfg, err := ParseFlagsWithEnv(flagSet, args, getenv)
			if err != nil && tc.expectError {
				// Expected error during parsing
				return
			}
			if err != nil && !tc.expectError {
				t.Errorf("Unexpected parsing error: %v", err)
				return
			}

			// Check validation errors
			mockLogger := &testLogger{}
			err = ValidateInputsWithEnv(cfg, mockLogger, getenv)

			if tc.expectError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestLoggingOutputRouting(t *testing.T) {
	baseArgs := []string{"--instructions", "test.txt", "--model", "test-model", "test-path"}

	tests := []struct {
		name                string
		flags               []string
		expectFileLogging   bool
		expectStderrLogging bool
	}{
		{
			name:                "Default routes to file",
			flags:               []string{},
			expectFileLogging:   true,
			expectStderrLogging: false,
		},
		{
			name:                "json-logs routes to stderr",
			flags:               []string{"--json-logs"},
			expectFileLogging:   false,
			expectStderrLogging: true,
		},
		{
			name:                "verbose routes to stderr",
			flags:               []string{"--verbose"},
			expectFileLogging:   false,
			expectStderrLogging: true,
		},
		{
			name:                "quiet still routes to file",
			flags:               []string{"--quiet"},
			expectFileLogging:   true,
			expectStderrLogging: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := append(tc.flags, baseArgs...)
			res := runCliTest(t, args, nil, true)

			hasFileLog := res.logFile != ""
			// For stderr logging, check for either INFO logs (when json-logs sends all to stderr)
			// or ERROR logs (when using stream separation)
			hasStderrLog := strings.Contains(res.stderr, `"level":"INFO"`) ||
				strings.Contains(res.stderr, `"level":"info"`) ||
				strings.Contains(res.stderr, `"level":"ERROR"`) ||
				strings.Contains(res.stderr, `"level":"error"`)

			if hasFileLog != tc.expectFileLogging {
				t.Errorf("Expected file logging=%v, got %v", tc.expectFileLogging, hasFileLog)
			}
			if hasStderrLog != tc.expectStderrLogging {
				t.Errorf("Expected stderr logging=%v, got %v", tc.expectStderrLogging, hasStderrLog)
			}
		})
	}
}
