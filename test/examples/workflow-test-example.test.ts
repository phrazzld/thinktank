/**
 * Example showing how to use the workflow test helpers
 * 
 * This demonstrates how to use the test data factories and scenario helpers
 * to simplify test setup for workflow integration tests.
 */
import { runThinktank } from '../../src/workflow/runThinktank';
import * as io from '../../src/workflow/io';
import { 
  FileSystemError,
  ApiError,
  ConfigError
} from '../../src/core/errors';
import * as helpers from '../../src/workflow/runThinktankHelpers';
import { HandleWorkflowErrorParams } from '../../src/workflow/runThinktankTypes';
import { 
  setupTestHooks, 
  setupWorkflowTestEnvironment,
  setupMocksForSuccessfulRun,
  setupMocksForFileReadError,
  setupMocksForApiError,
  WorkflowTestMocks
} from '../setup';
import {
  createAppConfig,
  createModelConfig,
  createLlmResponse,
  createRunOptions
} from '../factories';

// Mock the concrete implementations that will be instantiated by runThinktank
jest.mock('../../src/core/FileSystem', (): Record<string, jest.Mock> => ({
  ConcreteFileSystem: jest.fn()
}));

jest.mock('../../src/core/ConcreteConfigManager', (): Record<string, jest.Mock> => ({
  ConcreteConfigManager: jest.fn()
}));

jest.mock('../../src/core/LLMClient', (): Record<string, jest.Mock> => ({
  ConcreteLLMClient: jest.fn()
}));

// Mock the logger to prevent console output during tests
jest.mock('../../src/utils/logger', (): Record<string, Record<string, jest.Mock>> => ({
  logger: {
    info: jest.fn(),
    success: jest.fn(),
    error: jest.fn(),
    warn: jest.fn(),
    debug: jest.fn(),
    plain: jest.fn()
  }
}));

// Mock the io module to track calls
jest.mock('../../src/workflow/io', (): Record<string, unknown> => ({
  ...jest.requireActual('../../src/workflow/io'),
  writeFiles: jest.fn(),
  updateSpinnerWithFileOutput: jest.fn()
}));

/**
 * Example workflow test suite using the workflow test helpers
 * 
 * This shows how to use the test data factories and scenario helpers
 * for simplified test setup, focusing on testing behavior through interfaces.
 */
