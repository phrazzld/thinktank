package models

import (
	"os"
	"strings"
	"testing"
)

func TestGetModelInfo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		modelName      string
		wantProvider   string
		wantAPIModelID string
		wantError      bool
		errorContains  string
	}{
		// Valid models - OpenAI (migrated to OpenRouter)
		{
			name:           "gpt-4.1 valid model",
			modelName:      "gpt-4.1",
			wantProvider:   "openrouter",
			wantAPIModelID: "openai/gpt-4.1",
			wantError:      false,
		},
		{
			name:           "o4-mini valid model",
			modelName:      "o4-mini",
			wantProvider:   "openrouter",
			wantAPIModelID: "openai/o4-mini",
			wantError:      false,
		},
		// Valid models - Gemini (migrated to OpenRouter)
		{
			name:           "gemini-2.5-pro valid model",
			modelName:      "gemini-2.5-pro",
			wantProvider:   "openrouter",
			wantAPIModelID: "google/gemini-2.5-pro",
			wantError:      false,
		},
		{
			name:           "gemini-2.5-flash valid model",
			modelName:      "gemini-2.5-flash",
			wantProvider:   "openrouter",
			wantAPIModelID: "google/gemini-2.5-flash",
			wantError:      false,
		},
		// Valid models - OpenRouter
		{
			name:           "openrouter model",
			modelName:      "openrouter/meta-llama/llama-4-maverick",
			wantProvider:   "openrouter",
			wantAPIModelID: "meta-llama/llama-4-maverick",
			wantError:      false,
		},
		// Error cases
		{
			name:          "empty model name",
			modelName:     "",
			wantError:     true,
			errorContains: "unknown model",
		},
		{
			name:          "unknown model",
			modelName:     "unknown-model",
			wantError:     true,
			errorContains: "unknown model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := GetModelInfo(tt.modelName)

			if tt.wantError {
				if err == nil {
					t.Errorf("GetModelInfo(%q) expected error, got nil", tt.modelName)
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("GetModelInfo(%q) error = %q, want error containing %q", tt.modelName, err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("GetModelInfo(%q) unexpected error: %v", tt.modelName, err)
				return
			}

			if info.Provider != tt.wantProvider {
				t.Errorf("GetModelInfo(%q).Provider = %q, want %q", tt.modelName, info.Provider, tt.wantProvider)
			}

			if info.APIModelID != tt.wantAPIModelID {
				t.Errorf("GetModelInfo(%q).APIModelID = %q, want %q", tt.modelName, info.APIModelID, tt.wantAPIModelID)
			}

			// Verify context window is positive
			if info.ContextWindow <= 0 {
				t.Errorf("GetModelInfo(%q).ContextWindow = %d, want > 0", tt.modelName, info.ContextWindow)
			}

			// Verify max output tokens is positive
			if info.MaxOutputTokens <= 0 {
				t.Errorf("GetModelInfo(%q).MaxOutputTokens = %d, want > 0", tt.modelName, info.MaxOutputTokens)
			}
		})
	}
}

func TestGetProviderDefaultRateLimitCore(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		provider string
		expected int
	}{
		{
			name:     "openai provider (obsolete)",
			provider: "openai",
			expected: 60, // Obsolete provider returns default rate limit
		},
		{
			name:     "gemini provider",
			provider: "gemini",
			expected: 60,
		},
		{
			name:     "openrouter provider",
			provider: "openrouter",
			expected: 20,
		},
		{
			name:     "unknown provider",
			provider: "unknown",
			expected: 60,
		},
		{
			name:     "empty provider",
			provider: "",
			expected: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetProviderDefaultRateLimit(tt.provider)
			if result != tt.expected {
				t.Errorf("GetProviderDefaultRateLimit(%q) = %d, want %d", tt.provider, result, tt.expected)
			}
		})
	}
}

