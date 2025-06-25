package cli

import (
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/models"
)

// APIKeyDetector provides API key detection and validation services
// optimized for CLI tool startup performance.
type APIKeyDetector struct {
	logger logutil.LoggerInterface

	// Cache with minimal overhead - single execution optimization
	cache      *detectionCache
	cacheMutex sync.RWMutex
}

// detectionCache stores API key detection results with timestamp
type detectionCache struct {
	result    *DetectionResult
	timestamp time.Time
	ttl       time.Duration
}

// DetectionResult contains the detected API keys and provider capabilities
type DetectionResult struct {
	// DetectedProviders maps provider name to API key presence
	DetectedProviders map[string]bool

	// ValidKeys maps provider name to validation status
	ValidKeys map[string]bool

	// ProviderCapabilities maps provider name to capability info
	ProviderCapabilities map[string]ProviderCapability

	// DetectionTimestamp records when detection was performed
	DetectionTimestamp time.Time
}

// ProviderCapability contains rate limiting and capability information
type ProviderCapability struct {
	Provider         string
	HasValidKey      bool
	DefaultRateLimit int
	ModelCount       int
	AvailableModels  []string
}

// Compile-time regex patterns for maximum performance
// These are compiled once at package init, not per-request
var (
	// API key validation patterns based on known provider formats
	providerPatterns = map[string]*regexp.Regexp{
		"openai":     regexp.MustCompile(`^sk-[a-zA-Z0-9]{20,}$|^sk-proj-[a-zA-Z0-9]{20,}$`),
		"gemini":     regexp.MustCompile(`^[a-zA-Z0-9_-]{39}$`),
		"openrouter": regexp.MustCompile(`^sk-or-v1-[a-zA-Z0-9]{64}$`),
	}

	// Single environment read for all providers - minimize system calls
	envKeys = []string{"OPENAI_API_KEY", "GEMINI_API_KEY", "OPENROUTER_API_KEY"}
)

// NewAPIKeyDetector creates a new API key detector with optimized caching
func NewAPIKeyDetector(logger logutil.LoggerInterface) *APIKeyDetector {
	return &APIKeyDetector{
		logger: logger,
		cache: &detectionCache{
			ttl: 5 * time.Minute, // Configurable TTL
		},
	}
}

// DetectAndValidate performs API key detection and validation with caching
// Optimized for CLI tool performance - single batch environment read
func (d *APIKeyDetector) DetectAndValidate() (*DetectionResult, error) {
	d.cacheMutex.RLock()
	if d.cache.result != nil && time.Since(d.cache.timestamp) < d.cache.ttl {
		d.logger.Debug("Using cached API key detection result, cache age: %.2fs", time.Since(d.cache.timestamp).Seconds())
		d.cacheMutex.RUnlock()
		return d.cache.result, nil
	}
	d.cacheMutex.RUnlock()

	// Perform fresh detection
	result, err := d.performDetection()
	if err != nil {
		return nil, err
	}

	// Update cache atomically
	d.cacheMutex.Lock()
	d.cache.result = result
	d.cache.timestamp = time.Now()
	d.cacheMutex.Unlock()

	return result, nil
}

// performDetection executes the actual API key detection logic
// Optimized for minimal system calls and allocation
func (d *APIKeyDetector) performDetection() (*DetectionResult, error) {
	startTime := time.Now()

	// Single batch read of all environment variables
	envValues := make(map[string]string, len(envKeys))
	for _, key := range envKeys {
		envValues[key] = os.Getenv(key)
	}

	d.logger.Debug("Environment variables read in %d microseconds, checked %d keys",
		time.Since(startTime).Microseconds(), len(envKeys))

	// Initialize result with pre-allocated maps
	result := &DetectionResult{
		DetectedProviders:    make(map[string]bool, 3),
		ValidKeys:            make(map[string]bool, 3),
		ProviderCapabilities: make(map[string]ProviderCapability, 3),
		DetectionTimestamp:   time.Now(),
	}

	// Process each provider efficiently
	providers := []string{"openai", "gemini", "openrouter"}
	for _, provider := range providers {
		envVar := models.GetAPIKeyEnvVar(provider)
		keyValue := envValues[envVar]

		// Check presence
		hasKey := keyValue != ""
		result.DetectedProviders[provider] = hasKey

		// Validate format if present
		isValid := false
		if hasKey {
			if pattern, exists := providerPatterns[provider]; exists {
				isValid = pattern.MatchString(keyValue)
			}
		}
		result.ValidKeys[provider] = isValid

		// Build capability info
		capability := ProviderCapability{
			Provider:         provider,
			HasValidKey:      isValid,
			DefaultRateLimit: models.GetProviderDefaultRateLimit(provider),
			AvailableModels:  models.ListModelsForProvider(provider),
		}
		capability.ModelCount = len(capability.AvailableModels)
		result.ProviderCapabilities[provider] = capability

		d.logger.Debug("Provider %s: has_key=%v, valid_format=%v, model_count=%d, rate_limit=%d",
			provider, hasKey, isValid, capability.ModelCount, capability.DefaultRateLimit)
	}

	totalDuration := time.Since(startTime)
	d.logger.Debug("API key detection completed in %d microseconds, processed %d providers, found %d valid keys",
		totalDuration.Microseconds(), len(providers), d.countValidKeys(result))

	return result, nil
}

// GetAvailableProviders returns a list of providers with valid API keys
func (d *APIKeyDetector) GetAvailableProviders() ([]string, error) {
	result, err := d.DetectAndValidate()
	if err != nil {
		return nil, err
	}

	var available []string
	for provider, isValid := range result.ValidKeys {
		if isValid {
			available = append(available, provider)
		}
	}

	return available, nil
}

// GetProviderCapability returns capability information for a specific provider
func (d *APIKeyDetector) GetProviderCapability(provider string) (*ProviderCapability, error) {
	result, err := d.DetectAndValidate()
	if err != nil {
		return nil, err
	}

	if capability, exists := result.ProviderCapabilities[provider]; exists {
		return &capability, nil
	}

	return nil, nil // Provider not found
}

// InvalidateCache forces a fresh detection on the next call
func (d *APIKeyDetector) InvalidateCache() {
	d.cacheMutex.Lock()
	d.cache.result = nil
	d.cacheMutex.Unlock()

	d.logger.Debug("API key detection cache invalidated")
}

// Helper function to count valid keys for logging
func (d *APIKeyDetector) countValidKeys(result *DetectionResult) int {
	count := 0
	for _, isValid := range result.ValidKeys {
		if isValid {
			count++
		}
	}
	return count
}
