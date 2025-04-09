/**
 * Example showing how to use the new workflow test helpers
 * 
 * This demonstrates how to use the test data factories and scenario helpers
 * to simplify test setup for workflow integration tests.
 */
import { runThinktank } from '../../src/workflow/runThinktank';
import { 
  FileSystemError,
  ApiError
} from '../../src/core/errors';
import * as helpers from '../../src/workflow/runThinktankHelpers';
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

/**
 * Example workflow test suite using the new helpers
 * 
 * This shows how to refactor workflow integration tests to use the
 * new test data factories and scenario helpers for simplified setup.
 */
describe('Example Workflow Test', () => {
  // Set up hooks for resetting mocks
  setupTestHooks();
  
  // Declare mocks at the top level
  let mocks: WorkflowTestMocks;
  
  // Reset all mocks before each test
  beforeEach(() => {
    mocks = setupWorkflowTestEnvironment();
    
    // Set up default spy behavior for _handleWorkflowError
    jest.spyOn(helpers, '_handleWorkflowError').mockImplementation((params: any) => {
      throw params.error;
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
      const options = createRunOptions({ input: 'prompt.txt' });
      
      // Run the workflow
      const result = await runThinktank(options);
      
      // Verify expected interactions with mocks
      expect(mocks.mockFileSystem.writeFile).toHaveBeenCalled();
      expect(mocks.mockSpinner.succeed).toHaveBeenCalled();
      expect(result).toBeDefined();
    });
    
    it('should handle custom options', async () => {
      const options = createRunOptions({ 
        input: 'prompt.txt',
        includeMetadata: true 
      });
      
      await runThinktank(options);
      
      // Specific assertions for this test case
      expect(mocks.mockFileSystem.writeFile).toHaveBeenCalledWith(
        expect.any(String),
        expect.stringContaining('metadata')
      );
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
      const options = createRunOptions({ input: 'missing.txt' });
      
      // Expect error to be thrown
      await expect(runThinktank(options)).rejects.toThrow(FileSystemError);
      
      // Verify error handler was called
      expect(helpers._handleWorkflowError).toHaveBeenCalled();
      expect(mocks.mockSpinner.fail).toHaveBeenCalled();
    });
    
    it('should throw ApiError when API call fails', async () => {
      // Configure mocks for API error scenario
      setupMocksForApiError(
        mocks,
        new ApiError('Rate limit exceeded', { providerId: 'openai' })
      );
      
      // Use factory to create run options
      const options = createRunOptions();
      
      // Expect error to be thrown
      await expect(runThinktank(options)).rejects.toThrow(ApiError);
      
      // Verify error handler was called
      expect(helpers._handleWorkflowError).toHaveBeenCalled();
      expect(mocks.mockSpinner.fail).toHaveBeenCalled();
    });
  });
});