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

// OSFileSystem implements the FileSystem interface using the standard os package
type OSFileSystem struct{}

// CreateTemp creates a temporary file with the given pattern
func (fs *OSFileSystem) CreateTemp(dir, pattern string) (*os.File, error) {
	return os.CreateTemp(dir, pattern)
}

// WriteFile writes data to the named file, creating it if necessary
func (fs *OSFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

// ReadFile reads and returns the content of the named file
func (fs *OSFileSystem) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// Remove removes the named file or directory
func (fs *OSFileSystem) Remove(name string) error {
	return os.Remove(name)
}

// MkdirAll creates a directory and all necessary parents
func (fs *OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// OpenFile opens the named file with specified flag and permissions
func (fs *OSFileSystem) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

// OSExitHandler implements the ExitHandler interface using os.Exit and handleError
type OSExitHandler struct{}

// Exit terminates the process with the given exit code
func (eh *OSExitHandler) Exit(code int) {
	os.Exit(code)
}

// HandleError processes an error and determines the appropriate exit code
// It calls the existing handleError function to maintain compatibility
func (eh *OSExitHandler) HandleError(ctx context.Context, err error, logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, operation string) {
	handleError(ctx, err, logger, auditLogger, operation)
}

// NewProductionRunConfig creates a RunConfig with production implementations
func NewProductionRunConfig(ctx context.Context, config *config.CliConfig, logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, apiService interfaces.APIService, consoleWriter logutil.ConsoleWriter) *RunConfig {
	return &RunConfig{
		Context:       ctx,
		Config:        config,
		Logger:        logger,
		AuditLogger:   auditLogger,
		APIService:    apiService,
		ConsoleWriter: consoleWriter,
		FileSystem:    &OSFileSystem{},
		ExitHandler:   &OSExitHandler{},
	}
}

// NewProductionMainConfig creates a MainConfig with production implementations
func NewProductionMainConfig() *MainConfig {
	return &MainConfig{
		FileSystem:  &OSFileSystem{},
		ExitHandler: &OSExitHandler{},
		Args:        os.Args,
		Getenv:      os.Getenv,
		Now:         time.Now,
	}
}
