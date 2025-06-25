package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/ratelimit"
	"github.com/phrazzld/thinktank/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigAdapter_BasicStructure tests the basic adapter structure and interface
func TestConfigAdapter_BasicStructure(t *testing.T) {
	simplified := &SimplifiedConfig{
		InstructionsFile: "test.md",
		TargetPath:       "src/",
		Flags:            FlagVerbose,
	}

	// This will fail initially - we need to create the adapter
	adapter := NewConfigAdapter(simplified)
	require.NotNil(t, adapter, "NewConfigAdapter should not return nil")

	// Test basic conversion functionality
	complex := adapter.ToComplexConfig()
	require.NotNil(t, complex, "ToComplexConfig should not return nil")

	// Verify basic field mapping
	assert.Equal(t, "test.md", complex.InstructionsFile)
	assert.Equal(t, []string{"src/"}, complex.Paths)
	assert.True(t, complex.Verbose)
}

// TestConfigAdapter_IntelligentDefaults tests intelligent default assignment
func TestConfigAdapter_IntelligentDefaults(t *testing.T) {
	simplified := &SimplifiedConfig{
		InstructionsFile: "test.md",
		TargetPath:       "src/",
		Flags:            0, // No special flags
	}

	adapter := NewConfigAdapter(simplified)
	complex := adapter.ToComplexConfig()

	// Test intelligent defaults are applied
	assert.Greater(t, complex.MaxConcurrentRequests, 0, "Should have non-zero concurrent request limit")
	assert.Greater(t, complex.RateLimitRequestsPerMinute, 0, "Should have non-zero rate limit")
	assert.Greater(t, complex.Timeout, time.Duration(0), "Should have non-zero timeout")

	// Test provider-specific rate limits
	assert.GreaterOrEqual(t, complex.GetProviderRateLimit("openai"), 0)
	assert.GreaterOrEqual(t, complex.GetProviderRateLimit("gemini"), 0)
	assert.GreaterOrEqual(t, complex.GetProviderRateLimit("openrouter"), 0)
}

// TestConfigAdapter_SynthesisModeDefaults tests synthesis-specific defaults
func TestConfigAdapter_SynthesisModeDefaults(t *testing.T) {
	simplified := &SimplifiedConfig{
		InstructionsFile: "test.md",
		TargetPath:       "src/",
		Flags:            FlagSynthesis,
	}

	adapter := NewConfigAdapter(simplified)
	complex := adapter.ToComplexConfig()

	// Synthesis mode should have different defaults
	assert.NotEmpty(t, complex.SynthesisModel, "Synthesis mode should set synthesis model")
	assert.Greater(t, len(complex.ModelNames), 1, "Synthesis mode should use multiple models")
	assert.Contains(t, complex.OutputDir, "synthesis", "Synthesis mode should use synthesis output directory")
}

