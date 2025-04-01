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

thinktank uses a simple, intuitive command line interface with positional arguments:

```bash
# Send a prompt to all models in the default group
thinktank prompt.txt

# Send a prompt to models in a specific group
thinktank prompt.txt coding

# Send a prompt to one specific model
thinktank prompt.txt openai:gpt-4o

# List all available models
thinktank models
```

### Command Examples

#### Running with a Prompt File

Send your prompt to all enabled models in the default group:

```bash
thinktank path/to/prompt.txt
```

#### Running with a Specific Group

Send your prompt to all models in a named group (as defined in your config file):

```bash
thinktank path/to/prompt.txt coding
thinktank path/to/prompt.txt creative
thinktank path/to/prompt.txt analysis
```

#### Running with a Specific Model

Send your prompt to a single model using the provider:model format:

```bash
thinktank path/to/prompt.txt openai:gpt-4o
thinktank path/to/prompt.txt anthropic:claude-3-opus-20240229
thinktank path/to/prompt.txt openrouter:anthropic/claude-3-5-sonnet-20240620
```

#### Listing Available Models

List all available models from all configured providers:

```bash
thinktank models
```

List models from a specific provider:

```bash
thinktank models openai
thinktank models anthropic
thinktank models openrouter
```

### Additional Options

| Option | Description |
|--------|-------------|
| `--help`, `-h` | Show help information |
| `--version`, `-v` | Show version number |
| `--no-color` | Disable colored output |

## Configuration

thinktank uses a JSON configuration file to define which LLM providers and models to use.

### Default Configuration

By default, thinktank will look for a `thinktank.config.json` file in the current directory. If not found, it will use a default configuration with common models (disabled by default).

### Configuration File Format

```json
{
  "defaultGroup": "general",
  "groups": {
    "general": [
      "openai:gpt-4o",
      "anthropic:claude-3-opus-20240229"
    ],
    "coding": [
      "openai:gpt-4o",
      "anthropic:claude-3-sonnet-20240229",
      "gemini:gemini-1.5-pro"
    ],
    "creative": [
      "anthropic:claude-3-opus-20240229",
      "openai:gpt-4-turbo"
    ]
  },
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
      "enabled": true,
      "options": {
        "temperature": 0.7,
        "maxTokens": 1000
      }
    },
    {
      "provider": "anthropic",
      "modelId": "claude-3-sonnet-20240229", 
      "enabled": true,
      "options": {
        "temperature": 0.7,
        "maxTokens": 1000
      }
    },
    {
      "provider": "gemini",
      "modelId": "gemini-1.5-pro", 
      "enabled": true,
      "options": {
        "temperature": 0.7,
        "maxTokens": 1000
      }
    }
  ]
}
```

### Configuration Options

#### Group Configuration

The `groups` object allows you to define named collections of models that can be activated with a single command:

| Property | Type | Description | Required |
|----------|------|-------------|----------|
| `defaultGroup` | string | The name of the default group to use when no group is specified | Yes |
| `groups` | object | An object where keys are group names and values are arrays of model identifiers | Yes |

