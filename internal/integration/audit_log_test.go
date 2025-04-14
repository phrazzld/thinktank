// internal/integration/audit_log_test.go
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/architect"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/logutil"
)

// TestAuditLogFunctionality tests the audit logging functionality
// using a table-driven approach to reduce repetitive test code.
func TestAuditLogFunctionality(t *testing.T) {
	// Define the test case struct for audit logging scenarios
	type auditLogTestCase struct {
		name                string
		instructionsContent string
		srcFiles            map[string]string
		loggerSetup         func(*TestEnv, *testing.T) (auditlog.AuditLogger, error)
		expectedLogEntries  []string
		validateLogFunc     func([]map[string]interface{}, *testing.T) bool
		shouldCreateLogFile bool
		outputShouldExist   bool
	}

	// Define test cases based on the original audit logging tests
	tests := []auditLogTestCase{
		{
			name:                "Valid audit log file",
			instructionsContent: "Implement a new feature to multiply two numbers",
			srcFiles: map[string]string{
				"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}`,
			},
			loggerSetup: func(env *TestEnv, t *testing.T) (auditlog.AuditLogger, error) {
				// Set up the audit log file path
				auditLogFile := filepath.Join(env.TestDir, "audit.log")

				// Create a custom FileAuditLogger
				testLogger := env.GetBufferedLogger(logutil.DebugLevel, "[test] ")
				return auditlog.NewFileAuditLogger(auditLogFile, testLogger)
			},
			expectedLogEntries: []string{
				"ExecuteStart",
				"ReadInstructions",
				"GatherContextStart",
				"GatherContextEnd",
				"CheckTokens",
				"GenerateContentStart",
				"GenerateContentEnd",
				"SaveOutputStart",
				"SaveOutputEnd",
				"ExecuteEnd",
			},
			validateLogFunc: func(entries []map[string]interface{}, t *testing.T) bool {
				// Validate ExecuteStart operation in detail
				for _, entry := range entries {
					operation, ok := entry["operation"].(string)
					if !ok || operation != "ExecuteStart" {
						continue
					}

					// Check status
					status, ok := entry["status"].(string)
					if !ok || status != "InProgress" {
						t.Errorf("Expected ExecuteStart status to be 'InProgress', got '%v'", status)
						return false
					}

					// Check timestamp
					_, hasTimestamp := entry["timestamp"]
					if !hasTimestamp {
						t.Error("ExecuteStart entry missing timestamp")
						return false
					}

					// Check inputs (should include CLI flags)
					inputs, ok := entry["inputs"].(map[string]interface{})
					if !ok {
						t.Error("ExecuteStart entry missing inputs")
						return false
					}

					// Verify specific input field (model names)
					modelNames, ok := inputs["model_names"]
					if !ok {
						t.Errorf("ExecuteStart inputs missing model_names, got: %v", modelNames)
						return false
					}

					// Verify the correct model name is included in the slice
					modelNamesSlice, ok := modelNames.([]interface{})
					if !ok || len(modelNamesSlice) == 0 || modelNamesSlice[0] != "test-model" {
						t.Errorf("ExecuteStart inputs incorrect model_names, got: %v", modelNames)
						return false
					}

					return true
				}
				t.Error("ExecuteStart operation not found in audit log")
				return false
			},
			shouldCreateLogFile: true,
			outputShouldExist:   true,
		},
		{
			name:                "Fallback to NoOpAuditLogger",
			instructionsContent: "Test instructions",
			srcFiles: map[string]string{
				"main.go": `package main

func main() {}`,
			},
			loggerSetup: func(env *TestEnv, t *testing.T) (auditlog.AuditLogger, error) {
				// Set up a test logger that captures errors
				testLogger := env.GetBufferedLogger(logutil.DebugLevel, "[test] ")

				// Attempt to create a FileAuditLogger with invalid path (will fail)
				invalidDir := filepath.Join(env.TestDir, "nonexistent-dir")
				invalidLogFile := filepath.Join(invalidDir, "audit.log")
				_, err := auditlog.NewFileAuditLogger(invalidLogFile, testLogger)

				// Verify error message contains expected text
				if err != nil {
					if !strings.Contains(err.Error(), "failed to open audit log file") {
						t.Errorf("Expected error message to contain 'failed to open audit log file', got: %s", err.Error())
					}
				} else {
					t.Error("Expected error when creating FileAuditLogger with invalid path, got nil")
				}

				// Return a NoOpAuditLogger as fallback
				return auditlog.NewNoOpAuditLogger(), nil
			},
			expectedLogEntries: nil, // No entries expected with NoOpAuditLogger
			validateLogFunc: func(entries []map[string]interface{}, t *testing.T) bool {
				// NoOpAuditLogger doesn't create any log entries, so we just verify
				// that the application ran successfully
				return true
			},
			shouldCreateLogFile: false,
			outputShouldExist:   true,
		},
	}

	// Execute each test case
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set up test environment
			env := NewTestEnv(t)
			defer env.Cleanup()

			// Set up the mock client
			env.SetupMockGeminiClient()

			// Create source files from the map
			for filename, content := range tc.srcFiles {
				env.CreateTestFile(t, "src/"+filename, content)
			}

			// Create instructions file
			instructionsFile := env.CreateTestFile(t, "instructions.md", tc.instructionsContent)

			// Set up the output directory and model-specific output file path
			outputDir := filepath.Join(env.TestDir, "output")
			modelName := "test-model"
			outputFile := filepath.Join(outputDir, modelName+".md")
			auditLogFile := filepath.Join(env.TestDir, "audit.log")

			// Set up the custom audit logger using the test case's setup function
			auditLogger, err := tc.loggerSetup(env, t)
			if err != nil {
				// If logger setup fails, the test case should handle this
				// and provide a fallback logger
				if auditLogger == nil {
					t.Fatalf("Logger setup failed and no fallback provided: %v", err)
				}
			}

			// If we got a FileAuditLogger, make sure to close it
			if fileLogger, ok := auditLogger.(*auditlog.FileAuditLogger); ok {
				defer fileLogger.Close()
			}

			// Replace the default NoOpAuditLogger with our test logger
			env.AuditLogger = auditLogger

			// Create a test configuration with the audit log file path
			testConfig := &config.CliConfig{
				InstructionsFile: instructionsFile,
				OutputDir:        outputDir,
				ModelNames:       []string{modelName},
				APIKey:           "test-api-key",
				Paths:            []string{env.TestDir + "/src"},
				LogLevel:         logutil.InfoLevel,
				AuditLogFile:     auditLogFile,
			}

			// Create a mock API service
			mockApiService := createMockAPIService(env)

			// Run the application with our test configuration
			ctx := context.Background()
			err = architect.Execute(
				ctx,
				testConfig,
				env.Logger,
				env.AuditLogger,
				mockApiService,
			)

			// Verify execution succeeded
			if err != nil {
				t.Fatalf("architect.Execute failed: %v", err)
			}

			// Check if output file exists based on expectation
			if tc.outputShouldExist {
				if _, err := os.Stat(outputFile); os.IsNotExist(err) {
					t.Errorf("Expected output file to exist, but it doesn't: %s", outputFile)
				}
			} else {
				if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
					t.Errorf("Output file was created when it shouldn't have been: %s", outputFile)
				}
			}

			// Validate audit log if we expect it to be created
			if tc.shouldCreateLogFile {
				// Check that the audit log file exists
				if _, err := os.Stat(auditLogFile); os.IsNotExist(err) {
					t.Errorf("Audit log file was not created at %s", auditLogFile)
					return
				}

				// Read and parse the audit log file
				content, err := os.ReadFile(auditLogFile)
				if err != nil {
					t.Fatalf("Failed to read audit log file: %v", err)
				}

				// Parse JSON entries
				lines := bytes.Split(content, []byte{'\n'})
				var entries []map[string]interface{}

				for _, line := range lines {
					if len(line) == 0 {
						continue
					}

					var entry map[string]interface{}
					if err := json.Unmarshal(line, &entry); err != nil {
						t.Errorf("Failed to parse JSON line: %v", err)
						t.Errorf("Line content: %s", string(line))
						continue
					}
					entries = append(entries, entry)
				}

				// Check that we have at least some entries
				if len(entries) == 0 {
					t.Error("Audit log file is empty or contains no valid JSON entries")
					return
				}

				// Validate expected log entries
				if tc.expectedLogEntries != nil {
					// Create a map to track which operations we found
					foundOperations := make(map[string]bool)

					// Find operations in the log entries
					for _, entry := range entries {
						operation, ok := entry["operation"].(string)
						if ok {
							foundOperations[operation] = true
						}
					}

					// Check that all expected operations are present
					for _, expectedOp := range tc.expectedLogEntries {
						if !foundOperations[expectedOp] {
							t.Errorf("Expected operation '%s' not found in audit log", expectedOp)
						}
					}
				}

				// Run custom validation function if provided
				if tc.validateLogFunc != nil {
					if !tc.validateLogFunc(entries, t) {
						t.Error("Custom validation of audit log entries failed")
					}
				}
			}
		})
	}
}