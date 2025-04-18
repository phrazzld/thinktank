# Architect Configuration Files

This directory contains default configuration files for the Architect tool.

## Models Configuration

The `models.yaml` file defines the LLM providers and models available to the Architect tool. It includes:

- API key environment variable mappings
- Provider definitions (OpenAI, Gemini, OpenRouter)
- Model configurations with context sizes and output limits
- Default parameter settings for each model

## Installation

You can install the configuration files using the provided installation script:

```bash
./config/install.sh
```

This will create the necessary directories and copy the configuration files to `~/.config/architect/`.

## Manual Installation

If you prefer to install manually:

1. Create the configuration directory:
   ```bash
   mkdir -p ~/.config/architect
   ```

2. Copy the models.yaml file:
   ```bash
   cp config/models.yaml ~/.config/architect/models.yaml
   ```

3. Set your API keys as environment variables:
   ```bash
   export OPENAI_API_KEY="your-openai-api-key"
   export GEMINI_API_KEY="your-gemini-api-key"
   export OPENROUTER_API_KEY="your-openrouter-api-key"
   ```

## Customization

You can customize the `models.yaml` file to:

- Add new models as they become available
- Adjust token limits to match model updates
- Configure default parameters for each model
- Add custom API endpoints (for self-hosted models or proxies)

After modifying the configuration, restart Architect for the changes to take effect.
