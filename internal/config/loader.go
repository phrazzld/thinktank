// Package config provides configuration management for the architect application
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/adrg/xdg"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/spf13/viper"
)

// ConfigFilename is the name of the configuration file
const ConfigFilename = "config.toml"

// Manager is responsible for loading and providing application configuration
type Manager struct {
	logger        logutil.LoggerInterface
	userConfigDir string
	sysConfigDirs []string
	config        *AppConfig
	viperInst     *viper.Viper
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
		viperInst:     viper.New(),
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

// GetConfigDirs returns all configuration directories
func (m *Manager) GetConfigDirs() ConfigDirectories {
	return ConfigDirectories{
		UserConfigDir:    m.userConfigDir,
		SystemConfigDirs: m.sysConfigDirs,
	}
}

// Template-related functions have been removed

// LoadFromFiles loads configuration from files (user, system) according to precedence
func (m *Manager) LoadFromFiles() error {
	v := m.viperInst
	v.SetConfigType("toml")
	v.SetConfigName(strings.TrimSuffix(ConfigFilename, filepath.Ext(ConfigFilename)))

	// Set up Viper with default values
	m.setViperDefaults(v)

	// Add config paths in precedence order (lowest to highest)
	// System-wide configs (processed in reverse order, so add them in reversed precedence)
	for i := len(m.sysConfigDirs) - 1; i >= 0; i-- {
		v.AddConfigPath(m.sysConfigDirs[i])
		m.logger.Debug("Added system config path: %s", m.sysConfigDirs[i])
	}
	// User config has highest precedence
	v.AddConfigPath(m.userConfigDir)
	m.logger.Debug("Added user config path: %s", m.userConfigDir)

	// Attempt to read config files
	err := v.ReadInConfig()
	if err != nil {
		// Check if the error is specifically ConfigFileNotFoundError
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			m.logger.Info("No configuration file found. Initializing default configuration...")

			// Ensure config directories exist before writing
			if ensureErr := m.EnsureConfigDirs(); ensureErr != nil {
				// Log warning but proceed with defaults in memory
				m.logger.Warn("Failed to create configuration directories: %v. Using default settings.", ensureErr)
				// Return nil here because we can still operate with defaults,
				// even if we couldn't write the initial file.
				return nil
			}

			// Write the default configuration file
			if writeErr := m.WriteDefaultConfig(); writeErr != nil {
				// Log warning but proceed with defaults in memory
				m.logger.Warn("Failed to write default configuration file: %v. Using default settings.", writeErr)
			} else {
				// Display success message only if write was successful
				m.displayInitializationMessage()
			}
			// Even if writing failed, we proceed with defaults loaded via setViperDefaults.
			// No need to unmarshal again as viper already has the defaults.
			return nil // Indicate success (defaults are loaded)
		}
		// Other errors should be reported
		return fmt.Errorf("error reading config file: %w", err)
	}

	// File was found and read successfully
	m.logger.Debug("Loaded configuration from %s", v.ConfigFileUsed())

	// Unmarshal into our config struct
	if err := v.Unmarshal(m.config); err != nil {
		return fmt.Errorf("failed to unmarshal config data: %w", err)
	}

	// Debug display config
	m.logger.Debug("Loaded config: OutputFile=%s, ModelName=%s",
		m.config.OutputFile, m.config.ModelName)

	return nil
}

// setViperDefaults initializes viper with default values from the DefaultConfig
func (m *Manager) setViperDefaults(v *viper.Viper) {
	defaultConfig := DefaultConfig()

	// Set defaults in Viper
	v.SetDefault("output_file", defaultConfig.OutputFile)
	v.SetDefault("model", defaultConfig.ModelName)
	v.SetDefault("format", defaultConfig.Format)
	v.SetDefault("verbose", defaultConfig.Verbose)
	v.SetDefault("log_level", defaultConfig.LogLevel)
	v.SetDefault("confirm_tokens", defaultConfig.ConfirmTokens)
	v.SetDefault("include", defaultConfig.Include)

	// Template settings have been removed

	v.SetDefault("excludes.extensions", defaultConfig.Excludes.Extensions)
	v.SetDefault("excludes.names", defaultConfig.Excludes.Names)
}

