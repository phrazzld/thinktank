package config

import (
	"testing"

	"github.com/phrazzld/architect/internal/logutil"
)

func TestNewDefaultCliConfig(t *testing.T) {
	cfg := NewDefaultCliConfig()

	// Verify default values
	if cfg.Format != DefaultFormat {
		t.Errorf("Expected Format to be %q, got %q", DefaultFormat, cfg.Format)
	}

	if cfg.Exclude != DefaultExcludes {
		t.Errorf("Expected Exclude to be %q, got %q", DefaultExcludes, cfg.Exclude)
	}

	if cfg.ExcludeNames != DefaultExcludeNames {
		t.Errorf("Expected ExcludeNames to be %q, got %q", DefaultExcludeNames, cfg.ExcludeNames)
	}

	if len(cfg.ModelNames) != 1 || cfg.ModelNames[0] != DefaultModel {
		t.Errorf("Expected ModelNames to be [%q], got %v", DefaultModel, cfg.ModelNames)
	}

	if cfg.LogLevel != logutil.InfoLevel {
		t.Errorf("Expected LogLevel to be %v, got %v", logutil.InfoLevel, cfg.LogLevel)
	}

	if cfg.MaxConcurrentRequests != DefaultMaxConcurrentRequests {
		t.Errorf("Expected MaxConcurrentRequests to be %d, got %d", DefaultMaxConcurrentRequests, cfg.MaxConcurrentRequests)
	}

	if cfg.RateLimitRequestsPerMinute != DefaultRateLimitRequestsPerMinute {
		t.Errorf("Expected RateLimitRequestsPerMinute to be %d, got %d", DefaultRateLimitRequestsPerMinute, cfg.RateLimitRequestsPerMinute)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Verify default values for all fields
	if cfg.OutputFile != DefaultOutputFile {
		t.Errorf("Expected OutputFile to be %q, got %q", DefaultOutputFile, cfg.OutputFile)
	}

	if cfg.ModelName != DefaultModel {
		t.Errorf("Expected ModelName to be %q, got %q", DefaultModel, cfg.ModelName)
	}

	if cfg.Format != DefaultFormat {
		t.Errorf("Expected Format to be %q, got %q", DefaultFormat, cfg.Format)
	}

	if cfg.LogLevel != logutil.InfoLevel {
		t.Errorf("Expected LogLevel to be %v, got %v", logutil.InfoLevel, cfg.LogLevel)
	}

	if cfg.ConfirmTokens != 0 {
		t.Errorf("Expected ConfirmTokens to be 0, got %d", cfg.ConfirmTokens)
	}

	// Verify nested ExcludeConfig values
	if cfg.Excludes.Extensions != DefaultExcludes {
		t.Errorf("Expected Excludes.Extensions to be %q, got %q", DefaultExcludes, cfg.Excludes.Extensions)
	}

	if cfg.Excludes.Names != DefaultExcludeNames {
		t.Errorf("Expected Excludes.Names to be %q, got %q", DefaultExcludeNames, cfg.Excludes.Names)
	}

	// Verify default values for fields that should be empty or zero
	if cfg.Include != "" {
		t.Errorf("Expected Include to be empty, got %q", cfg.Include)
	}

	if cfg.Verbose != false {
		t.Errorf("Expected Verbose to be false, got %v", cfg.Verbose)
	}
}

// TestValidateConfig is a future test we'll implement when we move the validation logic
func TestValidateConfig(t *testing.T) {
	// Skipping this test as it will be implemented in the next task
	t.Skip("Will be implemented in the next task")
}
