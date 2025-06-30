package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/models"
	"github.com/phrazzld/thinktank/internal/thinktank"
)

// LogModelSelectionAudit logs detailed model selection information to audit logs
// according to Phase 7.2 requirements for structured logging enhancement.
func LogModelSelectionAudit(
	ctx context.Context,
	auditLogger auditlog.AuditLogger,
	tokenReq thinktank.TokenCountingRequest,
	compatibleModels []thinktank.ModelCompatibility,
	err error,
) error {
	// Calculate token counts from the request
	totalTokens := calculateTotalTokens(tokenReq)
	instructionTokens := len(tokenReq.Instructions) / 4 // Rough estimation for audit
	fileTokens := totalTokens - instructionTokens

	// Categorize models
	var selectedModels []string
	var skippedModels []string
	compatibleCount := 0
	accurateCount := 0
	primaryTokenizer := "estimation" // Default fallback

	for _, model := range compatibleModels {
		if model.IsCompatible {
			selectedModels = append(selectedModels, model.ModelName)
			compatibleCount++
			if model.IsAccurate {
				accurateCount++
				// Use the tokenizer from the first accurate model as primary
				if primaryTokenizer == "estimation" && model.IsAccurate {
					primaryTokenizer = model.TokenizerUsed
				}
			}
		} else {
			skippedModels = append(skippedModels, model.ModelName)
		}
	}

	// Determine status
	status := "Success"
	if err != nil || len(selectedModels) == 0 {
		status = "Failure"
	}

	// Prepare input data for audit log
	inputs := map[string]interface{}{
		"instruction_tokens": instructionTokens,
		"file_tokens":        fileTokens,
		"total_tokens":       totalTokens,
		"provider_count":     len(getUniqueProviders(compatibleModels)),
		"safety_margin":      tokenReq.SafetyMarginPercent,
		"file_count":         len(tokenReq.Files),
	}

	// Add correlation ID if available
	if correlationID := logutil.GetCorrelationID(ctx); correlationID != "" {
		inputs["correlation_id"] = correlationID
	}

	// Prepare output data for audit log
	outputs := map[string]interface{}{
		"selected_models":  selectedModels,
		"skipped_models":   skippedModels,
		"tokenizer_method": primaryTokenizer,
		"compatible_count": compatibleCount,
		"accurate_count":   accurateCount,
		"total_models":     len(compatibleModels),
	}

	// Create token count info for audit entry
	tokenCounts := &auditlog.TokenCountInfo{
		PromptTokens: int32(totalTokens),
		TotalTokens:  int32(totalTokens),
	}

	// Create audit entry
	entry := auditlog.AuditEntry{
		Timestamp:   time.Now().UTC(),
		Operation:   "model_selection",
		Status:      status,
		Inputs:      inputs,
		Outputs:     outputs,
		TokenCounts: tokenCounts,
	}

	if err != nil {
		entry.Error = &auditlog.ErrorInfo{
			Message: err.Error(),
			Type:    "ModelSelectionError",
		}
	}

	// Log the audit entry
	return auditLogger.Log(ctx, entry)
}

// calculateTotalTokens provides a rough estimation of total tokens from a request
// This is used for audit logging purposes only
func calculateTotalTokens(req thinktank.TokenCountingRequest) int {
	// Simple estimation: 1 token per 4 characters
	totalChars := len(req.Instructions)
	for _, file := range req.Files {
		totalChars += len(file.Content)
	}
	return totalChars / 4
}

// getUniqueProviders returns the unique providers from a list of model compatibilities
func getUniqueProviders(models []thinktank.ModelCompatibility) []string {
	providerMap := make(map[string]bool)
	var providers []string

	for _, model := range models {
		if !providerMap[model.Provider] {
			providerMap[model.Provider] = true
			providers = append(providers, model.Provider)
		}
	}

	return providers
}

// auditModelSelection performs model selection and logs detailed audit information.
// This function is called from Main() after context and audit logging are set up.
func auditModelSelection(
	ctx context.Context,
	minimalConfig *config.MinimalConfig,
	logger logutil.LoggerInterface,
	simplifiedConfig *SimplifiedConfig,
	tokenService thinktank.TokenCountingService,
) error {
	// Create audit logger for this operation
	var auditLogger auditlog.AuditLogger
	if minimalConfig.DryRun {
		auditLogger = auditlog.NewNoOpAuditLogger()
	} else {
		auditLogPath := filepath.Join(minimalConfig.OutputDir, "audit.jsonl")
		var err error
		auditLogger, err = auditlog.NewFileAuditLogger(auditLogPath, logger)
		if err != nil {
			return fmt.Errorf("failed to create audit logger: %w", err)
		}
		defer func() {
			if closeErr := auditLogger.Close(); closeErr != nil {
				logger.ErrorContext(ctx, "Failed to close audit logger: %v", closeErr)
			}
		}()
	}

	// Re-run model selection to get detailed compatibility information for audit logging
	compatibleModels, tokenReq, err := performModelSelectionWithDetails(simplifiedConfig, tokenService)
	if err != nil {
		// Log the failure to both audit and structured logs
		auditErr := LogModelSelectionAudit(ctx, auditLogger, tokenReq, nil, err)
		structuredErr := LogModelSelectionStructured(ctx, logger, tokenReq, nil, err)
		if auditErr != nil {
			return auditErr
		}
		return structuredErr
	}

	// Log successful model selection with detailed information
	err = LogModelSelectionAudit(ctx, auditLogger, tokenReq, compatibleModels, nil)
	if err != nil {
		return err
	}

	// Also log structured summary to main logger (thinktank.log)
	return LogModelSelectionStructured(ctx, logger, tokenReq, compatibleModels, nil)
}

