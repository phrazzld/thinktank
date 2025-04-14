// internal/gemini/helper_methods_test.go
// Tests for helper methods in the gemini client

package gemini

import (
	"testing"

	genai "github.com/google/generative-ai-go/genai"
)

func TestHelperMethods(t *testing.T) {
	// This test verifies helper methods:
	// - mapSafetyRatings correctly converts between types
	// - GetModelName returns the correct model name
	// - GetTemperature returns the correct temperature
	// - GetMaxOutputTokens returns the correct token limit
	// - GetTopP returns the correct topP value

	t.Run("mapSafetyRatings with nil ratings", func(t *testing.T) {
		// When passed nil, should return nil
		result := mapSafetyRatings(nil)
		if result != nil {
			t.Errorf("Expected nil result for nil input, got %+v", result)
		}
	})

	t.Run("mapSafetyRatings with empty ratings", func(t *testing.T) {
		// When passed empty slice, should return empty slice
		result := mapSafetyRatings([]*genai.SafetyRating{})
		if result == nil {
			t.Error("Expected empty slice, got nil")
		}
		if len(result) != 0 {
			t.Errorf("Expected empty slice, got slice with %d elements", len(result))
		}
	})

	t.Run("mapSafetyRatings with actual ratings", func(t *testing.T) {
		// Create sample genai safety ratings
		ratings := []*genai.SafetyRating{
			{
				Category:    genai.HarmCategoryHarassment,
				Probability: genai.HarmProbabilityMedium,
				Blocked:     true,
			},
			{
				Category:    genai.HarmCategoryDangerousContent,
				Probability: genai.HarmProbabilityLow,
				Blocked:     false,
			},
		}

		// Map to our internal format
		result := mapSafetyRatings(ratings)

		// Verify the result
		if len(result) != 2 {
			t.Fatalf("Expected 2 ratings, got %d", len(result))
		}

		// Check first rating
		if string(result[0].Category) != string(genai.HarmCategoryHarassment) {
			t.Errorf("Expected category %q, got %q", genai.HarmCategoryHarassment, result[0].Category)
		}
		if result[0].Score != float32(genai.HarmProbabilityMedium) {
			t.Errorf("Expected score %f, got %f", float32(genai.HarmProbabilityMedium), result[0].Score)
		}
		if !result[0].Blocked {
			t.Error("Expected blocked to be true, got false")
		}

		// Check second rating
		if string(result[1].Category) != string(genai.HarmCategoryDangerousContent) {
			t.Errorf("Expected category %q, got %q", genai.HarmCategoryDangerousContent, result[1].Category)
		}
		if result[1].Score != float32(genai.HarmProbabilityLow) {
			t.Errorf("Expected score %f, got %f", float32(genai.HarmProbabilityLow), result[1].Score)
		}
		if result[1].Blocked {
			t.Error("Expected blocked to be false, got true")
		}
	})

	t.Run("GetModelName returns correct value", func(t *testing.T) {
		// Test with the actual implementation
		const expectedModelName = "test-model-name"

		client := &geminiClient{
			modelName: expectedModelName,
		}

		modelName := client.GetModelName()
		if modelName != expectedModelName {
			t.Errorf("Expected model name %q, got %q", expectedModelName, modelName)
		}

		// Test with MockClient
		mockClient := &MockClient{
			GetModelNameFunc: func() string {
				return "mock-model-name"
			},
		}

		mockModelName := mockClient.GetModelName()
		if mockModelName != "mock-model-name" {
			t.Errorf("Expected mock model name %q, got %q", "mock-model-name", mockModelName)
		}

		// Test default mock implementation
		defaultMockClient := NewMockClient()
		defaultMockName := defaultMockClient.GetModelName()
		if defaultMockName != "mock-model" {
			t.Errorf("Expected default mock model name %q, got %q", "mock-model", defaultMockName)
		}
	})

	t.Run("GetTemperature returns correct value", func(t *testing.T) {
		defaultTemp := DefaultModelConfig().Temperature
		customTemp := float32(0.42)

		// Test actual implementation with nil model
		clientWithNilModel := &geminiClient{
			model:  nil,
			logger: getTestLogger(),
		}

		tempFromNilModel := clientWithNilModel.GetTemperature()
		if tempFromNilModel != defaultTemp {
			t.Errorf("Expected default temperature %f from nil model, got %f", defaultTemp, tempFromNilModel)
		}

		// Test with MockClient with custom function
		mockClient := &MockClient{
			GetTemperatureFunc: func() float32 {
				return customTemp
			},
		}

		mockTemp := mockClient.GetTemperature()
		if mockTemp != customTemp {
			t.Errorf("Expected mock temperature %f, got %f", customTemp, mockTemp)
		}

		// Test default mock implementation
		defaultMockClient := NewMockClient()
		defaultMockTemp := defaultMockClient.GetTemperature()
		if defaultMockTemp != defaultTemp {
			t.Errorf("Expected default mock temperature %f, got %f", defaultTemp, defaultMockTemp)
		}
	})

	t.Run("GetMaxOutputTokens returns correct value", func(t *testing.T) {
		defaultTokens := DefaultModelConfig().MaxOutputTokens
		customTokens := int32(4096)

		// Test actual implementation with nil model
		clientWithNilModel := &geminiClient{
			model:  nil,
			logger: getTestLogger(),
		}

		tokensFromNilModel := clientWithNilModel.GetMaxOutputTokens()
		if tokensFromNilModel != defaultTokens {
			t.Errorf("Expected default tokens %d from nil model, got %d", defaultTokens, tokensFromNilModel)
		}

		// Test with MockClient with custom function
		mockClient := &MockClient{
			GetMaxOutputTokensFunc: func() int32 {
				return customTokens
			},
		}

		mockTokens := mockClient.GetMaxOutputTokens()
		if mockTokens != customTokens {
			t.Errorf("Expected mock tokens %d, got %d", customTokens, mockTokens)
		}

		// Test default mock implementation
		defaultMockClient := NewMockClient()
		defaultMockTokens := defaultMockClient.GetMaxOutputTokens()
		if defaultMockTokens != defaultTokens {
			t.Errorf("Expected default mock tokens %d, got %d", defaultTokens, defaultMockTokens)
		}
	})

	t.Run("GetTopP returns correct value", func(t *testing.T) {
		defaultTopP := DefaultModelConfig().TopP
		customTopP := float32(0.75)

		// Test actual implementation with nil model
		clientWithNilModel := &geminiClient{
			model:  nil,
			logger: getTestLogger(),
		}

		topPFromNilModel := clientWithNilModel.GetTopP()
		if topPFromNilModel != defaultTopP {
			t.Errorf("Expected default topP %f from nil model, got %f", defaultTopP, topPFromNilModel)
		}

		// Test with MockClient with custom function
		mockClient := &MockClient{
			GetTopPFunc: func() float32 {
				return customTopP
			},
		}

		mockTopP := mockClient.GetTopP()
		if mockTopP != customTopP {
			t.Errorf("Expected mock topP %f, got %f", customTopP, mockTopP)
		}

		// Test default mock implementation
		defaultMockClient := NewMockClient()
		defaultMockTopP := defaultMockClient.GetTopP()
		if defaultMockTopP != defaultTopP {
			t.Errorf("Expected default mock topP %f, got %f", defaultTopP, defaultMockTopP)
		}
	})
}
