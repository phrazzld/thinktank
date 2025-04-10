// Package config handles loading and managing application configuration
package config

import (
	"github.com/phrazzld/architect/internal/logutil"
)

// Configuration constants
const (
	// App name used for XDG paths
	AppName = "architect"

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

// TemplateConfig defines template-specific configuration options
type TemplateConfig struct {
	// Default template used for generating content (logical name or path)
	Default string `mapstructure:"default" toml:"default"`
	// Clarification template used for task analysis
	Clarify string `mapstructure:"clarify" toml:"clarify"`
	// Refinement template used for refining task description
	Refine string `mapstructure:"refine" toml:"refine"`
	// Test template for integration testing
	Test string `mapstructure:"test" toml:"test"`
	// Custom template for integration testing
	Custom string `mapstructure:"custom" toml:"custom"`
	// Directory to look for custom templates, relative to config dir or absolute
	Dir string `mapstructure:"dir" toml:"dir"`
}

// ExcludeConfig defines file exclusion configuration
type ExcludeConfig struct {
	// File extensions to exclude
	Extensions string `mapstructure:"extensions" toml:"extensions"`
	// File and directory names to exclude
	Names string `mapstructure:"names" toml:"names"`
}

// AppConfig holds configuration settings loaded from config files, env vars, and flags
type AppConfig struct {
	// Task-related settings
	TaskDescription string `mapstructure:"task_description" toml:"-"` // Not saved to config
	TaskFile        string `mapstructure:"task_file" toml:"-"`        // Not saved to config
	OutputFile      string `mapstructure:"output_file" toml:"output_file"`
	ModelName       string `mapstructure:"model" toml:"model"`

	// File handling settings
	Include       string `mapstructure:"include" toml:"include"`
	Format        string `mapstructure:"format" toml:"format"`
	ConfirmTokens int    `mapstructure:"confirm_tokens" toml:"confirm_tokens"`

	// Logging and display settings
	Verbose   bool             `mapstructure:"verbose" toml:"verbose"`
	LogLevel  logutil.LogLevel `mapstructure:"log_level" toml:"log_level"`
	UseColors bool             `mapstructure:"use_colors" toml:"use_colors"`
	DryRun    bool             `mapstructure:"dry_run" toml:"-"` // Not saved to config

	// Template settings (hierarchical)
	Templates TemplateConfig `mapstructure:"templates" toml:"templates"`

	// Exclude settings (hierarchical)
	Excludes ExcludeConfig `mapstructure:"excludes" toml:"excludes"`

	// Input paths (not saved to config file)
	Paths []string `mapstructure:"paths" toml:"-"`

	// API key (not saved to config file for security)
	APIKey string `mapstructure:"api_key" toml:"-"`
}

// DefaultConfig returns a new AppConfig instance with default values
func DefaultConfig() *AppConfig {
	return &AppConfig{
		OutputFile:    DefaultOutputFile,
		ModelName:     DefaultModel,
		Format:        DefaultFormat,
		UseColors:     true,
		LogLevel:      logutil.InfoLevel,
		ConfirmTokens: 0, // Disabled by default
		Excludes: ExcludeConfig{
			Extensions: DefaultExcludes,
			Names:      DefaultExcludeNames,
		},
		Templates: TemplateConfig{
			Default: "default.tmpl",
			Clarify: "clarify.tmpl",
			Refine:  "refine.tmpl",
		},
	}
}
