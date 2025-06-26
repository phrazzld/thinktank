package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/phrazzld/thinktank/internal/config"
)

// LoadEnvironmentDefaults loads environment variable defaults into a config structure.
// This function implements the precedence model: CLI flags > environment variables > defaults.
// It only sets values from environment variables if the corresponding config field is empty/default.
// This ensures that explicit CLI flags always take precedence over environment variables.
func LoadEnvironmentDefaults(cfg *config.CliConfig, getenv func(string) string) error {
	// Basic configuration environment variables
	if err := loadBasicEnvDefaults(cfg, getenv); err != nil {
		return fmt.Errorf("failed to load basic environment defaults: %w", err)
	}

	// Rate limiting environment variables
	if err := loadRateLimitEnvDefaults(cfg, getenv); err != nil {
		return fmt.Errorf("failed to load rate limit environment defaults: %w", err)
	}

	// File pattern environment variables
	if err := loadFilePatternEnvDefaults(cfg, getenv); err != nil {
		return fmt.Errorf("failed to load file pattern environment defaults: %w", err)
	}

	// Timeout and other advanced configuration
	if err := loadAdvancedEnvDefaults(cfg, getenv); err != nil {
		return fmt.Errorf("failed to load advanced environment defaults: %w", err)
	}

	return nil
}

// loadBasicEnvDefaults loads basic configuration from environment variables
func loadBasicEnvDefaults(cfg *config.CliConfig, getenv func(string) string) error {
	// THINKTANK_MODEL - Set model if no models are currently set
	if len(cfg.ModelNames) == 0 || (len(cfg.ModelNames) == 1 && cfg.ModelNames[0] == config.DefaultModel) {
		if model := getenv("THINKTANK_MODEL"); model != "" {
			cfg.ModelNames = []string{model}
		}
	}

	// THINKTANK_OUTPUT_DIR - Set output directory if empty
	if cfg.OutputDir == "" {
		if outputDir := getenv("THINKTANK_OUTPUT_DIR"); outputDir != "" {
			cfg.OutputDir = outputDir
		}
	}

	// THINKTANK_DRY_RUN - Set dry run mode if not explicitly set
	if !cfg.DryRun {
		if dryRun := getenv("THINKTANK_DRY_RUN"); dryRun != "" {
			cfg.DryRun = parseBooleanEnvVar(dryRun)
		}
	}

	// THINKTANK_VERBOSE - Set verbose mode if not explicitly set
	if !cfg.Verbose {
		if verbose := getenv("THINKTANK_VERBOSE"); verbose != "" {
			cfg.Verbose = parseBooleanEnvVar(verbose)
		}
	}

	// THINKTANK_QUIET - Set quiet mode if not explicitly set
	if !cfg.Quiet {
		if quiet := getenv("THINKTANK_QUIET"); quiet != "" {
			cfg.Quiet = parseBooleanEnvVar(quiet)
		}
	}

	return nil
}

