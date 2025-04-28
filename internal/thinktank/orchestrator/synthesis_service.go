// Package orchestrator is responsible for coordinating the core application workflow.
// It brings together various components like context gathering, API interaction,
// token management, and output writing to execute the main task defined
// by user instructions and configuration.
package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
	"github.com/phrazzld/thinktank/internal/thinktank/prompt"
)

// SynthesisService handles the combining of multiple model outputs into a single result
type SynthesisService interface {
	// SynthesizeResults processes multiple model outputs through a synthesis model
	// It combines model outputs with the original instructions into a unified response
	SynthesizeResults(ctx context.Context, instructions string, modelOutputs map[string]string) (string, error)
}

// DefaultSynthesisService implements the SynthesisService interface
type DefaultSynthesisService struct {
	apiService  interfaces.APIService
	auditLogger auditlog.AuditLogger
	logger      logutil.LoggerInterface
	modelName   string // The name of the synthesis model to use
}

// NewSynthesisService creates a new SynthesisService instance with the specified dependencies
func NewSynthesisService(
	apiService interfaces.APIService,
	auditLogger auditlog.AuditLogger,
	logger logutil.LoggerInterface,
	modelName string,
) SynthesisService {
	return &DefaultSynthesisService{
		apiService:  apiService,
		auditLogger: auditLogger,
		logger:      logger,
		modelName:   modelName,
	}
}

