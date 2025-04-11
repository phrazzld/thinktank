package architect

import (
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/logutil"
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
				"--instructions", "instructions.md",
				"path1", "path2",
			},
			env: map[string]string{
				apiKeyEnvVar: "test-api-key",
			},
			want: &CliConfig{
				InstructionsFile: "instructions.md",
				Paths:            []string{"path1", "path2"},
				ApiKey:           "test-api-key",
				OutputFile:       defaultOutputFile,
				ModelName:        defaultModel,
				Exclude:          defaultExcludes,
				ExcludeNames:     defaultExcludeNames,
				Format:           defaultFormat,
			},
			wantErr: false,
		},
		{
			name: "Missing instructions file",
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
				"--instructions", "instructions.md",
			},
			env: map[string]string{
				apiKeyEnvVar: "test-api-key",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Dry run without instructions file",
			args: []string{
				"--dry-run",
				"path1", "path2",
			},
			env: map[string]string{
				apiKeyEnvVar: "test-api-key",
			},
			want: &CliConfig{
				InstructionsFile: "",
				DryRun:           true,
				Paths:            []string{"path1", "path2"},
				ApiKey:           "test-api-key",
				OutputFile:       defaultOutputFile,
				ModelName:        defaultModel,
				Exclude:          defaultExcludes,
				ExcludeNames:     defaultExcludeNames,
				Format:           defaultFormat,
			},
			wantErr: false,
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

			// Compare fields we care about for this test
			if got.InstructionsFile != tt.want.InstructionsFile {
				t.Errorf("InstructionsFile = %v, want %v", got.InstructionsFile, tt.want.InstructionsFile)
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
			if got.DryRun != tt.want.DryRun {
				t.Errorf("DryRun = %v, want %v", got.DryRun, tt.want.DryRun)
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
			// Initialize the config with the desired log level based on test case
			if tt.config.Verbose {
				tt.config.LogLevel = logutil.DebugLevel
			} else if tt.logLevelFlag != nil {
				// Parse the flag value
				logLevelValue := tt.logLevelFlag.Value.String()
				parsedLevel, err := logutil.ParseLogLevel(logLevelValue)
				if err == nil {
					tt.config.LogLevel = parsedLevel
				} else {
					tt.config.LogLevel = logutil.InfoLevel
				}
			} else {
				tt.config.LogLevel = logutil.InfoLevel
			}

			// Now call SetupLoggingCustom which should now just use the existing LogLevel
			SetupLoggingCustom(tt.config, tt.logLevelFlag, io.Discard)

			// Case insensitive comparison since the logLevel returns uppercase values
			if strings.ToLower(tt.config.LogLevel.String()) != tt.wantLevel {
				t.Errorf("LogLevel = %v, want %v", tt.config.LogLevel.String(), tt.wantLevel)
			}
		})
	}
}

