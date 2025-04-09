# Troubleshooting Guide

This guide addresses common issues you might encounter when using thinktank.

## Common Issues

### "Input file not found"
- Ensure the file path is correct and the file exists
- Check if you're in the right working directory
- Try using an absolute path instead of a relative one

### "No models found in group"
- Verify that the group name exists in your configuration
- Check that models in the group are properly configured and enabled
- Try using a specific model identifier instead

### "Invalid model format"
- Ensure you're using the correct format: `provider:modelId`
- For OpenRouter models, use: `openrouter:provider/modelId`
- Check for typos in provider or model names

### "Missing API keys for models"
- Ensure you have set the correct environment variables in your `.env` file
- Check that the API keys are valid
- Follow the provider-specific instructions for obtaining API keys

### "Failed to create output directory"
- Check if you have write permissions for the specified output directory
- Ensure the path exists or can be created

## API Key Issues

If you're having issues with API keys:

1. Confirm your API keys are correctly set in the `.env` file
2. Verify that the environment variables match what the providers expect:
   - OpenAI: `OPENAI_API_KEY`
   - Anthropic: `ANTHROPIC_API_KEY`
   - Google: `GEMINI_API_KEY`
   - OpenRouter: `OPENROUTER_API_KEY`
3. You can override the environment variable name in the config using `apiKeyEnvVar`

## Installation Issues

### Global Installation Problems
- Ensure you've built the project: `pnpm run build`
- Make sure the CLI file is executable: `chmod +x dist/src/cli/index.js`
- Try running from the project directory: `./dist/src/cli/index.js`
- Check pnpm link status: `pnpm list -g`

### Dependency Issues
- Update dependencies: `pnpm update`
- Clean install: `rm -rf node_modules && pnpm install`
- Check Node.js version (requires 18.x+): `node --version`

## API Response Issues

### Timeouts
- Reduce the complexity or length of your prompt
- Check your internet connection
- Try a different model or provider

### Error Responses
- Check the error message for specific guidance
- Verify your API key has sufficient permissions and quota
- Review the provider's documentation for rate limits

## If All Else Fails

If you've tried the above solutions and still have issues:

1. Run with verbose logging: `DEBUG=thinktank:* thinktank run prompt.txt`
2. Check the thinktank-errors.log file
3. [Open an issue](https://github.com/phrazzld/thinktank/issues) with details of the problem and steps to reproduce