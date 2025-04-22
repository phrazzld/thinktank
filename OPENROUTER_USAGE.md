# Using OpenRouter with thinktank

To use OpenRouter models with the thinktank tool, follow these steps:

1. Ensure your OpenRouter API key is in the environment:
   export OPENROUTER_API_KEY="your-openrouter-api-key"

2. Confirm the key is correct by checking that it starts with "sk-or"

3. Use the full model name including the provider prefix:
   thinktank --instructions your_instructions.txt --model openrouter/deepseek/deepseek-chat-v3-0324 ./path/to/files

You should see debug logs indicating that:
- The OpenRouter API key is loaded from the environment
- The key starts with "sk-or" (not any other prefix)
- The provider is correctly identified as "openrouter"