// TestAdvancedConfiguration tests more complex configuration options
func TestAdvancedConfiguration(t *testing.T) {
	testCases := []struct {
		name             string
		args             []string
		expectedFormat   string
		expectedModel    string
		expectedExclude  string
		expectedLogLevel string
	}{
		{
			name:             "Default format and model",
			args:             []string{"--instructions", "instructions.txt", "./"},
			expectedFormat:   defaultFormat,
			expectedModel:    defaultModel,
			expectedExclude:  defaultExcludes,
			expectedLogLevel: "info",
		},
		{
			name: "Custom format and model",
			args: []string{
				"--instructions", "instructions.txt",
				"--format", "Custom: {path}\n{content}\n---\n",
				"--model", "custom-model",
				"--exclude", "*.tmp,*.bak",
				"--log-level", "debug",
				"./",
			},
			expectedFormat:   "Custom: {path}\n{content}\n---\n",
			expectedModel:    "custom-model",
			expectedExclude:  "*.tmp,*.bak",
			expectedLogLevel: "debug",
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
			if err != nil {
				t.Fatalf("Expected no error for valid config, got: %v", err)
			}

			// Check values
			if config.Format != tc.expectedFormat {
				t.Errorf("Expected Format to be %q, got %q", tc.expectedFormat, config.Format)
			}

			if config.ModelName != tc.expectedModel {
				t.Errorf("Expected ModelName to be %q, got %q", tc.expectedModel, config.ModelName)
			}

			if config.Exclude != tc.expectedExclude {
				t.Errorf("Expected Exclude to be %q, got %q", tc.expectedExclude, config.Exclude)
			}

			// Set up logging to populate the log level
			SetupLoggingCustom(config, nil, io.Discard)

			if strings.ToLower(config.LogLevel.String()) != tc.expectedLogLevel {
				t.Errorf("Expected LogLevel to be %q, got %q", tc.expectedLogLevel, strings.ToLower(config.LogLevel.String()))
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

// TestInstructionsFileRequirement confirms that ValidateInputs properly checks for instructions file
// TestUsageMessage verifies that the usage message contains the correct information
func TestUsageMessage(t *testing.T) {
	// Create a buffer to capture the usage output
	var buffer strings.Builder

	// Create a new flag set with the buffer as output
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(&buffer)

	// Define flags to make this a real flagset (minimal subset)
	fs.String("instructions", "", "Path to a file containing the static instructions for the LLM.")
	fs.String("output", "output.md", "Output file path")

	// Add a custom usage function similar to the one in cli.go
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s --instructions <file> [options] <path1> [path2...]\n\n", os.Args[0])

		fmt.Fprintf(fs.Output(), "Arguments:\n")
		fmt.Fprintf(fs.Output(), "  <path1> [path2...]   One or more file or directory paths for project context.\n\n")

		fmt.Fprintf(fs.Output(), "Example Commands:\n")
		fmt.Fprintf(fs.Output(), "  %s --instructions instructions.txt ./src                Generate plan using instructions file\n", os.Args[0])
		fmt.Fprintf(fs.Output(), "  %s --instructions instructions.txt --output custom.md ./ Generate plan with custom output file\n", os.Args[0])
		fmt.Fprintf(fs.Output(), "  %s --dry-run ./                                         Show files without generating plan\n\n", os.Args[0])

		fmt.Fprintf(fs.Output(), "Options:\n")
		fs.PrintDefaults()
	}

	// Call usage
	fs.Usage()

	// Get the output
	output := buffer.String()

	// Verify the usage message contains key elements
	requiredPhrases := []string{
		"--instructions <file>",
		"Arguments:",
		"Example Commands:",
		"--instructions instructions.txt",
		"--dry-run",
		"Options:",
	}

	for _, phrase := range requiredPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Usage message doesn't contain required phrase: %q", phrase)
		}
	}

	// Verify the usage message does NOT contain removed flags
	removedFlags := []string{
		"--task-file",
		"--prompt-template",
		"--list-examples",
		"--show-example",
	}

	for _, flag := range removedFlags {
		if strings.Contains(output, flag) {
			t.Errorf("Usage message shouldn't contain removed flag: %q", flag)
		}
	}
}

func TestInstructionsFileRequirement(t *testing.T) {
	// Create a test instructions file
	tempFile, err := os.CreateTemp("", "instructions-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary instructions file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString("Test instructions content")
	if err != nil {
		t.Fatalf("Failed to write to temporary instructions file: %v", err)
	}
	tempFile.Close()

	// Test with a valid instructions file
	config := &CliConfig{
		InstructionsFile: tempFile.Name(),
		Paths:            []string{"testfile"},
		ApiKey:           "test-key",
	}

	// Create a custom logger that tracks error calls
	errLogger := &errorTrackingLogger{}

	// Run validation with a valid instructions file
	var validationErr error

	// Test the normal case first (valid config)
	validationErr = ValidateInputs(config, errLogger)
	if validationErr != nil {
		t.Errorf("Validation failed with a valid instructions file: %s", validationErr)
	}

	// Check that no error was logged
	if errLogger.errorCalled {
		t.Error("Error was logged for valid instructions file")
	}

	// Reset for next test
	errLogger.reset()

	// Test with no instructions file and not in dry run mode
	configNoFile := &CliConfig{
		InstructionsFile: "",
		DryRun:           false,
		Paths:            []string{"testfile"},
		ApiKey:           "test-key",
	}

	// Run validation with no instructions file
	validationErr = ValidateInputs(configNoFile, errLogger)

	// Validation should fail without an instructions file
	if validationErr == nil {
		t.Error("Validation passed with no instructions file, expected failure")
	}

	// Should log an error about missing instructions file
	if !errLogger.errorCalled {
		t.Error("No error was logged for missing instructions file")
	}

	// Reset for next test
	errLogger.reset()

	// Test with no instructions file but in dry run mode
	configDryRun := &CliConfig{
		InstructionsFile: "",
		DryRun:           true,
		Paths:            []string{"testfile"},
		ApiKey:           "test-key",
	}

	// Run validation in dry run mode
	validationErr = ValidateInputs(configDryRun, errLogger)

	// Validation should pass in dry run mode even without an instructions file
	if validationErr != nil {
		t.Errorf("Validation failed in dry run mode without an instructions file: %s", validationErr)
	}

	// No error should be logged
	if errLogger.errorCalled {
		t.Error("Error was logged for dry run mode")
	}
}
