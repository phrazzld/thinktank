package architect

import (
	"flag"
	"io"
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
				"--task", "Test task",
				"path1", "path2",
			},
			env: map[string]string{
				apiKeyEnvVar: "test-api-key",
			},
			want: &CliConfig{
				TaskDescription: "Test task",
				Paths:           []string{"path1", "path2"},
				ApiKey:          "test-api-key",
				OutputFile:      defaultOutputFile,
				ModelName:       defaultModel,
				UseColors:       true, // default value
				Exclude:         defaultExcludes,
				ExcludeNames:    defaultExcludeNames,
				Format:          defaultFormat,
			},
			wantErr: false,
		},
		{
			name: "Missing task description",
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
				"--task", "Test task",
			},
			env: map[string]string{
				apiKeyEnvVar: "test-api-key",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Task file instead of task description",
			args: []string{
				"--task-file", "task.md",
				"path1",
			},
			env: map[string]string{
				apiKeyEnvVar: "test-api-key",
			},
			want: &CliConfig{
				TaskFile:     "task.md",
				Paths:        []string{"path1"},
				ApiKey:       "test-api-key",
				OutputFile:   defaultOutputFile,
				ModelName:    defaultModel,
				UseColors:    true, // default value
				Exclude:      defaultExcludes,
				ExcludeNames: defaultExcludeNames,
				Format:       defaultFormat,
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

			// Compare only the fields we care about for this test
			if got.TaskDescription != tt.want.TaskDescription {
				t.Errorf("TaskDescription = %v, want %v", got.TaskDescription, tt.want.TaskDescription)
			}
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
