# thinktank Cascading Configuration Guide

thinktank implements a powerful cascading configuration system that intelligently resolves model options from multiple sources, ensuring sensible defaults while giving you fine-grained control over model behavior.

## Understanding the Configuration Hierarchy

Options for each model are resolved through a six-layer hierarchy, with each subsequent layer overriding settings from previous layers:

1. **Base Defaults** - Global options for all models
2. **Provider Defaults** - Defaults specific to a provider (e.g., Anthropic, OpenAI)
3. **Model-Specific Defaults** - Defaults for a specific model (e.g., Claude 3 Opus, GPT-4o)
4. **User Config Options** - Options set in your `thinktank.config.json` file
5. **Group-Specific Options** - Options set for a model group
6. **CLI Options** - Options provided via command-line flags

This layered approach means you don't need to specify every option for every model. Instead, the system automatically applies appropriate defaults and only overrides what you explicitly specify.

## Configuration Layers Explained

### 1. Base Defaults

These are the foundational options applied to all models:

```json
{
  "temperature": 0.7,
  "maxTokens": 1000
}
```

### 2. Provider Defaults

These options are specific to a provider:

```json
{
  "anthropic": {
    "thinking": {
      "type": "enabled",
      "budget_tokens": 10000
    }
  },
  "openai": {
    "temperature": 0.7
  },
  "google": {
    "temperature": 0.7
  },
  "openrouter": {
    "temperature": 0.7
  }
}
```

### 3. Model-Specific Defaults

These options are tailored to specific models:

```json
{
  "anthropic:claude-3-opus-20240229": {
    "temperature": 0.7,
    "maxTokens": 4000
  },
  "anthropic:claude-3-sonnet-20240229": {
    "temperature": 0.7,
    "maxTokens": 4000
  },
  "anthropic:claude-3-haiku-20240307": {
    "temperature": 0.8,
    "maxTokens": 2000
  },
  "openai:gpt-4o": {
    "temperature": 0.7,
    "maxTokens": 4000
  },
  "openai:gpt-3.5-turbo": {
    "temperature": 0.8,
    "maxTokens": 2000
  }
}
```

### 4. User Config Options

These are options you define in your `thinktank.config.json` file:

```json
{
  "models": [
    {
      "provider": "openai",
      "modelId": "gpt-4o",
      "enabled": true,
      "options": {
        "temperature": 0.6,
        "maxTokens": 2000
      }
    }
  ]
}
```

### 5. Group-Specific Options

Options defined for a specific group take precedence over user config options:

```json
{
  "groups": {
    "coding": {
      "name": "coding",
      "systemPrompt": {
        "text": "You are an expert software engineer..."
      },
      "models": [
        {
          "provider": "openai",
          "modelId": "gpt-4o",
          "enabled": true,
          "options": {
            "temperature": 0.3  // Lower temperature for coding tasks
          }
        }
      ]
    }
  }
}
```

### 6. CLI Options

Options provided via command-line flags have the highest priority:

```bash
thinktank run prompt.txt --temperature 0.9 --max-tokens 3000
```

## How Option Resolution Works

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

## Using the Cascading Configuration System

### Via Configuration File

Define options at the model level in your `thinktank.config.json`:

```json
{
  "models": [
    {
      "provider": "anthropic",
      "modelId": "claude-3-opus-20240229",
      "enabled": true,
      "options": {
        "temperature": 0.7,
        "maxTokens": 4000,
        "thinking": {
          "type": "enabled",
          "budget_tokens": 12000
        }
      }
    }
  ]
}
```

### Via Groups

Define groups with specialized configurations:

```json
{
  "groups": {
    "creative": {
      "name": "creative",
      "systemPrompt": {
        "text": "You are a creative writer with a flair for storytelling..."
      },
      "models": [
        {
          "provider": "anthropic",
          "modelId": "claude-3-opus-20240229",
          "enabled": true,
          "options": {
            "temperature": 0.9  // Higher temperature for creative tasks
          }
        },
        {
          "provider": "openai",
          "modelId": "gpt-4o",
          "enabled": true,
          "options": {
            "temperature": 0.9
          }
        }
      ]
    },
    "analytical": {
      "name": "analytical",
      "systemPrompt": {
        "text": "You are an expert analyst skilled at breaking down complex problems..."
      },
      "models": [
        {
          "provider": "anthropic",
          "modelId": "claude-3-opus-20240229",
          "enabled": true,
          "options": {
            "temperature": 0.2  // Lower temperature for analytical tasks
          }
        },
        {
          "provider": "openai",
          "modelId": "gpt-4o",
          "enabled": true,
          "options": {
            "temperature": 0.2
          }
        }
      ]
    }
  }
}
```

### Via CLI

Override any option when running thinktank:

