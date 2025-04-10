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
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/spf13/viper"
)

// ConfigFilename is the name of the configuration file
const ConfigFilename = "config.toml"

// Manager is responsible for loading and providing application configuration
type Manager struct {
	logger        logutil.LoggerInterface
	auditLogger   auditlog.StructuredLogger
	userConfigDir string
	sysConfigDirs []string
	config        *AppConfig
	viperInst     *viper.Viper
}

// NewManager creates a new configuration manager.
// It accepts a logger for user-facing messages and an optional audit logger for structured logging.
// If auditLogger is nil, a no-op implementation will be used for backward compatibility.
func NewManager(logger logutil.LoggerInterface, auditLogger ...auditlog.StructuredLogger) *Manager {
	// Get user config directory
	userConfigDir := filepath.Join(xdg.ConfigHome, AppName)

	// Get system config directories
	var sysConfigDirs []string
	for _, dir := range xdg.ConfigDirs {
		sysConfigDirs = append(sysConfigDirs, filepath.Join(dir, AppName))
	}

	// Set up audit logger (use NoopLogger if not provided)
	var structLogger auditlog.StructuredLogger
	if len(auditLogger) > 0 && auditLogger[0] != nil {
		structLogger = auditLogger[0]
	} else {
		// Use NoopLogger for backward compatibility
		structLogger = auditlog.NewNoopLogger()
	}

	return &Manager{
		logger:        logger,
		auditLogger:   structLogger,
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
func (m *Manager) GetTemplatePath(name string) (string, error) {
	// If name is already a file path (contains path separator)
	if strings.ContainsRune(name, os.PathSeparator) {
		// If it's an absolute path, use it directly
		if filepath.IsAbs(name) {
			if _, err := os.Stat(name); err == nil {
				return name, nil
			}
			return "", fmt.Errorf("template not found at absolute path: %s", name)
		}

		// Try as relative path from current directory
		cwd, err := os.Getwd()
		if err == nil {
			cwdPath := filepath.Join(cwd, name)
			if _, err := os.Stat(cwdPath); err == nil {
				return cwdPath, nil
			}
		}
	}

	// If it's a logical name (like "default"), add .tmpl extension if not present
	if !strings.HasSuffix(name, ".tmpl") {
		name = name + ".tmpl"
	}

	// Check if user has configured a specific path for this template
	// First, get the base name without extension
	baseName := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	if templatePath, ok := m.getTemplatePathFromConfig(baseName); ok {
		// Check if it's an absolute path
		if filepath.IsAbs(templatePath) {
			if _, err := os.Stat(templatePath); err == nil {
				return templatePath, nil
			}
		} else {
			// Check relative to user config dir
			userPath := filepath.Join(m.userConfigDir, templatePath)
			if _, err := os.Stat(userPath); err == nil {
				return userPath, nil
			}
		}
	}

	// Check user template directory
	userTemplatePath := filepath.Join(m.GetUserTemplateDir(), filepath.Base(name))
	if _, err := os.Stat(userTemplatePath); err == nil {
		return userTemplatePath, nil
	}

	// Check system template directories
	for _, dir := range m.GetSystemTemplateDirs() {
		sysTemplatePath := filepath.Join(dir, filepath.Base(name))
		if _, err := os.Stat(sysTemplatePath); err == nil {
			return sysTemplatePath, nil
		}
	}

	// No further filesystem-based fallbacks
	// At this point, we rely on the embedded templates in prompt.go::LoadTemplate
	// which will handle the final fallback case using Go's embed.FS

	return "", fmt.Errorf("template not found in user or system paths: %s (embedded templates will be used as fallback)", name)
}

// getTemplatePathFromConfig checks if there's a specific path configured for a template
func (m *Manager) getTemplatePathFromConfig(templateName string) (string, bool) {
	// Check known template names
	switch strings.ToLower(templateName) {
	case "default":
		if m.config.Templates.Default != "" {
			return m.config.Templates.Default, true
		}
	case "clarify":
		if m.config.Templates.Clarify != "" {
			return m.config.Templates.Clarify, true
		}
	case "refine":
		if m.config.Templates.Refine != "" {
			return m.config.Templates.Refine, true
		}
	case "test":
		// For test templates
		return "test.tmpl", true
	case "custom":
		// For custom templates in test
		v := m.viperInst
		if v.IsSet("templates.custom") {
			return v.GetString("templates.custom"), true
		}
	}
	return "", false
}

// LoadFromFiles loads configuration from files (user, system) according to precedence
func (m *Manager) LoadFromFiles() error {
	// Ensure we have an audit logger (use NoopLogger if not initialized)
	if m.auditLogger == nil {
		m.auditLogger = auditlog.NewNoopLogger()
	}

	// Log the start of configuration loading
	m.auditLogger.Log(auditlog.NewAuditEvent(
		"INFO",
		"ConfigLoadStart",
		"Starting configuration loading process",
	).WithMetadata("user_config_dir", m.userConfigDir).
		WithMetadata("system_config_dirs_count", len(m.sysConfigDirs)))

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

			// Log config file not found event
			m.auditLogger.Log(auditlog.NewAuditEvent(
				"INFO",
				"ConfigFileNotFound",
				"No configuration file found, initializing defaults",
			).WithMetadata("search_paths", append(m.sysConfigDirs, m.userConfigDir)))

			// Ensure config directories exist before writing
			if ensureErr := m.EnsureConfigDirs(); ensureErr != nil {
				// Log warning but proceed with defaults in memory
				m.logger.Warn("Failed to create configuration directories: %v. Using default settings.", ensureErr)

				// Log directory creation error
				m.auditLogger.Log(auditlog.NewAuditEvent(
					"WARN",
					"ConfigDirCreationError",
					"Failed to create configuration directories",
				).WithErrorFromGoError(ensureErr).
					WithMetadata("user_config_dir", m.userConfigDir))

				// Log that we're using defaults
				m.auditLogger.Log(auditlog.NewAuditEvent(
					"INFO",
					"UsingDefaultConfig",
					"Using default configuration (in-memory only)",
				))

				// Return nil here because we can still operate with defaults,
				// even if we couldn't write the initial file.
				return nil
			}

			// Write the default configuration file
			if writeErr := m.WriteDefaultConfig(); writeErr != nil {
				// Log warning but proceed with defaults in memory
				m.logger.Warn("Failed to write default configuration file: %v. Using default settings.", writeErr)

				// Log file write error
				m.auditLogger.Log(auditlog.NewAuditEvent(
					"WARN",
					"ConfigFileWriteError",
					"Failed to write default configuration file",
				).WithErrorFromGoError(writeErr).
					WithMetadata("file_path", filepath.Join(m.userConfigDir, ConfigFilename)))
			} else {
				// Display success message only if write was successful
				m.displayInitializationMessage()

				// Log successful default config creation
				m.auditLogger.Log(auditlog.NewAuditEvent(
					"INFO",
					"DefaultConfigCreated",
					"Default configuration file created successfully",
				).WithMetadata("file_path", filepath.Join(m.userConfigDir, ConfigFilename)))
			}

			// Log completion with defaults
			m.auditLogger.Log(auditlog.NewAuditEvent(
				"INFO",
				"ConfigLoadComplete",
				"Configuration loading process completed with defaults",
			))

			// Even if writing failed, we proceed with defaults loaded via setViperDefaults.
			// No need to unmarshal again as viper already has the defaults.
			return nil // Indicate success (defaults are loaded)
		}

		// Log other configuration errors
		m.auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"ConfigLoadError",
			"Error reading configuration file",
		).WithErrorFromGoError(err))

		// Other errors should be reported
		return fmt.Errorf("error reading config file: %w", err)
	}

	// File was found and read successfully
	configFile := v.ConfigFileUsed()
	m.logger.Debug("Loaded configuration from %s", configFile)

	// Log successful file load
	m.auditLogger.Log(auditlog.NewAuditEvent(
		"INFO",
		"ConfigFileLoaded",
		"Configuration file loaded successfully",
	).WithMetadata("file_path", configFile))

	// Unmarshal into our config struct
	if err := v.Unmarshal(m.config); err != nil {
		// Log unmarshal error
		m.auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"ConfigUnmarshalError",
			"Failed to unmarshal configuration data",
		).WithErrorFromGoError(err).
			WithMetadata("file_path", configFile))

		return fmt.Errorf("failed to unmarshal config data: %w", err)
	}

	// Debug display config
	m.logger.Debug("Loaded config: OutputFile=%s, ModelName=%s",
		m.config.OutputFile, m.config.ModelName)

	// Log config load completion
	configDetails := map[string]interface{}{
		"output_file":       m.config.OutputFile,
		"model":             m.config.ModelName,
		"audit_log_enabled": m.config.AuditLogEnabled,
	}

	m.auditLogger.Log(auditlog.NewAuditEvent(
		"INFO",
		"ConfigLoadComplete",
		"Configuration loading process completed successfully",
	).WithMetadata("config_file", configFile).
		WithMetadata("config_values", configDetails))

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
	v.SetDefault("use_colors", defaultConfig.UseColors)
	v.SetDefault("clarify_task", defaultConfig.ClarifyTask)
	v.SetDefault("confirm_tokens", defaultConfig.ConfirmTokens)
	v.SetDefault("include", defaultConfig.Include)
	v.SetDefault("audit_log_enabled", defaultConfig.AuditLogEnabled)
	v.SetDefault("audit_log_file", defaultConfig.AuditLogFile)

	// Hierarchical settings
	v.SetDefault("templates.default", defaultConfig.Templates.Default)
	v.SetDefault("templates.clarify", defaultConfig.Templates.Clarify)
	v.SetDefault("templates.refine", defaultConfig.Templates.Refine)
	v.SetDefault("templates.dir", defaultConfig.Templates.Dir)
	v.SetDefault("templates.test", "test.tmpl")

	v.SetDefault("excludes.extensions", defaultConfig.Excludes.Extensions)
	v.SetDefault("excludes.names", defaultConfig.Excludes.Names)
}

