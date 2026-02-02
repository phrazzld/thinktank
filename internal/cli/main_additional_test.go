package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/misty-step/thinktank/internal/config"
	"github.com/misty-step/thinktank/internal/logutil"
	"github.com/misty-step/thinktank/internal/thinktank"
)

func TestApplyEnvironmentVars(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		config  *config.MinimalConfig
		wantErr bool
	}{
		{
			name: "valid config - no env vars to apply",
			config: &config.MinimalConfig{
				InstructionsFile: "test.txt",
				TargetPaths:      []string{"test"},
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: false, // Function currently returns nil for any input
		},
		{
			name:    "empty config",
			config:  &config.MinimalConfig{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := applyEnvironmentVars(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyEnvironmentVars() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfigAdditionalCases(t *testing.T) {
	t.Parallel()
	// Create temporary files for testing
	tempDir := t.TempDir()
	instructionsFile := filepath.Join(tempDir, "instructions.txt")
	targetFile := filepath.Join(tempDir, "target.txt")

	// Create test files
	if err := os.WriteFile(instructionsFile, []byte("test instructions"), 0644); err != nil {
		t.Fatalf("Failed to create test instructions file: %v", err)
	}
	if err := os.WriteFile(targetFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test target file: %v", err)
	}

	tests := []struct {
		name          string
		config        *config.MinimalConfig
		wantErr       bool
		errorContains string
	}{
		{
			name: "nonexistent instructions file",
			config: &config.MinimalConfig{
				InstructionsFile: "/nonexistent/instructions.txt",
				TargetPaths:      []string{targetFile},
			},
			wantErr:       true,
			errorContains: "instructions file not found",
		},
		{
			name: "nonexistent target path",
			config: &config.MinimalConfig{
				InstructionsFile: instructionsFile,
				TargetPaths:      []string{"/nonexistent/path"},
			},
			wantErr:       true,
			errorContains: "target path not found",
		},
		{
			name: "multiple target paths - one invalid",
			config: &config.MinimalConfig{
				InstructionsFile: instructionsFile,
				TargetPaths:      []string{targetFile, "/nonexistent/path"},
			},
			wantErr:       true,
			errorContains: "target path not found",
		},
		{
			name: "all valid paths",
			config: &config.MinimalConfig{
				InstructionsFile: instructionsFile,
				TargetPaths:      []string{targetFile, tempDir},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errorContains != "" {
				if err == nil || !contains(err.Error(), tt.errorContains) {
					t.Errorf("validateConfig() error = %v, want error containing %q", err, tt.errorContains)
				}
			}
		})
	}
}

func TestSetupGracefulShutdownComplete(t *testing.T) {
	t.Parallel()
	logger := logutil.NewSlogLoggerFromLogLevel(nil, logutil.InfoLevel)

	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "background context",
			ctx:  context.Background(),
		},
		{
			name: "context with timeout",
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				return ctx
			}(),
		},
		{
			name: "already cancelled context",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := setupGracefulShutdown(tt.ctx, logger)
			if result == nil {
				t.Error("setupGracefulShutdown() returned nil context")
				return
			}

			// Verify we get a context that can be cancelled
			if result == tt.ctx {
				t.Error("setupGracefulShutdown() returned the same context - should return a new cancellable context")
			}

			// Test that context operations work
			select {
			case <-result.Done():
				// If original context was already cancelled, result should be done too
				if tt.ctx.Err() == nil {
					t.Error("result context is done but original context was not cancelled")
				}
			default:
				// Expected for non-cancelled contexts
				if tt.ctx.Err() != nil {
					t.Error("result context not done but original context was cancelled")
				}
			}
		})
	}
}

func TestRunApplicationDryRun(t *testing.T) {
	t.Parallel()
	// Create temporary files for testing
	tempDir := t.TempDir()
	instructionsFile := filepath.Join(tempDir, "instructions.txt")
	targetFile := filepath.Join(tempDir, "target.txt")

	// Create test files
	if err := os.WriteFile(instructionsFile, []byte("test instructions"), 0644); err != nil {
		t.Fatalf("Failed to create test instructions file: %v", err)
	}
	if err := os.WriteFile(targetFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test target file: %v", err)
	}

	logger := logutil.NewSlogLoggerFromLogLevel(nil, logutil.InfoLevel)
	ctx := context.Background()

	tests := []struct {
		name    string
		config  *config.MinimalConfig
		wantErr bool
	}{
		{
			name: "valid dry run config",
			config: &config.MinimalConfig{
				InstructionsFile: instructionsFile,
				TargetPaths:      []string{targetFile},
				ModelNames:       []string{"gpt-5.2"},
				OutputDir:        tempDir,
				DryRun:           true,
			},
			wantErr: false,
		},
		{
			name: "invalid config - missing instructions",
			config: &config.MinimalConfig{
				InstructionsFile: "",
				TargetPaths:      []string{targetFile},
				OutputDir:        tempDir,
				DryRun:           true,
			},
			wantErr: true,
		},
		{
			name: "invalid config - nonexistent instructions file",
			config: &config.MinimalConfig{
				InstructionsFile: "/nonexistent/file.txt",
				TargetPaths:      []string{targetFile},
				OutputDir:        tempDir,
				DryRun:           true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock token service for testing
			tokenService := thinktank.NewTokenCountingService()
			err := runApplication(ctx, tt.config, logger, tokenService, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("runApplication() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(substr) > 0 && findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
