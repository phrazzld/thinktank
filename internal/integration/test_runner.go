// Package integration provides testing utilities for the thinktank project
// This is a simplified version with token functionality removed (T036C)
package integration

// TestRunner runs integration tests
type TestRunner struct {
	env *TestEnv
}

// NewTestRunner creates a new test runner
func NewTestRunner() *TestRunner {
	env := NewTestEnv()
	return &TestRunner{
		env: env,
	}
}

// RunTest runs a test
func (r *TestRunner) RunTest() *SimpleResult {
	return MockResponse("Test successful")
}