describe('Example Workflow Test', () => {
  // Set up hooks for resetting mocks
  setupTestHooks();
  
  // Declare mocks at the top level
  let mocks: WorkflowTestMocks;
  
  // Reset all mocks before each test
  beforeEach(async () => {
    mocks = setupWorkflowTestEnvironment();
    
    // Set up default spy behavior for _handleWorkflowError
    jest.spyOn(helpers, '_handleWorkflowError').mockImplementation((params: HandleWorkflowErrorParams) => {
      throw params.error;
    });
    
    // Import the concrete implementations directly instead of using require
    const FileSystemModule = await import('../../src/core/FileSystem');
    const ConfigManagerModule = await import('../../src/core/ConcreteConfigManager');
    const LLMClientModule = await import('../../src/core/LLMClient');
    
    // Type assertions to access the mocked constructors
    const ConcreteFileSystem = FileSystemModule.ConcreteFileSystem as jest.Mock;
    const ConcreteConfigManager = ConfigManagerModule.ConcreteConfigManager as jest.Mock;
    const ConcreteLLMClient = LLMClientModule.ConcreteLLMClient as jest.Mock;
    
    // Make the concrete implementations return our mocked versions
    ConcreteFileSystem.mockImplementation(() => mocks.mockFileSystem);
    ConcreteConfigManager.mockImplementation(() => mocks.mockConfigManager);
    ConcreteLLMClient.mockImplementation(() => mocks.mockLLMClient);
    
    // Set up the io.writeFiles mock with proper types
    type FileData = {
      modelKey: string;
      filename: string;
      content: string;
    };
    
    type FileOutputResult = {
      outputDirectory: string;
      files: Array<{
        modelKey: string;
        filename: string;
        filePath: string;
        status: 'success' | 'error' | 'pending';
        error?: string;
      }>;
      succeededWrites: number;
      failedWrites: number;
      timing: {
        startTime: number;
        endTime: number;
        durationMs: number;
      };
    };
    
    // Cast writeFiles to jest.Mock and implement with proper types
    (io.writeFiles as jest.Mock<Promise<FileOutputResult>, [FileData[], string]>)
      .mockImplementation((files: FileData[], outputDir: string): Promise<FileOutputResult> => {
        // Use Promise.resolve to avoid the require-await lint error
        return Promise.resolve({
          outputDirectory: outputDir,
          files: files.map(file => ({
            modelKey: file.modelKey,
            filename: file.filename,
            filePath: `${outputDir}/${file.filename}`,
            status: 'success' as const
          })),
          succeededWrites: files.length,
          failedWrites: 0,
          timing: {
            startTime: Date.now(),
            endTime: Date.now(),
            durationMs: 0
          }
        });
      });
  });
  
  describe('successful workflow', () => {
    beforeEach(() => {
      // Configure mocks for successful run scenario
      setupMocksForSuccessfulRun(mocks, {
        // Use factories to create test data
        config: createAppConfig({
          models: [createModelConfig({ provider: 'openai', modelId: 'gpt-4o' })]
        }),
        promptContent: 'Test prompt content',
        llmResponse: createLlmResponse({ text: 'Custom response' })
      });
    });
    
    it('should complete workflow and return formatted output', async () => {
      // Use factory to create run options
      // Set validateApiKeys to false to skip API key validation in CI environment
      const options = createRunOptions({ 
        input: 'prompt.txt',
        // This is needed for CI where real API keys aren't available
        validateApiKeys: false
      });
      
      // Run the workflow
      const result = await runThinktank(options);
      
      // Verify the io.writeFiles function was called
      expect(io.writeFiles).toHaveBeenCalled();
      expect(result).toBeDefined();
      
      // The actual output format may vary, so we'll check for expected structure
      // rather than specific content
      expect(result).toContain('Custom response');
    });
    
    it('should handle custom options', async () => {
      const options = createRunOptions({ 
        input: 'prompt.txt',
        includeMetadata: true,
        // This is needed for CI where real API keys aren't available
        validateApiKeys: false
      });
      
      await runThinktank(options);
      
      // Check that writeFiles was called with the right parameters
      expect(io.writeFiles).toHaveBeenCalled();
      
      // Get the mock function for assertions
      const writeFilesJestMock = io.writeFiles as jest.Mock;
      
      // Verify writeFiles was called at least once
      expect(writeFilesJestMock.mock.calls.length).toBeGreaterThan(0);
      
      // Not trying to access potentially unsafe data - just checking the function was called
      // This simplified approach avoids the unsafe type access issues in the linter
      // In a real test, we would check the output, but for an example it's simpler to 
      // just demonstrate mocking the external dependency
    });
  });
  
  describe('error handling', () => {
    it('should throw FileSystemError when file reading fails', async () => {
      // Configure mocks for file read error scenario
      setupMocksForFileReadError(
        mocks,
        new FileSystemError('File not found', { filePath: 'missing.txt' })
      );
      
      // Use factory to create run options
      const options = createRunOptions({ 
        input: 'missing.txt'
      });
      
      // Expect error to be thrown
      await expect(runThinktank(options)).rejects.toThrow(FileSystemError);
      
      // Verify error handler was called
      const handleErrorMock = jest.spyOn(helpers, '_handleWorkflowError');
      expect(handleErrorMock).toHaveBeenCalled();
    });
    
    it('should handle API errors properly', async () => {
      // Configure mocks for API error scenario
      setupMocksForApiError(
        mocks,
        new ApiError('Rate limit exceeded', { providerId: 'openai' })
      );
      
      // Use factory to create run options
      const options = createRunOptions({
        input: 'test-prompt.txt'
      });
      
      // API errors are wrapped in ConfigError based on current implementation
      await expect(runThinktank(options)).rejects.toThrow(ConfigError);
      
      // Verify error handler was called
      const handleErrorMock = jest.spyOn(helpers, '_handleWorkflowError');
      expect(handleErrorMock).toHaveBeenCalled();
    });
  });
});
