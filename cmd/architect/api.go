// Package architect provides the command-line interface for the architect tool
package architect

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// Define package-level error types for better error handling
var (
	// ErrEmptyResponse indicates the API returned an empty response
	ErrEmptyResponse = errors.New("received empty response from Gemini")

	// ErrWhitespaceContent indicates the API returned only whitespace content
	ErrWhitespaceContent = errors.New("Gemini returned an empty plan text")

	// ErrSafetyBlocked indicates content was blocked by safety filters
	ErrSafetyBlocked = errors.New("content blocked by Gemini safety filters")

	// ErrAPICall indicates a general API call error
	ErrAPICall = errors.New("error calling Gemini API")

	// ErrClientInitialization indicates client initialization failed
	ErrClientInitialization = errors.New("error creating Gemini client")
)

// APIService defines the interface for API-related operations
type APIService interface {
	// InitClient initializes and returns a Gemini client
	InitClient(ctx context.Context, apiKey, modelName string) (gemini.Client, error)

	// ProcessResponse processes the API response and extracts content
	ProcessResponse(result *gemini.GenerationResult) (string, error)

	// IsEmptyResponseError checks if an error is related to empty API responses
	IsEmptyResponseError(err error) bool

	// IsSafetyBlockedError checks if an error is related to safety filters
	IsSafetyBlockedError(err error) bool

	// GetErrorDetails extracts detailed information from an error
	GetErrorDetails(err error) string
}

// apiService implements the APIService interface
type apiService struct {
	logger logutil.LoggerInterface
	// For testing
	newClientFunc func(ctx context.Context, apiKey, modelName string) (gemini.Client, error)
}

// NewAPIService is a variable that holds the factory function for creating a new APIService
// It's defined as a variable to allow mocking in tests
var NewAPIService = func(logger logutil.LoggerInterface) APIService {
	return &apiService{
		logger:        logger,
		newClientFunc: gemini.NewClient, // Default to the real implementation
	}
}

// InitClient initializes and returns a Gemini client
func (s *apiService) InitClient(ctx context.Context, apiKey, modelName string) (gemini.Client, error) {
	// Check for empty required parameters
	if apiKey == "" {
		return nil, fmt.Errorf("%w: API key is required", ErrClientInitialization)
	}
	if modelName == "" {
		return nil, fmt.Errorf("%w: model name is required", ErrClientInitialization)
	}

	// Check for context cancellation
	if ctx.Err() != nil {
		return nil, fmt.Errorf("%w: %v", ErrClientInitialization, ctx.Err())
	}

	// Initialize the client (using a function that can be swapped out in tests)
	client, err := s.newClientFunc(ctx, apiKey, modelName)
	if err != nil {
		// Check if it's already an API error with enhanced details
		if apiErr, ok := gemini.IsAPIError(err); ok {
			return nil, fmt.Errorf("%w: %s", ErrClientInitialization, apiErr.UserFacingError())
		}

		// Wrap the original error
		return nil, fmt.Errorf("%w: %v", ErrClientInitialization, err)
	}

	return client, nil
}

// ProcessResponse processes the API response and extracts content
func (s *apiService) ProcessResponse(result *gemini.GenerationResult) (string, error) {
	// Check for nil result
	if result == nil {
		return "", fmt.Errorf("%w: result is nil", ErrEmptyResponse)
	}

	// Check for empty content
	if result.Content == "" {
		var errDetails strings.Builder

		// Add finish reason if available
		if result.FinishReason != "" {
			fmt.Fprintf(&errDetails, " (Finish Reason: %s)", result.FinishReason)
		}

		// Check for safety blocks
		if len(result.SafetyRatings) > 0 {
			blocked := false
			safetyInfo := ""
			for _, rating := range result.SafetyRatings {
				if rating.Blocked {
					blocked = true
					safetyInfo += fmt.Sprintf(" Blocked by Safety Category: %s;", rating.Category)
				}
			}

			if blocked {
				if errDetails.Len() > 0 {
					errDetails.WriteString(" ")
				}
				errDetails.WriteString("Safety Blocking:")
				errDetails.WriteString(safetyInfo)

				// If we have safety blocks, use the specific safety error
				return "", fmt.Errorf("%w%s", ErrSafetyBlocked, errDetails.String())
			}
		}

		// If we don't have safety blocks, use the generic empty response error
		return "", fmt.Errorf("%w%s", ErrEmptyResponse, errDetails.String())
	}

	// Check for whitespace-only content
	if strings.TrimSpace(result.Content) == "" {
		return "", ErrWhitespaceContent
	}

	return result.Content, nil
}

// IsEmptyResponseError checks if an error is related to empty API responses
func (s *apiService) IsEmptyResponseError(err error) bool {
	return errors.Is(err, ErrEmptyResponse) || errors.Is(err, ErrWhitespaceContent)
}

// IsSafetyBlockedError checks if an error is related to safety filters
func (s *apiService) IsSafetyBlockedError(err error) bool {
	return errors.Is(err, ErrSafetyBlocked)
}

// GetErrorDetails extracts detailed information from an error
func (s *apiService) GetErrorDetails(err error) string {
	// Handle nil error case
	if err == nil {
		return ""
	}

	// Check if it's an API error with enhanced details
	if apiErr, ok := gemini.IsAPIError(err); ok {
		return apiErr.UserFacingError()
	}

	// Return the error string for other error types
	return err.Error()
}
