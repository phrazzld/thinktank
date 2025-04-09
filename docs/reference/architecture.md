# Architecture

thinktank follows a domain-oriented architecture with clear separation of concerns:

```
src/
├── core/        # Core types, interfaces, config management, and registry
├── providers/   # LLM provider implementations
├── cli/         # Command-line interface with commander.js
├── utils/       # Common utility functions
└── workflow/    # Main workflow orchestration modules
```

## Key Components

- **ConfigManager**: Handles loading, validating, and modifying configuration with full CLI management
- **LLMRegistry**: Manages provider registration and retrieval with cascading configuration support
- **Providers**: Implementation of various LLM APIs (OpenAI, Anthropic, Google, OpenRouter)
- **CLI**: Provides a comprehensive command-line interface with commander.js
- **Workflow Modules**:
  - **InputHandler**: Processes prompts from files or direct input
  - **ModelSelector**: Determines which models to use based on configuration and CLI flags
  - **QueryExecutor**: Manages parallel API calls with proper error handling
  - **OutputHandler**: Formats and writes results to files and console

## Architectural Principles

- **Modularity**: Each component has a single responsibility and clear interfaces
- **Testability**: Components are designed for easy testing with dependency injection
- **Error Handling**: Comprehensive error system with categorization and helpful messages
- **Configuration**: Flexible, cascading configuration system with sensible defaults
- **Extension Points**: Clear patterns for adding new providers and features

## Adding a New LLM Provider

1. Create a new file in `src/providers/<provider-name>.ts`
2. Implement the `LLMProvider` interface
3. Register the provider in the LLM registry
4. Import the provider in `src/workflow/runThinktank.ts`

Example implementation:

```typescript
import { LLMProvider, LLMResponse, ModelOptions } from '../core/types';
import { registerProvider } from '../core/llmRegistry';

export class NewProvider implements LLMProvider {
  public readonly providerId = 'new-provider';
  
  constructor(private readonly apiKey?: string) {
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
}

// Export a default instance
export const newProvider = new NewProvider();
```