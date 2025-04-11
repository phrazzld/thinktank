package architect

import (
	"flag"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestParseFlagsWithEnv(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		env     map[string]string
		want    *CliConfig
		wantErr bool
	}{
		{
			name: "Basic valid configuration",
			args: []string{
				"--task-file", "task.md",
				"path1", "path2",
			},
			env: map[string]string{
				apiKeyEnvVar: "test-api-key",
			},
			want: &CliConfig{
				TaskFile:     "task.md",
				Paths:        []string{"path1", "path2"},
				ApiKey:       "test-api-key",
				OutputFile:   defaultOutputFile,
				ModelName:    defaultModel,
				Exclude:      defaultExcludes,
				ExcludeNames: defaultExcludeNames,
				Format:       defaultFormat,
			},
			wantErr: false,
		},
		{
			name: "Missing task file",
			args: []string{
				"path1",
			},
			env: map[string]string{
				apiKeyEnvVar: "test-api-key",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Missing paths",
			args: []string{
				"--task-file", "task.md",
			},
			env: map[string]string{
				apiKeyEnvVar: "test-api-key",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new FlagSet for each test
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			// Disable output to avoid cluttering test output
			fs.SetOutput(io.Discard)

			// Create a mock environment getter
			getenv := func(key string) string {
				return tt.env[key]
			}

			got, err := ParseFlagsWithEnv(fs, tt.args, getenv)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlagsWithEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Compare only the fields we care about for this test
			if got.TaskFile != tt.want.TaskFile {
				t.Errorf("TaskFile = %v, want %v", got.TaskFile, tt.want.TaskFile)
			}
			if !reflect.DeepEqual(got.Paths, tt.want.Paths) {
				t.Errorf("Paths = %v, want %v", got.Paths, tt.want.Paths)
			}
			if got.ApiKey != tt.want.ApiKey {
				t.Errorf("ApiKey = %v, want %v", got.ApiKey, tt.want.ApiKey)
			}
			if got.OutputFile != tt.want.OutputFile {
				t.Errorf("OutputFile = %v, want %v", got.OutputFile, tt.want.OutputFile)
			}
		})
	}
}

func TestSetupLoggingCustom(t *testing.T) {
	tests := []struct {
		name         string
		config       *CliConfig
		logLevelFlag *flag.Flag
		wantLevel    string
	}{
		{
			name: "Use verbose flag",
			config: &CliConfig{
				Verbose: true,
			},
			logLevelFlag: nil,
			wantLevel:    "debug",
		},
		{
			name: "Use default when no verbose or flag",
			config: &CliConfig{
				Verbose: false,
			},
			logLevelFlag: nil,
			wantLevel:    "info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily check the logger itself, but we can check
			// that the config's LogLevel was set correctly
			SetupLoggingCustom(tt.config, tt.logLevelFlag, io.Discard)

			// Case insensitive comparison since the logLevel returns uppercase values
			if strings.ToLower(tt.config.LogLevel.String()) != tt.wantLevel {
				t.Errorf("LogLevel = %v, want %v", tt.config.LogLevel.String(), tt.wantLevel)
			}
		})
	}
}

