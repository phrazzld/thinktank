package openai

import (
	"errors"
	"testing"

	"github.com/phrazzld/architect/internal/llm"
)

// Test that APIError implements the llm.CategorizedError interface correctly
func TestAPIError_Category(t *testing.T) {
	tests := []struct {
		name             string
		errorType        ErrorType
		expectedCategory llm.ErrorCategory
	}{
		{"Auth Error", ErrorTypeAuth, llm.CategoryAuth},
		{"Rate Limit Error", ErrorTypeRateLimit, llm.CategoryRateLimit},
		{"Invalid Request Error", ErrorTypeInvalidRequest, llm.CategoryInvalidRequest},
		{"Not Found Error", ErrorTypeNotFound, llm.CategoryNotFound},
		{"Server Error", ErrorTypeServer, llm.CategoryServer},
		{"Network Error", ErrorTypeNetwork, llm.CategoryNetwork},
		{"Cancelled Error", ErrorTypeCancelled, llm.CategoryCancelled},
		{"Input Limit Error", ErrorTypeInputLimit, llm.CategoryInputLimit},
		{"Content Filtered Error", ErrorTypeContentFiltered, llm.CategoryContentFiltered},
		{"Unknown Error", ErrorTypeUnknown, llm.CategoryUnknown},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			apiErr := &APIError{
				Type: tc.errorType,
			}

			// Verify it implements the interface
			var catErr llm.CategorizedError = apiErr

			// Check that the category is correctly mapped
			if category := catErr.Category(); category != tc.expectedCategory {
				t.Errorf("Expected category %v, got %v", tc.expectedCategory, category)
			}
		})
	}
}

// Test that an APIError can be detected using llm.IsCategorizedError
func TestAPIError_IsCategorizedError(t *testing.T) {
	apiErr := &APIError{
		Type:    ErrorTypeRateLimit,
		Message: "Rate limit exceeded",
	}

	// Test with direct error
	catErr, ok := llm.IsCategorizedError(apiErr)
	if !ok {
		t.Errorf("Expected APIError to be detected as CategorizedError")
		return
	}
	if catErr.Category() != llm.CategoryRateLimit {
		t.Errorf("Expected category RateLimit, got %v", catErr.Category())
	}

	// Test with wrapped error
	wrappedErr := errors.New("wrapper: " + apiErr.Error())
	_, ok = llm.IsCategorizedError(wrappedErr)
	if ok {
		t.Errorf("Expected wrapped standard error not to be detected as CategorizedError")
	}

	// Test with fmt.Errorf and %w
	wrappedWithW := errors.New("wrapped with %%w")
	fmtErr := FormatAPIError(wrappedWithW, 429) // This should be a rate limit error
	catErr, ok = llm.IsCategorizedError(fmtErr)
	if !ok {
		t.Errorf("Expected FormatAPIError result to be detected as CategorizedError")
		return
	}
	if catErr.Category() != llm.CategoryRateLimit {
		t.Errorf("Expected category RateLimit, got %v", catErr.Category())
	}
}
