// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"os"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
)

func TestApplyEnvironmentVars(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"THINKTANK_MODEL",
		"THINKTANK_MODELS",
		"THINKTANK_OUTPUT_DIR",
		"THINKTANK_SYNTHESIS_MODEL",
		"THINKTANK_DRY_RUN",
		"THINKTANK_VERBOSE",
		"THINKTANK_QUIET",
		"THINKTANK_NO_PROGRESS",
		"THINKTANK_TIMEOUT",
		"THINKTANK_EXCLUDE",
		"THINKTANK_EXCLUDE_NAMES",
	}

	for _, env := range envVars {
		originalEnv[env] = os.Getenv(env)
		_ = os.Unsetenv(env)
	}

	defer func() {
		for _, env := range envVars {
			if val, exists := originalEnv[env]; exists && val != "" {
				_ = os.Setenv(env, val)
			} else {
				_ = os.Unsetenv(env)
			}
		}
	}()

	tests := []struct {
		name     string
		envVars  map[string]string
		initial  *config.MinimalConfig
		expected *config.MinimalConfig
		wantErr  bool
	}{
		{
			name: "apply single model",
			envVars: map[string]string{
				"THINKTANK_MODEL": "gpt-4",
			},
			initial: &config.MinimalConfig{
				ModelNames: []string{config.DefaultModel},
			},
			expected: &config.MinimalConfig{
				ModelNames: []string{"gpt-4"},
			},
		},
		{
			name: "apply multiple models",
			envVars: map[string]string{
				"THINKTANK_MODELS": "gpt-4,gemini-1.5-flash",
			},
			initial: &config.MinimalConfig{
				ModelNames: []string{config.DefaultModel},
			},
			expected: &config.MinimalConfig{
				ModelNames: []string{"gpt-4", "gemini-1.5-flash"},
			},
		},
		{
			name: "apply output directory",
			envVars: map[string]string{
				"THINKTANK_OUTPUT_DIR": "/tmp/test",
			},
			initial: &config.MinimalConfig{
				OutputDir: "",
			},
			expected: &config.MinimalConfig{
				OutputDir: "/tmp/test",
			},
		},
		{
			name: "apply synthesis model",
			envVars: map[string]string{
				"THINKTANK_SYNTHESIS_MODEL": "gpt-4",
			},
			initial: &config.MinimalConfig{
				SynthesisModel: "",
			},
			expected: &config.MinimalConfig{
				SynthesisModel: "gpt-4",
			},
		},
		{
			name: "apply boolean flags",
			envVars: map[string]string{
				"THINKTANK_DRY_RUN":     "true",
				"THINKTANK_VERBOSE":     "1",
				"THINKTANK_QUIET":       "yes",
				"THINKTANK_NO_PROGRESS": "on",
			},
			initial: &config.MinimalConfig{
				DryRun:     false,
				Verbose:    false,
				Quiet:      false,
				NoProgress: false,
				LogLevel:   logutil.InfoLevel,
			},
			expected: &config.MinimalConfig{
				DryRun:     true,
				Verbose:    true,
				Quiet:      true,
				NoProgress: true,
				LogLevel:   logutil.DebugLevel, // Should be set when verbose is true
			},
		},
		{
			name: "apply timeout",
			envVars: map[string]string{
				"THINKTANK_TIMEOUT": "30m",
			},
			initial: &config.MinimalConfig{
				Timeout: config.DefaultTimeout,
			},
			expected: &config.MinimalConfig{
				Timeout: 30 * time.Minute,
			},
		},
		{
			name: "apply file patterns",
			envVars: map[string]string{
				"THINKTANK_EXCLUDE":       "*.log,*.tmp",
				"THINKTANK_EXCLUDE_NAMES": "node_modules,target",
			},
			initial: &config.MinimalConfig{
				Exclude:      config.DefaultExcludes,
				ExcludeNames: config.DefaultExcludeNames,
			},
			expected: &config.MinimalConfig{
				Exclude:      "*.log,*.tmp",
				ExcludeNames: "node_modules,target",
			},
		},
		{
			name: "invalid timeout format",
			envVars: map[string]string{
				"THINKTANK_TIMEOUT": "invalid",
			},
			initial: &config.MinimalConfig{
				Timeout: config.DefaultTimeout,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}
			defer func() {
				for key := range tt.envVars {
					_ = os.Unsetenv(key)
				}
			}()

			// Apply environment variables
			err := applyEnvironmentVars(tt.initial)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check expected values
			if tt.expected != nil {
				if len(tt.expected.ModelNames) > 0 && len(tt.initial.ModelNames) > 0 {
					if len(tt.initial.ModelNames) != len(tt.expected.ModelNames) {
						t.Errorf("ModelNames length mismatch: got %d, want %d", len(tt.initial.ModelNames), len(tt.expected.ModelNames))
					} else {
						for i, model := range tt.expected.ModelNames {
							if tt.initial.ModelNames[i] != model {
								t.Errorf("ModelNames[%d] = %s, want %s", i, tt.initial.ModelNames[i], model)
							}
						}
					}
				}

				if tt.expected.OutputDir != "" && tt.initial.OutputDir != tt.expected.OutputDir {
					t.Errorf("OutputDir = %s, want %s", tt.initial.OutputDir, tt.expected.OutputDir)
				}

				if tt.expected.SynthesisModel != "" && tt.initial.SynthesisModel != tt.expected.SynthesisModel {
					t.Errorf("SynthesisModel = %s, want %s", tt.initial.SynthesisModel, tt.expected.SynthesisModel)
				}

				if tt.initial.DryRun != tt.expected.DryRun {
					t.Errorf("DryRun = %v, want %v", tt.initial.DryRun, tt.expected.DryRun)
				}

				if tt.initial.Verbose != tt.expected.Verbose {
					t.Errorf("Verbose = %v, want %v", tt.initial.Verbose, tt.expected.Verbose)
				}

				if tt.initial.Quiet != tt.expected.Quiet {
					t.Errorf("Quiet = %v, want %v", tt.initial.Quiet, tt.expected.Quiet)
				}

				if tt.initial.NoProgress != tt.expected.NoProgress {
					t.Errorf("NoProgress = %v, want %v", tt.initial.NoProgress, tt.expected.NoProgress)
				}

				if tt.expected.Timeout != 0 && tt.initial.Timeout != tt.expected.Timeout {
					t.Errorf("Timeout = %v, want %v", tt.initial.Timeout, tt.expected.Timeout)
				}

				if tt.expected.Exclude != "" && tt.initial.Exclude != tt.expected.Exclude {
					t.Errorf("Exclude = %s, want %s", tt.initial.Exclude, tt.expected.Exclude)
				}

				if tt.expected.ExcludeNames != "" && tt.initial.ExcludeNames != tt.expected.ExcludeNames {
					t.Errorf("ExcludeNames = %s, want %s", tt.initial.ExcludeNames, tt.expected.ExcludeNames)
				}

				if tt.expected.LogLevel != 0 && tt.initial.LogLevel != tt.expected.LogLevel {
					t.Errorf("LogLevel = %v, want %v", tt.initial.LogLevel, tt.expected.LogLevel)
				}
			}
		})
	}
}

