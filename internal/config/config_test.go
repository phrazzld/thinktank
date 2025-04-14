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

// TestValidateConfig is a future test we'll implement when we move the validation logic
func TestValidateConfig(t *testing.T) {
	// Skipping this test as it will be implemented in the next task
	t.Skip("Will be implemented in the next task")
}