// TestConfigAdapter_RoundTripBehaviorEquivalence tests behavior equivalence
// between simplified and complex config paths
func TestConfigAdapter_RoundTripBehaviorEquivalence(t *testing.T) {
	testCases := []struct {
		name   string
		config SimplifiedConfig
	}{
		{
			name: "basic_configuration",
			config: SimplifiedConfig{
				InstructionsFile: "test.md",
				TargetPath:       "src/",
				Flags:            0,
			},
		},
		{
			name: "verbose_mode",
			config: SimplifiedConfig{
				InstructionsFile: "verbose.md",
				TargetPath:       "code/",
				Flags:            FlagVerbose,
			},
		},
		{
			name: "synthesis_mode",
			config: SimplifiedConfig{
				InstructionsFile: "synthesis.md",
				TargetPath:       "project/",
				Flags:            FlagSynthesis,
			},
		},
		{
			name: "dry_run_mode",
			config: SimplifiedConfig{
				InstructionsFile: "dry.md",
				TargetPath:       "test/",
				Flags:            FlagDryRun,
			},
		},
		{
			name: "multiple_flags",
			config: SimplifiedConfig{
				InstructionsFile: "multi.md",
				TargetPath:       "app/",
				Flags:            FlagVerbose | FlagSynthesis,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert using adapter
			adapter := NewConfigAdapter(&tc.config)
			adapterResult := adapter.ToComplexConfig()

			// Convert using direct method
			directResult := tc.config.ToCliConfig()

			// Core fields should be identical (behavior equivalence)
			assert.Equal(t, directResult.InstructionsFile, adapterResult.InstructionsFile)
			assert.Equal(t, directResult.Paths, adapterResult.Paths)
			assert.Equal(t, directResult.DryRun, adapterResult.DryRun)
			assert.Equal(t, directResult.Verbose, adapterResult.Verbose)
			assert.Equal(t, directResult.ModelNames, adapterResult.ModelNames)
			assert.Equal(t, directResult.SynthesisModel, adapterResult.SynthesisModel)
			assert.Equal(t, directResult.OutputDir, adapterResult.OutputDir)

			// Adapter should have enhanced defaults that direct conversion lacks
			assert.GreaterOrEqual(t, adapterResult.OpenAIRateLimit, directResult.OpenAIRateLimit)
			assert.GreaterOrEqual(t, adapterResult.GeminiRateLimit, directResult.GeminiRateLimit)
			assert.GreaterOrEqual(t, adapterResult.OpenRouterRateLimit, directResult.OpenRouterRateLimit)
			assert.GreaterOrEqual(t, adapterResult.Timeout, directResult.Timeout)
		})
	}
}

// TestConfigAdapter_ValidationBehaviorPreservation tests that validation
// behavior is preserved through the adapter conversion
func TestConfigAdapter_ValidationBehaviorPreservation(t *testing.T) {
	testCases := []struct {
		name                  string
		config                SimplifiedConfig
		expectedInstructions  string
		expectedPathsNotEmpty bool
	}{
		{
			name: "basic_config_preservation",
			config: SimplifiedConfig{
				InstructionsFile: "valid.md",
				TargetPath:       "src/",
				Flags:            0,
			},
			expectedInstructions:  "valid.md",
			expectedPathsNotEmpty: true,
		},
		{
			name: "dry_run_with_empty_instructions",
			config: SimplifiedConfig{
				InstructionsFile: "",
				TargetPath:       "src/",
				Flags:            FlagDryRun,
			},
			expectedInstructions:  "", // Dry run preserves empty instructions
			expectedPathsNotEmpty: true,
		},
		{
			name: "empty_instructions_preserved",
			config: SimplifiedConfig{
				InstructionsFile: "",
				TargetPath:       "src/",
				Flags:            0,
			},
			expectedInstructions:  "", // Empty instructions should be preserved in complex config
			expectedPathsNotEmpty: true,
		},
		{
			name: "empty_target_path_preserved",
			config: SimplifiedConfig{
				InstructionsFile: "test.md",
				TargetPath:       "",
				Flags:            0,
			},
			expectedInstructions:  "test.md",
			expectedPathsNotEmpty: false, // Empty target path results in empty paths array
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test adapter conversion preserves the input data
			adapter := NewConfigAdapter(&tc.config)
			complexConfig := adapter.ToComplexConfig()

			// Verify that field mapping is preserved correctly
			assert.Equal(t, tc.expectedInstructions, complexConfig.InstructionsFile,
				"Instructions file should be preserved from simplified to complex config")

			if tc.expectedPathsNotEmpty {
				assert.NotEmpty(t, complexConfig.Paths, "Paths should not be empty when target path is set")
				assert.Equal(t, tc.config.TargetPath, complexConfig.Paths[0], "Target path should be preserved in paths array")
			} else {
				// When target path is empty, we expect paths array to contain one empty string
				assert.Len(t, complexConfig.Paths, 1, "Paths array should have one element")
				assert.Equal(t, "", complexConfig.Paths[0], "Empty target path should result in empty first path element")
			}

			// Verify flag preservation
			assert.Equal(t, tc.config.HasFlag(FlagDryRun), complexConfig.DryRun, "DryRun flag should be preserved")
			assert.Equal(t, tc.config.HasFlag(FlagVerbose), complexConfig.Verbose, "Verbose flag should be preserved")
		})
	}
}

