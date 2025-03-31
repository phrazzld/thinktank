# thinktank

A powerful CLI tool for querying multiple Large Language Models (LLMs) with the same prompt and comparing their responses.

## Overview

thinktank allows you to send the same text prompt to multiple LLM providers (like OpenAI, Anthropic, etc.) simultaneously and view their responses side-by-side. This is useful for:

- Comparing how different models interpret and respond to the same prompt
- Finding the best model for specific types of queries
- Testing prompts across different providers before committing to one
- Educational purposes to understand model differences

Built with TypeScript and designed with extensibility in mind, thinktank follows Atomic Design principles to provide a robust, maintainable architecture that makes it easy to add new LLM providers.

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
GEMINI_API_KEY=your_gemini_api_key_here
OPENROUTER_API_KEY=your_openrouter_api_key_here
# Add other provider API keys as needed
```

## Usage

### Basic Usage

```bash
# Send a prompt file to all enabled models (saves responses to ./thinktank_outputs/)
thinktank -i prompt.txt

# Send to a specific model
thinktank -i prompt.txt -m openai:gpt-4o

# Use a custom config file
thinktank -i prompt.txt -c custom-config.json

# Specify a custom output directory
thinktank -i prompt.txt -o ./custom-outputs

# Include metadata in the output
thinktank -i prompt.txt --metadata

# Disable colored output
thinktank -i prompt.txt --no-color

# List all available models from configured providers
thinktank list-models

# List models from a specific provider
thinktank list-models -p anthropic
```

### Command Line Options

#### Default Command Options (for prompt querying)

| Option | Alias | Description | Type | Required |
|--------|-------|-------------|------|----------|
| `--input` | `-i` | Path to input prompt file | string | Yes |
| `--config` | `-c` | Path to configuration file | string | No |
| `--output` | `-o` | Path to custom output directory (default: './thinktank-reports/') | string | No |
| `--model` | `-m` | Models to use (provider:model, provider, or model) | array | No |
| `--metadata` | | Include metadata in output | boolean | No |
| `--no-color` | | Disable colored output | boolean | No |
| `--help` | `-h` | Show help | | |
| `--version` | `-v` | Show version number | | |

#### List-Models Command Options

| Option | Alias | Description | Type | Required |
|--------|-------|-------------|------|----------|
| `--provider` | `-p` | Filter models by provider ID | string | No |
| `--config` | `-c` | Path to configuration file | string | No |
| `--help` | `-h` | Show help | | |

## Configuration

thinktank uses a JSON configuration file to define which LLM providers and models to use.

### Default Configuration

By default, thinktank will look for a `thinktank.config.json` file in the current directory. If not found, it will use a default configuration with common models (disabled by default).

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

#### Provider-Specific Configuration

##### OpenRouter Configuration

OpenRouter provides access to models from multiple providers through a unified API. To use OpenRouter:

1. Sign up at [openrouter.ai](https://openrouter.ai) and get your API key
2. Set the `OPENROUTER_API_KEY` environment variable or use the `apiKeyEnvVar` property in your config
3. Configure models in your config using the format `provider/model-id` for the `modelId` property

Example configuration:

```json
{
  "models": [
    {
      "provider": "openrouter",
      "modelId": "openai/gpt-4o",
      "enabled": true,
      "options": {
        "temperature": 0.7,
        "maxTokens": 1000
      }
    },
    {
      "provider": "openrouter",
      "modelId": "anthropic/claude-3-opus-20240229",
      "enabled": true,
      "options": {
        "temperature": 0.7,
        "maxTokens": 2000
      }
    }
  ]
}
```

Additional options supported by OpenRouter:
- `top_p`: Controls diversity of generated output (0-1)
- `top_k`: Limits tokens considered to the top k most likely
- `stop`: Array of sequences that stop generation when encountered
- `presence_penalty`: Decreases repetition of tokens (>0 = less repetition)
- `frequency_penalty`: Decreases repetition of phrases (>0 = less repetition)

To see the full list of available models, run:
```bash
thinktank list-models -p openrouter
```

## List Models Feature

thinktank allows you to list all available models from configured providers, making it easy to discover which models you can use in your queries.

### Usage

```bash
# List all available models from all providers
thinktank list-models

# List models from a specific provider
thinktank list-models -p anthropic
thinktank list-models -p openrouter

# List models using a custom config file
thinktank list-models -c custom-config.json
```

### How It Works

1. The command queries each provider in your configuration for their available models
2. For providers that support the `listModels` API (like Anthropic and OpenAI), it retrieves the list of models dynamically
3. Results are displayed grouped by provider, with descriptions when available
4. Errors (like missing API keys) are clearly shown for troubleshooting

### Example Output

```
Available Models:

--- openai ---
  - gpt-4o (Owned by: system)
  - gpt-4-turbo (Owned by: system)
  - gpt-3.5-turbo (Owned by: openai)

--- anthropic ---
  - claude-3-opus-20240229 (Most powerful model)
  - claude-3-sonnet-20240229 (Balanced model)
  - claude-3-haiku-20240307 (Fast, efficient model)
  - claude-3-5-sonnet-20240620 (Latest Sonnet model)
  - claude-3-7-sonnet-20250219 (Latest advanced model)

