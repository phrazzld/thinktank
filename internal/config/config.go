// Package config handles loading and managing application configuration
package config

import (
	"github.com/phrazzld/architect/internal/logutil"
)

// Configuration constants
const (
	// Default values
	DefaultOutputFile = "PLAN.md"
	DefaultModel      = "gemini-2.5-pro-exp-03-25"
	APIKeyEnvVar      = "GEMINI_API_KEY"
	DefaultFormat     = "<{path}>\n```\n{content}\n```\n</{path}>\n\n"

	// Default excludes for file extensions
	DefaultExcludes = ".exe,.bin,.obj,.o,.a,.lib,.so,.dll,.dylib,.class,.jar,.pyc,.pyo,.pyd," +
		".zip,.tar,.gz,.rar,.7z,.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.odt,.ods,.odp," +
		".jpg,.jpeg,.png,.gif,.bmp,.tiff,.svg,.mp3,.wav,.ogg,.mp4,.avi,.mov,.wmv,.flv," +
		".iso,.img,.dmg,.db,.sqlite,.log"

	// Default excludes for file and directory names
	DefaultExcludeNames = ".git,.hg,.svn,node_modules,bower_components,vendor,target,dist,build," +
		"out,tmp,coverage,__pycache__,*.pyc,*.pyo,.DS_Store,~$*,desktop.ini,Thumbs.db," +
		"package-lock.json,yarn.lock,go.sum,go.work"
)

// ExcludeConfig defines file exclusion configuration
type ExcludeConfig struct {
	// File extensions to exclude
	Extensions string
	// File and directory names to exclude
	Names string
}

// AppConfig holds configuration settings loaded from defaults, env vars, and flags
type AppConfig struct {
	// Task-related settings
	TaskDescription string // Not saved to config
	TaskFile        string // Not saved to config
	OutputFile      string
	ModelName       string

	// File handling settings
	Include       string
	Format        string
	ConfirmTokens int

	// Logging and display settings
	Verbose  bool
	LogLevel logutil.LogLevel
	DryRun   bool // Not saved to config

	// Template settings have been removed

	// Exclude settings (hierarchical)
	Excludes ExcludeConfig

	// Input paths (not saved to config file)
	Paths []string

	// API key (not saved to config file for security)
	APIKey string
}

// DefaultConfig returns a new AppConfig instance with default values
func DefaultConfig() *AppConfig {
	return &AppConfig{
		OutputFile:    DefaultOutputFile,
		ModelName:     DefaultModel,
		Format:        DefaultFormat,
		LogLevel:      logutil.InfoLevel,
		ConfirmTokens: 0, // Disabled by default
		Excludes: ExcludeConfig{
			Extensions: DefaultExcludes,
			Names:      DefaultExcludeNames,
		},
	}
}
