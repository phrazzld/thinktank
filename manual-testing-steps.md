# Manual End-to-End Testing Instructions

This document outlines the steps for manual testing of the OpenRouter authentication fix, specifically verifying that API key isolation works correctly across different providers.

## Prerequisites

1. Ensure you have the following environment variables set with valid API keys:
   - `GEMINI_API_KEY` - for testing with Google's Gemini models
   - `OPENAI_API_KEY` - for testing with OpenAI models
   - `OPENROUTER_API_KEY` - for testing with OpenRouter models (should start with "sk-or")

2. Make sure you have the latest version of the application built:
   ```bash
   go build -o thinktank ./cmd/thinktank
   ```

## Test Cases

### Test Case 1: Gemini Model Authentication

1. Run the following command:
   ```bash
   LOGRUS_DEBUG=1 ./thinktank --instructions test-prompt.txt --model gemini-2.5-pro-preview-03-25 --output-dir ./test-output ./
   ```

2. Verify the following in the debug logs:
   - The application loaded the model configuration correctly
   - It used the GEMINI_API_KEY from the environment (check for log message: `Using API key from environment variable GEMINI_API_KEY for provider 'gemini'`)
   - The API call was successful (no authentication errors)

### Test Case 2: OpenRouter Model Authentication

1. Run the following command:
   ```bash
   LOGRUS_DEBUG=1 ./thinktank --instructions test-prompt.txt --model openrouter/deepseek/deepseek-r1 --output-dir ./test-output ./
   ```

2. Verify the following in the debug logs:
   - The application loaded the model configuration correctly
   - It used the OPENROUTER_API_KEY from the environment (check for log message: `Using API key from environment variable OPENROUTER_API_KEY for provider 'openrouter'`)
   - The API key prefix is shown as "sk-or" in the logs
   - The API call was successful (no authentication errors)

### Test Case 3: Multiple Models in Sequence

1. Run the following command with multiple models:
   ```bash
   LOGRUS_DEBUG=1 ./thinktank --instructions test-prompt.txt --model gemini-2.5-pro-preview-03-25 --model openrouter/deepseek/deepseek-r1 --output-dir ./test-output ./
   ```

2. Verify the following in the debug logs:
   - The application correctly switches API keys between providers
   - For the Gemini model, it uses the GEMINI_API_KEY
   - For the OpenRouter model, it uses the OPENROUTER_API_KEY
   - Both API calls are successful (no authentication errors)
   - There is no key leakage between providers

## Expected Results

- Each provider should use its own dedicated API key from the environment
- Debug logs should show which API key is being used for each provider
- The application should successfully authenticate with each provider
- When running multiple models in sequence, there should be no key leakage between providers

## Troubleshooting

If authentication fails:
1. Verify the environment variables are set correctly
2. Check that the API keys are valid and have the correct format
3. Look for error messages in the debug logs that might indicate issues with the API keys
4. Verify that the `InitLLMClient` method in `internal/thinktank/registry_api.go` is correctly prioritizing environment variables

## Test Completion Criteria

The test is considered successful when:
1. OpenRouter authentication works correctly with the OPENROUTER_API_KEY
2. Gemini authentication works correctly with the GEMINI_API_KEY
3. Both can work in the same session without key leakage