// SynthesizeResults processes multiple model outputs through a synthesis model.
// It builds a prompt that includes the original instructions and all model outputs,
// then sends this to the synthesis model to generate a consolidated result.
//
// This is a key component of the synthesis feature, which allows combining outputs
// from multiple models into a single coherent response. The method:
// 1. Creates a specially formatted synthesis prompt using StitchSynthesisPrompt
// 2. Initializes a client for the synthesis model
// 3. Calls the synthesis model API with the combined prompt
// 4. Processes and returns the synthesized result
//
// The method includes comprehensive audit logging at each step of the synthesis process,
// allowing for detailed tracking and debugging.
//
// Parameters:
//   - ctx: The context for API client communication
//   - originalInstructions: The user's original instructions
//   - modelOutputs: A map of model names to their generated content
//
// Returns:
//   - A string containing the synthesized result
//   - An error if any step of the synthesis process fails
func (s *DefaultSynthesisService) SynthesizeResults(
	ctx context.Context,
	originalInstructions string,
	modelOutputs map[string]string,
) (string, error) {
	startTime := time.Now()

	// Get logger with context
	contextLogger := s.logger.WithContext(ctx)

	// Log synthesis process start with audit logger
	s.logAuditEvent(auditlog.AuditEntry{
		Operation: "SynthesisStart",
		Status:    "InProgress",
		Inputs: map[string]interface{}{
			"synthesis_model": s.modelName,
			"model_count":     len(modelOutputs),
			"model_names":     getMapKeys(modelOutputs),
		},
		Message: fmt.Sprintf("Starting synthesis with model %s, processing %d outputs",
			s.modelName, len(modelOutputs)),
	})

	// Build synthesis prompt using the dedicated prompt function
	contextLogger.DebugContext(ctx, "Building synthesis prompt")
	synthesisPrompt := prompt.StitchSynthesisPrompt(originalInstructions, modelOutputs)
	contextLogger.DebugContext(ctx, "Synthesis prompt built, length: %d characters", len(synthesisPrompt))

	// Log prompt building completed
	s.logAuditEvent(auditlog.AuditEntry{
		Operation: "SynthesisPromptCreated",
		Status:    "Success",
		Inputs: map[string]interface{}{
			"synthesis_model": s.modelName,
			"prompt_length":   len(synthesisPrompt),
		},
		Message: "Synthesis prompt built successfully",
	})

	// Get model parameters
	contextLogger.DebugContext(ctx, "Getting model parameters for synthesis model: %s", s.modelName)
	modelParams, err := s.apiService.GetModelParameters(s.modelName)
	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		durationMs := duration

		// Log error with audit logger
		s.logAuditEvent(auditlog.AuditEntry{
			Operation:  "SynthesisModelParameters",
			Status:     "Failure",
			DurationMs: &durationMs,
			Inputs: map[string]interface{}{
				"synthesis_model": s.modelName,
			},
			Error: &auditlog.ErrorInfo{
				Message: err.Error(),
				Type:    "ModelParameterError",
			},
			Message: fmt.Sprintf("Failed to get parameters for synthesis model %s", s.modelName),
		})

		return "", fmt.Errorf("failed to get model parameters for synthesis model: %w", fmt.Errorf("%w: %v", ErrInvalidSynthesisModel, err))
	}

	// Log successful parameter retrieval
	s.logAuditEvent(auditlog.AuditEntry{
		Operation: "SynthesisModelParameters",
		Status:    "Success",
		Inputs: map[string]interface{}{
			"synthesis_model": s.modelName,
			"param_count":     len(modelParams),
		},
		Message: "Successfully retrieved synthesis model parameters",
	})

	// Get client for synthesis model
	contextLogger.DebugContext(ctx, "Initializing client for synthesis model: %s", s.modelName)
	client, err := s.apiService.InitLLMClient(ctx, "", s.modelName, "")
	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		durationMs := duration

		// Log error with audit logger
		s.logAuditEvent(auditlog.AuditEntry{
			Operation:  "SynthesisClientInit",
			Status:     "Failure",
			DurationMs: &durationMs,
			Inputs: map[string]interface{}{
				"synthesis_model": s.modelName,
			},
			Error: &auditlog.ErrorInfo{
				Message: err.Error(),
				Type:    "ClientInitializationError",
			},
			Message: fmt.Sprintf("Failed to initialize client for synthesis model %s", s.modelName),
		})

		return "", fmt.Errorf("failed to initialize synthesis model client: %w", fmt.Errorf("%w: %v", ErrInvalidSynthesisModel, err))
	}

	// Log successful client initialization
	s.logAuditEvent(auditlog.AuditEntry{
		Operation: "SynthesisClientInit",
		Status:    "Success",
		Inputs: map[string]interface{}{
			"synthesis_model": s.modelName,
		},
		Message: "Successfully initialized synthesis model client",
	})

	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			contextLogger.WarnContext(ctx, "Error closing synthesis model client: %v", closeErr)

			// Log client close error
			s.logAuditEvent(auditlog.AuditEntry{
				Operation: "SynthesisClientClose",
				Status:    "Failure",
				Inputs: map[string]interface{}{
					"synthesis_model": s.modelName,
				},
				Error: &auditlog.ErrorInfo{
					Message: closeErr.Error(),
					Type:    "ClientCloseError",
				},
				Message: "Error closing synthesis model client",
			})
		} else {
			// Log successful client close
			s.logAuditEvent(auditlog.AuditEntry{
				Operation: "SynthesisClientClose",
				Status:    "Success",
				Inputs: map[string]interface{}{
					"synthesis_model": s.modelName,
				},
				Message: "Successfully closed synthesis model client",
			})
		}
	}()

	// Log API call start
	apiCallStartTime := time.Now()
	s.logAuditEvent(auditlog.AuditEntry{
		Operation: "SynthesisAPICall",
		Status:    "InProgress",
		Inputs: map[string]interface{}{
			"synthesis_model": s.modelName,
			"prompt_length":   len(synthesisPrompt),
		},
		Message: fmt.Sprintf("Calling synthesis model API: %s", s.modelName),
	})

	// Call model API
	contextLogger.InfoContext(ctx, "Calling synthesis model API: %s", s.modelName)
	result, err := client.GenerateContent(ctx, synthesisPrompt, modelParams)

	// Calculate API call duration
	apiCallDurationMs := time.Since(apiCallStartTime).Milliseconds()

	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		durationMs := duration

		// Log API call failure
		s.logAuditEvent(auditlog.AuditEntry{
			Operation:  "SynthesisAPICall",
			Status:     "Failure",
			DurationMs: &apiCallDurationMs,
			Inputs: map[string]interface{}{
				"synthesis_model": s.modelName,
				"prompt_length":   len(synthesisPrompt),
			},
			Error: &auditlog.ErrorInfo{
				Message: err.Error(),
				Type:    "APICallError",
			},
			Message: fmt.Sprintf("Synthesis model API call failed for model %s", s.modelName),
		})

		// Log overall synthesis failure
		s.logAuditEvent(auditlog.AuditEntry{
			Operation:  "SynthesisEnd",
			Status:     "Failure",
			DurationMs: &durationMs,
			Inputs: map[string]interface{}{
				"synthesis_model": s.modelName,
			},
			Error: &auditlog.ErrorInfo{
				Message: err.Error(),
				Type:    "SynthesisError",
			},
			Message: fmt.Sprintf("Synthesis process failed with model %s", s.modelName),
		})

		return "", s.handleSynthesisError(ctx, err)
	}

	// Log successful API call
	s.logAuditEvent(auditlog.AuditEntry{
		Operation:  "SynthesisAPICall",
		Status:     "Success",
		DurationMs: &apiCallDurationMs,
		Inputs: map[string]interface{}{
			"synthesis_model": s.modelName,
		},
		Outputs: map[string]interface{}{
			"result_received": result != nil,
		},
		Message: "Synthesis model API call completed successfully",
	})

	// Process response
	responseStartTime := time.Now()
	contextLogger.DebugContext(ctx, "Processing synthesis model response")

	synthesisOutput, err := s.apiService.ProcessLLMResponse(result)

	// Calculate response processing duration
	responseDurationMs := time.Since(responseStartTime).Milliseconds()

	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		durationMs := duration

		// Log response processing failure
		s.logAuditEvent(auditlog.AuditEntry{
			Operation:  "SynthesisResponseProcessing",
			Status:     "Failure",
			DurationMs: &responseDurationMs,
			Inputs: map[string]interface{}{
				"synthesis_model": s.modelName,
			},
			Error: &auditlog.ErrorInfo{
				Message: err.Error(),
				Type:    "ResponseProcessingError",
			},
			Message: "Failed to process synthesis model response",
		})

		// Log overall synthesis failure
		s.logAuditEvent(auditlog.AuditEntry{
			Operation:  "SynthesisEnd",
			Status:     "Failure",
			DurationMs: &durationMs,
			Inputs: map[string]interface{}{
				"synthesis_model": s.modelName,
			},
			Error: &auditlog.ErrorInfo{
				Message: err.Error(),
				Type:    "SynthesisError",
			},
			Message: fmt.Sprintf("Synthesis process failed with model %s", s.modelName),
		})

		return "", fmt.Errorf("failed to process synthesis model response: %w", fmt.Errorf("%w: %v", ErrSynthesisFailed, err))
	}

	// Log successful response processing
	s.logAuditEvent(auditlog.AuditEntry{
		Operation:  "SynthesisResponseProcessing",
		Status:     "Success",
		DurationMs: &responseDurationMs,
		Inputs: map[string]interface{}{
			"synthesis_model": s.modelName,
		},
		Outputs: map[string]interface{}{
			"output_length": len(synthesisOutput),
		},
		Message: "Successfully processed synthesis model response",
	})

	// Calculate total duration and log overall success
	duration := time.Since(startTime).Milliseconds()
	durationMs := duration

	s.logAuditEvent(auditlog.AuditEntry{
		Operation:  "SynthesisEnd",
		Status:     "Success",
		DurationMs: &durationMs,
		Inputs: map[string]interface{}{
			"synthesis_model": s.modelName,
			"model_count":     len(modelOutputs),
		},
		Outputs: map[string]interface{}{
			"output_length": len(synthesisOutput),
		},
		Message: fmt.Sprintf("Synthesis completed successfully with model %s", s.modelName),
	})

	return synthesisOutput, nil
}

