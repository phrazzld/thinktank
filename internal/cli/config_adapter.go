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
}

// NewConfigAdapter creates a new config adapter for the given simplified config
func NewConfigAdapter(simplified *SimplifiedConfig) *ConfigAdapter {
	return &ConfigAdapter{
		simplified: simplified,
	}
}

// ToComplexConfig converts the simplified configuration to a full CliConfig
// with intelligent defaults for unmapped fields. This preserves all existing
// validation behavior while providing enhanced default assignment.
func (ca *ConfigAdapter) ToComplexConfig() *config.CliConfig {
	// Start with base conversion using existing logic
	cfg := ca.simplified.ToCliConfig()

	// Apply intelligent defaults for unmapped fields
	ca.applyIntelligentDefaults(cfg)

	// Set API key from environment for compatibility with existing validation
	ca.applyAPIKeyDefaults(cfg)

	return cfg
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
func (ca *ConfigAdapter) applyRateLimitDefaults(cfg *config.CliConfig) {
	if ca.simplified.HasFlag(FlagSynthesis) {
		// Synthesis mode: more conservative to avoid rate limits across providers
		// Apply 60% reduction to provider defaults
		cfg.OpenAIRateLimit = int(float64(getProviderDefaultOrFallback("openai", 3000)) * 0.6)       // 60% of OpenAI default
		cfg.GeminiRateLimit = int(float64(getProviderDefaultOrFallback("gemini", 60)) * 0.6)         // 60% of Gemini default
		cfg.OpenRouterRateLimit = int(float64(getProviderDefaultOrFallback("openrouter", 20)) * 0.6) // 60% of OpenRouter default
	} else {
		// Single model mode: use standard provider defaults
		cfg.OpenAIRateLimit = getProviderDefaultOrFallback("openai", 3000)
		cfg.GeminiRateLimit = getProviderDefaultOrFallback("gemini", 60)
		cfg.OpenRouterRateLimit = getProviderDefaultOrFallback("openrouter", 20)
	}
}

// applyTimeoutDefaults sets intelligent timeouts based on usage complexity
func (ca *ConfigAdapter) applyTimeoutDefaults(cfg *config.CliConfig) {
	if ca.simplified.HasFlag(FlagSynthesis) {
		// Synthesis mode: longer timeout for multiple model processing
		cfg.Timeout = 15 * time.Minute
	} else {
		// Single model mode: standard timeout
		cfg.Timeout = 10 * time.Minute
	}
}

// applyConcurrencyDefaults sets intelligent concurrency based on usage mode
func (ca *ConfigAdapter) applyConcurrencyDefaults(cfg *config.CliConfig) {
	if ca.simplified.HasFlag(FlagSynthesis) {
		// Synthesis mode: lower concurrency to reduce rate limit pressure
		cfg.MaxConcurrentRequests = 3
	} else {
		// Single model mode: standard concurrency
		cfg.MaxConcurrentRequests = 5
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
