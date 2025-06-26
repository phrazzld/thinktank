package cli

import (
	"os"
	"time"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/models"
)

// ConfigAdapter provides an adapter pattern for converting SimplifiedConfig
// to the comprehensive CliConfig structure with intelligent defaults.
// This follows the existing ToCliConfig() pattern but with enhanced
// default assignment and validation behavior preservation.
type ConfigAdapter struct {
	simplified *SimplifiedConfig
	baseConfig *config.CliConfig // Optional base config with pre-loaded environment defaults
}

// NewConfigAdapter creates a new config adapter for the given simplified config
func NewConfigAdapter(simplified *SimplifiedConfig) *ConfigAdapter {
	return &ConfigAdapter{
		simplified: simplified,
		baseConfig: nil,
	}
}

// NewConfigAdapterWithBase creates a new config adapter with a pre-configured base config
// This is used when environment variables have been pre-loaded into the base config
func NewConfigAdapterWithBase(simplified *SimplifiedConfig, baseConfig *config.CliConfig) *ConfigAdapter {
	return &ConfigAdapter{
		simplified: simplified,
		baseConfig: baseConfig,
	}
}

// ToComplexConfig converts the simplified configuration to a full CliConfig
// with intelligent defaults for unmapped fields. This preserves all existing
// validation behavior while providing enhanced default assignment.
func (ca *ConfigAdapter) ToComplexConfig() *config.CliConfig {
	var cfg *config.CliConfig

	if ca.baseConfig != nil {
		// Start with environment-loaded base config
		cfg = ca.copyConfig(ca.baseConfig)
		// Apply simplified config values on top of environment defaults
		ca.applySimplifiedConfigOverrides(cfg)
	} else {
		// Use existing logic for backward compatibility
		cfg = ca.simplified.ToCliConfig()
	}

	// Apply intelligent defaults for unmapped fields
	ca.applyIntelligentDefaults(cfg)

	// Set API key from environment for compatibility with existing validation
	ca.applyAPIKeyDefaults(cfg)

	return cfg
}

// copyConfig creates a copy of the given CliConfig
func (ca *ConfigAdapter) copyConfig(src *config.CliConfig) *config.CliConfig {
	dst := *src // Shallow copy

	// Deep copy slices to avoid sharing
	if src.ModelNames != nil {
		dst.ModelNames = make([]string, len(src.ModelNames))
		copy(dst.ModelNames, src.ModelNames)
	}

	if src.Paths != nil {
		dst.Paths = make([]string, len(src.Paths))
		copy(dst.Paths, src.Paths)
	}

	return &dst
}

// applySimplifiedConfigOverrides applies the simplified config values on top of the base config
func (ca *ConfigAdapter) applySimplifiedConfigOverrides(cfg *config.CliConfig) {
	// Apply simplified config fields that override environment defaults
	cfg.InstructionsFile = ca.simplified.InstructionsFile
	cfg.Paths = []string{ca.simplified.TargetPath}

	// Apply flags that were explicitly set in simplified config
	if ca.simplified.HasFlag(FlagDryRun) {
		cfg.DryRun = true
	}

	if ca.simplified.HasFlag(FlagVerbose) {
		cfg.Verbose = true
	}

	if ca.simplified.HasFlag(FlagSynthesis) {
		// Use gemini-2.5-pro as the default synthesis model for simplified interface
		cfg.SynthesisModel = "gemini-2.5-pro"
	}
}

// applyIntelligentDefaults applies context-aware defaults based on the
// simplified configuration flags and detected environment
func (ca *ConfigAdapter) applyIntelligentDefaults(cfg *config.CliConfig) {
	// Apply intelligent rate limiting based on usage mode
	ca.applyRateLimitDefaults(cfg)

	// Apply intelligent timeout defaults
	ca.applyTimeoutDefaults(cfg)

	// Apply intelligent concurrency defaults
	ca.applyConcurrencyDefaults(cfg)
}

// applyRateLimitDefaults sets intelligent rate limits based on synthesis mode
// Only applies defaults if the values are still at their default (0) settings
func (ca *ConfigAdapter) applyRateLimitDefaults(cfg *config.CliConfig) {
	if ca.simplified.HasFlag(FlagSynthesis) {
		// Synthesis mode: more conservative to avoid rate limits across providers
		// Apply 60% reduction to provider defaults
		if cfg.OpenAIRateLimit == 0 {
			cfg.OpenAIRateLimit = int(float64(getProviderDefaultOrFallback("openai", 3000)) * 0.6) // 60% of OpenAI default
		}
		if cfg.GeminiRateLimit == 0 {
			cfg.GeminiRateLimit = int(float64(getProviderDefaultOrFallback("gemini", 60)) * 0.6) // 60% of Gemini default
		}
		if cfg.OpenRouterRateLimit == 0 {
			cfg.OpenRouterRateLimit = int(float64(getProviderDefaultOrFallback("openrouter", 20)) * 0.6) // 60% of OpenRouter default
		}
	} else {
		// Single model mode: use standard provider defaults only if not set
		if cfg.OpenAIRateLimit == 0 {
			cfg.OpenAIRateLimit = getProviderDefaultOrFallback("openai", 3000)
		}
		if cfg.GeminiRateLimit == 0 {
			cfg.GeminiRateLimit = getProviderDefaultOrFallback("gemini", 60)
		}
		if cfg.OpenRouterRateLimit == 0 {
			cfg.OpenRouterRateLimit = getProviderDefaultOrFallback("openrouter", 20)
		}
	}
}

// applyTimeoutDefaults sets intelligent timeouts based on usage complexity
// Only applies defaults if the timeout is still at the default setting
func (ca *ConfigAdapter) applyTimeoutDefaults(cfg *config.CliConfig) {
	// Only override if still at default timeout (10 minutes)
	if cfg.Timeout == config.DefaultTimeout {
		if ca.simplified.HasFlag(FlagSynthesis) {
			// Synthesis mode: longer timeout for multiple model processing
			cfg.Timeout = 15 * time.Minute
		}
		// else: keep the default or environment-loaded timeout
	}
}

// applyConcurrencyDefaults sets intelligent concurrency based on usage mode
// Only applies defaults if the concurrency is still at the default setting
func (ca *ConfigAdapter) applyConcurrencyDefaults(cfg *config.CliConfig) {
	// Only override if still at default concurrency (5)
	if cfg.MaxConcurrentRequests == config.DefaultMaxConcurrentRequests {
		if ca.simplified.HasFlag(FlagSynthesis) {
			// Synthesis mode: lower concurrency to reduce rate limit pressure
			cfg.MaxConcurrentRequests = 3
		}
		// else: keep the default or environment-loaded concurrency
	}
}

// applyAPIKeyDefaults sets API keys from environment for compatibility with existing validation
func (ca *ConfigAdapter) applyAPIKeyDefaults(cfg *config.CliConfig) {
	// Set Gemini API key from environment (required for validation compatibility)
	// The validation logic checks config.APIKey for Gemini but uses getenv for others
	if os.Getenv("GEMINI_API_KEY") != "" {
		cfg.APIKey = os.Getenv("GEMINI_API_KEY")
	}

	// Set API endpoint if available
	if os.Getenv("GEMINI_API_URL") != "" {
		cfg.APIEndpoint = os.Getenv("GEMINI_API_URL")
	}
}

// getProviderDefaultOrFallback gets the provider default rate limit or returns fallback
func getProviderDefaultOrFallback(provider string, fallback int) int {
	if defaultRate := models.GetProviderDefaultRateLimit(provider); defaultRate > 0 {
		return defaultRate
	}
	return fallback
}