```bash
# Set temperature for all models in this run
thinktank run prompt.txt --temperature 0.4

# Set multiple options
thinktank run prompt.txt --temperature 0.4 --max-tokens 5000

# Use a specific group and override its options
thinktank run prompt.txt --group coding --temperature 0.5

# Use a specific model with overridden options
thinktank run prompt.txt --model openai:gpt-4o --temperature 0.8
```

## Common Configuration Patterns

### Task-Optimized Groups

Create different groups for different tasks:

```json
{
  "groups": {
    "coding": {
      "models": [
        {
          "provider": "anthropic",
          "modelId": "claude-3-opus-20240229",
          "options": {
            "temperature": 0.2,
            "maxTokens": 8000
          }
        }
      ]
    },
    "creative": {
      "models": [
        {
          "provider": "anthropic",
          "modelId": "claude-3-opus-20240229",
          "options": {
            "temperature": 0.9,
            "maxTokens": 4000
          }
        }
      ]
    },
    "chat": {
      "models": [
        {
          "provider": "anthropic",
          "modelId": "claude-3-haiku-20240307",
          "options": {
            "temperature": 0.7,
            "maxTokens": 2000
          }
        }
      ]
    }
  }
}
```

### Provider-Specific Settings

Configure options specific to each provider:

```json
{
  "models": [
    {
      "provider": "anthropic",
      "modelId": "claude-3-opus-20240229",
      "options": {
        "thinking": {
          "type": "enabled",
          "budget_tokens": 16000
        }
      }
    },
    {
      "provider": "openai",
      "modelId": "gpt-4o",
      "options": {
        "top_p": 0.95,
        "presence_penalty": 0.1
      }
    }
  ]
}
```

### Defaulting Strategy

Apply a conservative default with task-specific overrides:

```json
{
  "models": [
    {
      "provider": "anthropic",
      "modelId": "claude-3-opus-20240229",
      "options": {
        "temperature": 0.5,  // Middle ground default
        "maxTokens": 4000
      }
    }
  ],
  "groups": {
    "precise": {
      "models": [
        {
          "provider": "anthropic",
          "modelId": "claude-3-opus-20240229",
          "options": {
            "temperature": 0.1  // Lower for precision
          }
        }
      ]
    },
    "exploratory": {
      "models": [
        {
          "provider": "anthropic",
          "modelId": "claude-3-opus-20240229",
          "options": {
            "temperature": 0.9  // Higher for exploration
          }
        }
      ]
    }
  }
}
```

## Practical Examples

### Example 1: Software Development Setup

```json
{
  "models": [
    {
      "provider": "anthropic",
      "modelId": "claude-3-opus-20240229",
      "enabled": true,
      "options": {
        "temperature": 0.7,
        "maxTokens": 4000
      }
    },
    {
      "provider": "openai",
      "modelId": "gpt-4o",
      "enabled": true,
      "options": {
        "temperature": 0.7,
        "maxTokens": 4000
      }
    }
  ],
  "groups": {
    "coding": {
      "name": "coding",
      "systemPrompt": {
        "text": "You are an expert software engineer with deep knowledge of programming languages, algorithms, and design patterns. Write clean, efficient, and maintainable code."
      },
      "models": [
        {
          "provider": "anthropic",
          "modelId": "claude-3-opus-20240229",
          "options": {
            "temperature": 0.2
          }
        },
        {
          "provider": "openai",
          "modelId": "gpt-4o",
          "options": {
            "temperature": 0.2
          }
        }
      ]
    },
    "debug": {
      "name": "debug",
      "systemPrompt": {
        "text": "You are an expert at debugging code. Identify potential issues, suggest fixes, and explain your reasoning."
      },
      "models": [
        {
          "provider": "anthropic",
          "modelId": "claude-3-opus-20240229",
          "options": {
            "temperature": 0.1,
            "thinking": {
              "type": "enabled",
              "budget_tokens": 20000
            }
          }
        }
      ]
    },
    "architecture": {
      "name": "architecture",
      "systemPrompt": {
        "text": "You are a software architect with expertise in system design. Provide high-level architectural guidance."
      },
      "models": [
        {
          "provider": "anthropic",
          "modelId": "claude-3-opus-20240229",
          "options": {
            "temperature": 0.3,
            "maxTokens": 8000
          }
        }
      ]
    }
  }
}
```

Usage:
```bash
# Generate code
thinktank run code-prompt.txt --group coding

# Debug an issue
thinktank run debug-prompt.txt --group debug --thinking

# Architecture planning with increased creativity
thinktank run architecture-prompt.txt --group architecture --temperature 0.6
```

### Example 2: Content Creation Setup

