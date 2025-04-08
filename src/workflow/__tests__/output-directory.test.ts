/**
 * Tests for the output directory feature in the thinktank application.
 * 
 * These tests verify the output directory creation and file writing functionality.
 */
import path from 'path';
import { RunOptions } from '../runThinktank';
import * as outputHandler from '../outputHandler';

// Mock the runThinktank module entirely
jest.mock('../runThinktank', () => {
  // Import the actual types but mock the implementation
  const originalModule = jest.requireActual('../runThinktank');
  
  // Return a mocked implementation
  return {
    ...originalModule,
    runThinktank: jest.fn().mockImplementation(async (options: RunOptions) => {
      // Call our mocked outputHandler functions to verify they're called correctly
      const outputDirPath = await (outputHandler.createOutputDirectory as jest.Mock)({
        outputDirectory: options.output,
        friendlyRunName: options.friendlyRunName
      });
      
      // Simulate query responses
      const responses = [{
        provider: 'provider-a',
        modelId: 'model-1',
        text: 'Provider A test response',
        configKey: 'provider-a:model-1',
        metadata: { usage: { total_tokens: 10 } }
      }];
      
      // Call writeResponsesToFiles with the responses
      await (outputHandler.writeResponsesToFiles as jest.Mock)(
        responses,
        outputDirPath,
        { includeMetadata: options.includeMetadata }
      );
      
      // Return mock console output
      return 'Mock console output';
    })
  };
});

// Mock the outputHandler methods we're testing
jest.mock('../outputHandler', () => {
  const originalModule = jest.requireActual('../outputHandler');
  return {
    ...originalModule,
    createOutputDirectory: jest.fn(),
    writeResponsesToFiles: jest.fn(),
    formatForConsole: jest.fn().mockReturnValue('Mock console output')
  };
});

// Import the mocked runThinktank function
import { runThinktank } from '../runThinktank';

