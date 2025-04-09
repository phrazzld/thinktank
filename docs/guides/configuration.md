# Configuration Guide

thinktank uses a JSON configuration file to define which LLM providers and models to use. This guide explains how to configure and customize thinktank for your needs.

## Configuration Location

The configuration file is stored following the XDG Base Directory Specification:

- **Windows**: `%APPDATA%\thinktank\config.json`
- **macOS**: `~/.config/thinktank/config.json`
- **Linux**: `~/.config/thinktank/config.json`

View your configuration location:
```bash
thinktank config path
```

## Configuration Format

```json
{
  "defaultGroup": "general",
  "groups": {
    "general": {
      "name": "general",
      "models": [
        "openai:gpt-4o",
        "anthropic:claude-3-opus-20240229"
      ]
    },
    "coding": {
      "name": "coding",
      "models": [
        "openai:gpt-4o"
      ]
    }
  },
  "models": [
    {
      "provider": "openai",
      "modelId": "gpt-4o",
      "enabled": true,
      "options": {
        "temperature": 0.7,
        "maxTokens": 2000
      }
    },
    {
      "provider": "anthropic",
      "modelId": "claude-3-opus-20240229",
      "enabled": true,
      "options": {
        "temperature": 0.7,
        "maxTokens": 4000
      }
    }
  ]
}
```

## Managing Configuration

### View Configuration
```bash
thinktank config show
```

### Using a Custom Config
```bash
thinktank run prompt.txt --config /path/to/custom-config.json
```

### Configuring via CLI

Add a model:
```bash
thinktank config models add openai gpt-4o --options '{"temperature": 0.7}'
```

Create a group:
```bash
thinktank config groups create coding --models openai:gpt-4o,anthropic:claude-3-opus-20240229
```

Enable/disable models:
```bash
thinktank config models enable openai:gpt-4o
thinktank config models disable anthropic:claude-3-haiku
```

## Cascading Configuration System

thinktank resolves model options through a hierarchy:

1. Base defaults
2. Provider defaults
3. Model-specific defaults
4. User config options
5. Group-specific options
6. CLI options

## Provider-Specific Configuration

### OpenAI
```json
{
  "temperature": 0.7,
  "maxTokens": 4000,
  "top_p": 0.95,
  "presence_penalty": 0.0,
  "frequency_penalty": 0.0
}
```

### Anthropic
```json
{
  "temperature": 0.7,
  "maxTokens": 4000,
  "thinking": {
    "type": "enabled",
    "budget_tokens": 10000
  }
}
```

### Google Gemini
```json
{
  "temperature": 0.7,
  "maxTokens": 2048,
  "topK": 40,
  "topP": 0.95
}
```

### OpenRouter
```json
{
  "provider": "openrouter",
  "modelId": "anthropic/claude-3-opus-20240229",
  "enabled": true,
  "apiKeyEnvVar": "OPENROUTER_API_KEY",
  "options": {
    "temperature": 0.7,
    "maxTokens": 4000
  }
}
```

See [the README](../../README.md) for basic usage information.