func TestCreateAdapterConfig(t *testing.T) {
	minimalConfig := &config.MinimalConfig{
		InstructionsFile: "test.txt",
		TargetPaths:      []string{"./src", "./docs"},
		ModelNames:       []string{"gemini-1.5-flash"},
		OutputDir:        "/tmp/output",
		DryRun:           true,
		Verbose:          true,
		SynthesisModel:   "gpt-4",
		LogLevel:         logutil.DebugLevel,
		Timeout:          30 * time.Minute,
		Quiet:            false,
		NoProgress:       false,
		Format:           "markdown",
		Exclude:          "*.log",
		ExcludeNames:     "node_modules",
	}

	result := createAdapterConfig(minimalConfig)

	if result.InstructionsFile != minimalConfig.InstructionsFile {
		t.Errorf("InstructionsFile = %s, want %s", result.InstructionsFile, minimalConfig.InstructionsFile)
	}

	if len(result.Paths) != len(minimalConfig.TargetPaths) {
		t.Errorf("Paths length = %d, want %d", len(result.Paths), len(minimalConfig.TargetPaths))
	}

	if len(result.ModelNames) != len(minimalConfig.ModelNames) {
		t.Errorf("ModelNames length = %d, want %d", len(result.ModelNames), len(minimalConfig.ModelNames))
	}

	if result.OutputDir != minimalConfig.OutputDir {
		t.Errorf("OutputDir = %s, want %s", result.OutputDir, minimalConfig.OutputDir)
	}

	if result.DryRun != minimalConfig.DryRun {
		t.Errorf("DryRun = %v, want %v", result.DryRun, minimalConfig.DryRun)
	}

	// Check that smart defaults are applied
	if result.MaxConcurrentRequests != 5 {
		t.Errorf("MaxConcurrentRequests = %d, want 5", result.MaxConcurrentRequests)
	}

	if result.RateLimitRequestsPerMinute != 60 {
		t.Errorf("RateLimitRequestsPerMinute = %d, want 60", result.RateLimitRequestsPerMinute)
	}

	if result.DirPermissions != 0755 {
		t.Errorf("DirPermissions = %o, want 0755", result.DirPermissions)
	}

	if result.FilePermissions != 0644 {
		t.Errorf("FilePermissions = %o, want 0644", result.FilePermissions)
	}

	if result.PartialSuccessOk != false {
		t.Errorf("PartialSuccessOk = %v, want false", result.PartialSuccessOk)
	}
}