// TestConfigAdapter_ProviderSpecificRateLimits tests provider-specific rate limit assignment
func TestConfigAdapter_ProviderSpecificRateLimits(t *testing.T) {
	tests := []struct {
		name                    string
		config                  SimplifiedConfig
		expectedOpenAILimit     int
		expectedGeminiLimit     int
		expectedOpenRouterLimit int
	}{
		{
			name: "normal_mode_rate_limits",
			config: SimplifiedConfig{
				InstructionsFile: "test.md",
				TargetPath:       "src/",
				Flags:            0,
			},
			expectedOpenAILimit:     3000, // Full rate limit
			expectedGeminiLimit:     60,
			expectedOpenRouterLimit: 20,
		},
		{
			name: "synthesis_mode_conservative_limits",
			config: SimplifiedConfig{
				InstructionsFile: "synthesis.md",
				TargetPath:       "project/",
				Flags:            FlagSynthesis,
			},
			expectedOpenAILimit:     1800, // 60% of default (conservative)
			expectedGeminiLimit:     36,   // 60% of default
			expectedOpenRouterLimit: 12,   // 60% of default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(&tt.config)
			complex := adapter.ToComplexConfig()

			assert.Equal(t, tt.expectedOpenAILimit, complex.OpenAIRateLimit)
			assert.Equal(t, tt.expectedGeminiLimit, complex.GeminiRateLimit)
			assert.Equal(t, tt.expectedOpenRouterLimit, complex.OpenRouterRateLimit)
		})
	}
}

// === INTEGRATION TESTS ===
// These tests verify that adapter-converted configs work correctly with real CLI components

// TestConfigAdapter_RateLimiterIntegration tests that adapter configs work with real rate limiters
func TestConfigAdapter_RateLimiterIntegration(t *testing.T) {
	tests := []struct {
		name           string
		simplifiedCfg  SimplifiedConfig
		expectedLimits map[string]int
	}{
		{
			name: "synthesis_mode_conservative_limits",
			simplifiedCfg: SimplifiedConfig{
				InstructionsFile: "synthesis.md",
				TargetPath:       "project/",
				Flags:            FlagSynthesis,
			},
			expectedLimits: map[string]int{
				"openai":     1800, // 60% of 3000
				"gemini":     36,   // 60% of 60
				"openrouter": 12,   // 60% of 20
			},
		},
		{
			name: "normal_mode_standard_limits",
			simplifiedCfg: SimplifiedConfig{
				InstructionsFile: "test.md",
				TargetPath:       "src/",
				Flags:            0,
			},
			expectedLimits: map[string]int{
				"openai":     3000, // Full provider default
				"gemini":     60,
				"openrouter": 20,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert simplified config through adapter
			adapter := NewConfigAdapter(&tt.simplifiedCfg)
			complexConfig := adapter.ToComplexConfig()

			// Test integration: Do the rate limits work with real rate limiter component?
			rateLimiter := ratelimit.NewRateLimiter(
				complexConfig.MaxConcurrentRequests,
				complexConfig.RateLimitRequestsPerMinute,
			)
			require.NotNil(t, rateLimiter, "Rate limiter should be created")

			// Verify rate limits are applied correctly
			for provider, expectedLimit := range tt.expectedLimits {
				actualLimit := complexConfig.GetProviderRateLimit(provider)
				assert.Equal(t, expectedLimit, actualLimit,
					"Rate limit for %s should match adapter's intelligent defaults", provider)
			}

			// Test that rate limiter accepts the configuration values
			assert.GreaterOrEqual(t, complexConfig.MaxConcurrentRequests, 1,
				"Max concurrent requests should be reasonable")
			assert.GreaterOrEqual(t, complexConfig.RateLimitRequestsPerMinute, 1,
				"Rate limit should be reasonable")
		})
	}
}

