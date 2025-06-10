package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// BenchmarkParseFlags benchmarks the CLI flag parsing functionality
func BenchmarkParseFlags(b *testing.B) {
	// Save original args and restore after test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	benchmarks := []struct {
		name string
		args []string
	}{
		{
			name: "MinimalArgs",
			args: []string{"thinktank", "--instructions", "test.md", "file.go"},
		},
		{
			name: "TypicalArgs",
			args: []string{
				"thinktank",
				"--instructions", "instructions.md",
				"--output-dir", "./output",
				"--model", "gemini-2.5-pro",
				"--include", "*.go,*.md",
				"--exclude", "*.test.go",
				"--verbose",
				"--dry-run",
				"./src",
			},
		},
		{
			name: "ComplexArgs",
			args: []string{
				"thinktank",
				"--instructions", "instructions.md",
				"--output-dir", "./output",
				"--model", "gemini-2.5-pro",
				"--model", "o4-mini",
				"--include", "*.go,*.md,*.txt",
				"--exclude", "*.test.go,*.bench.go",
				"--exclude-names", "vendor,node_modules",
				"--verbose",
				"--dry-run",
				"--confirm-tokens", "5000",
				"--log-level", "debug",
				"--audit-log-file", "audit.log",
				"--force-overwrite",
				"--partial-success-ok",
				"./src", "./docs", "./config",
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				os.Args = bm.args
				_, err := ParseFlags()
				if err != nil {
					b.Fatalf("ParseFlags failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkValidateInputs benchmarks the input validation functionality
func BenchmarkValidateInputs(b *testing.B) {
	logger := &testLogger{}

	benchmarks := []struct {
		name   string
		config *config.CliConfig
	}{
		{
			name: "ValidConfig",
			config: &config.CliConfig{
				InstructionsFile: "test.md",
				Paths:            []string{"file.go"},
				ModelNames:       []string{"gemini-2.5-pro"},
				OutputDir:        "./output",
				Timeout:          30 * time.Second,
				Include:          "*.go",
				Exclude:          "*.test.go",
				ExcludeNames:     "vendor",
				LogLevel:         logutil.InfoLevel,
				AuditLogFile:     "",
				DryRun:           false,
				Verbose:          false,
				// ForceOverwrite field removed
			},
		},
		{
			name: "MultiModelConfig",
			config: &config.CliConfig{
				InstructionsFile: "test.md",
				Paths:            []string{"./src", "./docs"},
				ModelNames:       []string{"gemini-2.5-pro", "o4-mini", "gemini-2.5-flash"},
				OutputDir:        "./output",
				Timeout:          60 * time.Second,
				Include:          "*.go,*.md,*.txt",
				Exclude:          "*.test.go,*.bench.go",
				ExcludeNames:     "vendor,node_modules,dist",
				LogLevel:         logutil.DebugLevel,
				AuditLogFile:     "audit.log",
				DryRun:           true,
				Verbose:          true,
				// ForceOverwrite field removed
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = ValidateInputs(bm.config, logger)
			}
		})
	}
}

// BenchmarkSetupLogging benchmarks the logging setup functionality
func BenchmarkSetupLogging(b *testing.B) {
	benchmarks := []struct {
		name   string
		config *config.CliConfig
	}{
		{
			name: "InfoLevel",
			config: &config.CliConfig{
				LogLevel: logutil.InfoLevel,
				Verbose:  false,
			},
		},
		{
			name: "DebugLevel",
			config: &config.CliConfig{
				LogLevel: logutil.DebugLevel,
				Verbose:  true,
			},
		},
		{
			name: "WarnLevel",
			config: &config.CliConfig{
				LogLevel: logutil.WarnLevel,
				Verbose:  false,
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = SetupLogging(bm.config)
			}
		})
	}
}

// testLogger is a minimal logger implementation for benchmarking
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
func (tl *testLogger) WithContext(ctx context.Context) logutil.LoggerInterface {
	return tl
}
