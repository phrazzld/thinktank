// Package prompt handles loading and processing prompt templates
package prompt

import (
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/logutil"
)

// ConfigAdapter adapts the config.ManagerInterface to the prompt.ConfigManagerInterface
type ConfigAdapter struct {
	configManager config.ManagerInterface
	logger        logutil.LoggerInterface
}

// NewConfigAdapter creates a new adapter for the config manager
func NewConfigAdapter(configManager config.ManagerInterface, logger logutil.LoggerInterface) *ConfigAdapter {
	return &ConfigAdapter{
		configManager: configManager,
		logger:        logger,
	}
}

// GetTemplatePath implements the prompt.ConfigManagerInterface method by delegating to the config manager
func (a *ConfigAdapter) GetTemplatePath(name string) (string, error) {
	return a.configManager.GetTemplatePath(name)
}

// CreatePromptManager creates a prompt.Manager that uses the configuration system
func CreatePromptManager(configManager config.ManagerInterface, logger logutil.LoggerInterface) *Manager {
	adapter := NewConfigAdapter(configManager, logger)
	return NewManagerWithConfig(logger, adapter)
}
