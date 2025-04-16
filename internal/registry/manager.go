// Package registry provides a configuration-driven registry
// for LLM providers and models, allowing for flexible configuration
// and easier addition of new models and providers.
package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/providers/gemini"
	"github.com/phrazzld/architect/internal/providers/openai"
)

// Manager provides a singleton-like access to the registry.
// It handles initialization, configuration loading, and provider registration.
type Manager struct {
	registry *Registry
	logger   logutil.LoggerInterface
	mu       sync.RWMutex
	loaded   bool
}

var (
	// globalManager is the singleton instance of the registry manager
	globalManager *Manager
	// managerMu protects access to the global manager
	managerMu sync.Mutex
)

// NewManager creates a new registry manager.
func NewManager(logger logutil.LoggerInterface) *Manager {
	if logger == nil {
		logger = logutil.NewLogger(logutil.InfoLevel, nil, "[registry-manager] ")
	}

	return &Manager{
		registry: NewRegistry(logger),
		logger:   logger,
		loaded:   false,
	}
}

// GetGlobalManager returns the global registry manager instance,
// creating it if necessary.
func GetGlobalManager(logger logutil.LoggerInterface) *Manager {
	managerMu.Lock()
	defer managerMu.Unlock()

	if globalManager == nil {
		globalManager = NewManager(logger)
	}

	return globalManager
}

// Initialize loads the registry configuration and registers
// provider implementations.
func (m *Manager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.loaded {
		m.logger.Debug("Registry already initialized, skipping")
		return nil
	}

	m.logger.Info("Initializing registry")

	// Load configuration
	configLoader := NewConfigLoader()
	if err := m.registry.LoadConfig(configLoader); err != nil {
		// Check if the error is due to missing config file
		if os.IsNotExist(err) {
			m.logger.Warn("Configuration file not found. Attempting to install default configuration.")
			if err := m.installDefaultConfig(); err != nil {
				return fmt.Errorf("failed to install default configuration: %w", err)
			}

			// Try loading again after installation
			if err := m.registry.LoadConfig(configLoader); err != nil {
				return fmt.Errorf("failed to load configuration after installation: %w", err)
			}
		} else {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
	}

	// Register provider implementations
	if err := m.registerProviders(); err != nil {
		return fmt.Errorf("failed to register provider implementations: %w", err)
	}

	m.loaded = true
	m.logger.Info("Registry initialized successfully")
	return nil
}

// GetRegistry returns the underlying registry.
func (m *Manager) GetRegistry() *Registry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.registry
}

// registerProviders registers the provider implementations with the registry.
func (m *Manager) registerProviders() error {
	// Register Gemini provider implementation
	geminiProvider := gemini.NewProvider(m.logger)
	if err := m.registry.RegisterProviderImplementation("gemini", geminiProvider); err != nil {
		return fmt.Errorf("failed to register Gemini provider: %w", err)
	}
	m.logger.Debug("Registered Gemini provider implementation")

	// Register OpenAI provider implementation
	openaiProvider := openai.NewProvider(m.logger)
	if err := m.registry.RegisterProviderImplementation("openai", openaiProvider); err != nil {
		return fmt.Errorf("failed to register OpenAI provider: %w", err)
	}
	m.logger.Debug("Registered OpenAI provider implementation")

	return nil
}

// installDefaultConfig creates the config directory and copies the default models.yaml file.
func (m *Manager) installDefaultConfig() error {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Create config directory
	configDir := filepath.Join(homeDir, ConfigDirName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Path to the default models.yaml in the repository
	defaultConfigPath := filepath.Join("config", ModelsConfigFileName)

	// Make sure the default config exists
	if _, err := os.Stat(defaultConfigPath); os.IsNotExist(err) {
		// Try with full path
		defaultConfigPath = filepath.Join("/Users/phaedrus/Development/architect", "config", ModelsConfigFileName)
		if _, err := os.Stat(defaultConfigPath); os.IsNotExist(err) {
			return fmt.Errorf("default configuration file not found at %s: %w", defaultConfigPath, err)
		}
	}

	// Read the default configuration
	defaultConfig, err := os.ReadFile(defaultConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read default configuration file: %w", err)
	}

	// Write to the user's config directory
	targetConfigPath := filepath.Join(configDir, ModelsConfigFileName)
	if err := os.WriteFile(targetConfigPath, defaultConfig, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	m.logger.Info("Default configuration installed to %s", targetConfigPath)
	return nil
}
