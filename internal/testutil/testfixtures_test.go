package testutil

import (
	"errors"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
)

// TestModelFixtures verifies that model fixtures are correctly defined
func TestModelFixtures(t *testing.T) {
	// Verify CreateTestModels returns expected models
	models := CreateTestModels()

	// Check number of models
	if len(models) < 9 {
		t.Errorf("Expected at least 9 models, got %d", len(models))
	}

	// Check key models exist
	requiredModels := []string{
		"gpt-4",
		"gemini-1.5-pro",
		"claude-3-opus",
		"test-model",
	}

	for _, name := range requiredModels {
		if _, exists := models[name]; !exists {
			t.Errorf("Expected model %s not found in test models", name)
		}
	}

	// Validate that models have required fields
	for name, model := range models {
		if model.Name != name {
			t.Errorf("Model %s has incorrect name: %s", name, model.Name)
		}

		if model.Provider == "" {
			t.Errorf("Model %s has empty provider", name)
		}

		if model.APIModelID == "" {
			t.Errorf("Model %s has empty APIModelID", name)
		}

		if len(model.Parameters) == 0 {
			t.Errorf("Model %s has no parameters", name)
		}
	}
}

// TestProviderFixtures verifies that provider fixtures are correctly defined
func TestProviderFixtures(t *testing.T) {
	// Verify CreateTestProviders returns expected providers
	providers := CreateTestProviders()

	// Check number of providers
	if len(providers) < 6 {
		t.Errorf("Expected at least 6 providers, got %d", len(providers))
	}

	// Check key providers exist
	requiredProviders := []string{
		"openai",
		"gemini",
		"openrouter",
		"test-provider",
	}

	for _, name := range requiredProviders {
		if _, exists := providers[name]; !exists {
			t.Errorf("Expected provider %s not found in test providers", name)
		}
	}

	// Validate custom base URLs
	customProviders := map[string]bool{
		"openai-custom":     true,
		"openrouter-custom": true,
		"test-provider":     true,
	}

	for name, provider := range providers {
		if provider.Name != name {
			t.Errorf("Provider %s has incorrect name: %s", name, provider.Name)
		}

		if customProviders[name] && provider.BaseURL == "" {
			t.Errorf("Custom provider %s has empty BaseURL", name)
		}
	}
}

// TestResponseFixtures verifies that response fixtures are correctly defined
func TestResponseFixtures(t *testing.T) {
	// Test BasicSuccessResponse
	if BasicSuccessResponse.Content == "" {
		t.Error("BasicSuccessResponse has empty content")
	}
	if BasicSuccessResponse.FinishReason != "stop" {
		t.Errorf("BasicSuccessResponse has incorrect FinishReason: %s", BasicSuccessResponse.FinishReason)
	}
	if BasicSuccessResponse.Truncated {
		t.Error("BasicSuccessResponse should not be truncated")
	}

	// Test TruncatedResponse
	if TruncatedResponse.Content == "" {
		t.Error("TruncatedResponse has empty content")
	}
	if TruncatedResponse.FinishReason != "length" {
		t.Errorf("TruncatedResponse has incorrect FinishReason: %s", TruncatedResponse.FinishReason)
	}
	if !TruncatedResponse.Truncated {
		t.Error("TruncatedResponse should be truncated")
	}

	// Test SafetyBlockedResponse
	if SafetyBlockedResponse.Content != "" {
		t.Errorf("SafetyBlockedResponse should have empty content, got: %s", SafetyBlockedResponse.Content)
	}
	if SafetyBlockedResponse.FinishReason != "safety" {
		t.Errorf("SafetyBlockedResponse has incorrect FinishReason: %s", SafetyBlockedResponse.FinishReason)
	}
	if len(SafetyBlockedResponse.SafetyInfo) == 0 {
		t.Error("SafetyBlockedResponse has no safety info")
	} else if !SafetyBlockedResponse.SafetyInfo[0].Blocked {
		t.Error("SafetyBlockedResponse should have blocked=true in safety info")
	}

	// Test CreateSuccessResponse
	customContent := "Custom success response"
	resp := CreateSuccessResponse(customContent)
	if resp.Content != customContent {
		t.Errorf("CreateSuccessResponse returned incorrect content: %s", resp.Content)
	}
	if resp.FinishReason != "stop" {
		t.Errorf("CreateSuccessResponse returned incorrect FinishReason: %s", resp.FinishReason)
	}
	if resp.Truncated {
		t.Error("CreateSuccessResponse should not return truncated response")
	}
}

// TestErrorFixtures verifies that error fixtures are correctly defined
func TestErrorFixtures(t *testing.T) {
	// Test CreateAuthError
	authErr := CreateAuthError("test-provider")
	if authErr.Provider != "test-provider" {
		t.Errorf("CreateAuthError has incorrect provider: %s", authErr.Provider)
	}
	if authErr.Category() != llm.CategoryAuth {
		t.Errorf("CreateAuthError has incorrect category: %v", authErr.Category())
	}
	if !errors.Is(authErr, llm.ErrAPICall) {
		t.Error("CreateAuthError should wrap ErrAPICall")
	}

	// Test CreateRateLimitError
	rateLimitErr := CreateRateLimitError("test-provider")
	if rateLimitErr.Provider != "test-provider" {
		t.Errorf("CreateRateLimitError has incorrect provider: %s", rateLimitErr.Provider)
	}
	if rateLimitErr.Category() != llm.CategoryRateLimit {
		t.Errorf("CreateRateLimitError has incorrect category: %v", rateLimitErr.Category())
	}

	// Test CreateSafetyError
	safetyErr := CreateSafetyError("test-provider")
	if safetyErr.Provider != "test-provider" {
		t.Errorf("CreateSafetyError has incorrect provider: %s", safetyErr.Provider)
	}
	if safetyErr.Category() != llm.CategoryContentFiltered {
		t.Errorf("CreateSafetyError has incorrect category: %v", safetyErr.Category())
	}
	if !errors.Is(safetyErr, llm.ErrSafetyBlocked) {
		t.Error("CreateSafetyError should wrap ErrSafetyBlocked")
	}

	// Test CreateModelNotFoundError
	modelNotFoundErr := CreateModelNotFoundError("nonexistent-model")
	if modelNotFoundErr.Provider != "registry" {
		t.Errorf("CreateModelNotFoundError has incorrect provider: %s", modelNotFoundErr.Provider)
	}
	if modelNotFoundErr.Category() != llm.CategoryNotFound {
		t.Errorf("CreateModelNotFoundError has incorrect category: %v", modelNotFoundErr.Category())
	}
	if !errors.Is(modelNotFoundErr, llm.ErrModelNotFound) {
		t.Error("CreateModelNotFoundError should wrap ErrModelNotFound")
	}
}
