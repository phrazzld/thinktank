package main

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/testutil/perftest"
)

// BenchmarkParseFlags benchmarks the CLI flag parsing functionality
func BenchmarkParseFlags(b *testing.B) {
	benchmarks := []struct {
		name string
		args []string
	}{
		{
			name: "MinimalArgs",
			args: []string{"--instructions", "test.md", "file.go"},
		},
		{
			name: "TypicalArgs",
			args: []string{
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
				"--instructions", "instructions.md",
				"--output-dir", "./output",
				"--model", "gemini-2.5-pro",
				"--model", "o4-mini",
				"--include", "*.go,*.md,*.txt",
				"--exclude", "*.test.go,*.bench.go",
				"--exclude-names", "vendor,node_modules",
				"--verbose",
				"--dry-run",
				"--log-level", "debug",
				"--audit-log-file", "audit.log",
				"--partial-success-ok",
				"--max-concurrent", "3",
				"--rate-limit", "30",
				"--timeout", "5m",
				"./src", "./docs", "./config",
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			perftest.RunBenchmark(b, "ParseFlags_"+bm.name, func(b *testing.B) {
				perftest.ReportAllocs(b)
				for i := 0; i < b.N; i++ {
					flagSet := flag.NewFlagSet("thinktank", flag.ContinueOnError)
					_, err := ParseFlagsWithEnv(flagSet, bm.args, os.Getenv)
					if err != nil {
						b.Fatalf("ParseFlagsWithEnv failed: %v", err)
					}
				}
			})
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
			perftest.RunBenchmark(b, "ValidateInputs_"+bm.name, func(b *testing.B) {
				perftest.ReportAllocs(b)
				for i := 0; i < b.N; i++ {
					_ = ValidateInputs(bm.config, logger)
				}
			})
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
			perftest.RunBenchmark(b, "SetupLogging_"+bm.name, func(b *testing.B) {
				perftest.ReportAllocs(b)
				for i := 0; i < b.N; i++ {
					_ = SetupLogging(bm.config)
				}
			})
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

// runBenchmark simulates a full processing lifecycle for benchmarking.
func runBenchmark(b *testing.B, isInteractive, noProgress bool) {
	cw := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return isInteractive },
	})
	cw.SetNoProgress(noProgress)

	perftest.ReportAllocs(b)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cw.StartProcessing(10)
		for j := 1; j <= 10; j++ {
			cw.ModelStarted(j, 10, "model")
			if j%3 == 0 {
				cw.ModelFailed(j, 10, "model", "simulated error")
			} else {
				cw.ModelCompleted(j, 10, "model", 1500*time.Millisecond)
			}
		}
		cw.SynthesisStarted()
		cw.SynthesisCompleted("output/path")
	}
}

// BenchmarkConsoleWriterLifecycle benchmarks the full ConsoleWriter lifecycle
func BenchmarkConsoleWriterLifecycle(b *testing.B) {
	b.Run("Interactive-WithProgress", func(b *testing.B) {
		perftest.RunBenchmark(b, "ConsoleWriterLifecycle_Interactive_WithProgress", func(b *testing.B) {
			runBenchmark(b, true, false)
		})
	})
	b.Run("Interactive-NoProgress", func(b *testing.B) {
		perftest.RunBenchmark(b, "ConsoleWriterLifecycle_Interactive_NoProgress", func(b *testing.B) {
			runBenchmark(b, true, true)
		})
	})
	b.Run("CI_Mode", func(b *testing.B) {
		perftest.RunBenchmark(b, "ConsoleWriterLifecycle_CI_Mode", func(b *testing.B) {
			runBenchmark(b, false, false) // NoProgress is implicit in CI
		})
	})
}

// BenchmarkConsoleWriterConcurrent benchmarks concurrent ConsoleWriter operations
func BenchmarkConsoleWriterConcurrent(b *testing.B) {
	perftest.RunBenchmark(b, "ConsoleWriterConcurrent", func(b *testing.B) {
		cw := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
			IsTerminalFunc: func() bool { return true },
		})
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				cw.ModelCompleted(1, 1, "concurrent-model", 1*time.Second)
			}
		})
	})
}

// BenchmarkSetupLoggingNew benchmarks the new logging setup with ConsoleWriter
func BenchmarkSetupLoggingNew(b *testing.B) {
	benchmarks := []struct {
		name   string
		config *config.CliConfig
	}{
		{
			name: "FileLogging",
			config: &config.CliConfig{
				LogLevel:  logutil.InfoLevel,
				Verbose:   false,
				JsonLogs:  false,
				OutputDir: "/tmp",
			},
		},
		{
			name: "StderrLogging_JsonLogs",
			config: &config.CliConfig{
				LogLevel:  logutil.InfoLevel,
				JsonLogs:  true,
				OutputDir: "/tmp",
			},
		},
		{
			name: "StderrLogging_Verbose",
			config: &config.CliConfig{
				LogLevel:  logutil.DebugLevel,
				Verbose:   true,
				OutputDir: "/tmp",
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			perftest.RunBenchmark(b, "SetupLoggingNew_"+bm.name, func(b *testing.B) {
				perftest.ReportAllocs(b)
				for i := 0; i < b.N; i++ {
					_ = SetupLogging(bm.config)
				}
			})
		})
	}
}

// BenchmarkConsoleWriterMethods benchmarks individual ConsoleWriter methods
func BenchmarkConsoleWriterMethods(b *testing.B) {
	cw := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return true },
	})

	b.Run("StartProcessing", func(b *testing.B) {
		perftest.RunBenchmark(b, "ConsoleWriterMethods_StartProcessing", func(b *testing.B) {
			perftest.ReportAllocs(b)
			for i := 0; i < b.N; i++ {
				cw.StartProcessing(10)
			}
		})
	})

	b.Run("ModelCompleted_Success", func(b *testing.B) {
		perftest.RunBenchmark(b, "ConsoleWriterMethods_ModelCompleted_Success", func(b *testing.B) {
			perftest.ReportAllocs(b)
			for i := 0; i < b.N; i++ {
				cw.ModelCompleted(1, 1, "test-model", 500*time.Millisecond)
			}
		})
	})

	b.Run("ModelFailed_Error", func(b *testing.B) {
		perftest.RunBenchmark(b, "ConsoleWriterMethods_ModelFailed_Error", func(b *testing.B) {
			perftest.ReportAllocs(b)
			for i := 0; i < b.N; i++ {
				cw.ModelFailed(1, 1, "test-model", "benchmark error")
			}
		})
	})

	b.Run("StatusMessage", func(b *testing.B) {
		perftest.RunBenchmark(b, "ConsoleWriterMethods_StatusMessage", func(b *testing.B) {
			perftest.ReportAllocs(b)
			for i := 0; i < b.N; i++ {
				cw.StatusMessage("Processing files...")
			}
		})
	})
}