// handleSynthesisError processes a synthesis error and generates a user-friendly error message.
// It categorizes the error based on its type and adds helpful guidance for the user.
//
// This method is a key part of the synthesis feature's error handling, providing
// detailed, actionable feedback when synthesis fails. It analyzes the error type
// (e.g., rate limiting, content filtering, connectivity, authentication) and
// provides specific guidance for each case. For example, rate limit errors include
// tips on waiting and retrying, while content filtering errors suggest reviewing prompts.
//
// This ensures that synthesis errors are presented in a clear, actionable format,
// improving the user experience when problems occur.
func (s *DefaultSynthesisService) handleSynthesisError(ctx context.Context, err error) error {
	// Get correlation ID for logging
	correlationID := logutil.GetCorrelationID(ctx)

	// Get logger with context
	contextLogger := s.logger.WithContext(ctx)

	// Log the detailed error for debugging purposes
	if correlationID != "" {
		contextLogger.ErrorContext(ctx, "Synthesis error occurred: %v", err)
	} else {
		// Fallback to regular logging if context doesn't have correlation ID
		s.logger.Error("Synthesis error occurred: %v", err)
	}

	var errMsg string

	// Check for specific error types
	switch {
	case strings.Contains(err.Error(), "rate limit"):
		// Rate limiting error
		errMsg = fmt.Sprintf("Synthesis model '%s' encountered rate limiting: %v", s.modelName, err)
		errMsg += "\n\nTip: If you're encountering rate limit errors, consider waiting a moment before retrying or using a different synthesis model."

	case strings.Contains(err.Error(), "safety") || strings.Contains(err.Error(), "content filter") || s.apiService.IsSafetyBlockedError(err):
		// Content filtering/safety error
		errMsg = fmt.Sprintf("Synthesis model '%s' was blocked by content safety filters: %v", s.modelName, err)
		errMsg += "\n\nTip: Your input or the model outputs may contain content that triggered safety filters. Consider reviewing your prompts and instructions."

	case strings.Contains(err.Error(), "connect") || strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline"):
		// Connectivity error
		errMsg = fmt.Sprintf("Connectivity issue when calling synthesis model '%s': %v", s.modelName, err)
		errMsg += "\n\nTip: Check your internet connection and try again. If the issue persists, the service might be experiencing temporary problems."

	case strings.Contains(err.Error(), "auth") || strings.Contains(err.Error(), "key") || strings.Contains(err.Error(), "credential"):
		// Authentication error
		errMsg = fmt.Sprintf("Authentication failed for synthesis model '%s': %v", s.modelName, err)
		errMsg += "\n\nTip: Check your API key and ensure it has proper permissions for this model."

	case s.apiService.IsEmptyResponseError(err):
		// Empty response error
		errMsg = fmt.Sprintf("Synthesis model '%s' returned an empty response: %v", s.modelName, err)
		errMsg += "\n\nTip: The model might be having trouble with your instructions or the outputs from other models. Try simplifying the instructions."

	default:
		// Generic error with model details
		errMsg = fmt.Sprintf("Error synthesizing results with model '%s': %v", s.modelName, err)

		// Add any additional details from API service
		if details := s.apiService.GetErrorDetails(err); details != "" {
			errMsg += "\nDetails: " + details
		}

		errMsg += "\n\nTip: If this error persists, try using a different synthesis model or check the model's documentation for limitations."
	}

	// Return a new error with the formatted message
	// Use errors package for consistency with other error handling in this file
	return fmt.Errorf("%w: %s", ErrSynthesisFailed, errMsg)
}

// logAuditEvent writes an audit log entry and logs any errors that occur.
// This helper ensures proper error handling for all audit log operations.
// It converts an AuditEntry to LogOp parameters for consistent logging.
func (s *DefaultSynthesisService) logAuditEvent(entry auditlog.AuditEntry) {
	// Create an error from entry.Error if present
	var err error
	if entry.Error != nil {
		err = fmt.Errorf("%s: %s", entry.Error.Type, entry.Error.Message)
	}

	if logErr := s.auditLogger.LogOp(entry.Operation, entry.Status, entry.Inputs, entry.Outputs, err); logErr != nil {
		s.logger.Warn("Failed to write audit log: %v", logErr)
	}
}

// getMapKeys extracts and returns all keys from a map as a slice.
// This helper function is used for audit logging to capture all model names.
func getMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