// TestConfigAdapter_ValidationPipelineIntegration tests adapter with real validation pipeline
func TestConfigAdapter_ValidationPipelineIntegration(t *testing.T) {
	// Create temporary test files for validation
	tempDir := t.TempDir()
	validInstFile := filepath.Join(tempDir, "instructions.md")
	require.NoError(t, os.WriteFile(validInstFile, []byte("test instructions"), 0644))

	validTargetDir := filepath.Join(tempDir, "src")
	require.NoError(t, os.Mkdir(validTargetDir, 0755))

	testCases := []struct {
		name          string
		simplifiedCfg SimplifiedConfig
		setupEnv      func(t *testing.T)
		expectValid   bool
		expectedError string
	}{
		{
			name: "synthesis_mode_validation_success",
			simplifiedCfg: SimplifiedConfig{
				InstructionsFile: validInstFile,
				TargetPath:       validTargetDir,
				Flags:            FlagSynthesis,
			},
			setupEnv: func(t *testing.T) {
				t.Setenv("GEMINI_API_KEY", "test-key")
				t.Setenv("OPENAI_API_KEY", "test-key-2")
			},
			expectValid: true,
		},
		{
			name: "synthesis_mode_missing_openai_key",
			simplifiedCfg: SimplifiedConfig{
				InstructionsFile: validInstFile,
				TargetPath:       validTargetDir,
				Flags:            FlagSynthesis,
			},
			setupEnv: func(t *testing.T) {
				t.Setenv("GEMINI_API_KEY", "test-key")
				// Missing OPENAI_API_KEY - this should cause validation to fail
				_ = os.Unsetenv("OPENAI_API_KEY")
			},
			expectValid:   false,
			expectedError: "openAI API key not set",
		},
		{
			name: "normal_mode_validation_success",
			simplifiedCfg: SimplifiedConfig{
				InstructionsFile: validInstFile,
				TargetPath:       validTargetDir,
				Flags:            0,
			},
			setupEnv: func(t *testing.T) {
				t.Setenv("GEMINI_API_KEY", "test-key")
			},
			expectValid: true,
		},
		{
			name: "dry_run_mode_skips_api_validation",
			simplifiedCfg: SimplifiedConfig{
				InstructionsFile: validInstFile,
				TargetPath:       validTargetDir,
				Flags:            FlagDryRun,
			},
			setupEnv:    func(t *testing.T) {}, // No API keys
			expectValid: true,                  // Should pass because dry-run skips API validation
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupEnv(t)

			// Test: SimplifiedConfig -> ConfigAdapter -> CliConfig -> Validation
			adapter := NewConfigAdapter(&tc.simplifiedCfg)
			complexConfig := adapter.ToComplexConfig()

			// Integration test: Use real validation pipeline with mock logger and custom env
			mockLogger := &testutil.MockLogger{}
			mockGetenv := func(key string) string {
				return os.Getenv(key) // This will see the test environment variables
			}
			err := config.ValidateConfigWithEnv(complexConfig, mockLogger, mockGetenv)

			if tc.expectValid {
				assert.NoError(t, err, "Config should validate successfully")
			} else {
				assert.Error(t, err, "Config should fail validation")
				if tc.expectedError != "" && err != nil {
					assert.Contains(t, err.Error(), tc.expectedError)
				}
			}
		})
	}
}