// performModelSelectionWithDetails performs model selection and returns detailed compatibility information
// for audit logging purposes. This duplicates some logic from selectModelsForConfigWithService but
// returns the detailed ModelCompatibility data needed for comprehensive audit logging.
func performModelSelectionWithDetails(
	simplifiedConfig *SimplifiedConfig,
	tokenService thinktank.TokenCountingService,
) ([]thinktank.ModelCompatibility, thinktank.TokenCountingRequest, error) {
	// Get available providers (those with API keys set)
	availableProviders := models.GetAvailableProviders()
	if len(availableProviders) == 0 {
		return nil, thinktank.TokenCountingRequest{}, fmt.Errorf("no API keys available")
	}

	// Read instructions content for TokenCountingService
	var instructionsContent string
	if content, err := os.ReadFile(simplifiedConfig.InstructionsFile); err == nil {
		instructionsContent = string(content)
	} else {
		// Fallback to empty instructions if file read fails
		instructionsContent = ""
	}

	// Create token counting request
	tokenReq := thinktank.TokenCountingRequest{
		Instructions:        instructionsContent,
		Files:               []thinktank.FileContent{}, // Empty for now - will be enhanced later
		SafetyMarginPercent: simplifiedConfig.SafetyMargin,
	}

	// Use TokenCountingService to get compatible models with accurate tokenization
	ctx := context.Background()
	compatibleModels, err := tokenService.GetCompatibleModels(ctx, tokenReq, availableProviders)
	if err != nil {
		return nil, tokenReq, fmt.Errorf("failed to get compatible models: %w", err)
	}

	return compatibleModels, tokenReq, nil
}

// LogModelSelectionStructured logs model selection summary to the main structured logger (thinktank.log)
// according to Phase 7.2 requirements for enhanced structured logging.
func LogModelSelectionStructured(
	ctx context.Context,
	logger logutil.LoggerInterface,
	tokenReq thinktank.TokenCountingRequest,
	compatibleModels []thinktank.ModelCompatibility,
	err error,
) error {
	// Calculate token counts and metrics
	totalTokens := calculateTotalTokens(tokenReq)
	instructionTokens := len(tokenReq.Instructions) / 4 // Rough estimation

	// Categorize models
	var selectedModels []string
	var skippedModels []string
	compatibleCount := 0
	accurateCount := 0
	primaryTokenizer := "estimation" // Default fallback

	for _, model := range compatibleModels {
		if model.IsCompatible {
			selectedModels = append(selectedModels, model.ModelName)
			compatibleCount++
			if model.IsAccurate {
				accurateCount++
				// Use the tokenizer from the first accurate model as primary
				if primaryTokenizer == "estimation" && model.IsAccurate {
					primaryTokenizer = model.TokenizerUsed
				}
			}
		} else {
			skippedModels = append(skippedModels, model.ModelName)
		}
	}

	// Log structured summary message with all required fields from Phase 7.2
	if err != nil || len(selectedModels) == 0 {
		logger.ErrorContext(ctx, "Model selection failed",
			slog.Int("input_tokens", instructionTokens),
			slog.String("tokenizer_method", primaryTokenizer),
			slog.Any("selected_models", selectedModels),
			slog.Any("skipped_models", skippedModels),
			slog.Int("compatible_count", compatibleCount),
			slog.Int("accurate_count", accurateCount),
			slog.Int("total_models", len(compatibleModels)),
		)
	} else {
		// Success case - log the summary as specified in Phase 7.2
		logger.InfoContext(ctx, fmt.Sprintf("Token counting summary: %d tokens, %s accuracy, %d compatible, %d skipped",
			totalTokens, primaryTokenizer, compatibleCount, len(skippedModels)),
			slog.Int("input_tokens", instructionTokens),
			slog.String("tokenizer_method", primaryTokenizer),
			slog.Any("selected_models", selectedModels),
			slog.Any("skipped_models", skippedModels),
			slog.Int("compatible_count", compatibleCount),
			slog.Int("accurate_count", accurateCount),
			slog.Int("total_models", len(compatibleModels)),
		)
	}

	return nil
}