--- legacy-provider ---
  Error fetching models: Provider 'legacy-provider' does not support listing models
```

## Output Directory Feature

thinktank automatically saves individual model responses to separate files in a dedicated output directory.

### Usage

```bash
# Run with default output directory (./thinktank-reports/)
thinktank -i prompt.txt

# Specify a custom output directory
thinktank -i prompt.txt -o ./custom-outputs
```

### How It Works

1. thinktank always creates a timestamped directory for each run (e.g., `./thinktank-reports/thinktank_run_20250331_123456_789/` by default).
2. The `-o/--output` option allows you to specify a custom base directory instead of the default `./thinktank-reports/`.
3. Each model's response is saved as a separate Markdown (.md) file within this directory, with filenames based on the provider and model ID (e.g., `openai-gpt-4o.md`).
4. Console output will show progress information and the location of the output directory.

### Output File Format

Each output file is formatted as Markdown with the following sections:

```markdown
# provider:model-id

Generated: YYYY-MM-DDThh:mm:ss.sssZ

## Response

The text response from the model goes here...

## Metadata (Optional, if --metadata is specified)

```json
{
  "usage": { "total_tokens": 123 },
  // Other provider-specific metadata
}
```
```

### Error Handling

If a model returns an error, the output file will include the error details:

```markdown
# provider:model-id

Generated: YYYY-MM-DDThh:mm:ss.sssZ

## Error

```
Error message from the API
```
```

## Architecture

thinktank follows the Atomic Design methodology with a clear separation of concerns:

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
- **Runthinktank**: Orchestrates the main workflow
- **CLI**: Provides the command-line interface

## Extending thinktank

### Adding a New LLM Provider

1. Create a new file in `src/molecules/llmProviders/<provider-name>.ts`
2. Implement the `LLMProvider` interface
3. Register the provider in the LLM registry

Here's an example implementation for a new provider:

```typescript
import { LLMProvider, LLMResponse, ModelOptions, LLMAvailableModel } from '../../atoms/types';
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
  
  // Optional method to list available models from this provider
  public async listModels(apiKey: string): Promise<LLMAvailableModel[]> {
    try {
      // Implementation to fetch available models from the provider's API
      // ...
      
      return [
        { id: 'model-1', description: 'First model' },
        { id: 'model-2', description: 'Second model' },
      ];
    } catch (error) {
      if (error instanceof Error) {
        throw new Error(`${this.providerId} API error when listing models: ${error.message}`);
      }
      throw new Error(`Unknown error occurred when listing ${this.providerId} models`);
    }
  }
}

// Export a default instance
export const newProvider = new NewProvider();
```

4. Import and use the provider in `src/templates/runthinktank.ts`:

```typescript
// Import provider modules to ensure they're registered
import '../molecules/llmProviders/openai';
import '../molecules/llmProviders/new-provider';
// Future providers will be imported here
```

5. Add an example configuration in `templates/thinktank.config.default.json`

### Example: OpenRouter Provider Implementation

OpenRouter is included as an example implementation that demonstrates how to add a new provider. It leverages the OpenAI SDK since OpenRouter's API is compatible with OpenAI's interface but gives access to many different models from various providers.

#### Key Features of the OpenRouter Implementation:

- **API Compatibility**: Uses the OpenAI SDK with a custom base URL to communicate with OpenRouter
- **Model Access**: Provides access to models from OpenAI, Anthropic, Google, and many other providers through a single API
- **Custom Headers**: Includes required HTTP headers for OpenRouter API attribution
- **Detailed Model Information**: When listing models, includes context length and pricing information when available

#### Configuration Example:

```json
{
  "provider": "openrouter",
  "modelId": "openai/gpt-4o",
  "enabled": true,
  "apiKeyEnvVar": "OPENROUTER_API_KEY",
  "options": {
    "temperature": 0.7,
    "maxTokens": 1000
  }
}
```

#### Usage:

```bash
# List all models available through OpenRouter
thinktank list-models -p openrouter

# Query a specific model via OpenRouter
thinktank -i prompt.txt -m openrouter:anthropic/claude-3-opus-20240229

# Query multiple OpenRouter models
thinktank -i prompt.txt -m openrouter:openai/gpt-4o -m openrouter:anthropic/claude-3-haiku
```

#### OpenRouter Model IDs:

OpenRouter uses a `provider/model-id` format for model names, for example:
- `openai/gpt-4o` (OpenAI's GPT-4o model)
- `anthropic/claude-3-opus-20240229` (Anthropic's Claude 3 Opus)
- `google/gemini-pro` (Google's Gemini Pro)

When configuring models in your config file or when using the `-m` option, use this format to specify which model you want to use through OpenRouter.

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
   - Check that the provider module is properly imported in `runthinktank.ts`

5. **"Failed to create output directory"**
   - Check if you have write permissions for the specified output directory
   - Ensure the path exists or can be created
   - Try providing an absolute path instead of a relative one

### API Key Issues

If you're having issues with API keys:

1. Confirm your API keys are correctly set in the `.env` file
2. Verify that the environment variables match what the providers expect:
   - OpenAI: `OPENAI_API_KEY`
   - Anthropic: `ANTHROPIC_API_KEY`
   - Google: `GEMINI_API_KEY`
   - OpenRouter: `OPENROUTER_API_KEY`
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