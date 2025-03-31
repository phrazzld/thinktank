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

## List Models Feature

Thinktank allows you to list all available models from configured providers, making it easy to discover which models you can use in your queries.

### Usage

```bash
# List all available models from all providers
thinktank list-models

# List models from a specific provider
thinktank list-models -p anthropic

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
  - gpt-4o
  - gpt-4-turbo (Legacy GPT-4 Turbo)
  - gpt-3.5-turbo

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

Thinktank automatically saves individual model responses to separate files in a dedicated output directory.

### Usage

```bash
# Run with default output directory (./thinktank-reports/)
thinktank -i prompt.txt

# Specify a custom output directory
thinktank -i prompt.txt -o ./custom-outputs
```

### How It Works

1. Thinktank always creates a timestamped directory for each run (e.g., `./thinktank-reports/thinktank_run_20250331_123456_789/` by default).
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