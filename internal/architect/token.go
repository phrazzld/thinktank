// Package architect contains the core application logic for the architect tool
package architect

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
)

// TokenResult holds information about token counts and limits
type TokenResult struct {
	TokenCount   int32
	InputLimit   int32
	ExceedsLimit bool
	LimitError   string
	Percentage   float64
}

// TokenManager defines the interface for token counting and management
type TokenManager interface {
	// GetTokenInfo retrieves token count information and checks limits
	GetTokenInfo(ctx context.Context, prompt string) (*TokenResult, error)

	// CheckTokenLimit verifies the prompt doesn't exceed the model's token limit
	CheckTokenLimit(ctx context.Context, prompt string) error

	// PromptForConfirmation asks for user confirmation to proceed if token count exceeds threshold
	PromptForConfirmation(tokenCount int32, threshold int) bool
}

// tokenManager implements the TokenManager interface
type tokenManager struct {
	logger      logutil.LoggerInterface
	auditLogger auditlog.AuditLogger
	client      llm.LLMClient
}

// NewTokenManager creates a new TokenManager instance
func NewTokenManager(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient) (TokenManager, error) {
	if client == nil {
		return nil, fmt.Errorf("client cannot be nil for TokenManager")
	}
	return NewTokenManagerWithClient(logger, auditLogger, client), nil
}

// NewTokenManagerWithClient creates a TokenManager with a specific client.
// This is defined as a variable to allow it to be mocked in tests.
var NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient) TokenManager {
	return &tokenManager{
		logger:      logger,
		auditLogger: auditLogger,
		client:      client,
	}
}

