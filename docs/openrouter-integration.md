# OpenRouter Integration in thinktank

This document explains how thinktank integrates with OpenRouter, a unified gateway service that provides access to various LLM models from different providers through a single consistent API.

## Overview

thinktank supports OpenRouter as a first-class provider alongside OpenAI and Gemini. OpenRouter allows you to access a wide range of models from providers like DeepSeek, Anthropic, xAI (Grok), and others through a single API with a consistent interface.

The OpenRouter integration in thinktank includes:
- Hardcoded model definitions for reliability and simplicity
- Uniform API key handling
- Parameter mapping that preserves model-specific options
- Comprehensive error handling with helpful suggestions
- Support for concurrent requests across multiple models

## Configuration

### API Key Setup

To use OpenRouter models with thinktank, you need to set up an API key:

1. Create an account on [OpenRouter](https://openrouter.ai/)
2. Navigate to the [API Keys page](https://openrouter.ai/keys)
3. Create a new API key
4. Set the API key as an environment variable:
   ```bash
   export OPENROUTER_API_KEY="your-openrouter-api-key"
   ```

The key can also be passed programmatically when creating a new client, but the environment variable approach is recommended for security.

### Model Configuration

OpenRouter models are hardcoded in `internal/models/models.go` for reliability and simplicity. The current supported OpenRouter models include:

- `openrouter/deepseek/deepseek-chat-v3-0324` - DeepSeek Chat model with 65k context
- `openrouter/deepseek/deepseek-r1` - DeepSeek R1 model with 131k context
- `openrouter/x-ai/grok-3-beta` - xAI's Grok model with 131k context

Each model is defined with:
- Provider set to `openrouter`
- API model ID for the actual OpenRouter API calls (e.g., `deepseek/deepseek-r1`)
- Context window and max output token limits
- Default parameters like temperature and top_p

To add new OpenRouter models, edit the `ModelDefinitions` map in `internal/models/models.go`. See [CLAUDE.md](./CLAUDE.md#adding-new-models) for detailed instructions.

## Usage Examples

### Basic Usage

```bash
# Use an OpenRouter model
thinktank --instructions task.txt --model openrouter/deepseek/deepseek-r1 ./src

# Compare outputs from multiple providers
thinktank --instructions task.txt \
  --model gemini-2.5-pro \
  --model gpt-4-turbo \
  --model openrouter/x-ai/grok-3-beta \
  ./src
```

### Model Selection

OpenRouter provides access to many models. Here are some examples included in the default configuration:

- `openrouter/deepseek/deepseek-chat-v3-0324` - DeepSeek Chat model with 64k context
- `openrouter/deepseek/deepseek-r1` - DeepSeek R1 model with 128k context
- `openrouter/x-ai/grok-3-beta` - xAI's Grok model with 131k context

You can add more models by editing the `ModelDefinitions` map in `internal/models/models.go`.

## Implementation Details

### Provider Implementation

The OpenRouter provider is implemented in `internal/providers/openrouter/provider.go`. It:
1. Implements the `Provider` interface from `internal/providers/provider.go`
2. Creates OpenRouter clients configured with appropriate API keys and endpoints
3. Validates model IDs to ensure they follow the required format

### Client Implementation

The OpenRouter client is implemented in `internal/providers/openrouter/client.go`. Key features:

1. **Request Parameters**: Local parameter handling for thread safety during concurrent requests
2. **API Mapping**: Maps thinktank's parameter format to OpenRouter's API format
3. **Response Processing**: Translates OpenRouter API responses into thinktank's internal format
4. **Error Handling**: Comprehensive error detection, categorization, and reporting

### Parameter Mapping

The OpenRouter client maps the following parameters from thinktank to the OpenRouter API:

| thinktank Parameter | OpenRouter API Parameter | Notes |
|---------------------|--------------------------|-------|
| `temperature`       | `temperature`            | Range 0.0-1.0, controls randomness |
| `top_p`             | `top_p`                  | Range 0.0-1.0, controls diversity |
| `presence_penalty`  | `presence_penalty`       | Range 0.0-2.0, discourages repetition |
| `frequency_penalty` | `frequency_penalty`      | Range 0.0-2.0, discourages frequent tokens |
| `max_tokens`        | `max_tokens`             | Maximum generation length |

Parameters are defined in the default configuration in `internal/models/models.go` or can be passed directly in the request.

## Error Handling

The OpenRouter integration includes sophisticated error handling in `internal/providers/openrouter/errors.go` that:

1. Categorizes errors into standard types (auth, rate limit, invalid request, etc.)
2. Provides detailed error messages with context about what went wrong
3. Includes helpful suggestions for resolving common issues
4. Maps OpenRouter-specific error codes to standardized categories

Common error types and their resolution:

| Error Category | Possible Causes | Resolution |
|----------------|-----------------|------------|
| Authentication | Invalid or expired API key | Check your API key and ensure OPENROUTER_API_KEY is set correctly |
| Rate Limit | Too many requests | Wait and try again; consider using `--max-concurrent` and `--rate-limit` flags |
| Insufficient Credits | Account balance too low | Add credits to your OpenRouter account |
| Invalid Request | Malformed request parameters | Check input format and parameters |
| Not Found | Invalid model ID | Verify the model name is correct with format 'provider/model' |
| Server | Issues with OpenRouter or model provider | Wait and try again later |
| Network | Connection issues | Check your internet connection |

## Limitations

- OpenRouter may have different rate limits and pricing for different models
- Some advanced features may not be available for all models
- Performance can vary based on the underlying model provider's infrastructure
- Each model has its own context window limitations

## Troubleshooting

### API Key Issues
- Ensure `OPENROUTER_API_KEY` is set correctly
- Verify the key has not expired on the OpenRouter dashboard
- Check if the key has appropriate permissions

### Model Selection Issues
- Make sure the model ID follows the correct format (typically `provider/model`)
- Check that the model is available on OpenRouter
- Verify the model is defined in `internal/models/models.go`

### Request Failures
- Check the error message for specific details
- For context window errors, reduce the number of files or try a model with a larger context window
- For rate limit errors, adjust the `--max-concurrent` and `--rate-limit` flags
- For credit issues, check your OpenRouter account balance

### Performance Optimization
- Use the `--include` flag to limit files to only what's needed
- Consider concurrent requests with multiple models using repeatable `--model` flags
- Adjust timeout settings if working with large inputs or complex tasks

## Further Resources

- [OpenRouter Website](https://openrouter.ai/)
- [OpenRouter API Documentation](https://openrouter.ai/docs)
- [thinktank README](../README.md)
- [General OpenRouter API Guide](./openrouter-docs.md)
