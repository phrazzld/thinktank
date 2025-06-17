// Package models provides hardcoded model definitions and lookup functions,
// replacing the complex registry system with simple map-based operations.
//
// This package contains metadata for all supported LLM models including their
// providers, API model IDs, token limits, and default parameters. It eliminates
// the need for YAML configuration files and complex initialization logic.
//
// Usage:
//
//	info, err := models.GetModelInfo("gpt-4.1")
//	if err != nil {
//	    // handle unknown model
//	}
//
//	provider, err := models.GetProviderForModel("gemini-2.5-pro")
//	envVar := models.GetAPIKeyEnvVar(provider)
//
// The package supports 7 models across 3 providers:
//   - OpenAI: gpt-4.1, o4-mini
//   - Gemini: gemini-2.5-pro, gemini-2.5-flash
//   - OpenRouter: deepseek-chat-v3-0324, deepseek-r1, grok-3-beta
package models