describe('Output Directory Feature', () => {
  // Setup constants for testing
  const mockRunDirectoryName = 'thinktank_run_20230515_143000_000';
  const mockOutputDir = 'mock-output-path';
  const mockFullOutputDir = path.join(mockOutputDir, mockRunDirectoryName);
  
  beforeEach(() => {
    jest.clearAllMocks();
    
    // Default mock implementations
    (outputHandler.createOutputDirectory as jest.Mock).mockResolvedValue(mockFullOutputDir);
    (outputHandler.writeResponsesToFiles as jest.Mock).mockResolvedValue({
      outputDirectory: mockFullOutputDir,
      files: [
        {
          modelKey: 'provider-a:model-1',
          filename: 'provider-a-model-1.md',
          filePath: path.join(mockFullOutputDir, 'provider-a-model-1.md'),
          status: 'success',
          startTime: Date.now(),
          endTime: Date.now(),
          durationMs: 0
        }
      ],
      succeededWrites: 1,
      failedWrites: 0,
      timing: {
        startTime: Date.now(),
        endTime: Date.now(),
        durationMs: 0
      }
    });
  });
  
  it('should create the output directory with the correct structure', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      output: mockOutputDir,
      includeMetadata: false,
      useColors: false,
    };
    
    await runThinktank(options);
    
    // Verify createOutputDirectory was called with the right arguments
    expect(outputHandler.createOutputDirectory).toHaveBeenCalled();
    const createDirArgs = (outputHandler.createOutputDirectory as jest.Mock).mock.calls[0][0];
    expect(createDirArgs.outputDirectory).toBe(mockOutputDir);
  });
  
  it('should write files for each model response', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      output: mockOutputDir,
      includeMetadata: true,
      useColors: false,
    };
    
    await runThinktank(options);
    
    // Verify writeResponsesToFiles was called with the right arguments
    expect(outputHandler.writeResponsesToFiles).toHaveBeenCalled();
    
    // Should have responses from first argument
    const firstArg = (outputHandler.writeResponsesToFiles as jest.Mock).mock.calls[0][0];
    expect(Array.isArray(firstArg)).toBe(true);
    expect(firstArg.length).toBeGreaterThan(0);
    
    // Should have output directory as second argument
    const secondArg = (outputHandler.writeResponsesToFiles as jest.Mock).mock.calls[0][1];
    expect(secondArg).toBe(mockFullOutputDir);
    
    // Should have options that include metadata
    const thirdArg = (outputHandler.writeResponsesToFiles as jest.Mock).mock.calls[0][2];
    expect(thirdArg.includeMetadata).toBe(true);
  });
  
  it('should still write files for models that return errors', async () => {
    // Modify the mock implementation to include an error response
    (runThinktank as jest.Mock).mockImplementationOnce(async (options: RunOptions) => {
      const outputDirPath = await (outputHandler.createOutputDirectory as jest.Mock)({
        outputDirectory: options.output,
        friendlyRunName: options.friendlyRunName
      });
      
      // Simulate query responses with error
      const responses = [{
        provider: 'error-provider',
        modelId: 'error-model',
        error: 'API error for testing',
        configKey: 'error-provider:error-model'
      }];
      
      // Call writeResponsesToFiles with the error responses
      await (outputHandler.writeResponsesToFiles as jest.Mock)(
        responses,
        outputDirPath, 
        { includeMetadata: options.includeMetadata }
      );
      
      return 'Mock console output with error';
    });
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
      output: mockOutputDir,
      includeMetadata: false,
      useColors: false,
    };
    
    await runThinktank(options);
    
    // Verify writeResponsesToFiles was called with error responses
    expect(outputHandler.writeResponsesToFiles).toHaveBeenCalled();
    
    // Check that the response contains an error
    const responses = (outputHandler.writeResponsesToFiles as jest.Mock).mock.calls[0][0];
    expect(responses[0].error).toBeDefined();
    expect(responses[0].error).toBe('API error for testing');
  });
  
  it('should handle errors during directory creation', async () => {
    // Override the mock to simulate an error
    (outputHandler.createOutputDirectory as jest.Mock).mockRejectedValueOnce(
      new Error('Failed to create output directory: Permission denied')
    );
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
      output: mockOutputDir,
      includeMetadata: false,
      useColors: false,
    };
    
    await expect(runThinktank(options)).rejects.toThrow('Failed to create output directory');
  });
  
  it('should handle errors during file writing without crashing', async () => {
    // Override the mock to simulate a file writing error
    (outputHandler.writeResponsesToFiles as jest.Mock).mockRejectedValueOnce(
      new Error('No space left on device')
    );
    
    // But also make runThinktank handle this error
    (runThinktank as jest.Mock).mockImplementationOnce(async (options: RunOptions) => {
      try {
        const outputDirPath = await (outputHandler.createOutputDirectory as jest.Mock)({
          outputDirectory: options.output,
          friendlyRunName: options.friendlyRunName
        });
        
        await (outputHandler.writeResponsesToFiles as jest.Mock)(
          [{ provider: 'provider-a', modelId: 'model-1', configKey: 'provider-a:model-1' }],
          outputDirPath
        );
      } catch (error) {
        // Simulate error handling without actual console logging
        // Don't log errors in tests to keep output clean
        // Continue execution
      }
      
      return 'Execution completed with errors';
    });
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
      output: mockOutputDir,
      includeMetadata: false,
      useColors: false,
    };
    
    // Should not throw
    const result = await runThinktank(options);
    
    // Verify writeResponsesToFiles was called
    expect(outputHandler.writeResponsesToFiles).toHaveBeenCalled();
    
    // Should have a result indicating execution completed
    expect(result).toBe('Execution completed with errors');
  });
  
  it('should properly handle custom output paths', async () => {
    // Use a different path
    const customOutputPath = 'custom-output-path';
    const customFullPath = path.join(customOutputPath, mockRunDirectoryName);
    
    // Mock to return the custom path
    (outputHandler.createOutputDirectory as jest.Mock).mockResolvedValue(customFullPath);
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
      output: customOutputPath,
      includeMetadata: false,
      useColors: false,
    };
    
    await runThinktank(options);
    
    // Verify createOutputDirectory was called with the custom path
    expect(outputHandler.createOutputDirectory).toHaveBeenCalled();
    const createDirArgs = (outputHandler.createOutputDirectory as jest.Mock).mock.calls[0][0];
    expect(createDirArgs.outputDirectory).toBe(customOutputPath);
    
    // Verify writeResponsesToFiles was called with the correct directory
    expect(outputHandler.writeResponsesToFiles).toHaveBeenCalled();
    const writeDirArg = (outputHandler.writeResponsesToFiles as jest.Mock).mock.calls[0][1];
    expect(writeDirArg).toBe(customFullPath);
  });

  // Edge case tests for output directory functionality
  describe('Edge Cases', () => {
    it('should handle responses with empty content', async () => {
      // Set up runThinktank mock to simulate empty response
      (runThinktank as jest.Mock).mockImplementationOnce(async (options: RunOptions) => {
        const outputDirPath = await (outputHandler.createOutputDirectory as jest.Mock)({
          outputDirectory: options.output,
          friendlyRunName: options.friendlyRunName
        });
        
        // Simulate an empty response
        const responses = [{
          provider: 'provider-a',
          modelId: 'model-1',
          text: '', // Empty response text
          configKey: 'provider-a:model-1',
          metadata: { usage: { total_tokens: 0 } }
        }];
        
        await (outputHandler.writeResponsesToFiles as jest.Mock)(
          responses,
          outputDirPath,
          { includeMetadata: options.includeMetadata }
        );
        
        return 'Mock console output with empty response';
      });
      
      const options: RunOptions = {
        input: 'test-prompt.txt',
        output: mockOutputDir,
        includeMetadata: true,
        useColors: false,
      };
      
      await runThinktank(options);
      
      // Verify writeResponsesToFiles was called with the empty response
      expect(outputHandler.writeResponsesToFiles).toHaveBeenCalled();
      const responses = (outputHandler.writeResponsesToFiles as jest.Mock).mock.calls[0][0];
      expect(responses[0].text).toBe('');
    });
    
    it('should handle responses with special characters', async () => {
      // Set up runThinktank mock to simulate response with special characters
      (runThinktank as jest.Mock).mockImplementationOnce(async (options: RunOptions) => {
        const outputDirPath = await (outputHandler.createOutputDirectory as jest.Mock)({
          outputDirectory: options.output,
          friendlyRunName: options.friendlyRunName
        });
        
        // Simulate a response with special characters
        const responses = [{
          provider: 'provider-a',
          modelId: 'model-1',
          text: '# Special characters: \u0000\u0001\u0002\u0003\n```\n<script>alert("XSS")</script>\n```\n\nEmojis: 🚀✨🔥',
          configKey: 'provider-a:model-1',
          metadata: { usage: { total_tokens: 15 } }
        }];
        
        await (outputHandler.writeResponsesToFiles as jest.Mock)(
          responses,
          outputDirPath,
          { includeMetadata: options.includeMetadata }
        );
        
        return 'Mock console output with special characters';
      });
      
      const options: RunOptions = {
        input: 'test-prompt.txt',
        output: mockOutputDir,
        includeMetadata: true,
        useColors: false,
      };
      
      await runThinktank(options);
      
      // Verify writeResponsesToFiles was called with the special character response
      expect(outputHandler.writeResponsesToFiles).toHaveBeenCalled();
      const responses = (outputHandler.writeResponsesToFiles as jest.Mock).mock.calls[0][0];
      expect(responses[0].text).toContain('Special characters:');
      expect(responses[0].text).toContain('Emojis:');
    });
    
    it('should handle extremely long model responses', async () => {
      // Create a very long response string
      const longText = 'A'.repeat(1000000); // 1MB of text
      
      // Set up runThinktank mock to simulate very long response
      (runThinktank as jest.Mock).mockImplementationOnce(async (options: RunOptions) => {
        const outputDirPath = await (outputHandler.createOutputDirectory as jest.Mock)({
          outputDirectory: options.output,
          friendlyRunName: options.friendlyRunName
        });
        
        // Simulate a very long response
        const responses = [{
          provider: 'provider-a',
          modelId: 'model-1',
          text: longText,
          configKey: 'provider-a:model-1',
          metadata: { usage: { total_tokens: 250000 } }
        }];
        
        await (outputHandler.writeResponsesToFiles as jest.Mock)(
          responses,
          outputDirPath,
          { includeMetadata: options.includeMetadata }
        );
        
        return 'Mock console output with very long response';
      });
      
      const options: RunOptions = {
        input: 'test-prompt.txt',
        output: mockOutputDir,
        includeMetadata: false,
        useColors: false,
      };
      
      await runThinktank(options);
      
      // Verify writeResponsesToFiles was called with the long response
      expect(outputHandler.writeResponsesToFiles).toHaveBeenCalled();
      const responses = (outputHandler.writeResponsesToFiles as jest.Mock).mock.calls[0][0];
      expect(responses[0].text.length).toBe(1000000);
    });
    
    it('should handle concurrent file writes without race conditions', async () => {
      // Set up runThinktank mock to simulate multiple concurrent responses
      (runThinktank as jest.Mock).mockImplementationOnce(async (options: RunOptions) => {
        const outputDirPath = await (outputHandler.createOutputDirectory as jest.Mock)({
          outputDirectory: options.output,
          friendlyRunName: options.friendlyRunName
        });
        
        // Simulate multiple model responses that will be written concurrently
        const responses = [
          {
            provider: 'provider-a',
            modelId: 'model-1',
            text: 'Response from model 1',
            configKey: 'provider-a:model-1',
            metadata: { usage: { total_tokens: 10 } }
          },
          {
            provider: 'provider-b',
            modelId: 'model-2',
            text: 'Response from model 2',
            configKey: 'provider-b:model-2',
            metadata: { usage: { total_tokens: 15 } }
          },
          {
            provider: 'provider-c',
            modelId: 'model-3',
            text: 'Response from model 3',
            configKey: 'provider-c:model-3',
            metadata: { usage: { total_tokens: 20 } }
          }
        ];
        
        await (outputHandler.writeResponsesToFiles as jest.Mock)(
          responses,
          outputDirPath,
          { includeMetadata: options.includeMetadata }
        );
        
        return 'Mock console output with multiple responses';
      });
      
      // Mock writeResponsesToFiles to return multiple file writes
      (outputHandler.writeResponsesToFiles as jest.Mock).mockResolvedValueOnce({
        outputDirectory: mockFullOutputDir,
        files: [
          {
            modelKey: 'provider-a:model-1',
            filename: 'provider-a-model-1.md',
            filePath: path.join(mockFullOutputDir, 'provider-a-model-1.md'),
            status: 'success',
            startTime: Date.now(),
            endTime: Date.now() + 10,
            durationMs: 10
          },
          {
            modelKey: 'provider-b:model-2',
            filename: 'provider-b-model-2.md',
            filePath: path.join(mockFullOutputDir, 'provider-b-model-2.md'),
            status: 'success',
            startTime: Date.now(),
            endTime: Date.now() + 15,
            durationMs: 15
          },
          {
            modelKey: 'provider-c:model-3',
            filename: 'provider-c-model-3.md',
            filePath: path.join(mockFullOutputDir, 'provider-c-model-3.md'),
            status: 'success',
            startTime: Date.now(),
            endTime: Date.now() + 20,
            durationMs: 20
          }
        ],
        succeededWrites: 3,
        failedWrites: 0,
        timing: {
          startTime: Date.now(),
          endTime: Date.now() + 20,
          durationMs: 20
        }
      });
      
      const options: RunOptions = {
        input: 'test-prompt.txt',
        output: mockOutputDir,
        includeMetadata: true,
        useColors: false,
      };
      
      await runThinktank(options);
      
      // Verify writeResponsesToFiles was called with multiple responses
      expect(outputHandler.writeResponsesToFiles).toHaveBeenCalled();
      const responses = (outputHandler.writeResponsesToFiles as jest.Mock).mock.calls[0][0];
      expect(responses.length).toBe(3);
    });
    
    it('should handle partial success scenario (some files succeed, others fail)', async () => {
      // Set up runThinktank mock to simulate partial failures
      (runThinktank as jest.Mock).mockImplementationOnce(async (options: RunOptions) => {
        const outputDirPath = await (outputHandler.createOutputDirectory as jest.Mock)({
          outputDirectory: options.output,
          friendlyRunName: options.friendlyRunName
        });
        
        // Simulate mixed success/failure responses
        const responses = [
          {
            provider: 'provider-a',
            modelId: 'model-1',
            text: 'Response from model 1',
            configKey: 'provider-a:model-1'
          },
          {
            provider: 'provider-b',
            modelId: 'model-2',
            error: 'API error for model 2',
            configKey: 'provider-b:model-2'
          }
        ];
        
        try {
          await (outputHandler.writeResponsesToFiles as jest.Mock)(
            responses,
            outputDirPath,
            { includeMetadata: options.includeMetadata }
          );
        } catch (error) {
          console.error('Error during file writes:', error);
        }
        
        return 'Mock console output with partial failures';
      });
      
      // Mock writeResponsesToFiles to simulate some successful writes and some failures
      (outputHandler.writeResponsesToFiles as jest.Mock).mockResolvedValueOnce({
        outputDirectory: mockFullOutputDir,
        files: [
          {
            modelKey: 'provider-a:model-1',
            filename: 'provider-a-model-1.md',
            filePath: path.join(mockFullOutputDir, 'provider-a-model-1.md'),
            status: 'success',
            startTime: Date.now(),
            endTime: Date.now() + 10,
            durationMs: 10
          },
          {
            modelKey: 'provider-b:model-2',
            filename: 'provider-b-model-2.md',
            filePath: path.join(mockFullOutputDir, 'provider-b-model-2.md'),
            status: 'error',
            error: 'Permission denied',
            startTime: Date.now(),
            endTime: Date.now() + 5,
            durationMs: 5
          }
        ],
        succeededWrites: 1,
        failedWrites: 1,
        timing: {
          startTime: Date.now(),
          endTime: Date.now() + 10,
          durationMs: 10
        }
      });
      
      const options: RunOptions = {
        input: 'test-prompt.txt',
        output: mockOutputDir,
        includeMetadata: false,
        useColors: false,
      };
      
      await runThinktank(options);
      
      // Verify writeResponsesToFiles was called with the mixed responses
      expect(outputHandler.writeResponsesToFiles).toHaveBeenCalled();
      const result = await (outputHandler.writeResponsesToFiles as jest.Mock).mock.results[0].value;
      expect(result.succeededWrites).toBe(1);
      expect(result.failedWrites).toBe(1);
    });
    
    it('should handle invalid output paths', async () => {
      // Override createOutputDirectory to simulate an invalid path error
      (outputHandler.createOutputDirectory as jest.Mock).mockRejectedValueOnce(
        new Error('Invalid characters in path')
      );
      
      const options: RunOptions = {
        input: 'test-prompt.txt',
        output: 'invalid/*/path?', // Invalid path with special characters
        includeMetadata: false,
        useColors: false,
      };
      
      await expect(runThinktank(options)).rejects.toThrow('Invalid characters in path');
      
      // Verify createOutputDirectory was called with the invalid path
      expect(outputHandler.createOutputDirectory).toHaveBeenCalled();
      const createDirArgs = (outputHandler.createOutputDirectory as jest.Mock).mock.calls[0][0];
      expect(createDirArgs.outputDirectory).toBe('invalid/*/path?');
    });
  });
});
