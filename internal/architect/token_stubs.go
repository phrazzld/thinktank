// Package architect contains the core application logic for the architect tool
package architect

// This file formerly contained token-related stubs that are no longer needed
// after the token handling removal refactoring (T036D).
//
// Kept as a minimal placeholder to maintain imports in existing tests.
// This file should be removed in a future cleanup once all dependent tests
// are updated or disabled.
//
// NOTE: T036C has already disabled most token-related tests, but some tests
// still have references to token types. A future task may remove these
// references completely.

// TokenResult is a stub retained for test compatibility only
type TokenResult struct {
	TokenCount   int32
	InputLimit   int32
	ExceedsLimit bool
	LimitError   string
	Percentage   float64
}
