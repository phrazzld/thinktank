# API Key Resolution Documentation

This document describes how the thinktank application resolves API keys for different LLM providers.

## Overview

Thinktank uses a centralized API key resolution system that provides a consistent and secure way to manage API keys across all supported providers. The resolution follows a clear precedence order and ensures proper isolation between different providers' keys.

## Resolution Precedence

The system resolves API keys in the following order:

### 1. Environment Variables (Highest Priority - Recommended)

Environment variables are the **recommended** method for providing API keys. Each provider has a specific environment variable:

| Provider | Environment Variable |
|----------|---------------------|
| OpenAI | `OPENAI_API_KEY` |
| Gemini | `GEMINI_API_KEY` |
| OpenRouter | `OPENROUTER_API_KEY` |

Custom providers can define their own environment variable mappings in `~/.config/thinktank/models.yaml`.

**Example:**
```bash
export OPENAI_API_KEY="sk-..."
export GEMINI_API_KEY="..."
export OPENROUTER_API_KEY="sk-or-..."
```

### 2. Explicitly Provided API Key (Fallback)

If an environment variable is not set, the system will use an API key provided directly to the function. This method is:
- Primarily intended for testing and development
- Discouraged for production use
- Only used if the environment variable is empty or not set

## Security Considerations

1. **Never hardcode API keys** in your source code
2. **Use environment variables** for production deployments
3. **Each provider requires its own API key** - never use one provider's key for another
4. **Keys are validated** to ensure they match expected formats where applicable

## Provider-Specific Requirements

### OpenAI
- Keys should start with `sk-`
- Uses `OPENAI_API_KEY` environment variable
- Example: `sk-1234567890abcdef...`

### Gemini
- No specific prefix requirement
- Uses `GEMINI_API_KEY` environment variable
- Keys should be at least 20 characters long

### OpenRouter
- Keys must start with `sk-or`
- Uses `OPENROUTER_API_KEY` environment variable
- Example: `sk-or-v1-1234567890abcdef...`

## Custom Provider Configuration

Custom providers can be configured in `~/.config/thinktank/models.yaml`:

```yaml
apikey_sources:
  customprovider: "CUSTOM_PROVIDER_API_KEY"
  anotherprovider: "ANOTHER_API_KEY"
```

## Implementation Details

The API key resolution is handled by the `APIKeyResolver` in `internal/providers/apikey.go`. This centralizes all key resolution logic and provides:

- Consistent error messages
- Proper logging (without exposing keys)
- Validation of key formats
- Clear precedence rules

## Troubleshooting

### Common Issues

1. **"API key required but not found"**
   - Ensure the correct environment variable is set
   - Check the variable name matches the provider (e.g., `OPENAI_API_KEY` not `OPENAI_KEY`)

2. **"Invalid API key format"**
   - Verify the key matches the provider's expected format
   - Ensure you're not using another provider's key

3. **Authentication failures**
   - Double-check you're using the correct key for the provider
   - Verify the key hasn't expired or been revoked

### Debugging

To debug API key issues:

1. Enable verbose logging: `--verbose` or `--log-level=debug`
2. Check which environment variables are set: `env | grep API_KEY`
3. Verify key format matches provider requirements

## Best Practices

1. **Use environment variables** exclusively in production
2. **Rotate keys regularly** for security
3. **Use separate keys** for development, staging, and production
4. **Never commit keys** to version control
5. **Monitor key usage** through provider dashboards

## Migration Guide

If you're updating from an older version that used different key resolution:

1. Update environment variable names to match the standard format
2. Remove any hardcoded keys from configuration files
3. Update deployment scripts to set the correct environment variables

## Related Documentation

- [Configuration Guide](configuration.md)
- [Security Best Practices](security.md)
- [Provider Setup Guide](providers.md)