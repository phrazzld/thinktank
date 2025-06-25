// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"context"
	"os"
	"time"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
)

// FileSystem defines the interface for file system operations
// This interface abstracts os package operations to enable testing
type FileSystem interface {
	// CreateTemp creates a temporary file with the given pattern
	CreateTemp(dir, pattern string) (*os.File, error)

	// WriteFile writes data to the named file, creating it if necessary
	WriteFile(filename string, data []byte, perm os.FileMode) error

	// ReadFile reads and returns the content of the named file
	ReadFile(filename string) ([]byte, error)

	// Remove removes the named file or directory
	Remove(name string) error

	// MkdirAll creates a directory and all necessary parents
	MkdirAll(path string, perm os.FileMode) error

	// OpenFile opens the named file with specified flag and permissions
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
}

// ExitHandler defines the interface for handling process termination
// This interface abstracts os.Exit() and error handling to enable testing
type ExitHandler interface {
	// Exit terminates the process with the given exit code
	Exit(code int)

	// HandleError processes an error and determines the appropriate exit code
	// It logs the error and calls Exit() with the determined code
	HandleError(ctx context.Context, err error, logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, operation string)
}

// RunConfig holds all dependencies needed for the Run() function
// This follows the same pattern as thinktank.Execute() for consistency
type RunConfig struct {
	// Context (first parameter - Go convention)
	Context context.Context

	// Configuration (parsed from flags)
	Config *config.CliConfig

	// Core operational dependencies (following Execute pattern)
	Logger        logutil.LoggerInterface
	AuditLogger   auditlog.AuditLogger
	APIService    interfaces.APIService
	ConsoleWriter logutil.ConsoleWriter

	// New dependencies for main() logic abstraction
	FileSystem      FileSystem
	ExitHandler     ExitHandler
	ContextGatherer interfaces.ContextGatherer // Optional - for testing file filtering behavior
}

// RunResult holds the result of the Run() function execution
// This enables testing by returning status instead of calling os.Exit()
type RunResult struct {
	// ExitCode is the exit code that would be passed to os.Exit()
	ExitCode int

	// Error is the error that caused failure, if any
	Error error

	// Stats provides additional execution metadata for testing
	Stats *ExecutionStats
}

// ExecutionStats holds metadata about the execution for testing and monitoring
type ExecutionStats struct {
	// Duration is how long the execution took
	Duration time.Duration

	// FilesProcessed is the number of files that were processed
	FilesProcessed int

	// APICalls is the number of API calls made (0 for dry-run)
	APICalls int

	// AuditEntriesWritten is the number of audit log entries written
	AuditEntriesWritten int
}

// ErrorProcessingResult holds the result of error processing for testing
// This enables testing error handling logic without os.Exit() side effects
type ErrorProcessingResult struct {
	// ExitCode is the exit code that should be used
	ExitCode int

	// UserMessage is the user-friendly error message to display
	UserMessage string

	// ShouldExit indicates whether the process should terminate
	ShouldExit bool

	// AuditLogged indicates whether audit logging was attempted
	AuditLogged bool

	// AuditError is any error that occurred during audit logging
	AuditError error
}

// MainConfig holds all dependencies needed for the RunMain() function
// This follows the RunConfig pattern but focuses on bootstrap/initialization dependencies
type MainConfig struct {
	// System dependencies for bootstrap phase
	FileSystem  FileSystem
	ExitHandler ExitHandler

	// Environment and argument access (for testing flag parsing)
	Args   []string            // Command line arguments (os.Args equivalent)
	Getenv func(string) string // Environment variable access (os.Getenv equivalent)

	// System operations (for testing signal handling and context setup)
	Now func() time.Time // Current time (time.Now equivalent)
}

// MainResult holds the result of the RunMain() function execution
// This enables testing bootstrap logic by returning status instead of calling os.Exit()
type MainResult struct {
	// ExitCode is the exit code that would be passed to os.Exit()
	ExitCode int

	// Error is the error that caused failure during bootstrap, if any
	Error error

	// RunResult is the result from the execution phase, if it was reached
	RunResult *RunResult

	// ParsingMode indicates which parser was used (for observability)
	ParsingMode ParsingMode

	// HasDeprecationWarning indicates if deprecation warnings were generated
	HasDeprecationWarning bool
}