// TestConfigAdapter_TimeoutConfigurationIntegration tests timeout behavior integration
func TestConfigAdapter_TimeoutConfigurationIntegration(t *testing.T) {
	tests := []struct {
		name            string
		simplifiedCfg   SimplifiedConfig
		expectedTimeout time.Duration
	}{
		{
			name: "synthesis_mode_extended_timeout",
			simplifiedCfg: SimplifiedConfig{
				InstructionsFile: "synthesis.md",
				TargetPath:       "project/",
				Flags:            FlagSynthesis,
			},
			expectedTimeout: 15 * time.Minute,
		},
		{
			name: "normal_mode_standard_timeout",
			simplifiedCfg: SimplifiedConfig{
				InstructionsFile: "test.md",
				TargetPath:       "src/",
				Flags:            0,
			},
			expectedTimeout: 10 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(&tt.simplifiedCfg)
			complexConfig := adapter.ToComplexConfig()

			// Test integration: Timeout configuration is applied correctly
			assert.Equal(t, tt.expectedTimeout, complexConfig.Timeout,
				"Timeout should match adapter's intelligent defaults")

			// Verify timeout is reasonable for CLI operations
			assert.GreaterOrEqual(t, complexConfig.Timeout, 5*time.Minute,
				"Timeout should be long enough for CLI operations")
			assert.LessOrEqual(t, complexConfig.Timeout, 30*time.Minute,
				"Timeout should not be excessively long")
		})
	}
}

// TestConfigAdapter_ConcurrencyConfigurationIntegration tests concurrency settings integration
func TestConfigAdapter_ConcurrencyConfigurationIntegration(t *testing.T) {
	tests := []struct {
		name                         string
		simplifiedCfg                SimplifiedConfig
		expectedMaxConcurrent        int
		expectedConservativeBehavior bool
	}{
		{
			name: "synthesis_mode_conservative_concurrency",
			simplifiedCfg: SimplifiedConfig{
				InstructionsFile: "synthesis.md",
				TargetPath:       "project/",
				Flags:            FlagSynthesis,
			},
			expectedMaxConcurrent:        3, // Lower for synthesis
			expectedConservativeBehavior: true,
		},
		{
			name: "normal_mode_standard_concurrency",
			simplifiedCfg: SimplifiedConfig{
				InstructionsFile: "test.md",
				TargetPath:       "src/",
				Flags:            0,
			},
			expectedMaxConcurrent:        5, // Standard
			expectedConservativeBehavior: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(&tt.simplifiedCfg)
			complexConfig := adapter.ToComplexConfig()

			// Test integration: Concurrency configuration is applied correctly
			assert.Equal(t, tt.expectedMaxConcurrent, complexConfig.MaxConcurrentRequests,
				"Max concurrent requests should match adapter's intelligent defaults")

			// Verify concurrency settings are reasonable
			assert.GreaterOrEqual(t, complexConfig.MaxConcurrentRequests, 1,
				"Must allow at least 1 concurrent request")
			assert.LessOrEqual(t, complexConfig.MaxConcurrentRequests, 10,
				"Should not exceed reasonable concurrency limit")

			// Test synthesis mode has more conservative settings
			if tt.expectedConservativeBehavior {
				assert.LessOrEqual(t, complexConfig.MaxConcurrentRequests, 3,
					"Synthesis mode should use conservative concurrency")
			}
		})
	}
}

// TestConfigAdapter_ModelSelectionIntegration tests model selection behavior integration
func TestConfigAdapter_ModelSelectionIntegration(t *testing.T) {
	tests := []struct {
		name                string
		simplifiedCfg       SimplifiedConfig
		expectedModels      []string
		expectedSynthesis   string
		shouldHaveSynthesis bool
	}{
		{
			name: "synthesis_mode_multi_model_selection",
			simplifiedCfg: SimplifiedConfig{
				InstructionsFile: "synthesis.md",
				TargetPath:       "project/",
				Flags:            FlagSynthesis,
			},
			expectedModels:      []string{"gemini-2.5-pro", "gpt-4.1"},
			expectedSynthesis:   "gemini-2.5-pro",
			shouldHaveSynthesis: true,
		},
		{
			name: "normal_mode_single_model_selection",
			simplifiedCfg: SimplifiedConfig{
				InstructionsFile: "test.md",
				TargetPath:       "src/",
				Flags:            0,
			},
			expectedModels:      []string{"gemini-2.5-pro"},
			expectedSynthesis:   "",
			shouldHaveSynthesis: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(&tt.simplifiedCfg)
			complexConfig := adapter.ToComplexConfig()

			// Test integration: Model selection works correctly
			assert.Equal(t, tt.expectedModels, complexConfig.ModelNames,
				"Model names should match adapter's intelligent defaults")

			if tt.shouldHaveSynthesis {
				assert.Equal(t, tt.expectedSynthesis, complexConfig.SynthesisModel,
					"Synthesis model should be set for synthesis mode")
				assert.Greater(t, len(complexConfig.ModelNames), 1,
					"Synthesis mode should use multiple models")
			} else {
				assert.Empty(t, complexConfig.SynthesisModel,
					"Normal mode should not have synthesis model")
				assert.Equal(t, 1, len(complexConfig.ModelNames),
					"Normal mode should use single model")
			}
		})
	}
}

