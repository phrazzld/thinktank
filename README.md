# thinktank

A powerful CLI tool for querying multiple Large Language Models (LLMs) with the same prompt and comparing their responses.

> 📚 **Documentation**: The repository contains comprehensive documentation including the [testing philosophy](TESTING_PHILOSOPHY.md), [context formatting](#including-context-files-and-directories), [GitHub Actions CI/CD](GITHUB_ACTIONS.md), and [project planning](docs/planning/master-plan.md).

## Overview

thinktank allows you to send the same text prompt to multiple LLM providers (like OpenAI, Anthropic, etc.) simultaneously and view their responses side-by-side. This is useful for:

- Comparing how different models interpret and respond to the same prompt
- Finding the best model for specific types of queries
- Testing prompts across different providers before committing to one
- Educational purposes to understand model differences

Built with TypeScript and designed with extensibility in mind, thinktank provides a domain-oriented architecture that makes it easy to add new LLM providers and configure your experience via a convenient CLI interface.

## Installation

### Prerequisites

- Node.js 18.x or higher
- pnpm ([installation instructions](https://pnpm.io/installation))

### Install from source

```bash
# Clone the repository
git clone https://github.com/phrazzld/thinktank.git
cd thinktank

# Install dependencies
pnpm install

# Build the project
pnpm run build

# Install globally
pnpm add -g .
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

thinktank uses a powerful command line interface built with commander.js:

```bash
# Send a prompt to all models in the default group
thinktank run prompt.txt

# Send a prompt to models in a specific group
thinktank run prompt.txt --group coding

# Send a prompt to one specific model
thinktank run prompt.txt --model openai:gpt-4o

# Send a prompt to multiple specific models
thinktank run prompt.txt --models openai:gpt-4o,anthropic:claude-3-opus

# Include context files or directories with your prompt
thinktank run prompt.txt context-file.js another-file.md path/to/directory/

# List all available models
thinktank models list

# Configure models and groups
thinktank config models add openai gpt-4o --options '{"temperature": 0.7}' 
thinktank config groups create coding --models openai:gpt-4o,anthropic:claude-3-opus

# Get help
thinktank --help
thinktank run --help
thinktank config --help
```

### Command Examples

#### Running with a Prompt File

Send your prompt to all enabled models in the default group:

```bash
thinktank run path/to/prompt.txt
```

#### Including Context Files and Directories

You can include additional files or directories as context for the prompt:

```bash
# Include a single file as context
thinktank run prompt.txt path/to/context-file.js

# Include multiple files as context
thinktank run prompt.txt file1.js file2.md file3.txt

# Include a directory (recursively reads all files in the directory)
thinktank run prompt.txt path/to/directory/

# Include a mix of files and directories
thinktank run prompt.txt file1.js path/to/directory/ file2.md

# Combine with other options
thinktank run prompt.txt file1.js directory/ --models openai:gpt-4o --verbose
```

When directories are provided, thinktank will:
- Recursively traverse the directory
- Respect `.gitignore` patterns to skip ignored files
- Skip binary files automatically
- Apply size limits to prevent overwhelming the LLM

This feature is particularly useful for:
- Analyzing code with relevant files as context
- Providing reference materials for complex queries
- Including configuration files along with your prompt
- Giving the LLM comprehensive background information

For details on how context files are formatted and presented to LLMs, see the [section above](#including-context-files-and-directories).

#### Running with a Specific Group

Send your prompt to all models in a named group (as defined in your config file):

```bash
thinktank run path/to/prompt.txt --group coding
thinktank run path/to/prompt.txt --group creative
thinktank run path/to/prompt.txt --group analysis
```

#### Running with Specific Models

Send your prompt to a single model:

```bash
thinktank run path/to/prompt.txt --model openai:gpt-4o
```

Send your prompt to multiple models at once:

```bash
thinktank run path/to/prompt.txt --models openai:gpt-4o,anthropic:claude-3-opus-20240229,openrouter:anthropic/claude-3-5-sonnet-20240620
```

#### Listing Available Models

List all available models from all configured providers:

```bash
thinktank models list
```

List models from a specific provider:

```bash
thinktank models list --provider openai
thinktank models list --provider anthropic
thinktank models list --provider openrouter
```

#### Configuring thinktank

Show the current configuration:

```bash
thinktank config show
```

Add a new model to your configuration:

```bash
thinktank config models add openai gpt-4o --options '{"temperature": 0.7, "maxTokens": 4000}'
```

Create a new group of models:

```bash
thinktank config groups create coding --models openai:gpt-4o,anthropic:claude-3-opus
```

Enable or disable a model:

```bash
thinktank config models enable openai:gpt-4o
thinktank config models disable openai:gpt-4o
```

### Additional Options

| Option | Description |
|--------|-------------|
| `--help`, `-h` | Show help information |
| `--version`, `-v` | Show version number |
| `--no-color` | Disable colored output |
| `--thinking` | Enable Claude's thinking capability (for supported models) |
| `--show-thinking` | Display thinking output in the results |

### Fun Run Names

When you run thinktank, it will generate a memorable "adjective-noun" style name for each run (e.g., `clever-meadow`, `swift-stream`) using Google's Gemini API. This makes it easier to identify and reference specific runs in the console output.

To enable this feature:
- Set the `GEMINI_API_KEY` environment variable with your Google API key
- The tool will automatically generate a friendly name for each run
- If the API key is missing or an error occurs, it will fall back to a timestamp-based name

Example output:
```
$ thinktank run prompt.txt
i Run name: swift-meadow
i Output directory: /path/to/thinktank-output/run-20250401-123045 (Run: swift-meadow)
...
+ Run 'swift-meadow' completed. 3 responses saved to /path/to/thinktank-output/run-20250401-123045
```

## Configuration

thinktank uses a JSON configuration file to define which LLM providers and models to use. The configuration file is stored in a standardized location following the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html) for each operating system:

- **Windows**: `%APPDATA%\thinktank\config.json`
  - Example: `C:\Users\Username\AppData\Roaming\thinktank\config.json`
  - Falls back to `%USERPROFILE%\AppData\Roaming\thinktank\config.json` if `APPDATA` is not set

- **macOS**: `~/.config/thinktank/config.json`
  - Example: `/Users/username/.config/thinktank/config.json`
  - Uses `$XDG_CONFIG_HOME/thinktank/config.json` if the `XDG_CONFIG_HOME` environment variable is set

- **Linux**: `~/.config/thinktank/config.json`
  - Example: `/home/username/.config/thinktank/config.json`
  - Uses `$XDG_CONFIG_HOME/thinktank/config.json` if the `XDG_CONFIG_HOME` environment variable is set

You can view your configuration location with:

```bash
thinktank config path
```

This command will show you exactly where your configuration is stored and whether the file already exists.

### Working with Configuration Files

The configuration system automatically creates a default configuration file the first time you run thinktank. You can:

1. **View the current configuration**:
   ```bash
   thinktank config show
   ```

2. **Use a custom configuration file**:
   ```bash
   thinktank run prompt.txt --config /path/to/custom-config.json
   ```

3. **Back up your configuration**:
   ```bash
   # Copy your config to a backup file
   cp "$(thinktank config path)" ~/thinktank-config-backup.json
   ```

4. **Edit the configuration directly**:
   ```bash
   # Open the config in your default editor
   nano "$(thinktank config path)"
   ```

5. **Use the configuration commands**:
   ```bash
   # Add a model
   thinktank config models add openai gpt-4o
   
   # Create a group
   thinktank config groups create coding --models openai:gpt-4o,anthropic:claude-3-opus
   ```

The standardized location makes it easier to find, back up, and manage your configuration across different environments.

### Cascading Configuration System

thinktank implements a powerful cascading configuration system that intelligently resolves model options from multiple sources, ensuring sensible defaults while giving you fine-grained control over model behavior.

#### Configuration Hierarchy

Options for each model are resolved through a six-layer hierarchy, with each subsequent layer overriding settings from previous layers:

1. **Base Defaults** - Global options for all models
2. **Provider Defaults** - Defaults specific to a provider (e.g., Anthropic, OpenAI)
3. **Model-Specific Defaults** - Defaults for a specific model (e.g., Claude 3 Opus, GPT-4o)
4. **User Config Options** - Options set in your `thinktank.config.json` file
5. **Group-Specific Options** - Options set for a model group
6. **CLI Options** - Options provided via command-line flags

This layered approach means you don't need to specify every option for every model. Instead, the system automatically applies appropriate defaults and only overrides what you explicitly specify.

#### Option Resolution Example

Let's trace how the system resolves options for `openai:gpt-4o` when used with the `coding` group:

1. **Start with base defaults**:
   ```json
   {
     "temperature": 0.7,
     "maxTokens": 1000
   }
   ```

2. **Apply provider defaults** (OpenAI):
   ```json
   {
     "temperature": 0.7,
     "maxTokens": 1000
   }
   ```

3. **Apply model-specific defaults** (GPT-4o):
   ```json
   {
     "temperature": 0.7,
     "maxTokens": 4000  // Updated from 1000
   }
   ```

4. **Apply user config options**:
   ```json
   {
     "temperature": 0.6,  // Updated from 0.7
     "maxTokens": 2000    // Updated from 4000
   }
   ```

5. **Apply group-specific options** (coding group):
   ```json
   {
     "temperature": 0.3,  // Updated from 0.6
     "maxTokens": 2000
   }
   ```

6. **Apply CLI options** (if provided):
   ```json
   {
     "temperature": 0.9,  // Updated from 0.3
     "maxTokens": 3000    // Updated from 2000
   }
   ```

The final resolved options would be:
```json
{
  "temperature": 0.9,
  "maxTokens": 3000
}
```

#### Using the Cascading Configuration System

You can customize model behavior in three main ways:

1. **Via Configuration File**
   - Define options at the model level in `thinktank.config.json`
   - Set provider-specific and model-specific options

2. **Via Groups**
   - Create specialized groups for different tasks
   - Override options on a per-group basis

3. **Via CLI**
   - Override any option when running thinktank:
   ```bash
   # Set temperature for all models in this run
   thinktank run prompt.txt --temperature 0.4

   # Use a specific group and override its options
   thinktank run prompt.txt --group coding --temperature 0.5
   
   # Use a specific model with overridden options
   thinktank run prompt.txt --model openai:gpt-4o --temperature 0.8
   ```

#### Best Practices

- **Task-Specific Configurations**
  - Use lower temperatures (0.1-0.3) for factual, precise tasks
  - Use medium temperatures (0.4-0.7) for balanced responses
  - Use higher temperatures (0.8-1.0) for creative, exploratory tasks

- **Layered Approach**
  - Define sensible defaults at the user config level
  - Create specialized groups for different tasks
  - Use CLI overrides for one-off adjustments

### Default Configuration

When you first run thinktank, a default configuration will be automatically created in your system's XDG-compliant configuration directory. This eliminates the need to manually set up a configuration file before using the tool.

The default configuration includes a minimal set of commonly used models from major providers, organized into basic groups that you can customize later:

```json
{
  "groups": {
    "default": {
      "name": "default",
      "systemPrompt": {
        "text": "You are a helpful, accurate, and intelligent assistant. Provide clear, concise, and correct information. If you are unsure about something, admit it rather than making up an answer."
      },
      "models": [
        {
          "provider": "openai",
          "modelId": "gpt-4o",
          "enabled": true
        }
      ],
      "description": "Default model group"
    },
    "coding": {
      "name": "coding",
      "systemPrompt": {
        "text": "You are a skilled software engineer with expertise in multiple programming languages and frameworks. Help write, explain, and debug code. Provide clear explanations and follow best practices for the language/framework being discussed."
      },
      "models": [],
      "description": "Models optimized for programming tasks"
    },
    "creative": {
      "name": "creative",
      "systemPrompt": {
        "text": "You are a creative assistant with a talent for generating unique and imaginative content. Help with writing, brainstorming, and ideation. Be inventive while still providing content that matches the user's request."
      },
      "models": [],
      "description": "Models for creative writing and generation"
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
    }
  ]
}
```

You can specify a custom configuration file path with the `--config` flag:

```bash
thinktank run prompt.txt --config ./custom-config.json
```

The configuration system prioritizes in the following order:
1. Explicitly specified configuration file with `--config`
2. XDG-compliant user configuration directory
3. Auto-generated default configuration if no configuration exists

#### Custom Configuration Path

You can specify a custom configuration file path using the `--config` flag:

```bash
thinktank run prompt.txt --config ./custom-config.json
```

This allows teams to share configuration settings via version control or use different configurations for different projects.

To see the path to the current configuration:

```bash
thinktank config path
```

The output will show the XDG-compliant configuration path that's being used.

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

#### Provider-Specific Options

Different providers support different options. Here are the most common ones for each:

##### Anthropic (Claude)

```json
{
  "temperature": 0.7,          // 0.0-1.0, controls randomness
  "maxTokens": 4000,           // Maximum output tokens
  "thinking": {                // Claude's thinking capability
    "type": "enabled",
    "budget_tokens": 10000
  }
}
```

##### OpenAI (GPT)

```json
{
  "temperature": 0.7,         // 0.0-2.0, controls randomness
  "maxTokens": 4000,          // Maximum output tokens
  "top_p": 0.95,              // Alternative to temperature
  "presence_penalty": 0.0,    // -2.0 to 2.0, penalizes repeated tokens
  "frequency_penalty": 0.0,   // -2.0 to 2.0, penalizes frequent tokens
  "seed": 12345,              // For reproducible outputs (if supported)
  "stop": ["###"]             // Stop sequences
}
```

##### Google (Gemini)

```json
{
  "temperature": 0.7,         // 0.0-1.0, controls randomness
  "maxTokens": 2048,          // Maximum output tokens
  "topK": 40,                 // Limits token selection pool
  "topP": 0.95                // Alternative to temperature
}
```

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

## Claude's Thinking Capability

Claude supports a "thinking" feature that allows the model to show its reasoning process before providing a final answer. This is especially helpful for complex tasks or when you want to understand how Claude arrived at its conclusion.

### Using Claude's Thinking

There are three ways to enable Claude's thinking:

1. **Use the `thinking` group:**
   ```bash
   thinktank prompt.txt --group thinking
   ```
   This uses a group specifically configured for Claude with thinking enabled.

2. **Add the `--thinking` flag:**
   ```bash
   thinktank prompt.txt --thinking
   ```
   This enables thinking for any Claude models in the group.

3. **Configure it directly in your config file:**
   ```json
   {
     "provider": "anthropic",
     "modelId": "claude-3-opus-20240229",
     "options": {
       "thinking": {
         "type": "enabled",
         "budget_tokens": 16000
       }
     }
   }
   ```

To display the thinking output, use the `--show-thinking` flag:
```bash
thinktank prompt.txt --thinking --show-thinking
```

### Important Temperature Limitation

**When using Claude's thinking capability, the temperature will automatically be set to 1, regardless of what value you configured.** This is a technical requirement from Anthropic's API.

For example, if you have:
```json
{
  "provider": "anthropic",
  "modelId": "claude-3-7-sonnet-20250219",
  "enabled": true,
  "options": {
    "temperature": 0.7,
    "thinking": {
      "type": "enabled",
      "budget_tokens": 16000
    }
  }
}
```

The temperature will be forced to 1 when making the API request, regardless of the 0.7 value specified.

### When to Use Thinking

Claude's thinking capability is most valuable for:

- Complex reasoning tasks
- Mathematical problem-solving
- Multi-step decisions
- Tasks requiring careful analysis
- Understanding model reasoning

Note that using thinking may increase token usage, as the model generates additional content for its reasoning process.

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

thinktank follows a domain-oriented architecture with a clear separation of concerns, designed for maintainability and extensibility:

```
src/
├── core/        # Core types, interfaces, config management, and registry
├── providers/   # LLM provider implementations
├── cli/         # Command-line interface with commander.js
├── utils/       # Common utility functions
└── workflow/    # Main workflow orchestration modules
```

### Key Components

- **ConfigManager**: Handles loading, validating, and modifying configuration with full CLI management
- **LLMRegistry**: Manages provider registration and retrieval with cascading configuration support
- **Providers**: Implementation of various LLM APIs (OpenAI, Anthropic, Google, OpenRouter)
- **CLI**: Provides a comprehensive command-line interface with commander.js
- **Workflow Modules**:
  - **InputHandler**: Processes prompts from files or direct input
  - **ModelSelector**: Determines which models to use based on configuration and CLI flags
  - **QueryExecutor**: Manages parallel API calls with proper error handling
  - **OutputHandler**: Formats and writes results to files and console
  
### Architectural Principles

The architecture is guided by several key principles:

- **Modularity**: Each component has a single responsibility and clear interfaces
- **Testability**: Components are designed for easy testing with dependency injection
- **Error Handling**: Comprehensive error system with categorization and helpful messages
- **Configuration**: Flexible, cascading configuration system with sensible defaults
- **Extension Points**: Clear patterns for adding new providers and features

## Extending thinktank

### Adding a New LLM Provider

1. Create a new file in `src/providers/<provider-name>.ts`
2. Implement the `LLMProvider` interface
3. Register the provider in the LLM registry

Here's an example implementation for a new provider:

```typescript
import { LLMProvider, LLMResponse, ModelOptions, LLMAvailableModel } from '../core/types';
import { registerProvider } from '../core/llmRegistry';

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

4. Import the provider in `src/workflow/runThinktank.ts`:

```typescript
// Import provider modules to ensure they're registered
import '../providers/openai';
import '../providers/anthropic';
import '../providers/google';
import '../providers/openrouter';
import '../providers/new-provider';
// Future providers will be imported here
```

5. Add an example configuration using the CLI:

```bash
thinktank config models add new-provider model-1 --options '{"temperature": 0.7}'
```

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

## Development Philosophy

thinktank is built with a focus on maintainability, reliability, and performance:

- **Test-Driven Development**: We write tests first to define expected behavior clearly
- **Type Safety**: We use TypeScript's strict type checking for reliability
- **Clean Architecture**: Domain-oriented architecture with clear separation of concerns
- **Continuous Integration**: We use GitHub Actions for automated linting, testing, and building
- **Pragmatic Simplicity**: We prioritize readability and maintainability
- **Minimal Abstraction**: We avoid premature abstraction and reassess regularly
- **Atomic Design**: Components are built from atoms→molecules→organisms→templates→runtime

For more detailed information on our development approach, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

For detailed contributor guidelines, best practices, and development philosophy, please review [CONTRIBUTING.md](CONTRIBUTING.md).
