// Package logutil provides console output functionality.
package logutil

import "time"

// ProgressOutput provides model processing progress reporting.
type ProgressOutput interface {
	StartProcessing(modelCount int)
	ModelStarted(modelIndex, totalModels int, modelName string)
	ModelCompleted(modelIndex, totalModels int, modelName string, duration time.Duration)
	ModelFailed(modelIndex, totalModels int, modelName string, reason string)
	ModelRateLimited(modelIndex, totalModels int, modelName string, retryAfter time.Duration)
}

// SummaryOutput provides final summary display capabilities.
type SummaryOutput interface {
	ShowSummarySection(summary SummaryData)
	ShowOutputFiles(files []OutputFile)
	ShowFailedModels(failed []FailedModel)
}

// StatusOutput provides workflow status messaging.
type StatusOutput interface {
	StatusMessage(message string)
	ShowFileOperations(message string)
	SynthesisStarted()
	SynthesisCompleted(outputPath string)
	ErrorMessage(message string)
	WarningMessage(message string)
	SuccessMessage(message string)
}

// EnvironmentAware provides environment detection capabilities.
type EnvironmentAware interface {
	IsInteractive() bool
	GetTerminalWidth() int
	FormatMessage(message string) string
}

// QuietModeController provides quiet mode control.
type QuietModeController interface {
	SetQuiet(quiet bool)
	SetNoProgress(noProgress bool)
}

var (
	_ ProgressOutput      = (*consoleWriter)(nil)
	_ SummaryOutput       = (*consoleWriter)(nil)
	_ StatusOutput        = (*consoleWriter)(nil)
	_ EnvironmentAware    = (*consoleWriter)(nil)
	_ QuietModeController = (*consoleWriter)(nil)
)