// TestParsingExampleFlags tests the parsing of the example-related flags
func TestParsingExampleFlags(t *testing.T) {
	testCases := []struct {
		name                string
		args                []string
		expectedListFlag    bool
		expectedShowExample string
	}{
		{
			name:                "No example flags",
			args:                []string{"--task-file", "task.txt", "./"},
			expectedListFlag:    false,
			expectedShowExample: "",
		},
		{
			name:                "With list-examples flag",
			args:                []string{"--list-examples"},
			expectedListFlag:    true,
			expectedShowExample: "",
		},
		{
			name:                "With show-example flag",
			args:                []string{"--show-example", "basic.tmpl"},
			expectedListFlag:    false,
			expectedShowExample: "basic.tmpl",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new FlagSet for each test
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			// Disable output to avoid cluttering test output
			fs.SetOutput(io.Discard)

			// Create a mock environment getter
			getenv := func(key string) string {
				return "test-api-key"
			}

			// Parse flags
			config, err := ParseFlagsWithEnv(fs, tc.args, getenv)

			// If we expect ListExamples or ShowExample, there shouldn't be an error
			// even if task-file is missing
			if tc.expectedListFlag || tc.expectedShowExample != "" {
				if err != nil {
					t.Fatalf("Expected no error for example flags, got: %v", err)
				}
			} else if err != nil {
				// For other test cases, errors are expected - we're testing the example flags specifically
				return
			}

			// Check flag values
			if config.ListExamples != tc.expectedListFlag {
				t.Errorf("Expected ListExamples to be %v, got %v", tc.expectedListFlag, config.ListExamples)
			}

			if config.ShowExample != tc.expectedShowExample {
				t.Errorf("Expected ShowExample to be %q, got %q", tc.expectedShowExample, config.ShowExample)
			}
		})
	}
}

// errorTrackingLogger is a minimal logger that tracks if error methods were called
type errorTrackingLogger struct {
	errorCalled bool
}

func (l *errorTrackingLogger) Error(format string, args ...interface{}) {
	l.errorCalled = true
}

func (l *errorTrackingLogger) reset() {
	l.errorCalled = false
}

func (l *errorTrackingLogger) Debug(format string, args ...interface{})  {}
func (l *errorTrackingLogger) Info(format string, args ...interface{})   {}
func (l *errorTrackingLogger) Warn(format string, args ...interface{})   {}
func (l *errorTrackingLogger) Fatal(format string, args ...interface{})  {}
func (l *errorTrackingLogger) Printf(format string, args ...interface{}) {}
func (l *errorTrackingLogger) Println(v ...interface{})                  {}

// TestTaskFileRequirementSimple confirms that ValidateInputs properly checks for task-file
func TestTaskFileRequirementSimple(t *testing.T) {
	// Create a test task file
	tempFile, err := os.CreateTemp("", "task-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary task file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString("Test task content")
	if err != nil {
		t.Fatalf("Failed to write to temporary task file: %v", err)
	}
	tempFile.Close()

	// Test with a valid task file
	config := &CliConfig{
		TaskFile: tempFile.Name(),
		Paths:    []string{"testfile"},
		ApiKey:   "test-key",
	}

	// Create a custom logger that tracks error calls
	errLogger := &errorTrackingLogger{}

	// Run validation with a valid task file
	var validationErr error

	// Test the normal case first (valid config)
	validationErr = ValidateInputs(config, errLogger)
	if validationErr != nil {
		t.Errorf("Validation failed with a valid task file: %s", validationErr)
	}

	// Check that no error was logged
	if errLogger.errorCalled {
		t.Error("Error was logged for valid task file")
	}

	// Reset for next test
	errLogger.reset()

	// Test with no task file and not in dry run mode
	configNoFile := &CliConfig{
		TaskFile: "",
		DryRun:   false,
		Paths:    []string{"testfile"},
		ApiKey:   "test-key",
	}

	// Run validation with no task file
	validationErr = ValidateInputs(configNoFile, errLogger)

	// Validation should fail without a task file
	if validationErr == nil {
		t.Error("Validation passed with no task file, expected failure")
	}

	// Should log an error about missing task file
	if !errLogger.errorCalled {
		t.Error("No error was logged for missing task file")
	}

	// Reset for next test
	errLogger.reset()

	// Test with no task file but in dry run mode
	configDryRun := &CliConfig{
		TaskFile: "",
		DryRun:   true,
		Paths:    []string{"testfile"},
		ApiKey:   "test-key",
	}

	// Run validation in dry run mode
	validationErr = ValidateInputs(configDryRun, errLogger)

	// Validation should pass in dry run mode even without a task file
	if validationErr != nil {
		t.Errorf("Validation failed in dry run mode without a task file: %s", validationErr)
	}

	// No error should be logged
	if errLogger.errorCalled {
		t.Error("Error was logged for dry run mode")
	}
}