// MergeWithFlags merges loaded configuration with command-line flags
func (m *Manager) MergeWithFlags(cliFlags map[string]interface{}) error {
	// Create a reflector to work with the config struct
	configVal := reflect.ValueOf(m.config).Elem()
	configType := configVal.Type()

	// Process each flag
	for flagName, flagValue := range cliFlags {
		// Skip if flag value is nil or empty string
		if flagValue == nil {
			continue
		}
		if strVal, ok := flagValue.(string); ok && strVal == "" {
			continue
		}

		// Convert flag name to the field name format (snake_case to camelCase or exact match)
		var found bool
		for i := 0; i < configType.NumField(); i++ {
			field := configType.Field(i)

			// Check for mapstructure tag match
			tag := field.Tag.Get("mapstructure")
			if tag == flagName {
				fieldVal := configVal.Field(i)
				if fieldVal.CanSet() {
					setValue(fieldVal, flagValue)
					found = true
					break
				}
			}

			// Try direct field name match (case-insensitive)
			if strings.EqualFold(field.Name, flagName) {
				fieldVal := configVal.Field(i)
				if fieldVal.CanSet() {
					setValue(fieldVal, flagValue)
					found = true
					break
				}
			}
		}

		// Handle nested fields for excludes (templates removed)
		if !found {
			// Template setting handling has been removed

			// Check if it's an excludes setting
			if strings.HasPrefix(flagName, "excludes.") {
				excludeField := strings.TrimPrefix(flagName, "excludes.")
				excludesVal := configVal.FieldByName("Excludes")

				if excludesVal.IsValid() && excludesVal.Kind() == reflect.Struct {
					for i := 0; i < excludesVal.NumField(); i++ {
						field := excludesVal.Type().Field(i)
						tag := field.Tag.Get("mapstructure")

						if tag == excludeField || strings.EqualFold(field.Name, excludeField) {
							fieldVal := excludesVal.Field(i)
							if fieldVal.CanSet() {
								setValue(fieldVal, flagValue)
								found = true
								break
							}
						}
					}
				}
			}
		}

		if !found {
			m.logger.Debug("Flag '%s' does not map to any config field", flagName)
		}
	}

	return nil
}

// setValue sets a reflected Value to the given interface{} value
func setValue(field reflect.Value, value interface{}) {
	if !field.CanSet() {
		return
	}

	switch field.Kind() {
	case reflect.String:
		if str, ok := value.(string); ok {
			field.SetString(str)
		}
	case reflect.Bool:
		if b, ok := value.(bool); ok {
			field.SetBool(b)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, ok := value.(int); ok {
			field.SetInt(int64(i))
		} else if i64, ok := value.(int64); ok {
			field.SetInt(i64)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if u, ok := value.(uint); ok {
			field.SetUint(uint64(u))
		} else if u64, ok := value.(uint64); ok {
			field.SetUint(u64)
		}
	case reflect.Float32, reflect.Float64:
		if f, ok := value.(float64); ok {
			field.SetFloat(f)
		}
	case reflect.Slice:
		// Handle string slices
		if strSlice, ok := value.([]string); ok && field.Type().Elem().Kind() == reflect.String {
			newSlice := reflect.MakeSlice(field.Type(), len(strSlice), len(strSlice))
			for i, str := range strSlice {
				newSlice.Index(i).SetString(str)
			}
			field.Set(newSlice)
		}
	case reflect.Struct:
		// We don't handle direct struct assignment,
		// as nested structs should be accessed via their fields
	}
}

// EnsureConfigDirs creates necessary configuration directories if they don't exist
func (m *Manager) EnsureConfigDirs() error {
	// Ensure user config directory exists
	if err := os.MkdirAll(m.userConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create user config directory: %w", err)
	}

	return nil
}

// WriteDefaultConfig writes the default configuration to the user's config file
func (m *Manager) WriteDefaultConfig() error {
	configPath := filepath.Join(m.userConfigDir, ConfigFilename)

	// Check if file already exists
	if _, err := os.Stat(configPath); !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			m.logger.Debug("Config file already exists at %s, skipping default creation", configPath)
			return nil // Not an error, just skip creation
		}
		return fmt.Errorf("failed to check for config file: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(m.userConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set up viper with defaults
	v := viper.New()
	m.setViperDefaults(v)

	// Write the config file
	if err := v.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	m.logger.Debug("Created default configuration at %s", configPath)
	return nil
}

// displayInitializationMessage prints information about the automatic config creation
func (m *Manager) displayInitializationMessage() {
	configPath := filepath.Join(m.userConfigDir, ConfigFilename)
	defaults := DefaultConfig() // Get a fresh set of defaults to display

	// Use logger.Printf to ensure color settings are respected
	m.logger.Printf("✓ Architect configuration initialized automatically.")
	m.logger.Printf("  Created default configuration file at: %s", configPath)
	m.logger.Printf("  Applying default settings:")
	m.logger.Printf("    - Output File: %s", defaults.OutputFile)
	m.logger.Printf("    - Model: %s", defaults.ModelName)
	m.logger.Printf("    - Log Level: %s", defaults.LogLevel)
	m.logger.Printf("  You can customize these settings by editing the file.")
	m.logger.Printf("  See documentation for all available options.")
}