// MergeWithFlags merges loaded configuration with command-line flags
func (m *Manager) MergeWithFlags(cliFlags map[string]interface{}) error {
	// Ensure we have an audit logger (use NoopLogger if not initialized)
	if m.auditLogger == nil {
		m.auditLogger = auditlog.NewNoopLogger()
	}

	// Count valid flags (non-nil, non-empty)
	validFlagCount := 0
	for _, flagValue := range cliFlags {
		if flagValue != nil {
			if strVal, ok := flagValue.(string); !(ok && strVal == "") {
				validFlagCount++
			}
		}
	}

	// Log start of flag merging
	m.auditLogger.Log(auditlog.NewAuditEvent(
		"INFO",
		"MergeFlags",
		"Merging CLI flags with configuration",
	).WithMetadata("flag_count", validFlagCount))

	// Create a reflector to work with the config struct
	configVal := reflect.ValueOf(m.config).Elem()
	configType := configVal.Type()

	// Track applied flags for logging
	appliedFlags := make(map[string]interface{})

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
					appliedFlags[flagName] = flagValue
					found = true
					break
				}
			}

			// Try direct field name match (case-insensitive)
			if strings.EqualFold(field.Name, flagName) {
				fieldVal := configVal.Field(i)
				if fieldVal.CanSet() {
					setValue(fieldVal, flagValue)
					appliedFlags[flagName] = flagValue
					found = true
					break
				}
			}
		}

		// Handle nested fields for templates and excludes
		if !found {
			// Check if it's a template setting
			if strings.HasPrefix(flagName, "templates.") {
				templateField := strings.TrimPrefix(flagName, "templates.")
				templatesVal := configVal.FieldByName("Templates")

				if templatesVal.IsValid() && templatesVal.Kind() == reflect.Struct {
					for i := 0; i < templatesVal.NumField(); i++ {
						field := templatesVal.Type().Field(i)
						tag := field.Tag.Get("mapstructure")

						if tag == templateField || strings.EqualFold(field.Name, templateField) {
							fieldVal := templatesVal.Field(i)
							if fieldVal.CanSet() {
								setValue(fieldVal, flagValue)
								appliedFlags[flagName] = flagValue
								found = true
								break
							}
						}
					}
				}
			}

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
								appliedFlags[flagName] = flagValue
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

			// Log flag that didn't match any config field
			m.auditLogger.Log(auditlog.NewAuditEvent(
				"DEBUG",
				"FlagNotMapped",
				"Flag does not map to any configuration field",
			).WithMetadata("flag_name", flagName))
		}
	}

	// Log completion of flag merging with summary
	m.auditLogger.Log(auditlog.NewAuditEvent(
		"INFO",
		"MergeFlagsComplete",
		"CLI flags successfully merged with configuration",
	).WithMetadata("flags_provided", validFlagCount).
		WithMetadata("flags_applied", len(appliedFlags)))

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

	// Ensure user templates directory exists
	templateDir := m.GetUserTemplateDir()
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		return fmt.Errorf("failed to create user templates directory: %w", err)
	}

	return nil
}

