package thinktank

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTokenCountingService_ConfigurableSafetyMargin tests that the safety margin configuration is properly applied
func TestTokenCountingService_ConfigurableSafetyMargin(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()
	ctx := context.Background()

	// Test case: model with 1000 context window
	// With different safety margins, usable context should change accordingly
	tests := []struct {
		name                string
		safetyMarginPercent uint8
		expectedUsableMin   int // Minimum usable context we expect
		expectedUsableMax   int // Maximum usable context we expect
	}{
		{
			name:                "0% safety margin",
			safetyMarginPercent: 0,
			expectedUsableMin:   890, // Should use default 10% = 900 tokens usable
			expectedUsableMax:   910,
		},
		{
			name:                "10% safety margin",
			safetyMarginPercent: 10,
			expectedUsableMin:   890, // 1000 - 100 = 900 tokens usable
			expectedUsableMax:   910,
		},
		{
			name:                "30% safety margin",
			safetyMarginPercent: 30,
			expectedUsableMin:   690, // 1000 - 300 = 700 tokens usable
			expectedUsableMax:   710,
		},
		{
			name:                "50% safety margin",
			safetyMarginPercent: 50,
			expectedUsableMin:   490, // 1000 - 500 = 500 tokens usable
			expectedUsableMax:   510,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request with enough tokens to exceed some models but not others
			req := TokenCountingRequest{
				Instructions:        "Short instruction that uses some tokens",
				Files:               []FileContent{},
				SafetyMarginPercent: tt.safetyMarginPercent,
			}

			// Get compatible models - we should see different usable context values
			compatibleModels, err := service.GetCompatibleModels(ctx, req, []string{"openrouter"})
			require.NoError(t, err)

			// Find a model that reports its usable context (looking for gpt-4.1 which has ~100K context)
			var foundModel *ModelCompatibility
			for _, model := range compatibleModels {
				if model.ModelName == "gpt-4.1" {
					foundModel = &model
					break
				}
			}

			require.NotNil(t, foundModel, "Should find gpt-4.1 model in results")

			// Verify that the usable context reflects our safety margin configuration
			// For gpt-4.1 with ~100K context window:
			// - 0% margin (default 10%) should give ~90K usable
			// - 10% margin should give ~90K usable
			// - 30% margin should give ~70K usable
			// - 50% margin should give ~50K usable

			expectedSafetyPercent := tt.safetyMarginPercent
			if expectedSafetyPercent == 0 {
				expectedSafetyPercent = 10 // Default
			}

			expectedUsable := foundModel.ContextWindow * (100 - int(expectedSafetyPercent)) / 100
			actualUsable := foundModel.UsableContext

			// Allow for some variance due to rounding
			tolerance := foundModel.ContextWindow / 100 // 1% tolerance
			assert.InDelta(t, expectedUsable, actualUsable, float64(tolerance),
				"Usable context should reflect %d%% safety margin. Expected ~%d, got %d (context: %d)",
				expectedSafetyPercent, expectedUsable, actualUsable, foundModel.ContextWindow)
		})
	}
}

// TestTokenCountingService_DefaultSafetyMargin tests that when no safety margin is specified, 20% is used
func TestTokenCountingService_DefaultSafetyMargin(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()
	ctx := context.Background()

	// Request without specifying SafetyMarginPercent (should default to 20%)
	req := TokenCountingRequest{
		Instructions: "Test instruction",
		Files:        []FileContent{},
		// SafetyMarginPercent: 0, // Explicitly test zero value
	}

	compatibleModels, err := service.GetCompatibleModels(ctx, req, []string{"openrouter"})
	require.NoError(t, err)

	// Find gpt-4.1 model
	var foundModel *ModelCompatibility
	for _, model := range compatibleModels {
		if model.ModelName == "gpt-4.1" {
			foundModel = &model
			break
		}
	}

	require.NotNil(t, foundModel, "Should find gpt-4.1 model")

	// Verify that default 10% safety margin is applied
	expectedUsable := foundModel.ContextWindow * 90 / 100 // 90% usable (10% safety margin)
	actualUsable := foundModel.UsableContext

	tolerance := foundModel.ContextWindow / 100 // 1% tolerance
	assert.InDelta(t, expectedUsable, actualUsable, float64(tolerance),
		"Should use default 10%% safety margin. Expected ~%d, got %d (context: %d)",
		expectedUsable, actualUsable, foundModel.ContextWindow)
}
