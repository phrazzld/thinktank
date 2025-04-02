// Package config provides configuration management for the architect application
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/phrazzld/architect/internal/logutil"
)

// ConfigFilename is the name of the configuration file
const ConfigFilename = "config.toml"

// Manager is responsible for loading and providing application configuration
type Manager struct {
	logger        logutil.LoggerInterface
	userConfigDir string
	sysConfigDirs []string
	config        *AppConfig
}

// NewManager creates a new configuration manager
func NewManager(logger logutil.LoggerInterface) *Manager {
	// Get user config directory
	userConfigDir := filepath.Join(xdg.ConfigHome, AppName)

	// Get system config directories
	var sysConfigDirs []string
	for _, dir := range xdg.ConfigDirs {
		sysConfigDirs = append(sysConfigDirs, filepath.Join(dir, AppName))
	}

	return &Manager{
		logger:        logger,
		userConfigDir: userConfigDir,
		sysConfigDirs: sysConfigDirs,
		config:        DefaultConfig(),
	}
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *AppConfig {
	return m.config
}

// GetUserConfigDir returns the user-specific configuration directory
func (m *Manager) GetUserConfigDir() string {
	return m.userConfigDir
}

// GetSystemConfigDirs returns the system-wide configuration directories
func (m *Manager) GetSystemConfigDirs() []string {
	return m.sysConfigDirs
}

// GetUserTemplateDir returns the directory for user-specific templates
func (m *Manager) GetUserTemplateDir() string {
	if m.config.Templates.Dir != "" {
		// If the template dir in config is absolute, use it directly
		if filepath.IsAbs(m.config.Templates.Dir) {
			return m.config.Templates.Dir
		}
		// Otherwise, it's relative to user config dir
		return filepath.Join(m.userConfigDir, m.config.Templates.Dir)
	}
	// Default templates directory within user config
	return filepath.Join(m.userConfigDir, "templates")
}

// GetSystemTemplateDirs returns the system-wide template directories
func (m *Manager) GetSystemTemplateDirs() []string {
	dirs := []string{}
	for _, dir := range m.sysConfigDirs {
		// If template dir is specified in config, use that (to be implemented)
		// For now, just use a standard "templates" subdir
		dirs = append(dirs, filepath.Join(dir, "templates"))
	}
	return dirs
}

// GetTemplatePath finds the path to a template file using the configured precedence
// This is a placeholder - will be implemented in the next step
func (m *Manager) GetTemplatePath(name string) (string, error) {
	// For now, just return a placeholder indicating this is not yet implemented
	return "", fmt.Errorf("GetTemplatePath not yet implemented")
}

// LoadFromFiles loads configuration from files (user, system) according to precedence
// This is a placeholder - will be implemented in the next step
func (m *Manager) LoadFromFiles() error {
	// For now, just log that we're using default config
	m.logger.Debug("LoadFromFiles not yet implemented, using default configuration")
	return nil
}

// MergeWithFlags merges loaded configuration with command-line flags
// This is a placeholder - will be implemented in the next step
func (m *Manager) MergeWithFlags(cliFlags map[string]interface{}) error {
	// For now, just log that we're not merging
	m.logger.Debug("MergeWithFlags not yet implemented, using CLI flags directly")
	return nil
}

// EnsureConfigDirs creates necessary configuration directories if they don't exist
func (m *Manager) EnsureConfigDirs() error {
	// Ensure user config directory exists
	if err := os.MkdirAll(m.userConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create user config directory: %w", err)
	}

	// Ensure user templates directory exists
	templateDir := m.GetUserTemplateDir()
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		return fmt.Errorf("failed to create user templates directory: %w", err)
	}

	return nil
}
