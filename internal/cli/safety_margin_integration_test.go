package cli

import (
	"context"
	"testing"

	"github.com/phrazzld/thinktank/internal/thinktank"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSafetyMarginIntegration tests end-to-end integration of CLI flag to model selection
func TestSafetyMarginIntegration(t *testing.T) {
	// Note: Cannot use t.Parallel() when using t.Setenv()

	tests := []struct {
		name           string
		args           []string
		expectedMargin uint8
	}{
		{
			name: "default safety margin",
			args: []string{
				"thinktank",
				"--dry-run",
				"../../README.md",
				".",
			},
			expectedMargin: 20, // Default
		},
		{
			name: "custom safety margin via separate flag",
			args: []string{
				"thinktank",
				"--token-safety-margin", "30",
				"--dry-run",
				"../../README.md",
				".",
			},
			expectedMargin: 30,
		},
		{
			name: "custom safety margin via equals syntax",
			args: []string{
				"thinktank",
				"--token-safety-margin=15",
				"--dry-run",
				"../../README.md",
				".",
			},
			expectedMargin: 15,
		},
		{
			name: "zero safety margin",
			args: []string{
				"thinktank",
				"--token-safety-margin", "0",
				"--dry-run",
				"../../README.md",
				".",
			},
			expectedMargin: 0,
		},
		{
			name: "maximum safety margin",
			args: []string{
				"thinktank",
				"--token-safety-margin", "50",
				"--dry-run",
				"../../README.md",
				".",
			},
			expectedMargin: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set a mock API key to ensure the function doesn't take the early return path
			t.Setenv("OPENAI_API_KEY", "test-key")

			// Parse CLI arguments
			config, err := ParseSimpleArgsWithArgs(tt.args)
			require.NoError(t, err, "CLI parsing should succeed")

			// Verify the safety margin was parsed correctly
			assert.Equal(t, tt.expectedMargin, config.SafetyMargin,
				"Safety margin should be correctly parsed from CLI arguments")

			// Test model selection integration (basic validation that it doesn't crash)
			// This exercises the selectModelsForConfigWithService path
			tokenService := NewMockTokenService()
			models, synthesis := selectModelsForConfigWithService(config, tokenService)

			// Basic validation that model selection worked
			assert.NotEmpty(t, models, "Should select at least one model")
			assert.NotEmpty(t, synthesis, "Should have a synthesis model")

			// Verify the mock received the correct safety margin
			assert.Equal(t, tt.expectedMargin, tokenService.LastSafetyMargin,
				"TokenCountingService should receive the correct safety margin")
		})
	}
}

// MockTokenService for testing integration without external dependencies
type MockTokenService struct {
	LastSafetyMargin uint8
}

func NewMockTokenService() *MockTokenService {
	return &MockTokenService{}
}

func (m *MockTokenService) CountTokens(ctx context.Context, req thinktank.TokenCountingRequest) (thinktank.TokenCountingResult, error) {
	m.LastSafetyMargin = req.SafetyMarginPercent
	return thinktank.TokenCountingResult{TotalTokens: 100}, nil
}

func (m *MockTokenService) CountTokensForModel(ctx context.Context, req thinktank.TokenCountingRequest, modelName string) (thinktank.ModelTokenCountingResult, error) {
	m.LastSafetyMargin = req.SafetyMarginPercent
	return thinktank.ModelTokenCountingResult{
		TokenCountingResult: thinktank.TokenCountingResult{TotalTokens: 100},
		ModelName:           modelName,
		TokenizerUsed:       "mock",
		Provider:            "mock",
		IsAccurate:          true,
	}, nil
}

func (m *MockTokenService) GetCompatibleModels(ctx context.Context, req thinktank.TokenCountingRequest, availableProviders []string) ([]thinktank.ModelCompatibility, error) {
	m.LastSafetyMargin = req.SafetyMarginPercent

	// Return a few mock compatible models
	return []thinktank.ModelCompatibility{
		{
			ModelName:     "gpt-4.1",
			IsCompatible:  true,
			TokenCount:    100,
			ContextWindow: 1000,
			UsableContext: 800, // Mock 20% safety margin
			Provider:      "openai",
			TokenizerUsed: "mock",
			IsAccurate:    true,
		},
		{
			ModelName:     "gemini-2.5-pro",
			IsCompatible:  true,
			TokenCount:    100,
			ContextWindow: 2000,
			UsableContext: 1600, // Mock 20% safety margin
			Provider:      "gemini",
			TokenizerUsed: "mock",
			IsAccurate:    true,
		},
	}, nil
}
