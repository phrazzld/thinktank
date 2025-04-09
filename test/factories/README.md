# Test Data Factories

This directory contains factory functions for creating test data objects with sensible defaults. These factories make tests more readable and maintainable by reducing boilerplate and standardizing test data creation.

## Available Factories

### App Configuration

`appConfigFactory.ts` provides functions for creating app configuration objects:

- **`createAppConfig()`**: Creates an `AppConfig` object with default values and optional overrides

```typescript
import { createAppConfig } from '../../test/factories';

// Create a default app config
const defaultConfig = createAppConfig();

// Create a customized app config
const customConfig = createAppConfig({
  models: [
    createModelConfig({ provider: 'openai', modelId: 'gpt-4o' }),
    createModelConfig({ provider: 'anthropic', modelId: 'claude-3-opus-20240229' })
  ]
});
```

### Model Configuration

`modelConfigFactory.ts` provides functions for creating model configuration objects:

- **`createModelConfig()`**: Creates a `ModelConfig` object with default values and optional overrides

```typescript
import { createModelConfig } from '../../test/factories';

// Create a default model config
const defaultModel = createModelConfig();

// Create a custom model config
const openaiModel = createModelConfig({
  provider: 'openai',
  modelId: 'gpt-4o',
  maxTokens: 8192
});
```

### LLM Responses

`llmResponseFactory.ts` provides functions for creating LLM response objects:

- **`createLlmResponse()`**: Creates an `LLMResponse` object with default values and optional overrides

```typescript
import { createLlmResponse } from '../../test/factories';

// Create a successful response
const successResponse = createLlmResponse({
  response: 'Generated text',
  model: 'openai:gpt-4o',
  tokensUsed: { total: 100, prompt: 50, completion: 50 }
});

// Create an error response
const errorResponse = createLlmResponse({
  error: new Error('API error'),
  model: 'anthropic:claude-3-opus-20240229'
});
```

### Run Options

`runOptionsFactory.ts` provides functions for creating run options objects:

- **`createRunOptions()`**: Creates a `RunOptions` object with default values and optional overrides

```typescript
import { createRunOptions } from '../../test/factories';

// Create default run options
const defaultOptions = createRunOptions();

// Create custom run options
const customOptions = createRunOptions({
  modelId: 'openai:gpt-4o',
  outputDir: '/custom/output',
  prompt: 'Custom prompt',
  contextFiles: ['/path/to/file.js']
});
```

## Best Practices

1. **Default First**: Start with the default factory and only override what you need for your specific test.

2. **Combine Factories**: Use multiple factories together to create complex test scenarios.

3. **Object Spread**: Use object spread syntax to combine defaults with overrides:
   ```typescript
   const customOptions = createRunOptions({
     ...createRunOptions(),
     modelId: 'custom-model',
     // other overrides...
   });
   ```

4. **Test Relevant Properties**: Only override properties that are relevant to your test case, let the factory handle the rest.

5. **Reuse Constants**: For frequently used test configurations, define constants at the describe level:
   ```typescript
   describe('Model Selection', () => {
     const TEST_MODEL = createModelConfig({
       provider: 'openai',
       modelId: 'gpt-4o'
     });
     
     // Use TEST_MODEL in multiple tests
   });
   ```
