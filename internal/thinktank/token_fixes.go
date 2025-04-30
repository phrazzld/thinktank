// Package thinktank contains the core application logic for the thinktank tool.
package thinktank

// TokenResult represents a simplified token counting result structure
// This is kept for backward compatibility with tests
type TokenResult struct {
	TokenCount   int32
	InputLimit   int32
	ExceedsLimit bool
	LimitError   string
	Percentage   float64
}
