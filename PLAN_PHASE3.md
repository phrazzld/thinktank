# Thinktank Test Suite Refactoring - Phase 3

## Phase 3: Decouple Dependencies with Interfaces

**Objective:** Reduce reliance on mocking by introducing interfaces and dependency injection.

### Steps:

1. **Identify Key Dependencies:**
   - Analyze modules interacting with external systems (APIs, filesystem, etc.)
   - Focus on `runThinktank.ts`, `runThinktankHelpers.ts`, and related functions

2. **Define Interfaces:**
   - Create interfaces for external dependencies:
     ```typescript
     // LLM API client interface
     interface LLMClient {
       generate(prompt: string, modelId: string, options?: ModelOptions): Promise<LLMResponse>;
     }

     // Config manager interface
     interface ConfigManagerInterface {
       loadConfig(): Promise<AppConfig>;
       saveConfig(config: AppConfig): Promise<void>;
       // ... other methods
     }

     // File system interface
     interface FileSystem {
       readDir(path: string): Promise<FileData[]>;
       writeFile(path: string, content: string): Promise<void>;
       // ... other methods
     }
     ```

3. **Implement Dependency Injection:**
   - Modify functions to accept interfaces as parameters:
     ```typescript
     export async function _executeQueries({
       spinner,
       config,
       models,
       combinedContent,
       options,
       llmClient, // Injected dependency
     }: ExecuteQueriesParams & { llmClient: LLMClient }): Promise<ExecuteQueriesResult> {
       // Use llmClient instead of direct API calls
     }
     ```

4. **Update Tests:**
   - Provide mock implementations:
     ```typescript
     const mockClient: LLMClient = { 
       generate: jest.fn().mockResolvedValue({ 
         text: 'mock response', 
         provider: 'mock', 
         modelId: 'test' 
       }) 
     };
     ```
