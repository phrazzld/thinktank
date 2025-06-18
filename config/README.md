# thinktank Configuration Files

This directory contains documentation for thinktank configuration.

## Model Configuration

As of the latest version, thinktank uses hardcoded model definitions in the `internal/models` package. Models are no longer configured via external YAML files, simplifying deployment and ensuring consistency.

## Supported Models

The following models are built into thinktank:

### OpenAI Models
- `gpt-4.1` - Latest GPT-4 model with 1M token context window
- `o4-mini` - Optimized OpenAI model with reasoning capabilities

### Gemini Models
- `gemini-2.5-pro` - Google's advanced model with 1M token context
- `gemini-2.5-flash` - Faster Gemini variant with 1M token context

### OpenRouter Models
- `openrouter/deepseek/deepseek-chat-v3-0324` - DeepSeek chat model
- `openrouter/deepseek/deepseek-r1` - DeepSeek reasoning model
- `openrouter/x-ai/grok-3-beta` - xAI's Grok model

## Setup

No configuration files need to be installed. Simply set the required API keys as environment variables:

```bash
export OPENAI_API_KEY="your-openai-api-key"
export GEMINI_API_KEY="your-gemini-api-key"
export OPENROUTER_API_KEY="your-openrouter-api-key"
```

## Adding New Models

To add new models, modify the `modelDefinitions` map in `internal/models/models.go` and submit a pull request.

## Usage

Use thinktank by specifying a model with the `--model` flag:

```bash
# Using OpenAI models
thinktank --model gpt-4.1 --instructions task.md src/

# Using Gemini models
thinktank --model gemini-2.5-pro --instructions task.md src/

# Using OpenRouter models
thinktank --model openrouter/deepseek/deepseek-r1 --instructions task.md src/
```

All model parameters (context windows, token limits, default settings) are configured automatically. No additional configuration is required.
