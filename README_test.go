package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestREADME_NoObsoleteProviderReferences ensures README reflects OpenRouter consolidation
func TestREADME_NoObsoleteProviderReferences(t *testing.T) {
	content, err := os.ReadFile("README.md")
	assert.NoError(t, err, "README.md should be readable")

	readmeText := string(content)

	// Should not contain references to multiple providers
	obsoleteReferences := []string{
		"using Gemini, OpenAI, or OpenRouter models",
		"Multiple LLM Models",
	}

	for _, reference := range obsoleteReferences {
		assert.NotContains(t, readmeText, reference,
			"README should not contain obsolete provider reference: %s", reference)
	}

	// Should contain OpenRouter-focused language
	shouldContain := []string{
		"OPENROUTER_API_KEY",
		"through OpenRouter's unified API",
		"https://openrouter.ai/keys",
	}

	for _, content := range shouldContain {
		assert.Contains(t, readmeText, content,
			"README should contain OpenRouter reference: %s", content)
	}
}

// TestREADME_QuickStartIsCorrect ensures Quick Start section shows proper setup
func TestREADME_QuickStartIsCorrect(t *testing.T) {
	content, err := os.ReadFile("README.md")
	assert.NoError(t, err, "README.md should be readable")

	readmeText := string(content)

	// Quick Start should be focused on single API key
	assert.Contains(t, readmeText, "export OPENROUTER_API_KEY=",
		"Quick Start should show OPENROUTER_API_KEY setup")
	assert.Contains(t, readmeText, "# Set API key (all models now use OpenRouter)",
		"Quick Start should explain unified API key approach")

	// Should not reference old providers in setup
	assert.NotContains(t, readmeText, "export OPENAI_API_KEY",
		"Quick Start should not reference OPENAI_API_KEY")
	assert.NotContains(t, readmeText, "export GEMINI_API_KEY",
		"Quick Start should not reference GEMINI_API_KEY")
}
