// Package prompt handles loading and processing prompt templates
package prompt

import (
	"fmt"

	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/logutil"
)

// SetupPromptManagerWithConfig initializes a prompt manager with the application's configuration
func SetupPromptManagerWithConfig(logger logutil.LoggerInterface, configManager config.ManagerInterface) (ManagerInterface, error) {
	// Create a prompt manager that uses the configuration system
	promptManager := CreatePromptManager(configManager, logger)

	// Pre-load the default templates
	for _, tmplName := range []string{"default.tmpl", "clarify.tmpl", "refine.tmpl"} {
		err := promptManager.LoadTemplate(tmplName)
		if err != nil {
			return nil, fmt.Errorf("failed to load template %s: %w", tmplName, err)
		}
	}

	return promptManager, nil
}