func TestListModelsForProvider(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		provider string
		expected []string
	}{
		{
			name:     "openai provider",
			provider: "openai",
			expected: []string{},
		},
		{
			name:     "gemini provider",
			provider: "gemini",
			expected: []string{},
		},
		{
			name:     "openrouter provider",
			provider: "openrouter",
			expected: []string{
				"gemini-2.5-flash",
				"gemini-2.5-pro",
				"gpt-4.1",
				"o3",
				"o4-mini",
				"openrouter/deepseek/deepseek-chat-v3-0324",
				"openrouter/deepseek/deepseek-chat-v3-0324:free",
				"openrouter/deepseek/deepseek-r1-0528",
				"openrouter/deepseek/deepseek-r1-0528:free",
				"openrouter/google/gemma-3-27b-it",
				"openrouter/meta-llama/llama-3.3-70b-instruct",
				"openrouter/meta-llama/llama-4-maverick",
				"openrouter/meta-llama/llama-4-scout",
				"openrouter/x-ai/grok-3-beta",
				"openrouter/x-ai/grok-3-mini-beta",
			},
		},
		{
			name:     "test provider",
			provider: "test",
			expected: []string{"model1", "model2", "model3", "synthesis-model"},
		},
		{
			name:     "unknown provider",
			provider: "unknown",
			expected: []string{},
		},
		{
			name:     "empty provider",
			provider: "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ListModelsForProvider(tt.provider)

			if len(result) != len(tt.expected) {
				t.Errorf("ListModelsForProvider(%q) returned %d models, want %d\nGot: %v\nWant: %v",
					tt.provider, len(result), len(tt.expected), result, tt.expected)
				return
			}

			// Create sets for comparison (order might vary)
			resultSet := make(map[string]bool)
			for _, model := range result {
				resultSet[model] = true
			}

			expectedSet := make(map[string]bool)
			for _, model := range tt.expected {
				expectedSet[model] = true
			}

			// Verify all expected models are present
			for _, expected := range tt.expected {
				if !resultSet[expected] {
					t.Errorf("ListModelsForProvider(%q) missing expected model: %s", tt.provider, expected)
				}
			}

			// Verify no unexpected models are present
			for _, actual := range result {
				if !expectedSet[actual] {
					t.Errorf("ListModelsForProvider(%q) returned unexpected model: %s", tt.provider, actual)
				}
			}
		})
	}
}

// Test benchmarks to verify performance
func BenchmarkGetModelInfo(b *testing.B) {
	modelName := "gpt-4.1"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetModelInfo(modelName)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetProviderDefaultRateLimit(b *testing.B) {
	provider := "openai"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetProviderDefaultRateLimit(provider)
	}
}

func BenchmarkListModelsForProvider(b *testing.B) {
	provider := "openai"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ListModelsForProvider(provider)
	}
}

func TestIsModelSupported(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		modelName string
		expected  bool
	}{
		{
			name:      "valid OpenAI model",
			modelName: "gpt-4.1",
			expected:  true,
		},
		{
			name:      "valid Gemini model",
			modelName: "gemini-2.5-pro",
			expected:  true,
		},
		{
			name:      "valid OpenRouter model",
			modelName: "openrouter/meta-llama/llama-4-maverick",
			expected:  true,
		},
		{
			name:      "invalid model",
			modelName: "invalid-model",
			expected:  false,
		},
		{
			name:      "empty model name",
			modelName: "",
			expected:  false,
		},
		{
			name:      "case sensitive check",
			modelName: "GPT-4.1",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsModelSupported(tt.modelName)
			if result != tt.expected {
				t.Errorf("IsModelSupported(%q) = %v, want %v", tt.modelName, result, tt.expected)
			}
		})
	}
}

