package models

import (
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
		// Valid models - OpenAI
		{
			name:           "gpt-4.1 valid model",
			modelName:      "gpt-4.1",
			wantProvider:   "openai",
			wantAPIModelID: "gpt-4.1",
			wantError:      false,
		},
		{
			name:           "o4-mini valid model",
			modelName:      "o4-mini",
			wantProvider:   "openai",
			wantAPIModelID: "o4-mini",
			wantError:      false,
		},
		// Valid models - Gemini
		{
			name:           "gemini-2.5-pro valid model",
			modelName:      "gemini-2.5-pro",
			wantProvider:   "gemini",
			wantAPIModelID: "gemini-2.5-pro",
			wantError:      false,
		},
		{
			name:           "gemini-2.5-flash valid model",
			modelName:      "gemini-2.5-flash",
			wantProvider:   "gemini",
			wantAPIModelID: "gemini-2.5-flash",
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
			name:     "openai provider",
			provider: "openai",
			expected: 3000,
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
			expected: []string{"gpt-4.1", "o3", "o4-mini"},
		},
		{
			name:     "gemini provider",
			provider: "gemini",
			expected: []string{"gemini-2.5-pro", "gemini-2.5-flash"},
		},
		{
			name:     "openrouter provider",
			provider: "openrouter",
			expected: []string{
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
