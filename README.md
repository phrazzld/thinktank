# Thinktank

A powerful CLI tool for querying multiple Large Language Models (LLMs) with the same prompt and comparing their responses.

## Overview

Thinktank allows you to send the same text prompt to multiple LLM providers (like OpenAI, Anthropic, etc.) simultaneously and view their responses side-by-side. This is useful for:

- Comparing how different models interpret and respond to the same prompt
- Finding the best model for specific types of queries
- Testing prompts across different providers before committing to one
- Educational purposes to understand model differences

Built with TypeScript and designed with extensibility in mind, Thinktank follows Atomic Design principles to provide a robust, maintainable architecture that makes it easy to add new LLM providers.

## Installation

### Prerequisites

- Node.js 18.x or higher
- npm or yarn

### Install from source

```bash
# Clone the repository
git clone https://github.com/phrazzld/thinktank.git
cd thinktank

# Install dependencies
npm install

# Build the project
npm run build

# Install globally
npm install -g .
```

### Environment Variables

Create a `.env` file in your project root with API keys for the LLM providers you wish to use:

```
OPENAI_API_KEY=your_openai_api_key_here
ANTHROPIC_API_KEY=your_anthropic_api_key_here
# Add other provider API keys as needed
```

## Usage

### Basic Usage

```bash
# Send a prompt file to all enabled models
thinktank -i prompt.txt

# Send to a specific model
thinktank -i prompt.txt -m openai:gpt-4o

# Use a custom config file
thinktank -i prompt.txt -c custom-config.json

# Write results to a file
thinktank -i prompt.txt -o results.txt

# Include metadata in the output
thinktank -i prompt.txt --metadata

# Disable colored output
thinktank -i prompt.txt --no-color
```

### Command Line Options

| Option | Alias | Description | Type | Required |
|--------|-------|-------------|------|----------|
| `--input` | `-i` | Path to input prompt file | string | Yes |
| `--config` | `-c` | Path to configuration file | string | No |
| `--output` | `-o` | Path to output file | string | No |
| `--model` | `-m` | Models to use (provider:model, provider, or model) | array | No |
| `--metadata` | | Include metadata in output | boolean | No |
| `--no-color` | | Disable colored output | boolean | No |
| `--help` | `-h` | Show help | | |
| `--version` | `-v` | Show version number | | |

## Configuration

Thinktank uses a JSON configuration file to define which LLM providers and models to use.

### Default Configuration

By default, Thinktank will look for a `thinktank.config.json` file in the current directory. If not found, it will use a default configuration with common models (disabled by default).

### Configuration File Format

```json
{
  "models": [
    {
      "provider": "openai",
      "modelId": "gpt-4o",
      "enabled": true,
      "options": {
        "temperature": 0.7,
        "maxTokens": 1000
      }
    },
    {
      "provider": "anthropic",
      "modelId": "claude-3-opus-20240229",
      "enabled": false,
      "options": {
        "temperature": 0.7,
        "maxTokens": 1000
      }
    }
  ]
}
```

### Configuration Options

#### Model Configuration

Each model in the `models` array can have the following properties:

| Property | Type | Description | Required |
|----------|------|-------------|----------|
| `provider` | string | The provider ID (e.g., "openai", "anthropic") | Yes |
| `modelId` | string | The specific model ID (e.g., "gpt-4o", "claude-3-opus-20240229") | Yes |
| `enabled` | boolean | Whether this model is enabled by default | Yes |
| `apiKeyEnvVar` | string | Custom environment variable name for the API key | No |
| `options` | object | Provider-specific options (see below) | No |

#### Common Options

The `options` object can include:

| Option | Type | Description |
|--------|------|-------------|
| `temperature` | number | Controls randomness (0-1, lower is more deterministic) |
| `maxTokens` | number | Maximum number of tokens to generate |

Provider-specific options can also be included and will be passed directly to the underlying API.

## Architecture

Thinktank follows the Atomic Design methodology with a clear separation of concerns:

```
src/
├── atoms/       # Core types, constants, and helpers
├── molecules/   # Basic functionality (file reading, providers)
├── organisms/   # Complex components (config, registry)
├── templates/   # Main workflow orchestration
└── runtime/     # CLI entry point
```

### Key Components

- **ConfigManager**: Handles loading and validating configuration
- **LLMRegistry**: Manages provider registration and retrieval
- **LLMProviders**: Implementation of various LLM APIs
- **RunThinktank**: Orchestrates the main workflow
- **CLI**: Provides the command-line interface

## Extending Thinktank

### Adding a New LLM Provider

1. Create a new file in `src/molecules/llmProviders/<provider-name>.ts`
2. Implement the `LLMProvider` interface
3. Register the provider in the LLM registry

Here's an example implementation for a new provider:

```typescript
import { LLMProvider, LLMResponse, ModelOptions } from '../../atoms/types';
import { registerProvider } from '../../organisms/llmRegistry';

export class NewProvider implements LLMProvider {
  public readonly providerId = 'new-provider';
  
  constructor(private readonly apiKey?: string) {
    // Auto-register this provider
    try {
      registerProvider(this);
    } catch (error) {
      // Ignore if already registered
      if (!(error instanceof Error && error.message.includes('already registered'))) {
        throw error;
      }
    }
  }
  
  public async generate(
    prompt: string,
    modelId: string,
    options?: ModelOptions
  ): Promise<LLMResponse> {
    try {
      // Implementation to call the provider's API
      // ...
      
      return {
        provider: this.providerId,
        modelId,
        text: 'Response text',
        metadata: {
          // Any provider-specific metadata
        },
      };
    } catch (error) {
      if (error instanceof Error) {
        throw new Error(`${this.providerId} API error: ${error.message}`);
      }
      throw new Error(`Unknown error occurred with ${this.providerId}`);
    }
  }
}

// Export a default instance
export const newProvider = new NewProvider();
```

4. Import and use the provider in `src/templates/runThinktank.ts`:

```typescript
// Import provider modules to ensure they're registered
import '../molecules/llmProviders/openai';
import '../molecules/llmProviders/new-provider';
// Future providers will be imported here
```

5. Add an example configuration in `templates/thinktank.config.default.json`

## Troubleshooting

### Common Issues

1. **"Input file not found"**
   - Ensure the file path is correct and the file exists
   - Try using an absolute path

2. **"No enabled models found in configuration"**
   - Check your config file to ensure at least one model has `"enabled": true`
   - Alternatively, specify models directly with the `-m` option

3. **"Missing API keys for models"**
   - Ensure you have set the correct environment variables in your `.env` file
   - Check that the API keys are valid

4. **"Provider not found"**
   - Make sure the provider ID in your config matches the registered provider
   - Check that the provider module is properly imported in `runThinktank.ts`

### API Key Issues

If you're having issues with API keys:

1. Confirm your API keys are correctly set in the `.env` file
2. Verify that the environment variables match what the providers expect:
   - OpenAI: `OPENAI_API_KEY`
   - Anthropic: `ANTHROPIC_API_KEY`
3. You can override the environment variable name in the config using `apiKeyEnvVar`

## License

[MIT](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request