func TestValidateFloatParameter(t *testing.T) {
	t.Parallel()
	constraint := ParameterConstraint{
		Type:     "float",
		MinValue: func() *float64 { v := 0.0; return &v }(),
		MaxValue: func() *float64 { v := 2.0; return &v }(),
	}

	tests := []struct {
		name          string
		paramName     string
		value         interface{}
		expectError   bool
		errorContains string
	}{
		// Valid cases
		{
			name:        "valid float64",
			paramName:   "temperature",
			value:       1.0,
			expectError: false,
		},
		{
			name:        "valid float32",
			paramName:   "temperature",
			value:       float32(1.5),
			expectError: false,
		},
		{
			name:        "valid int",
			paramName:   "temperature",
			value:       1,
			expectError: false,
		},
		{
			name:        "valid int32",
			paramName:   "temperature",
			value:       int32(1),
			expectError: false,
		},
		{
			name:        "valid int64",
			paramName:   "temperature",
			value:       int64(1),
			expectError: false,
		},
		{
			name:        "at minimum boundary",
			paramName:   "temperature",
			value:       0.0,
			expectError: false,
		},
		{
			name:        "at maximum boundary",
			paramName:   "temperature",
			value:       2.0,
			expectError: false,
		},
		// Error cases
		{
			name:          "non-numeric type",
			paramName:     "temperature",
			value:         "not_a_number",
			expectError:   true,
			errorContains: "must be a numeric value",
		},
		{
			name:          "below minimum",
			paramName:     "temperature",
			value:         -0.1,
			expectError:   true,
			errorContains: "must be >= 0.00",
		},
		{
			name:          "above maximum",
			paramName:     "temperature",
			value:         2.1,
			expectError:   true,
			errorContains: "must be <= 2.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFloatParameter(tt.paramName, tt.value, constraint)

			if tt.expectError {
				if err == nil {
					t.Errorf("validateFloatParameter(%q, %v, constraint) expected error, got nil", tt.paramName, tt.value)
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("validateFloatParameter(%q, %v, constraint) error = %q, want error containing %q",
						tt.paramName, tt.value, err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("validateFloatParameter(%q, %v, constraint) unexpected error: %v", tt.paramName, tt.value, err)
			}
		})
	}
}

func TestValidateIntParameter(t *testing.T) {
	t.Parallel()
	constraint := ParameterConstraint{
		Type:     "int",
		MinValue: func() *float64 { v := 1.0; return &v }(),
		MaxValue: func() *float64 { v := 1000.0; return &v }(),
	}

	tests := []struct {
		name          string
		paramName     string
		value         interface{}
		expectError   bool
		errorContains string
	}{
		// Valid cases
		{
			name:        "valid int",
			paramName:   "max_tokens",
			value:       500,
			expectError: false,
		},
		{
			name:        "valid int32",
			paramName:   "max_tokens",
			value:       int32(500),
			expectError: false,
		},
		{
			name:        "valid int64",
			paramName:   "max_tokens",
			value:       int64(500),
			expectError: false,
		},
		{
			name:        "valid float64 whole number",
			paramName:   "max_tokens",
			value:       500.0,
			expectError: false,
		},
		{
			name:        "valid float32 whole number",
			paramName:   "max_tokens",
			value:       float32(500.0),
			expectError: false,
		},
		{
			name:        "at minimum boundary",
			paramName:   "max_tokens",
			value:       1,
			expectError: false,
		},
		{
			name:        "at maximum boundary",
			paramName:   "max_tokens",
			value:       1000,
			expectError: false,
		},
		// Error cases
		{
			name:          "non-numeric type",
			paramName:     "max_tokens",
			value:         "not_a_number",
			expectError:   true,
			errorContains: "must be an integer value",
		},
		{
			name:          "float64 with decimal",
			paramName:     "max_tokens",
			value:         500.5,
			expectError:   true,
			errorContains: "must be an integer",
		},
		{
			name:          "float32 with decimal",
			paramName:     "max_tokens",
			value:         float32(500.5),
			expectError:   true,
			errorContains: "must be an integer",
		},
		{
			name:          "below minimum",
			paramName:     "max_tokens",
			value:         0,
			expectError:   true,
			errorContains: "must be >= 1",
		},
		{
			name:          "above maximum",
			paramName:     "max_tokens",
			value:         1001,
			expectError:   true,
			errorContains: "must be <= 1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIntParameter(tt.paramName, tt.value, constraint)

			if tt.expectError {
				if err == nil {
					t.Errorf("validateIntParameter(%q, %v, constraint) expected error, got nil", tt.paramName, tt.value)
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("validateIntParameter(%q, %v, constraint) error = %q, want error containing %q",
						tt.paramName, tt.value, err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("validateIntParameter(%q, %v, constraint) unexpected error: %v", tt.paramName, tt.value, err)
			}
		})
	}
}

func TestValidateStringParameter(t *testing.T) {
	t.Parallel()
	constraintWithEnum := ParameterConstraint{
		Type:       "string",
		EnumValues: []string{"low", "medium", "high"},
	}
	constraintWithoutEnum := ParameterConstraint{
		Type: "string",
	}

	tests := []struct {
		name          string
		paramName     string
		value         interface{}
		constraint    ParameterConstraint
		expectError   bool
		errorContains string
	}{
		// Valid cases with enum
		{
			name:        "valid enum value - low",
			paramName:   "priority",
			value:       "low",
			constraint:  constraintWithEnum,
			expectError: false,
		},
		{
			name:        "valid enum value - medium",
			paramName:   "priority",
			value:       "medium",
			constraint:  constraintWithEnum,
			expectError: false,
		},
		{
			name:        "valid enum value - high",
			paramName:   "priority",
			value:       "high",
			constraint:  constraintWithEnum,
			expectError: false,
		},
		// Valid cases without enum
		{
			name:        "any string value when no enum",
			paramName:   "description",
			value:       "any description",
			constraint:  constraintWithoutEnum,
			expectError: false,
		},
		{
			name:        "empty string when no enum",
			paramName:   "description",
			value:       "",
			constraint:  constraintWithoutEnum,
			expectError: false,
		},
		// Error cases
		{
			name:          "non-string type",
			paramName:     "priority",
			value:         123,
			constraint:    constraintWithEnum,
			expectError:   true,
			errorContains: "must be a string",
		},
		{
			name:          "invalid enum value",
			paramName:     "priority",
			value:         "invalid",
			constraint:    constraintWithEnum,
			expectError:   true,
			errorContains: "must be one of",
		},
		{
			name:          "case sensitive enum check",
			paramName:     "priority",
			value:         "Low",
			constraint:    constraintWithEnum,
			expectError:   true,
			errorContains: "must be one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStringParameter(tt.paramName, tt.value, tt.constraint)

			if tt.expectError {
				if err == nil {
					t.Errorf("validateStringParameter(%q, %v, constraint) expected error, got nil", tt.paramName, tt.value)
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("validateStringParameter(%q, %v, constraint) error = %q, want error containing %q",
						tt.paramName, tt.value, err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("validateStringParameter(%q, %v, constraint) unexpected error: %v", tt.paramName, tt.value, err)
			}
		})
	}
}

// TestTestModelsExist verifies that test models required by integration tests are supported
// This test drives the requirement to add test models to the ModelDefinitions map
// TestOpenRouterConsolidation tests the migration of OpenAI and Gemini models to OpenRouter
// This test will initially fail, driving the implementation of the migration
func TestOpenRouterConsolidation(t *testing.T) {
	t.Parallel()
	migratedModels := []struct {
		name               string
		modelName          string
		expectedProvider   string
		expectedAPIModelID string
		// Fields that must be preserved during migration
		minContextWindow int
		minOutputTokens  int
	}{
		{
			name:               "gpt-4.1 migrated to OpenRouter",
			modelName:          "gpt-4.1",
			expectedProvider:   "openrouter",
			expectedAPIModelID: "openai/gpt-4.1",
			minContextWindow:   1000000,
			minOutputTokens:    200000,
		},
		{
			name:               "o4-mini migrated to OpenRouter",
			modelName:          "o4-mini",
			expectedProvider:   "openrouter",
			expectedAPIModelID: "openai/o4-mini",
			minContextWindow:   200000,
			minOutputTokens:    200000,
		},
		{
			name:               "o3 migrated to OpenRouter",
			modelName:          "o3",
			expectedProvider:   "openrouter",
			expectedAPIModelID: "openai/o3",
			minContextWindow:   200000,
			minOutputTokens:    200000,
		},
		{
			name:               "gemini-2.5-pro migrated to OpenRouter",
			modelName:          "gemini-2.5-pro",
			expectedProvider:   "openrouter",
			expectedAPIModelID: "google/gemini-2.5-pro",
			minContextWindow:   1000000,
			minOutputTokens:    65000,
		},
		{
			name:               "gemini-2.5-flash migrated to OpenRouter",
			modelName:          "gemini-2.5-flash",
			expectedProvider:   "openrouter",
			expectedAPIModelID: "google/gemini-2.5-flash",
			minContextWindow:   1000000,
			minOutputTokens:    65000,
		},
	}

	for _, tt := range migratedModels {
		t.Run(tt.name, func(t *testing.T) {
			info, err := GetModelInfo(tt.modelName)
			if err != nil {
				t.Errorf("GetModelInfo(%s) failed: %v", tt.modelName, err)
				return
			}

			// Test the core migration requirements
			if info.Provider != tt.expectedProvider {
				t.Errorf("GetModelInfo(%s).Provider = %s, want %s", tt.modelName, info.Provider, tt.expectedProvider)
			}

			if info.APIModelID != tt.expectedAPIModelID {
				t.Errorf("GetModelInfo(%s).APIModelID = %s, want %s", tt.modelName, info.APIModelID, tt.expectedAPIModelID)
			}

			// Ensure migration preserves critical functionality
			if info.ContextWindow < tt.minContextWindow {
				t.Errorf("GetModelInfo(%s).ContextWindow = %d, want >= %d", tt.modelName, info.ContextWindow, tt.minContextWindow)
			}

			if info.MaxOutputTokens < tt.minOutputTokens {
				t.Errorf("GetModelInfo(%s).MaxOutputTokens = %d, want >= %d", tt.modelName, info.MaxOutputTokens, tt.minOutputTokens)
			}

			// Ensure default parameters are preserved
			if len(info.DefaultParams) == 0 {
				t.Errorf("GetModelInfo(%s).DefaultParams should not be empty after migration", tt.modelName)
			}

			// Ensure parameter constraints are preserved
			if len(info.ParameterConstraints) == 0 {
				t.Errorf("GetModelInfo(%s).ParameterConstraints should not be empty after migration", tt.modelName)
			}
		})
	}
}

func TestTestModelsExist(t *testing.T) {
	t.Parallel()
	testModels := []struct {
		name     string
		model    string
		provider string
	}{
		{
			name:     "model1 for integration tests",
			model:    "model1",
			provider: "test",
		},
		{
			name:     "model2 for integration tests",
			model:    "model2",
			provider: "test",
		},
		{
			name:     "model3 for integration tests",
			model:    "model3",
			provider: "test",
		},
		{
			name:     "synthesis-model for integration tests",
			model:    "synthesis-model",
			provider: "test",
		},
	}

	for _, tt := range testModels {
		t.Run(tt.name, func(t *testing.T) {
			// Test model should be supported
			if !IsModelSupported(tt.model) {
				t.Errorf("Test model %s should be supported for integration tests", tt.model)
			}

			// Test model should have correct provider and basic properties
			info, err := GetModelInfo(tt.model)
			if err != nil {
				t.Errorf("GetModelInfo(%s) failed: %v", tt.model, err)
				return
			}

			if info.Provider != tt.provider {
				t.Errorf("GetModelInfo(%s).Provider = %s, want %s", tt.model, info.Provider, tt.provider)
			}

			// Test models should have sufficient context for test scenarios (100 tokens + safety margin)
			minRequiredContext := 200 // Conservative minimum for test scenarios
			if info.ContextWindow < minRequiredContext {
				t.Errorf("GetModelInfo(%s).ContextWindow = %d, want >= %d for test scenarios",
					tt.model, info.ContextWindow, minRequiredContext)
			}

			// Test models should have positive output token limits
			if info.MaxOutputTokens <= 0 {
				t.Errorf("GetModelInfo(%s).MaxOutputTokens = %d, want > 0", tt.model, info.MaxOutputTokens)
			}
		})
	}
}

func TestGetAvailableProvidersWithHelpfulMessages(t *testing.T) {
	// Don't run in parallel to avoid environment variable interference

	// Save original environment variables
	originalOpenRouter := os.Getenv("OPENROUTER_API_KEY")
	originalOpenAI := os.Getenv("OPENAI_API_KEY")
	originalGemini := os.Getenv("GEMINI_API_KEY")
	originalTestModels := os.Getenv("THINKTANK_ENABLE_TEST_MODELS")

	// Clean up after test
	defer func() {
		if originalOpenRouter != "" {
			_ = os.Setenv("OPENROUTER_API_KEY", originalOpenRouter)
		} else {
			_ = os.Unsetenv("OPENROUTER_API_KEY")
		}
		if originalOpenAI != "" {
			_ = os.Setenv("OPENAI_API_KEY", originalOpenAI)
		} else {
			_ = os.Unsetenv("OPENAI_API_KEY")
		}
		if originalGemini != "" {
			_ = os.Setenv("GEMINI_API_KEY", originalGemini)
		} else {
			_ = os.Unsetenv("GEMINI_API_KEY")
		}
		if originalTestModels != "" {
			_ = os.Setenv("THINKTANK_ENABLE_TEST_MODELS", originalTestModels)
		} else {
			_ = os.Unsetenv("THINKTANK_ENABLE_TEST_MODELS")
		}
	}()

	t.Run("should log helpful message when old OPENAI_API_KEY is detected", func(t *testing.T) {
		// Clear all API keys first
		_ = os.Unsetenv("OPENROUTER_API_KEY")
		_ = os.Unsetenv("OPENAI_API_KEY")
		_ = os.Unsetenv("GEMINI_API_KEY")

		// Enable test models for the test
		_ = os.Setenv("THINKTANK_ENABLE_TEST_MODELS", "true")

		// Set only old OPENAI_API_KEY
		_ = os.Setenv("OPENAI_API_KEY", "sk-test123")

		// Capture stderr output to check for helpful message
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		providers := GetAvailableProviders()

		// Restore stderr
		_ = w.Close()
		os.Stderr = oldStderr

		// Read captured output
		buf := make([]byte, 1024)
		n, _ := r.Read(buf)
		output := string(buf[:n])

		// Should return only test provider (no openrouter since no API key)
		expectedProviders := []string{"test"}
		if len(providers) != len(expectedProviders) || providers[0] != expectedProviders[0] {
			t.Errorf("GetAvailableProviders() with OPENAI_API_KEY = %v, want %v", providers, expectedProviders)
		}

		// Should contain helpful message about old API key
		if !strings.Contains(output, "OPENAI_API_KEY detected but no longer used") {
			t.Errorf("Expected helpful message about OPENAI_API_KEY, got output: %q", output)
		}
		if !strings.Contains(output, "OPENROUTER_API_KEY") {
			t.Errorf("Expected mention of OPENROUTER_API_KEY, got output: %q", output)
		}
		if !strings.Contains(output, "https://openrouter.ai/keys") {
			t.Errorf("Expected OpenRouter URL, got output: %q", output)
		}
	})

	t.Run("should log helpful message when old GEMINI_API_KEY is detected", func(t *testing.T) {
		// Clear all API keys first
		_ = os.Unsetenv("OPENROUTER_API_KEY")
		_ = os.Unsetenv("OPENAI_API_KEY")
		_ = os.Unsetenv("GEMINI_API_KEY")

		// Enable test models for the test
		_ = os.Setenv("THINKTANK_ENABLE_TEST_MODELS", "true")

		// Set only old GEMINI_API_KEY
		_ = os.Setenv("GEMINI_API_KEY", "test-gemini-key")

		// Capture stderr output to check for helpful message
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		providers := GetAvailableProviders()

		// Restore stderr
		_ = w.Close()
		os.Stderr = oldStderr

		// Read captured output
		buf := make([]byte, 1024)
		n, _ := r.Read(buf)
		output := string(buf[:n])

		// Should return only test provider (no openrouter since no API key)
		expectedProviders := []string{"test"}
		if len(providers) != len(expectedProviders) || providers[0] != expectedProviders[0] {
			t.Errorf("GetAvailableProviders() with GEMINI_API_KEY = %v, want %v", providers, expectedProviders)
		}

		// Should contain helpful message about old API key but NOT about OPENAI_API_KEY
		if !strings.Contains(output, "GEMINI_API_KEY detected but no longer used") {
			t.Errorf("Expected helpful message about GEMINI_API_KEY, got output: %q", output)
		}
		if strings.Contains(output, "OPENAI_API_KEY") {
			t.Errorf("Should not contain OPENAI_API_KEY message when only GEMINI_API_KEY is set, got output: %q", output)
		}
		if !strings.Contains(output, "OPENROUTER_API_KEY") {
			t.Errorf("Expected mention of OPENROUTER_API_KEY, got output: %q", output)
		}
		if !strings.Contains(output, "https://openrouter.ai/keys") {
			t.Errorf("Expected OpenRouter URL, got output: %q", output)
		}
	})

	t.Run("should not show message when OPENROUTER_API_KEY is properly set", func(t *testing.T) {
		// Clear all API keys first
		_ = os.Unsetenv("OPENROUTER_API_KEY")
		_ = os.Unsetenv("OPENAI_API_KEY")
		_ = os.Unsetenv("GEMINI_API_KEY")

		// Enable test models for the test
		_ = os.Setenv("THINKTANK_ENABLE_TEST_MODELS", "true")

		// Set proper OPENROUTER_API_KEY
		_ = os.Setenv("OPENROUTER_API_KEY", "or-test-key")

		// Capture stderr output to check for helpful message
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		providers := GetAvailableProviders()

		// Restore stderr
		_ = w.Close()
		os.Stderr = oldStderr

		// Read captured output
		buf := make([]byte, 1024)
		n, _ := r.Read(buf)
		output := string(buf[:n])

		// Should return openrouter and test providers
		expectedProviders := []string{"openrouter", "test"}
		if len(providers) != 2 || providers[0] != "openrouter" || providers[1] != "test" {
			t.Errorf("GetAvailableProviders() with OPENROUTER_API_KEY = %v, want %v", providers, expectedProviders)
		}

		// Should NOT contain any helpful messages since API key is properly set
		if strings.Contains(output, "detected but no longer used") {
			t.Errorf("Should not show helpful message when OPENROUTER_API_KEY is set, got output: %q", output)
		}
	})
}