// TestConfigAdapter_BehaviorRegressionPrevention tests that adapter preserves existing CLI behavior
func TestConfigAdapter_BehaviorRegressionPrevention(t *testing.T) {
	// Create test environment
	tempDir := t.TempDir()
	instructionsFile := filepath.Join(tempDir, "instructions.md")
	require.NoError(t, os.WriteFile(instructionsFile, []byte("test instructions"), 0644))

	sourceDir := filepath.Join(tempDir, "src")
	require.NoError(t, os.Mkdir(sourceDir, 0755))

	testCases := []struct {
		name             string
		simplifiedCfg    SimplifiedConfig
		expectedBehavior map[string]interface{}
	}{
		{
			name: "dry_run_behavior_preservation",
			simplifiedCfg: SimplifiedConfig{
				InstructionsFile: instructionsFile,
				TargetPath:       sourceDir,
				Flags:            FlagDryRun,
			},
			expectedBehavior: map[string]interface{}{
				"DryRun":        true,
				"Verbose":       false,
				"EmptyAPICheck": true, // Dry run should skip API validation
			},
		},
		{
			name: "verbose_behavior_preservation",
			simplifiedCfg: SimplifiedConfig{
				InstructionsFile: instructionsFile,
				TargetPath:       sourceDir,
				Flags:            FlagVerbose,
			},
			expectedBehavior: map[string]interface{}{
				"DryRun":   false,
				"Verbose":  true,
				"LogLevel": "debug", // Verbose should set debug level
			},
		},
		{
			name: "combined_flags_behavior_preservation",
			simplifiedCfg: SimplifiedConfig{
				InstructionsFile: instructionsFile,
				TargetPath:       sourceDir,
				Flags:            FlagVerbose | FlagDryRun,
			},
			expectedBehavior: map[string]interface{}{
				"DryRun":        true,
				"Verbose":       true,
				"LogLevel":      "debug",
				"EmptyAPICheck": true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := NewConfigAdapter(&tc.simplifiedCfg)
			complexConfig := adapter.ToComplexConfig()

			// Test behavior preservation
			if expected, ok := tc.expectedBehavior["DryRun"].(bool); ok {
				assert.Equal(t, expected, complexConfig.DryRun,
					"DryRun behavior should be preserved")
			}

			if expected, ok := tc.expectedBehavior["Verbose"].(bool); ok {
				assert.Equal(t, expected, complexConfig.Verbose,
					"Verbose behavior should be preserved")
			}

			if _, checkEmptyAPI := tc.expectedBehavior["EmptyAPICheck"]; checkEmptyAPI {
				// For dry-run mode, API key validation should be skipped
				mockLogger := &testutil.MockLogger{}
				mockGetenv := func(key string) string {
					return os.Getenv(key) // This will see the test environment variables
				}
				err := config.ValidateConfigWithEnv(complexConfig, mockLogger, mockGetenv)
				if complexConfig.DryRun {
					// Dry run should not fail due to missing API keys
					assert.NoError(t, err, "Dry run should skip API key validation")
				}
			}

			// Verify field mapping is consistent
			assert.Equal(t, tc.simplifiedCfg.InstructionsFile, complexConfig.InstructionsFile,
				"Instructions file should be preserved")
			assert.Equal(t, []string{tc.simplifiedCfg.TargetPath}, complexConfig.Paths,
				"Target path should be preserved in paths array")
		})
	}
}

