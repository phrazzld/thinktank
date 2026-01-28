package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/misty-step/thinktank/internal/llm"
)

// TestErrorCategorization tests the error categorization functionality
func TestErrorCategorization(t *testing.T) {
	tests := []struct {
		name             string
		errorMessage     string
		httpStatus       int
		expectedCategory llm.ErrorCategory
	}{
		{
			name:             "Auth error via status code",
			errorMessage:     "some error",
			httpStatus:       401,
			expectedCategory: llm.CategoryAuth,
		},
		{
			name:             "Rate limit error via status code",
			errorMessage:     "some error",
			httpStatus:       429,
			expectedCategory: llm.CategoryRateLimit,
		},
		{
			name:             "Server error via status code",
			errorMessage:     "some error",
			httpStatus:       500,
			expectedCategory: llm.CategoryServer,
		},
		{
			name:             "Auth error via message",
			errorMessage:     "invalid API key",
			httpStatus:       0,
			expectedCategory: llm.CategoryAuth,
		},
		{
			name:             "Rate limit error via message",
			errorMessage:     "rate limit exceeded",
			httpStatus:       0,
			expectedCategory: llm.CategoryRateLimit,
		},
		{
			name:             "Content filtered error via message",
			errorMessage:     "content was filtered by safety settings",
			httpStatus:       0,
			expectedCategory: llm.CategoryContentFiltered,
		},
		{
			name:             "Input limit error via message",
			errorMessage:     "token limit exceeded",
			httpStatus:       0,
			expectedCategory: llm.CategoryInputLimit,
		},
		{
			name:             "Network error via message",
			errorMessage:     "network connection failed",
			httpStatus:       0,
			expectedCategory: llm.CategoryNetwork,
		},
		{
			name:             "Cancelled error via message",
			errorMessage:     "context cancelled",
			httpStatus:       0,
			expectedCategory: llm.CategoryCancelled,
		},
		{
			name:             "Not found error via message",
			errorMessage:     "model not found",
			httpStatus:       0,
			expectedCategory: llm.CategoryNotFound,
		},
		{
			name:             "Invalid request error via message",
			errorMessage:     "invalid request parameters",
			httpStatus:       0,
			expectedCategory: llm.CategoryInvalidRequest,
		},
		{
			name:             "Unknown error",
			errorMessage:     "some random error",
			httpStatus:       0,
			expectedCategory: llm.CategoryUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errorMessage)
			category := llm.DetectErrorCategory(err, tt.httpStatus)

			if category != tt.expectedCategory {
				t.Errorf("Expected category %v, got %v", tt.expectedCategory, category)
			}
		})
	}
}

// TestGetErrorCategoryFromStatusCode tests HTTP status code to error category mapping
func TestGetErrorCategoryFromStatusCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		statusCode       int
		expectedCategory llm.ErrorCategory
	}{
		{401, llm.CategoryAuth},
		{403, llm.CategoryAuth},
		{402, llm.CategoryInsufficientCredits},
		{429, llm.CategoryRateLimit},
		{400, llm.CategoryInvalidRequest},
		{404, llm.CategoryNotFound},
		{500, llm.CategoryServer},
		{502, llm.CategoryServer},
		{503, llm.CategoryServer},
		{504, llm.CategoryServer},
		{200, llm.CategoryUnknown}, // Success code should be Unknown category
		{302, llm.CategoryUnknown}, // Redirect code should be Unknown category
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("HTTP %d", tt.statusCode), func(t *testing.T) {
			category := llm.GetErrorCategoryFromStatusCode(tt.statusCode)
			if category != tt.expectedCategory {
				t.Errorf("Expected category %v for status code %d, got %v",
					tt.expectedCategory, tt.statusCode, category)
			}
		})
	}
}

// TestGetErrorCategoryFromMessage tests error message pattern to category mapping
func TestGetErrorCategoryFromMessage(t *testing.T) {
	tests := []struct {
		message          string
		expectedCategory llm.ErrorCategory
	}{
		{"API key invalid", llm.CategoryAuth},
		{"Unauthorized access", llm.CategoryAuth},
		{"Invalid authorization", llm.CategoryAuth},

		{"Rate limit exceeded", llm.CategoryRateLimit},
		{"Too many requests", llm.CategoryRateLimit},
		{"Quota exceeded", llm.CategoryRateLimit},

		{"Insufficient credits", llm.CategoryInsufficientCredits},
		{"Payment required", llm.CategoryInsufficientCredits},
		{"Billing account issue", llm.CategoryInsufficientCredits},

		{"Content blocked by safety filters", llm.CategoryContentFiltered},
		{"Filtered due to moderation", llm.CategoryContentFiltered},
		{"Content does not comply with safety", llm.CategoryContentFiltered},

		{"Token limit exceeded", llm.CategoryInputLimit},
		{"Maximum context length exceeded", llm.CategoryInputLimit},
		{"Tokens exceeds model limit", llm.CategoryInputLimit},

		{"Network connection failed", llm.CategoryNetwork},
		{"Timeout while connecting", llm.CategoryNetwork},
		{"Connection error", llm.CategoryNetwork},

		{"Operation cancelled", llm.CategoryCancelled},
		{"Context deadline exceeded", llm.CategoryCancelled},
		{"Request cancelled", llm.CategoryCancelled},

		{"Model not found", llm.CategoryNotFound},
		{"No such model exists", llm.CategoryNotFound},
		{"Unknown model", llm.CategoryNotFound},

		{"Invalid request parameters", llm.CategoryInvalidRequest},
		{"Bad request format", llm.CategoryInvalidRequest},
		{"Invalid parameters", llm.CategoryInvalidRequest},

		{"Unknown error", llm.CategoryUnknown},
		{"Something went wrong", llm.CategoryUnknown},
		{"General error", llm.CategoryUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			category := llm.GetErrorCategoryFromMessage(tt.message)
			if category != tt.expectedCategory {
				t.Errorf("Expected category %v for message %q, got %v",
					tt.expectedCategory, tt.message, category)
			}
		})
	}
}

// TestErrorCategoryString tests the string representation of error categories
func TestErrorCategoryString(t *testing.T) {
	tests := []struct {
		category llm.ErrorCategory
		expected string
	}{
		{llm.CategoryUnknown, "Unknown"},
		{llm.CategoryAuth, "Auth"},
		{llm.CategoryRateLimit, "RateLimit"},
		{llm.CategoryInvalidRequest, "InvalidRequest"},
		{llm.CategoryNotFound, "NotFound"},
		{llm.CategoryServer, "Server"},
		{llm.CategoryNetwork, "Network"},
		{llm.CategoryCancelled, "Cancelled"},
		{llm.CategoryInputLimit, "InputLimit"},
		{llm.CategoryContentFiltered, "ContentFiltered"},
		{llm.CategoryInsufficientCredits, "InsufficientCredits"},
		{llm.ErrorCategory(999), "Unknown"}, // Unknown category should return "Unknown"
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.category.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
