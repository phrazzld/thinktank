package logutil

import "testing"

// TestRoleInterfaceCompileTimeChecks verifies that consoleWriter implements all role interfaces.
// This is a compile-time check - if consoleWriter doesn't implement an interface, this won't compile.
func TestRoleInterfaceCompileTimeChecks(t *testing.T) {
	// The var _ statements in interfaces.go provide compile-time checks.
	// This test documents that those checks exist and provides runtime verification.
	cw := NewConsoleWriter()

	if _, ok := cw.(ProgressOutput); !ok {
		t.Error("ConsoleWriter should implement ProgressOutput")
	}
	if _, ok := cw.(SummaryOutput); !ok {
		t.Error("ConsoleWriter should implement SummaryOutput")
	}
	if _, ok := cw.(StatusOutput); !ok {
		t.Error("ConsoleWriter should implement StatusOutput")
	}
	if _, ok := cw.(EnvironmentAware); !ok {
		t.Error("ConsoleWriter should implement EnvironmentAware")
	}
	if _, ok := cw.(QuietModeController); !ok {
		t.Error("ConsoleWriter should implement QuietModeController")
	}
}
