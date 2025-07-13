// Package config provides minimal configuration for the simplified thinktank CLI.
// This replaces the complex 27-field CliConfig with only essential fields.
package config

import (
	"github.com/phrazzld/thinktank/internal/logutil"
	"time"
)

// MinimalConfig represents the essential configuration for thinktank execution.
// This struct replaces the complex CliConfig, focusing only on what's actually
// needed for the simplified CLI approach.
type MinimalConfig struct {
	// Core execution fields
	InstructionsFile string   // Path to instructions file
	TargetPaths      []string // Paths to analyze
	ModelNames       []string // Models to use for generation
	OutputDir        string   // Directory for output files

	// Execution modes
	DryRun         bool   // Show what would be processed without calling API
	Verbose        bool   // Enable verbose output
	SynthesisModel string // Optional model for synthesizing results

	// Minimal additional fields that are actually used
	LogLevel   logutil.LogLevel // Logging verbosity
	Timeout    time.Duration    // Global timeout for operation
	Quiet      bool             // Suppress non-error output
	NoProgress bool             // Disable progress indicators
	JsonLogs   bool             // Show JSON logs on stderr (preserves old behavior)

	// File handling (using smart defaults)
	Format       string // Format string for file content
	Exclude      string // File extensions to exclude
	ExcludeNames string // File/dir names to exclude

	// Token safety margin percentage (0-50%) - percentage of context window reserved for output
	TokenSafetyMargin uint8
}

// NewDefaultMinimalConfig returns a MinimalConfig with sensible defaults.
// Smart defaults are applied here rather than storing complex configuration.
func NewDefaultMinimalConfig() *MinimalConfig {
	return &MinimalConfig{
		ModelNames:        []string{DefaultModel},
		OutputDir:         "", // Will be set to timestamp-based dir by output manager
		LogLevel:          logutil.InfoLevel,
		Timeout:           DefaultTimeout,
		Format:            DefaultFormat,
		Exclude:           DefaultExcludes,
		ExcludeNames:      DefaultExcludeNames,
		TokenSafetyMargin: 10, // Default 10% safety margin
	}
}

// ConfigInterface defines the minimal interface needed by orchestrator and core components.
// This allows us to transition from CliConfig to MinimalConfig gradually.
type ConfigInterface interface {
	// Core getters
	GetInstructionsFile() string
	GetTargetPaths() []string
	GetModelNames() []string
	GetOutputDir() string

	// Mode getters
	IsDryRun() bool
	IsVerbose() bool
	GetSynthesisModel() string

	// Additional getters used by orchestrator
	GetLogLevel() logutil.LogLevel
	GetTimeout() time.Duration
	IsQuiet() bool
	ShouldShowProgress() bool
	ShouldShowJsonLogs() bool

	// File handling
	GetFormat() string
	GetExclude() string
	GetExcludeNames() string
}

// Implement ConfigInterface for MinimalConfig

func (m *MinimalConfig) GetInstructionsFile() string   { return m.InstructionsFile }
func (m *MinimalConfig) GetTargetPaths() []string      { return m.TargetPaths }
func (m *MinimalConfig) GetModelNames() []string       { return m.ModelNames }
func (m *MinimalConfig) GetOutputDir() string          { return m.OutputDir }
func (m *MinimalConfig) IsDryRun() bool                { return m.DryRun }
func (m *MinimalConfig) IsVerbose() bool               { return m.Verbose }
func (m *MinimalConfig) GetSynthesisModel() string     { return m.SynthesisModel }
func (m *MinimalConfig) GetLogLevel() logutil.LogLevel { return m.LogLevel }
func (m *MinimalConfig) GetTimeout() time.Duration     { return m.Timeout }
func (m *MinimalConfig) IsQuiet() bool                 { return m.Quiet }
func (m *MinimalConfig) ShouldShowProgress() bool      { return !m.NoProgress }
func (m *MinimalConfig) ShouldShowJsonLogs() bool      { return m.JsonLogs }
func (m *MinimalConfig) GetFormat() string             { return m.Format }
func (m *MinimalConfig) GetExclude() string            { return m.Exclude }
func (m *MinimalConfig) GetExcludeNames() string       { return m.ExcludeNames }

// Implement ConfigInterface for CliConfig (temporary shim for transition)

func (c *CliConfig) GetInstructionsFile() string   { return c.InstructionsFile }
func (c *CliConfig) GetTargetPaths() []string      { return c.Paths }
func (c *CliConfig) GetModelNames() []string       { return c.ModelNames }
func (c *CliConfig) GetOutputDir() string          { return c.OutputDir }
func (c *CliConfig) IsDryRun() bool                { return c.DryRun }
func (c *CliConfig) IsVerbose() bool               { return c.Verbose }
func (c *CliConfig) GetSynthesisModel() string     { return c.SynthesisModel }
func (c *CliConfig) GetLogLevel() logutil.LogLevel { return c.LogLevel }
func (c *CliConfig) GetTimeout() time.Duration     { return c.Timeout }
func (c *CliConfig) IsQuiet() bool                 { return c.Quiet }
func (c *CliConfig) ShouldShowProgress() bool      { return !c.NoProgress }
func (c *CliConfig) ShouldShowJsonLogs() bool      { return c.JsonLogs }
func (c *CliConfig) GetFormat() string             { return c.Format }
func (c *CliConfig) GetExclude() string            { return c.Exclude }
func (c *CliConfig) GetExcludeNames() string       { return c.ExcludeNames }
