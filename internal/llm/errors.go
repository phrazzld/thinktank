// Package llm provides a common interface for interacting with various LLM providers
package llm

import "errors"

// ErrorCategory represents different categories of errors that can occur when using LLM APIs
type ErrorCategory int

const (
	// CategoryUnknown represents an unknown or uncategorized error
	CategoryUnknown ErrorCategory = iota
	// CategoryAuth represents authentication and authorization errors
	CategoryAuth
	// CategoryRateLimit represents rate limiting or quota errors
	CategoryRateLimit
	// CategoryInvalidRequest represents invalid request errors
	CategoryInvalidRequest
	// CategoryNotFound represents model not found errors
	CategoryNotFound
	// CategoryServer represents server errors
	CategoryServer
	// CategoryNetwork represents network connectivity errors
	CategoryNetwork
	// CategoryCancelled represents cancelled context errors
	CategoryCancelled
	// CategoryInputLimit represents input token limit exceeded errors
	CategoryInputLimit
	// CategoryContentFiltered represents content filtered by safety settings errors
	CategoryContentFiltered
)

// String returns a string representation of the ErrorCategory
func (c ErrorCategory) String() string {
	switch c {
	case CategoryAuth:
		return "Auth"
	case CategoryRateLimit:
		return "RateLimit"
	case CategoryInvalidRequest:
		return "InvalidRequest"
	case CategoryNotFound:
		return "NotFound"
	case CategoryServer:
		return "Server"
	case CategoryNetwork:
		return "Network"
	case CategoryCancelled:
		return "Cancelled"
	case CategoryInputLimit:
		return "InputLimit"
	case CategoryContentFiltered:
		return "ContentFiltered"
	default:
		return "Unknown"
	}
}

// CategorizedError is an interface that extends error with the ability
// to provide error category information for more specific error handling.
type CategorizedError interface {
	error // Embed standard error interface
	Category() ErrorCategory
}

// IsCategorizedError checks if an error implements the CategorizedError interface
// and returns the implementing error and true if it does, or nil and false if not.
func IsCategorizedError(err error) (CategorizedError, bool) {
	var catErr CategorizedError
	if err == nil {
		return nil, false
	}
	if errors.As(err, &catErr) {
		return catErr, true
	}
	return nil, false
}
