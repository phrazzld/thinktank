# OpenRouter Integration in thinktank

This document explains how thinktank integrates with OpenRouter, a unified gateway service that provides access to various LLM models from different providers through a single consistent API.

## Overview

thinktank now uses OpenRouter exclusively as its unified provider. OpenRouter provides access to models from OpenAI, Google (Gemini), DeepSeek, Anthropic, xAI (Grok), and other providers through a single API with a consistent interface.

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

All models are now defined in `internal/models/models.go` and use OpenRouter for access. The current supported models include:

**OpenAI Models (via OpenRouter):**
- `gpt-5.2` (openai/gpt-5.2) - GPT-4.1 model
- `o3` (openai/o3) - O4 Mini model
- `o3` (openai/o3) - O3 model

**Google Models (via OpenRouter):**
- `gemini-3-flash` (google/gemini-3-flash) - Gemini 2.5 Flash model
- `gemini-3-flash` (google/gemini-3-flash) - Gemini 2.5 Pro model

**Native OpenRouter Models:**
- `openrouter/deepseek/deepseek-chat-v3-0324` - DeepSeek Chat model with 65k context
- `openrouter/deepseek/deepseek-r1` - DeepSeek R1 model with 131k context
- `openrouter/x-ai/grok-3-beta` - xAI's Grok model with 131k context

Each model is defined with:
- Provider always set to `openrouter` (unified architecture)
- API model ID for OpenRouter API calls (e.g., `openai/gpt-5.2`, `google/gemini-3-flash`, `deepseek/deepseek-r1`)
- Context window and max output token limits
- Default parameters like temperature and top_p

To add new OpenRouter models, edit the `ModelDefinitions` map in `internal/models/models.go`. See [CLAUDE.md](./CLAUDE.md#adding-new-models) for detailed instructions.

## Migration from Multi-Provider Architecture

### What Changed

Previously, thinktank used separate providers for different model families:
- OpenAI models required `OPENAI_API_KEY`
- Google Gemini models required `GEMINI_API_KEY`
- OpenRouter models required `OPENROUTER_API_KEY`

**After the consolidation, all models now use OpenRouter exclusively and require only `OPENROUTER_API_KEY`.**

### Migration Steps

1. **Set up OpenRouter API Key:**
   ```bash
   export OPENROUTER_API_KEY="your-openrouter-api-key"
   ```

2. **Remove old API keys (optional but recommended):**
   ```bash
   unset OPENAI_API_KEY
   unset GEMINI_API_KEY
   ```

3. **Update your scripts or CI/CD:** Replace any references to the old environment variables with `OPENROUTER_API_KEY`.

### Backward Compatibility

The system automatically detects old API keys and provides helpful migration messages:

```
Warning: OPENAI_API_KEY detected but no longer used.
Please set OPENROUTER_API_KEY instead.
Get your key at: https://openrouter.ai/keys
```

### What Stays the Same

- **CLI Commands:** All existing commands work identically
- **Model Names:** Model names remain the same (e.g., `gpt-5.2`, `gemini-3-flash`)
- **Features:** All functionality is preserved
- **Configuration:** All flags and options work the same way

### Benefits of the New Architecture

- **Simplified Setup:** Only one API key needed
- **Unified Billing:** All model usage through a single OpenRouter account
- **Consistent Interface:** All models use the same API format
- **Access to More Models:** OpenRouter provides access to additional models like DeepSeek, xAI Grok, etc.
- **Better Rate Limiting:** Unified rate limiting across all models

## Usage Examples

### Basic Usage

```bash
# Use any model (all models now use OpenRouter)
thinktank task.txt ./src --model gpt-5.2
thinktank task.txt ./src --model gemini-3-flash
thinktank task.txt ./src --model openrouter/deepseek/deepseek-r1

# Simplified interface with automatic model selection
thinktank task.txt ./src

# Force synthesis mode for multi-model analysis
thinktank task.txt ./src --synthesis
```

### Model Selection

All models are now accessed through OpenRouter's unified API:

**Popular Models:**
- `gpt-5.2` - OpenAI's GPT-4.1 model via OpenRouter
- `gemini-3-flash` - Google's Gemini 2.5 Pro via OpenRouter
- `openrouter/deepseek/deepseek-r1` - DeepSeek R1 with reasoning capabilities
- `openrouter/x-ai/grok-3-beta` - xAI's Grok model

**Adding New Models:**
Edit the `ModelDefinitions` map in `internal/models/models.go` with `Provider: "openrouter"` and the appropriate OpenRouter model ID.

## Implementation Details

### Provider Implementation

The OpenRouter provider is the sole provider implementation. It:
1. Implements the unified Provider interface
2. Creates OpenRouter clients configured with the `OPENROUTER_API_KEY`
3. Handles all model access through OpenRouter's API endpoints
4. Validates model IDs for all supported models

### Client Implementation

The unified OpenRouter client handles all model requests. Key features:

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

The unified error handling system provides:

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

## BYOK (Bring Your Own Key) Models

Some models on OpenRouter require you to bring your own API key from the original provider. These models are marked with `RequiresBYOK: true` in the model definitions.

### Currently Supported BYOK Models

- `o3` (openai/o3) - OpenAI's O3 model requires an OpenAI API key

### How to Use BYOK Models

1. **Add your provider API key to OpenRouter:**
   - Go to [OpenRouter Settings > Integrations](https://openrouter.ai/settings/integrations)
   - Add your provider's API key (e.g., OpenAI API key)
   - OpenRouter will automatically use your key when you request BYOK models

2. **Use the model normally in thinktank:**
   ```bash
   # Once your key is added to OpenRouter, use BYOK models like any other
   thinktank task.txt ./src --model o3
   ```

### BYOK Error Messages

If you try to use a BYOK model without adding your provider key to OpenRouter, you'll see:
```
Model openai/o3 requires you to add your own API key to OpenRouter
Please add your OpenAI API key at https://openrouter.ai/settings/integrations to use this model
```

### Benefits of BYOK

- Access to premium models without OpenRouter markup
- Use your existing provider relationships and pricing
- OpenRouter handles the unified interface while you maintain direct billing

## Limitations

- OpenRouter may have different rate limits and pricing for different models
- Some advanced features may not be available for all models
- Performance can vary based on the underlying model provider's infrastructure
- Each model has its own context window limitations
- BYOK models require additional setup through OpenRouter's website

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
