// Package architect provides the command-line interface for the architect tool
package architect

import (
	"context"
	"fmt"

	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// APIService defines the interface for API-related operations
type APIService interface {
	// InitClient initializes and returns a Gemini client
	InitClient(ctx context.Context, apiKey, modelName string) (gemini.Client, error)

	// ProcessResponse processes the API response and extracts content
	ProcessResponse(result *gemini.GenerationResult) (string, error)
}

// apiService implements the APIService interface
type apiService struct {
	logger logutil.LoggerInterface
}

// NewAPIService creates a new APIService instance
func NewAPIService(logger logutil.LoggerInterface) APIService {
	return &apiService{
		logger: logger,
	}
}

// InitClient initializes and returns a Gemini client
func (s *apiService) InitClient(ctx context.Context, apiKey, modelName string) (gemini.Client, error) {
	client, err := gemini.NewClient(ctx, apiKey, modelName)
	if err != nil {
		return nil, fmt.Errorf("error creating Gemini client: %w", err)
	}
	return client, nil
}

// ProcessResponse processes the API response and extracts content
func (s *apiService) ProcessResponse(result *gemini.GenerationResult) (string, error) {
	// Stub implementation - will be replaced with actual code from main.go
	return "", fmt.Errorf("not implemented yet")
}
