// Package models provides hardcoded model definitions and lookup functions,
// replacing the complex registry system with simple map-based operations.
//
// This package contains metadata for all supported LLM models including their
// providers, API model IDs, token limits, and default parameters. It eliminates
// the need for YAML configuration files and complex initialization logic.
//
// Usage:
//
//	info, err := models.GetModelInfo("gpt-5.2")
//	if err != nil {
//	    // handle unknown model
//	}
//
//	provider, err := models.GetProviderForModel("gemini-3-flash")
//	envVar := models.GetAPIKeyEnvVar(provider)
//
// The package supports multiple OpenRouter-hosted production models plus test-only models.
// See modelDefinitions in models.go for the current catalog.
package models