// TestConfigAdapter_EndToEndConfigurationFlow tests complete config flow integration
func TestConfigAdapter_EndToEndConfigurationFlow(t *testing.T) {
	// Create realistic test environment
	tempDir := t.TempDir()
	instructionsFile := filepath.Join(tempDir, "instructions.md")
	require.NoError(t, os.WriteFile(instructionsFile, []byte("Analyze this code"), 0644))

	sourceFile := filepath.Join(tempDir, "main.go")
	require.NoError(t, os.WriteFile(sourceFile, []byte("package main\nfunc main() {}"), 0644))

	testCases := []struct {
		name          string
		simplifiedCfg SimplifiedConfig
		setupEnv      func(t *testing.T)
		expectSuccess bool
	}{
		{
			name: "complete_dry_run_flow",
			simplifiedCfg: SimplifiedConfig{
				InstructionsFile: instructionsFile,
				TargetPath:       sourceFile,
				Flags:            FlagDryRun,
			},
			setupEnv: func(t *testing.T) {
				// Dry run doesn't need API keys
			},
			expectSuccess: true,
		},
		{
			name: "synthesis_mode_with_validation",
			simplifiedCfg: SimplifiedConfig{
				InstructionsFile: instructionsFile,
				TargetPath:       sourceFile,
				Flags:            FlagSynthesis | FlagDryRun, // Dry run to avoid API calls
			},
			setupEnv: func(t *testing.T) {
				t.Setenv("GEMINI_API_KEY", "test-key")
				t.Setenv("OPENAI_API_KEY", "test-key-2")
			},
			expectSuccess: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupEnv(t)

			// Test complete flow: SimplifiedConfig -> ConfigAdapter -> CliConfig -> Validation
			adapter := NewConfigAdapter(&tc.simplifiedCfg)
			complexConfig := adapter.ToComplexConfig()

			// Step 1: Validation should succeed
			mockLogger := &testutil.MockLogger{}
			mockGetenv := func(key string) string {
				return os.Getenv(key) // This will see the test environment variables
			}
			err := config.ValidateConfigWithEnv(complexConfig, mockLogger, mockGetenv)
			if tc.expectSuccess {
				assert.NoError(t, err, "Configuration should validate successfully")
			} else {
				assert.Error(t, err, "Configuration should fail validation")
				return // Skip further tests if validation fails
			}

			// Step 2: Configuration should have all required fields
			assert.NotEmpty(t, complexConfig.InstructionsFile, "Instructions file should be set")
			assert.NotEmpty(t, complexConfig.Paths, "Paths should be set")
			assert.NotEmpty(t, complexConfig.ModelNames, "Model names should be set")
			assert.Greater(t, complexConfig.MaxConcurrentRequests, 0, "Max concurrent should be positive")
			assert.Greater(t, complexConfig.Timeout, time.Duration(0), "Timeout should be positive")

			// Step 3: Provider configuration should be reasonable
			assert.GreaterOrEqual(t, complexConfig.GetProviderRateLimit("gemini"), 1,
				"Gemini rate limit should be reasonable")
			assert.GreaterOrEqual(t, complexConfig.GetProviderRateLimit("openai"), 1,
				"OpenAI rate limit should be reasonable")

			// Step 4: Synthesis mode should be configured correctly
			if tc.simplifiedCfg.HasFlag(FlagSynthesis) {
				assert.NotEmpty(t, complexConfig.SynthesisModel, "Synthesis model should be set")
				assert.Greater(t, len(complexConfig.ModelNames), 1, "Should have multiple models")
				assert.Equal(t, 3, complexConfig.MaxConcurrentRequests, "Should use conservative concurrency")
				assert.Equal(t, 15*time.Minute, complexConfig.Timeout, "Should use extended timeout")
			}
		})
	}
}
