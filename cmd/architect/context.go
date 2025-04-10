// Package architect provides the command-line interface for the architect tool
package architect

import (
	"context"
	"fmt"

	"github.com/phrazzld/architect/internal/fileutil"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// ContextStats holds information about processed files and context size
type ContextStats struct {
	ProcessedFilesCount int
	CharCount           int
	LineCount           int
	TokenCount          int32
	ProcessedFiles      []string
}

// ContextGatherer defines the interface for gathering project context
type ContextGatherer interface {
	// GatherContext collects and processes files based on configuration
	GatherContext(paths []string, include, exclude, excludeNames, format string) (string, *ContextStats, error)

	// DisplayDryRunInfo shows detailed information for dry run mode
	DisplayDryRunInfo(stats *ContextStats, ctx context.Context, client gemini.Client) error
}

// contextGatherer implements the ContextGatherer interface
type contextGatherer struct {
	logger       logutil.LoggerInterface
	dryRun       bool
	tokenManager TokenManager
}

// NewContextGatherer creates a new ContextGatherer instance
func NewContextGatherer(logger logutil.LoggerInterface, dryRun bool, tokenManager TokenManager) ContextGatherer {
	return &contextGatherer{
		logger:       logger,
		dryRun:       dryRun,
		tokenManager: tokenManager,
	}
}

// GatherContext collects and processes files based on configuration
func (cg *contextGatherer) GatherContext(paths []string, include, exclude, excludeNames, format string) (string, *ContextStats, error) {
	// Stub implementation - will be replaced with actual code from main.go
	return "", nil, fmt.Errorf("not implemented yet")
}

// DisplayDryRunInfo shows detailed information for dry run mode
func (cg *contextGatherer) DisplayDryRunInfo(stats *ContextStats, ctx context.Context, client gemini.Client) error {
	// Stub implementation - will be replaced with actual code from main.go
	return fmt.Errorf("not implemented yet")
}