```json
{
  "groups": {
    "writing": {
      "name": "writing",
      "systemPrompt": {
        "text": "You are a skilled writer. Create engaging, well-structured content."
      },
      "models": [
        {
          "provider": "anthropic",
          "modelId": "claude-3-opus-20240229",
          "options": {
            "temperature": 0.7,
            "maxTokens": 6000
          }
        }
      ]
    },
    "brainstorm": {
      "name": "brainstorm",
      "systemPrompt": {
        "text": "You are an idea generator. Provide diverse, creative ideas without self-censoring."
      },
      "models": [
        {
          "provider": "openai",
          "modelId": "gpt-4o",
          "options": {
            "temperature": 0.9,
            "maxTokens": 4000
          }
        },
        {
          "provider": "anthropic",
          "modelId": "claude-3-opus-20240229",
          "options": {
            "temperature": 0.9,
            "maxTokens": 4000
          }
        }
      ]
    },
    "edit": {
      "name": "edit",
      "systemPrompt": {
        "text": "You are an editor with an eye for detail. Improve writing while maintaining the author's voice."
      },
      "models": [
        {
          "provider": "anthropic",
          "modelId": "claude-3-opus-20240229",
          "options": {
            "temperature": 0.3,
            "maxTokens": 8000
          }
        }
      ]
    }
  }
}
```

Usage:
```bash
# Generate a first draft
thinktank run article-outline.txt --group writing

# Brainstorm ideas
thinktank run topic.txt --group brainstorm

# Edit a draft with lower creativity but higher precision
thinktank run draft.txt --group edit
```

## Provider-Specific Options

Different providers support different options. Here are the most common ones for each:

### Anthropic (Claude)

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

### OpenAI (GPT)

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

### Google (Gemini)

```json
{
  "temperature": 0.7,         // 0.0-1.0, controls randomness
  "maxTokens": 2048,          // Maximum output tokens
  "topK": 40,                 // Limits token selection pool
  "topP": 0.95                // Alternative to temperature
}
```

## Integration with Provider Capabilities

### Claude's Thinking Capability

Configure Claude's thinking feature:

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

Enable at runtime:
```bash
thinktank run prompt.txt --model anthropic:claude-3-opus-20240229 --thinking
```

**Note:** When using Claude's thinking capability, temperature is forced to 1.0 due to API limitations, regardless of configured value.

## Best Practices

1. **Start Simple, Refine Later**
   - Begin with minimal configuration
   - Gradually refine as you observe model outputs
   - Create specialized groups once patterns emerge

2. **Task-Specific Configurations**
   - Use lower temperatures (0.1-0.3) for factual, precise tasks
   - Use medium temperatures (0.4-0.7) for balanced responses
   - Use higher temperatures (0.8-1.0) for creative, exploratory tasks

3. **Consistent Provider Options**
   - Maintain consistent options across providers for comparable outputs
   - Align temperature scales appropriately (OpenAI vs. Anthropic)

4. **Layered Approach**
   - Define sensible defaults at the user config level
   - Create specialized groups for different tasks
   - Use CLI overrides for one-off adjustments

5. **Test Different Configurations**
   - Compare outputs with different settings
   - Use groups to maintain different configuration profiles
   - Document what works best for different tasks

## Troubleshooting

### Common Issues

1. **Unexpected Model Behavior**
   - Check the resolved options using `thinktank config show resolved openai:gpt-4o`
   - Verify that CLI overrides are being applied correctly
   - Remember that group options override user config options

2. **Temperature Issues with Claude's Thinking**
   - Remember that when using Claude's thinking capability, temperature is forced to 1.0

3. **Options Not Taking Effect**
   - Verify the option name is correct (e.g., `maxTokens` not `max_tokens`)
   - Check the configuration hierarchy to see which layer might be overriding your setting
   - Ensure you're using the correct CLI flag format (e.g., `--max-tokens` for CLI, but `maxTokens` in JSON)

## Configuration Management Commands

thinktank provides several commands to manage your configuration:

```bash
# Show the current configuration
thinktank config show

# Show resolved options for a specific model
thinktank config show resolved openai:gpt-4o

# Add or update a model with specific options
thinktank config models add openai gpt-4o --options '{"temperature": 0.7, "maxTokens": 4000}'

# Create a new group with specific system prompt
thinktank config groups create coding --system-prompt "You are an expert programmer..." --models openai:gpt-4o,anthropic:claude-3-opus

# Add a model to a group with specific options
thinktank config groups add-model coding openai gpt-4o --options '{"temperature": 0.2}'
```

## Conclusion

The cascading configuration system in thinktank provides a flexible yet powerful way to control LLM behavior. By understanding the hierarchy and how options resolve, you can create customized configurations for different tasks while maintaining reasonable defaults.

This approach allows you to:
- Define sensible defaults for all models
- Create task-specific configurations in groups
- Override specific options via the CLI when needed
- Maintain a clean, organized configuration file

By leveraging this system effectively, you can get the most out of your LLM interactions with minimal configuration overhead.