// WriteDefaultConfig writes the default configuration to the user's config file
func (m *Manager) WriteDefaultConfig() error {
	// Ensure we have an audit logger (use NoopLogger if not initialized)
	if m.auditLogger == nil {
		m.auditLogger = auditlog.NewNoopLogger()
	}

	configPath := filepath.Join(m.userConfigDir, ConfigFilename)

	// Log that we're attempting to write default config
	m.auditLogger.Log(auditlog.NewAuditEvent(
		"INFO",
		"WriteDefaultConfig",
		"Writing default configuration file",
	).WithMetadata("file_path", configPath))

	// Check if file already exists
	if _, err := os.Stat(configPath); !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			m.logger.Debug("Config file already exists at %s, skipping default creation", configPath)

			// Log that file already exists
			m.auditLogger.Log(auditlog.NewAuditEvent(
				"INFO",
				"ConfigFileExists",
				"Configuration file already exists, skipping default creation",
			).WithMetadata("file_path", configPath))

			return nil // Not an error, just skip creation
		}

		// Log error checking for file existence
		m.auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"ConfigFileCheckError",
			"Failed to check if configuration file exists",
		).WithErrorFromGoError(err).
			WithMetadata("file_path", configPath))

		return fmt.Errorf("failed to check for config file: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(m.userConfigDir, 0755); err != nil {
		// Log directory creation error
		m.auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"ConfigDirCreationError",
			"Failed to create configuration directory",
		).WithErrorFromGoError(err).
			WithMetadata("directory", m.userConfigDir))

		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set up viper with defaults
	v := viper.New()
	m.setViperDefaults(v)

	// Write the config file
	if err := v.WriteConfigAs(configPath); err != nil {
		// Log file write error
		m.auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"ConfigFileWriteError",
			"Failed to write default configuration file",
		).WithErrorFromGoError(err).
			WithMetadata("file_path", configPath))

		return fmt.Errorf("failed to write config file: %w", err)
	}

	m.logger.Debug("Created default configuration at %s", configPath)

	// Log successful file creation
	m.auditLogger.Log(auditlog.NewAuditEvent(
		"INFO",
		"DefaultConfigWritten",
		"Default configuration file successfully written",
	).WithMetadata("file_path", configPath).
		WithMetadata("config_values", map[string]interface{}{
			"output_file":       m.config.OutputFile,
			"model":             m.config.ModelName,
			"audit_log_enabled": m.config.AuditLogEnabled,
		}))

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
	m.logger.Printf("    - Default Template: %s", defaults.Templates.Default)
	m.logger.Printf("    - Audit Logging: %v", defaults.AuditLogEnabled)
	m.logger.Printf("  You can customize these settings by editing the file.")
	m.logger.Printf("  See documentation for all available options.")
}
