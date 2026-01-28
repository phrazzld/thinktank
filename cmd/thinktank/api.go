// Package main provides the command-line interface for the thinktank tool
package main

import (
	"github.com/misty-step/thinktank/internal/llm"
	"github.com/misty-step/thinktank/internal/logutil"
	"github.com/misty-step/thinktank/internal/thinktank"
	"github.com/misty-step/thinktank/internal/thinktank/interfaces"
)

// Re-export error types from internal/llm for backward compatibility with tests
var (
	ErrEmptyResponse        = llm.ErrEmptyResponse
	ErrWhitespaceContent    = llm.ErrWhitespaceContent
	ErrSafetyBlocked        = llm.ErrSafetyBlocked
	ErrAPICall              = llm.ErrAPICall
	ErrClientInitialization = llm.ErrClientInitialization
)

// APIService is an alias to the interfaces one
type APIService = interfaces.APIService

// NewAPIService is a wrapper for the internal one
// It uses the models-based implementation for simplicity
func NewAPIService(logger logutil.LoggerInterface) APIService {
	return thinktank.NewRegistryAPIService(logger)
}