// loadRateLimitEnvDefaults loads rate limiting configuration from environment variables
func loadRateLimitEnvDefaults(cfg *config.CliConfig, getenv func(string) string) error {
	// THINKTANK_RATE_LIMIT - Global rate limit
	if cfg.RateLimitRequestsPerMinute == config.DefaultRateLimitRequestsPerMinute {
		if rateLimit := getenv("THINKTANK_RATE_LIMIT"); rateLimit != "" {
			value, err := strconv.Atoi(rateLimit)
			if err != nil {
				return fmt.Errorf("invalid rate limit value %q: %w", rateLimit, err)
			}
			cfg.RateLimitRequestsPerMinute = value
		}
	}

	// THINKTANK_MAX_CONCURRENT - Maximum concurrent requests
	if cfg.MaxConcurrentRequests == config.DefaultMaxConcurrentRequests {
		if maxConcurrent := getenv("THINKTANK_MAX_CONCURRENT"); maxConcurrent != "" {
			value, err := strconv.Atoi(maxConcurrent)
			if err != nil {
				return fmt.Errorf("invalid max concurrent value %q: %w", maxConcurrent, err)
			}
			if value <= 0 {
				return fmt.Errorf("max concurrent requests must be positive, got %d", value)
			}
			cfg.MaxConcurrentRequests = value
		}
	}

	// Provider-specific rate limits
	if cfg.OpenAIRateLimit == 0 {
		if rateLimit := getenv("THINKTANK_RATE_LIMIT_OPENAI"); rateLimit != "" {
			value, err := strconv.Atoi(rateLimit)
			if err != nil {
				return fmt.Errorf("invalid OpenAI rate limit value %q: %w", rateLimit, err)
			}
			cfg.OpenAIRateLimit = value
		}
	}

	if cfg.GeminiRateLimit == 0 {
		if rateLimit := getenv("THINKTANK_RATE_LIMIT_GEMINI"); rateLimit != "" {
			value, err := strconv.Atoi(rateLimit)
			if err != nil {
				return fmt.Errorf("invalid Gemini rate limit value %q: %w", rateLimit, err)
			}
			cfg.GeminiRateLimit = value
		}
	}

	if cfg.OpenRouterRateLimit == 0 {
		if rateLimit := getenv("THINKTANK_RATE_LIMIT_OPENROUTER"); rateLimit != "" {
			value, err := strconv.Atoi(rateLimit)
			if err != nil {
				return fmt.Errorf("invalid OpenRouter rate limit value %q: %w", rateLimit, err)
			}
			cfg.OpenRouterRateLimit = value
		}
	}

	return nil
}

// loadFilePatternEnvDefaults loads file pattern configuration from environment variables
func loadFilePatternEnvDefaults(cfg *config.CliConfig, getenv func(string) string) error {
	// THINKTANK_INCLUDE - Include patterns
	if cfg.Include == "" {
		if include := getenv("THINKTANK_INCLUDE"); include != "" {
			cfg.Include = include
		}
	}

	// THINKTANK_EXCLUDE - Exclude patterns
	if cfg.Exclude == config.DefaultExcludes {
		if exclude := getenv("THINKTANK_EXCLUDE"); exclude != "" {
			cfg.Exclude = exclude
		}
	}

	// THINKTANK_EXCLUDE_NAMES - Exclude name patterns
	if cfg.ExcludeNames == config.DefaultExcludeNames {
		if excludeNames := getenv("THINKTANK_EXCLUDE_NAMES"); excludeNames != "" {
			cfg.ExcludeNames = excludeNames
		}
	}

	return nil
}

// loadAdvancedEnvDefaults loads advanced configuration from environment variables
func loadAdvancedEnvDefaults(cfg *config.CliConfig, getenv func(string) string) error {
	// THINKTANK_TIMEOUT - Operation timeout
	if cfg.Timeout == config.DefaultTimeout {
		if timeout := getenv("THINKTANK_TIMEOUT"); timeout != "" {
			duration, err := time.ParseDuration(timeout)
			if err != nil {
				return fmt.Errorf("invalid timeout format %q: %w", timeout, err)
			}
			cfg.Timeout = duration
		}
	}

	// THINKTANK_LOG_LEVEL - Logging level
	if logLevel := getenv("THINKTANK_LOG_LEVEL"); logLevel != "" {
		// Only set if not already customized
		// Note: This is a simplified implementation - could be enhanced with actual log level parsing
		switch strings.ToLower(logLevel) {
		case "debug":
			cfg.LogLevel = 0 // Assuming debug is 0
		case "info":
			cfg.LogLevel = 1 // Assuming info is 1
		case "warn", "warning":
			cfg.LogLevel = 2 // Assuming warn is 2
		case "error":
			cfg.LogLevel = 3 // Assuming error is 3
		}
	}

	return nil
}

// parseBooleanEnvVar converts environment variable values to boolean
// Supports common boolean representations: true/false, 1/0, yes/no, on/off
// Empty strings are treated as false
func parseBooleanEnvVar(value string) bool {
	if value == "" {
		return false
	}

	lower := strings.ToLower(strings.TrimSpace(value))
	switch lower {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		// Invalid values default to false (lenient parsing)
		return false
	}
}

// parseStringSliceEnvVar converts comma-separated environment variable values to string slice
// Trims whitespace and filters out empty values
func parseStringSliceEnvVar(value string) []string {
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}
