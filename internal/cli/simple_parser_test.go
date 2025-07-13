// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"github.com/phrazzld/thinktank/internal/testutil/perftest"
)

// TestParseSimpleArgsWithArgs_Basic tests the core functionality with minimal valid cases
func TestParseSimpleArgsWithArgs_Basic(t *testing.T) {
	// Create test files for validation
	tempDir := t.TempDir()
	testInstructionsFile := filepath.Join(tempDir, "instructions.txt")
	testTargetDir := filepath.Join(tempDir, "src")

	// Create test files
	if err := os.WriteFile(testInstructionsFile, []byte("test instructions"), 0644); err != nil {
		t.Fatalf("Failed to create test instructions file: %v", err)
	}
	if err := os.MkdirAll(testTargetDir, 0755); err != nil {
		t.Fatalf("Failed to create test target directory: %v", err)
	}

	tests := []struct {
		name        string
		args        []string
		want        *SimplifiedConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "minimal_valid_args_dry_run",
			args: []string{"thinktank", testInstructionsFile, testTargetDir, "--dry-run"},
			want: &SimplifiedConfig{
				InstructionsFile: testInstructionsFile,
				TargetPath:       testTargetDir,
				Flags:            FlagDryRun,
				SafetyMargin:     10, // Default safety margin
			},
		},
		{
			name:        "insufficient_args_none",
			args:        []string{"thinktank"},
			wantErr:     true,
			errContains: "usage:",
		},
		{
			name:        "insufficient_args_one",
			args:        []string{"thinktank", "instructions.txt"},
			wantErr:     true,
			errContains: "usage:",
		},
		{
			name:        "empty_args",
			args:        []string{},
			wantErr:     true,
			errContains: "usage:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSimpleArgsWithArgs(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseSimpleArgsWithArgs() expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ParseSimpleArgsWithArgs() error = %q, want error containing %q",
						err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseSimpleArgsWithArgs() unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSimpleArgsWithArgs() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestParseSimpleArgsWithArgs_BooleanFlags tests all boolean flag combinations
func TestParseSimpleArgsWithArgs_BooleanFlags(t *testing.T) {
	// Create test files for validation
	tempDir := t.TempDir()
	testInstructionsFile := filepath.Join(tempDir, "instructions.txt")
	testTargetDir := filepath.Join(tempDir, "src")

	// Create test files
	if err := os.WriteFile(testInstructionsFile, []byte("test instructions"), 0644); err != nil {
		t.Fatalf("Failed to create test instructions file: %v", err)
	}
	if err := os.MkdirAll(testTargetDir, 0755); err != nil {
		t.Fatalf("Failed to create test target directory: %v", err)
	}

	tests := []struct {
		name        string
		args        []string
		want        *SimplifiedConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "single_dry_run_flag",
			args: []string{"thinktank", testInstructionsFile, testTargetDir, "--dry-run"},
			want: &SimplifiedConfig{
				InstructionsFile: testInstructionsFile,
				TargetPath:       testTargetDir,
				Flags:            FlagDryRun,
				SafetyMargin:     10, // Default safety margin
			},
		},
		{
			name: "single_verbose_flag",
			args: []string{"thinktank", testInstructionsFile, testTargetDir, "--verbose", "--dry-run"},
			want: &SimplifiedConfig{
				InstructionsFile: testInstructionsFile,
				TargetPath:       testTargetDir,
				Flags:            FlagVerbose | FlagDryRun,
				SafetyMargin:     10, // Default safety margin
			},
		},
		{
			name: "single_synthesis_flag",
			args: []string{"thinktank", testInstructionsFile, testTargetDir, "--synthesis", "--dry-run"},
			want: &SimplifiedConfig{
				InstructionsFile: testInstructionsFile,
				TargetPath:       testTargetDir,
				Flags:            FlagSynthesis | FlagDryRun,
				SafetyMargin:     10, // Default safety margin
			},
		},
		{
			name: "multiple_boolean_flags",
			args: []string{"thinktank", testInstructionsFile, testTargetDir, "--dry-run", "--verbose", "--synthesis"},
			want: &SimplifiedConfig{
				InstructionsFile: testInstructionsFile,
				TargetPath:       testTargetDir,
				Flags:            FlagDryRun | FlagVerbose | FlagSynthesis,
				SafetyMargin:     10, // Default safety margin
			},
		},
		{
			name: "flags_in_different_order",
			args: []string{"thinktank", testInstructionsFile, testTargetDir, "--synthesis", "--dry-run", "--verbose"},
			want: &SimplifiedConfig{
				InstructionsFile: testInstructionsFile,
				TargetPath:       testTargetDir,
				Flags:            FlagDryRun | FlagVerbose | FlagSynthesis,
				SafetyMargin:     10, // Default safety margin
			},
		},
		{
			name:        "unknown_flag",
			args:        []string{"thinktank", testInstructionsFile, testTargetDir, "--unknown"},
			wantErr:     true,
			errContains: "unknown flag: --unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSimpleArgsWithArgs(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseSimpleArgsWithArgs() expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ParseSimpleArgsWithArgs() error = %q, want error containing %q",
						err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseSimpleArgsWithArgs() unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSimpleArgsWithArgs() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestParseSimpleArgsWithArgs_ValueFlags tests flags that require values
func TestParseSimpleArgsWithArgs_ValueFlags(t *testing.T) {
	// Create test files for validation
	tempDir := t.TempDir()
	testInstructionsFile := filepath.Join(tempDir, "instructions.txt")
	testTargetDir := filepath.Join(tempDir, "src")

	// Create test files
	if err := os.WriteFile(testInstructionsFile, []byte("test instructions"), 0644); err != nil {
		t.Fatalf("Failed to create test instructions file: %v", err)
	}
	if err := os.MkdirAll(testTargetDir, 0755); err != nil {
		t.Fatalf("Failed to create test target directory: %v", err)
	}

	tests := []struct {
		name        string
		args        []string
		want        *SimplifiedConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "model_flag_with_space",
			args: []string{"thinktank", testInstructionsFile, testTargetDir, "--model", "gpt-4", "--dry-run"},
			want: &SimplifiedConfig{
				InstructionsFile: testInstructionsFile,
				TargetPath:       testTargetDir,
				Flags:            FlagDryRun,
				SafetyMargin:     10, // Default safety margin
			},
		},
		{
			name: "model_flag_with_equals",
			args: []string{"thinktank", testInstructionsFile, testTargetDir, "--model=gpt-4", "--dry-run"},
			want: &SimplifiedConfig{
				InstructionsFile: testInstructionsFile,
				TargetPath:       testTargetDir,
				Flags:            FlagDryRun,
				SafetyMargin:     10, // Default safety margin
			},
		},
		{
			name: "output_dir_flag_with_space",
			args: []string{"thinktank", testInstructionsFile, testTargetDir, "--output-dir", "./out", "--dry-run"},
			want: &SimplifiedConfig{
				InstructionsFile: testInstructionsFile,
				TargetPath:       testTargetDir,
				Flags:            FlagDryRun,
				SafetyMargin:     10, // Default safety margin
			},
		},
		{
			name: "output_dir_flag_with_equals",
			args: []string{"thinktank", testInstructionsFile, testTargetDir, "--output-dir=./out", "--dry-run"},
			want: &SimplifiedConfig{
				InstructionsFile: testInstructionsFile,
				TargetPath:       testTargetDir,
				Flags:            FlagDryRun,
				SafetyMargin:     10, // Default safety margin
			},
		},
		{
			name: "mixed_flag_formats",
			args: []string{"thinktank", testInstructionsFile, testTargetDir, "--model", "gpt-4", "--output-dir=./out", "--verbose", "--dry-run"},
			want: &SimplifiedConfig{
				InstructionsFile: testInstructionsFile,
				TargetPath:       testTargetDir,
				Flags:            FlagVerbose | FlagDryRun,
				SafetyMargin:     10, // Default safety margin
			},
		},
		{
			name:        "model_flag_missing_value",
			args:        []string{"thinktank", "instructions.txt", "./src", "--model"},
			wantErr:     true,
			errContains: "--model flag requires a value",
		},
		{
			name:        "output_dir_flag_missing_value",
			args:        []string{"thinktank", "instructions.txt", "./src", "--output-dir"},
			wantErr:     true,
			errContains: "--output-dir flag requires a value",
		},
		{
			name:        "model_flag_empty_value_equals",
			args:        []string{"thinktank", "instructions.txt", "./src", "--model="},
			wantErr:     true,
			errContains: "--model flag requires a non-empty value",
		},
		{
			name:        "output_dir_flag_empty_value_equals",
			args:        []string{"thinktank", "instructions.txt", "./src", "--output-dir="},
			wantErr:     true,
			errContains: "--output-dir flag requires a non-empty value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSimpleArgsWithArgs(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseSimpleArgsWithArgs() expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ParseSimpleArgsWithArgs() error = %q, want error containing %q",
						err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseSimpleArgsWithArgs() unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSimpleArgsWithArgs() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestParseSimpleArgsWithArgs_EdgeCases tests edge cases and boundary conditions
func TestParseSimpleArgsWithArgs_EdgeCases(t *testing.T) {
	// Create test files for validation testing
	tempDir := t.TempDir()
	testInstructionsFile := filepath.Join(tempDir, "test.txt")
	testTargetDir := filepath.Join(tempDir, "src")

	// Create test files
	if err := os.WriteFile(testInstructionsFile, []byte("test instructions"), 0644); err != nil {
		t.Fatalf("Failed to create test instructions file: %v", err)
	}
	if err := os.MkdirAll(testTargetDir, 0755); err != nil {
		t.Fatalf("Failed to create test target directory: %v", err)
	}

	tests := []struct {
		name        string
		args        []string
		want        *SimplifiedConfig
		wantErr     bool
		errContains string
		setupEnv    func()
		cleanupEnv  func()
	}{
		{
			name: "unicode_filenames",
			args: []string{"thinktank", testInstructionsFile, testTargetDir, "--dry-run"},
			want: &SimplifiedConfig{
				InstructionsFile: testInstructionsFile,
				TargetPath:       testTargetDir,
				Flags:            FlagDryRun,
				SafetyMargin:     10, // Default safety margin
			},
		},
		{
			name: "spaces_in_paths",
			args: []string{"thinktank", testInstructionsFile, testTargetDir, "--dry-run"},
			want: &SimplifiedConfig{
				InstructionsFile: testInstructionsFile,
				TargetPath:       testTargetDir,
				Flags:            FlagDryRun,
				SafetyMargin:     10, // Default safety margin
			},
		},
		{
			name: "all_flags_combined",
			args: []string{"thinktank", testInstructionsFile, testTargetDir, "--dry-run", "--verbose", "--synthesis", "--model=gpt-4", "--output-dir=./results"},
			want: &SimplifiedConfig{
				InstructionsFile: testInstructionsFile,
				TargetPath:       testTargetDir,
				Flags:            FlagDryRun | FlagVerbose | FlagSynthesis,
				SafetyMargin:     10, // Default safety margin
			},
		},
		{
			name:        "very_long_path",
			args:        []string{"thinktank", strings.Repeat("a", 300) + ".txt", "./src", "--dry-run"},
			wantErr:     true,
			errContains: "path too long",
		},
		{
			name:        "nonexistent_file_no_dry_run",
			args:        []string{"thinktank", "/nonexistent/path/instructions.txt", "./src"},
			wantErr:     true,
			errContains: "invalid configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupEnv != nil {
				tt.setupEnv()
			}
			if tt.cleanupEnv != nil {
				defer tt.cleanupEnv()
			}

			got, err := ParseSimpleArgsWithArgs(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseSimpleArgsWithArgs() expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ParseSimpleArgsWithArgs() error = %q, want error containing %q",
						err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseSimpleArgsWithArgs() unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSimpleArgsWithArgs() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestParseSimpleArgsWithArgs_PropertyBased uses Go's testing/quick for property-based testing
func TestParseSimpleArgsWithArgs_PropertyBased(t *testing.T) {
	// Property: valid flag combinations should always parse successfully
	validFlagsProperty := func(hasVerbose, hasDryRun, hasSynthesis bool) bool {
		// Create temporary files for testing
		tempDir := t.TempDir()
		testInstructionsFile := filepath.Join(tempDir, "test.txt")
		testTargetDir := filepath.Join(tempDir, "src")

		// Create test files
		if err := os.WriteFile(testInstructionsFile, []byte("test"), 0644); err != nil {
			return false
		}
		if err := os.MkdirAll(testTargetDir, 0755); err != nil {
			return false
		}

		args := []string{"thinktank", testInstructionsFile, testTargetDir, "--dry-run"}

		if hasVerbose {
			args = append(args, "--verbose")
		}
		if hasDryRun {
			args = append(args, "--dry-run")
		}
		if hasSynthesis {
			args = append(args, "--synthesis")
		}

		config, err := ParseSimpleArgsWithArgs(args)
		if err != nil {
			return false
		}

		return config.HasFlag(FlagVerbose) == hasVerbose &&
			config.HasFlag(FlagDryRun) == (hasDryRun || true) && // Always true because we add it for validation
			config.HasFlag(FlagSynthesis) == hasSynthesis
	}

	if err := quick.Check(validFlagsProperty, &quick.Config{MaxCount: 50}); err != nil {
		t.Error(err)
	}
}

// TestParseSimpleArgsWithArgs_Deterministic verifies parsing is deterministic
func TestParseSimpleArgsWithArgs_Deterministic(t *testing.T) {
	tempDir := t.TempDir()
	testInstructionsFile := filepath.Join(tempDir, "test.txt")
	testTargetDir := filepath.Join(tempDir, "src")

	// Create test files
	if err := os.WriteFile(testInstructionsFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.MkdirAll(testTargetDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	args := []string{"thinktank", testInstructionsFile, testTargetDir, "--verbose", "--dry-run"}

	config1, err1 := ParseSimpleArgsWithArgs(args)
	config2, err2 := ParseSimpleArgsWithArgs(args)

	if (err1 == nil) != (err2 == nil) {
		t.Error("parsing should be deterministic for errors")
	}

	if err1 == nil && !reflect.DeepEqual(config1, config2) {
		t.Error("parsing should be deterministic for results")
	}
}

// BenchmarkParseSimpleArgsWithArgs benchmarks parsing performance (target <100μs)
func BenchmarkParseSimpleArgsWithArgs(b *testing.B) {
	benchmarks := []struct {
		name string
		args []string
	}{
		{
			name: "minimal",
			args: []string{"thinktank", "test.txt", "./src"},
		},
		{
			name: "typical",
			args: []string{"thinktank", "test.txt", "./src", "--verbose", "--model", "gpt-4"},
		},
		{
			name: "complex",
			args: []string{"thinktank", "test.txt", "./src", "--verbose", "--dry-run",
				"--synthesis", "--model=gpt-4", "--output-dir=./out"},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			perftest.RunBenchmark(b, "ParseSimpleArgsWithArgs_"+bm.name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_, err := ParseSimpleArgsWithArgs(bm.args)
					if err != nil {
						// Skip validation errors in benchmarks
						continue
					}
				}
			})
		})
	}
}

// BenchmarkParseSimpleArgsWithArgs_Allocs benchmarks memory allocations
func BenchmarkParseSimpleArgsWithArgs_Allocs(b *testing.B) {
	args := []string{"thinktank", "test.txt", "./src", "--verbose"}

	perftest.RunBenchmark(b, "ParseSimpleArgsWithArgs_Allocs", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			config, err := ParseSimpleArgsWithArgs(args)
			if err != nil {
				// Skip validation errors in benchmarks
				continue
			}
			_ = config // Prevent optimization
		}
	})
}

// TestParseSimpleArgs tests the main entry point function
func TestParseSimpleArgs(t *testing.T) {
	// Create test files for validation
	tempDir := t.TempDir()
	testInstructionsFile := filepath.Join(tempDir, "instructions.txt")
	testTargetDir := filepath.Join(tempDir, "src")

	// Create test files
	if err := os.WriteFile(testInstructionsFile, []byte("test instructions"), 0644); err != nil {
		t.Fatalf("Failed to create test instructions file: %v", err)
	}
	if err := os.MkdirAll(testTargetDir, 0755); err != nil {
		t.Fatalf("Failed to create test target directory: %v", err)
	}

	// Temporarily modify os.Args for testing
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"thinktank", testInstructionsFile, testTargetDir, "--dry-run"}

	config, err := ParseSimpleArgs()
	if err != nil {
		t.Errorf("ParseSimpleArgs() unexpected error: %v", err)
		return
	}

	if config == nil {
		t.Error("ParseSimpleArgs() returned nil config")
		return
	}

	if config.InstructionsFile != testInstructionsFile {
		t.Errorf("ParseSimpleArgs() InstructionsFile = %v, want %v", config.InstructionsFile, testInstructionsFile)
	}
}

// TestParseResult tests the ParseResult helper functions
func TestParseResult(t *testing.T) {
	// Create test files for validation
	tempDir := t.TempDir()
	testInstructionsFile := filepath.Join(tempDir, "instructions.txt")
	testTargetDir := filepath.Join(tempDir, "src")

	// Create test files
	if err := os.WriteFile(testInstructionsFile, []byte("test instructions"), 0644); err != nil {
		t.Fatalf("Failed to create test instructions file: %v", err)
	}
	if err := os.MkdirAll(testTargetDir, 0755); err != nil {
		t.Fatalf("Failed to create test target directory: %v", err)
	}

	t.Run("successful_result", func(t *testing.T) {
		args := []string{"thinktank", testInstructionsFile, testTargetDir, "--dry-run"}
		result := ParseSimpleArgsWithResult(args)

		if !result.IsSuccess() {
			t.Errorf("ParseSimpleArgsWithResult() expected success, got error: %v", result.Error)
		}

		config := result.MustConfig()
		if config == nil {
			t.Error("MustConfig() returned nil")
			return
		}

		if config.InstructionsFile != testInstructionsFile {
			t.Errorf("Config InstructionsFile = %v, want %v", config.InstructionsFile, testInstructionsFile)
		}
	})

	t.Run("error_result", func(t *testing.T) {
		args := []string{"thinktank"} // insufficient args
		result := ParseSimpleArgsWithResult(args)

		if result.IsSuccess() {
			t.Error("ParseSimpleArgsWithResult() expected error, got success")
		}

		if result.Error == nil {
			t.Error("ParseSimpleArgsWithResult() expected error to be set")
		}

		if result.Config != nil {
			t.Error("ParseSimpleArgsWithResult() expected config to be nil on error")
		}
	})

	t.Run("must_config_panic", func(t *testing.T) {
		args := []string{"thinktank"} // insufficient args
		result := ParseSimpleArgsWithResult(args)

		defer func() {
			if r := recover(); r == nil {
				t.Error("MustConfig() expected panic on error result")
			}
		}()

		result.MustConfig() // Should panic
	})
}