Each group is an array of model identifiers in the format `provider:modelId`. When you run thinktank with a group name, all models in that group will be used for the query.

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
thinktank models openrouter
```

### Configuration Examples

Below are example configurations for different use cases to help you get started.

#### Minimal Configuration

This is a minimal configuration with a default group and two models:

```json
{
  "defaultGroup": "basic",
  "groups": {
    "basic": [
      "openai:gpt-3.5-turbo",
      "anthropic:claude-3-haiku-20240307"
    ]
  },
  "models": [
    {
      "provider": "openai",
      "modelId": "gpt-3.5-turbo",
      "enabled": true
    },
    {
      "provider": "anthropic",
      "modelId": "claude-3-haiku-20240307",
      "enabled": true
    }
  ]
}
```

#### Development-Focused Configuration

This configuration is optimized for software development tasks:

```json
{
  "defaultGroup": "coding",
  "groups": {
    "coding": [
      "openai:gpt-4o",
      "anthropic:claude-3-opus-20240229"
    ],
    "review": [
      "openai:gpt-4o",
      "anthropic:claude-3-sonnet-20240229"
    ],
    "quick": [
      "openai:gpt-3.5-turbo"
    ]
  },
  "models": [
    {
      "provider": "openai",
      "modelId": "gpt-4o",
      "enabled": true,
      "options": {
        "temperature": 0.1,
        "maxTokens": 8000
      }
    },
    {
      "provider": "openai",
      "modelId": "gpt-3.5-turbo",
      "enabled": true,
      "options": {
        "temperature": 0.2,
        "maxTokens": 2000
      }
    },
    {
      "provider": "anthropic",
      "modelId": "claude-3-opus-20240229",
      "enabled": true,
      "options": {
        "temperature": 0.1,
        "maxTokens": 8000
      }
    },
    {
      "provider": "anthropic",
      "modelId": "claude-3-sonnet-20240229",
      "enabled": true,
      "options": {
        "temperature": 0.2,
        "maxTokens": 4000
      }
    }
  ]
}
```

#### Creative Writing Configuration

This configuration is tailored for creative writing tasks:

```json
{
  "defaultGroup": "creative",
  "groups": {
    "creative": [
      "anthropic:claude-3-opus-20240229",
      "openai:gpt-4o"
    ],
    "brainstorm": [
      "anthropic:claude-3-opus-20240229",
      "openai:gpt-4o",
      "openai:gpt-3.5-turbo"
    ],
    "draft": [
      "anthropic:claude-3-opus-20240229"
    ],
    "edit": [
      "openai:gpt-4o"
    ]
  },
  "models": [
    {
      "provider": "openai",
      "modelId": "gpt-4o",
      "enabled": true,
      "options": {
        "temperature": 0.8,
        "maxTokens": 4000,
        "top_p": 0.9
      }
    },
    {
      "provider": "openai",
      "modelId": "gpt-3.5-turbo",
      "enabled": true,
      "options": {
        "temperature": 0.9,
        "maxTokens": 2000,
        "top_p": 0.95
      }
    },
    {
      "provider": "anthropic",
      "modelId": "claude-3-opus-20240229",
      "enabled": true,
      "options": {
        "temperature": 0.8,
        "maxTokens": 4000
      }
    }
  ]
}
```

#### Mixed-Provider Configuration with OpenRouter

This configuration uses OpenRouter to access multiple providers through a single API:

```json
{
  "defaultGroup": "mixed",
  "groups": {
    "mixed": [
      "openrouter:anthropic/claude-3-opus-20240229",
      "openrouter:openai/gpt-4o",
      "openrouter:google/gemini-pro"
    ],
    "anthropic-only": [
      "openrouter:anthropic/claude-3-opus-20240229",
      "openrouter:anthropic/claude-3-sonnet-20240229"
    ],
    "openai-only": [
      "openrouter:openai/gpt-4o",
      "openrouter:openai/gpt-3.5-turbo"
    ]
  },
  "models": [
    {
      "provider": "openrouter",
      "modelId": "anthropic/claude-3-opus-20240229",
      "enabled": true,
      "apiKeyEnvVar": "OPENROUTER_API_KEY",
      "options": {
        "temperature": 0.7,
        "maxTokens": 4000
      }
    },
    {
      "provider": "openrouter",
      "modelId": "anthropic/claude-3-sonnet-20240229",
      "enabled": true,
      "apiKeyEnvVar": "OPENROUTER_API_KEY",
      "options": {
        "temperature": 0.7,
        "maxTokens": 4000
      }
    },
    {
      "provider": "openrouter",
      "modelId": "openai/gpt-4o",
      "enabled": true,
      "apiKeyEnvVar": "OPENROUTER_API_KEY",
      "options": {
        "temperature": 0.7,
        "maxTokens": 4000
      }
    },
    {
      "provider": "openrouter",
      "modelId": "openai/gpt-3.5-turbo",
      "enabled": true,
      "apiKeyEnvVar": "OPENROUTER_API_KEY",
      "options": {
        "temperature": 0.7,
        "maxTokens": 2000
      }
    },
    {
      "provider": "openrouter",
      "modelId": "google/gemini-pro",
      "enabled": true,
      "apiKeyEnvVar": "OPENROUTER_API_KEY",
      "options": {
        "temperature": 0.7,
        "maxTokens": 2000
      }
    }
  ]
}
```

## Output Directory

thinktank automatically saves individual model responses to separate files in a dedicated output directory.

### How It Works

1. For each run, thinktank creates a directory under `./thinktank-reports/` based on the group or model name.
2. Each model's response is saved as a separate Markdown (.md) file within this directory, with filenames based on the provider and model ID (e.g., `openai-gpt-4o.md`).
3. Console output will show progress information and the location of the output directory.

### Output File Format

Each output file is formatted as Markdown with the following sections:

```markdown
# provider:model-id

Generated: YYYY-MM-DDThh:mm:ss.sssZ

## Response

The text response from the model goes here...

## Metadata

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
thinktank models openrouter

# Query a specific model via OpenRouter
thinktank prompt.txt openrouter:anthropic/claude-3-opus-20240229

# Include OpenRouter models in a group config
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
   - Check if you're in the right working directory
   - Try using an absolute path instead of a relative one

2. **"No models found in group"**
   - Verify that the group name exists in your configuration
   - Check that models in the group are properly configured and enabled
   - Try using a specific model identifier instead

3. **"Invalid model format"**
   - Ensure you're using the correct format: `provider:modelId`
   - For OpenRouter models, use: `openrouter:provider/modelId`
   - Check for typos in provider or model names

4. **"Missing API keys for models"**
   - Ensure you have set the correct environment variables in your `.env` file
   - Check that the API keys are valid
   - Follow the provider-specific instructions for obtaining API keys

5. **"Failed to create output directory"**
   - Check if you have write permissions for the specified output directory
   - Ensure the path exists or can be created

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