// GetTokenInfo retrieves token count information and checks limits
func (tm *tokenManager) GetTokenInfo(ctx context.Context, prompt string) (*TokenResult, error) {
	// Get the model name from the injected client
	modelName := tm.client.GetModelName()

	// Log the start of token checking
	checkStartTime := time.Now()
	if logErr := tm.auditLogger.Log(auditlog.AuditEntry{
		Timestamp: checkStartTime,
		Operation: "CheckTokensStart",
		Status:    "InProgress",
		Inputs: map[string]interface{}{
			"prompt_length": len(prompt),
			"model_name":    modelName,
		},
		Message: "Starting token count check for model " + modelName,
	}); logErr != nil {
		tm.logger.Error("Failed to write audit log: %v", logErr)
	}

	// Create result structure
	result := &TokenResult{
		ExceedsLimit: false,
	}

	// Get model information (limits)
	modelInfo, err := tm.client.GetModelInfo(ctx)
	if err != nil {
		// Handle provider-agnostic error logging
		errorType := "TokenCheckError"
		// Using a separate variable declaration to avoid ineffectual assignment
		var errorMessage string

		// Get specific error details if we can recognize provider-specific errors
		// Note: This approach allows us to handle Gemini or other provider-specific errors
		// without direct dependency on provider-specific error types
		if apiService, ok := tm.client.(APIService); ok && apiService != nil {
			if apiService.IsSafetyBlockedError(err) {
				errorType = "SafetyBlockedError"
			}
			errorMessage = apiService.GetErrorDetails(err)
		} else {
			// Just use the error message directly
			errorMessage = err.Error()
		}

		// Log the token check failure
		if logErr := tm.auditLogger.Log(auditlog.AuditEntry{
			Timestamp: time.Now().UTC(),
			Operation: "CheckTokens",
			Status:    "Failure",
			Inputs: map[string]interface{}{
				"prompt_length": len(prompt),
				"model_name":    modelName,
			},
			Error: &auditlog.ErrorInfo{
				Message: errorMessage,
				Type:    errorType,
			},
			Message: "Token count check failed for model " + modelName,
		}); logErr != nil {
			tm.logger.Error("Failed to write audit log: %v", logErr)
		}

		// Return the original error
		return nil, fmt.Errorf("failed to get model info for token limit check: %w", err)
	}

	// Store input limit
	result.InputLimit = modelInfo.InputTokenLimit

	// Count tokens in the prompt
	tokenResult, err := tm.client.CountTokens(ctx, prompt)
	if err != nil {
		// Handle provider-agnostic error logging
		errorType := "TokenCheckError"
		// Using a separate variable declaration to avoid ineffectual assignment
		var errorMessage string

		// Get specific error details if we can recognize provider-specific errors
		if apiService, ok := tm.client.(APIService); ok && apiService != nil {
			if apiService.IsSafetyBlockedError(err) {
				errorType = "SafetyBlockedError"
			}
			errorMessage = apiService.GetErrorDetails(err)
		} else {
			// Just use the error message directly
			errorMessage = err.Error()
		}

		// Log the token check failure
		if logErr := tm.auditLogger.Log(auditlog.AuditEntry{
			Timestamp: time.Now().UTC(),
			Operation: "CheckTokens",
			Status:    "Failure",
			Inputs: map[string]interface{}{
				"prompt_length": len(prompt),
				"model_name":    modelName,
			},
			Error: &auditlog.ErrorInfo{
				Message: errorMessage,
				Type:    errorType,
			},
			Message: "Token count check failed for model " + modelName,
		}); logErr != nil {
			tm.logger.Error("Failed to write audit log: %v", logErr)
		}

		// Return the original error
		return nil, fmt.Errorf("failed to count tokens for token limit check: %w", err)
	}

	// Store token count
	result.TokenCount = tokenResult.Total

	// Calculate percentage of limit
	result.Percentage = float64(result.TokenCount) / float64(result.InputLimit) * 100

	// Log token usage information
	tm.logger.Debug("Token usage: %d / %d (%.1f%%)",
		result.TokenCount,
		result.InputLimit,
		result.Percentage)

	// Check if the prompt exceeds the token limit
	if result.TokenCount > result.InputLimit {
		result.ExceedsLimit = true
		result.LimitError = fmt.Sprintf("prompt exceeds token limit (%d tokens > %d token limit)",
			result.TokenCount, result.InputLimit)

		// Log the token limit exceeded case
		if logErr := tm.auditLogger.Log(auditlog.AuditEntry{
			Timestamp: time.Now().UTC(),
			Operation: "CheckTokens",
			Status:    "Failure",
			Inputs: map[string]interface{}{
				"prompt_length": len(prompt),
				"model_name":    modelName,
			},
			TokenCounts: &auditlog.TokenCountInfo{
				PromptTokens: result.TokenCount,
				TotalTokens:  result.TokenCount,
				Limit:        result.InputLimit,
			},
			Error: &auditlog.ErrorInfo{
				Message: result.LimitError,
				Type:    "TokenLimitExceededError",
			},
			Message: "Token limit exceeded for model " + modelName,
		}); logErr != nil {
			tm.logger.Error("Failed to write audit log: %v", logErr)
		}
	} else {
		// Log the successful token check
		if logErr := tm.auditLogger.Log(auditlog.AuditEntry{
			Timestamp: time.Now().UTC(),
			Operation: "CheckTokens",
			Status:    "Success",
			Inputs: map[string]interface{}{
				"prompt_length": len(prompt),
				"model_name":    modelName,
			},
			Outputs: map[string]interface{}{
				"percentage": result.Percentage,
			},
			TokenCounts: &auditlog.TokenCountInfo{
				PromptTokens: result.TokenCount,
				TotalTokens:  result.TokenCount,
				Limit:        result.InputLimit,
			},
			Message: fmt.Sprintf("Token check passed for model %s: %d / %d tokens (%.1f%% of limit)",
				modelName, result.TokenCount, result.InputLimit, result.Percentage),
		}); logErr != nil {
			tm.logger.Error("Failed to write audit log: %v", logErr)
		}
	}

	return result, nil
}

// CheckTokenLimit verifies the prompt doesn't exceed the model's token limit
func (tm *tokenManager) CheckTokenLimit(ctx context.Context, prompt string) error {
	tokenInfo, err := tm.GetTokenInfo(ctx, prompt)
	if err != nil {
		return err
	}

	if tokenInfo.ExceedsLimit {
		return fmt.Errorf(tokenInfo.LimitError)
	}

	return nil
}

// PromptForConfirmation asks for user confirmation to proceed
func (tm *tokenManager) PromptForConfirmation(tokenCount int32, threshold int) bool {
	if threshold <= 0 || int32(threshold) > tokenCount {
		// No confirmation needed if threshold is disabled (0) or token count is below threshold
		tm.logger.Debug("No confirmation needed: threshold=%d, tokenCount=%d", threshold, tokenCount)
		return true
	}

	tm.logger.Info("Token count (%d) exceeds confirmation threshold (%d).", tokenCount, threshold)
	tm.logger.Info("Do you want to proceed with the API call? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		tm.logger.Error("Error reading input: %v", err)
		return false
	}

	// Log the raw response for debugging
	tm.logger.Debug("User confirmation response (raw): %q", response)

	// Trim whitespace and convert to lowercase
	response = strings.ToLower(strings.TrimSpace(response))
	tm.logger.Debug("User confirmation response (processed): %q", response)

	// Only proceed if the user explicitly confirms with 'y' or 'yes'
	result := response == "y" || response == "yes"
	tm.logger.Debug("User confirmation result: %v", result)
	return result
}
