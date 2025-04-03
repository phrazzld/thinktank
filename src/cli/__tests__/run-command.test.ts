/**
 * Integration tests for the run command workflow
 * 
 * These tests verify that the CLI's run command correctly
 * integrates with the runThinktank module to process prompts
 * through LLM models.
 */
import * as runThinktankModule from '../../workflow/runThinktank';
import * as fileReader from '../../utils/fileReader';
import * as configManager from '../../core/configManager';
import fs from 'fs/promises';

// Mock dependencies
jest.mock('../../workflow/runThinktank');
jest.mock('../../utils/fileReader');
jest.mock('../../core/configManager');
jest.mock('fs/promises');

// Access the mocks
const runThinktank = runThinktankModule.runThinktank as jest.MockedFunction<typeof runThinktankModule.runThinktank>;
const fileExists = fileReader.fileExists as jest.MockedFunction<typeof fileReader.fileExists>;

describe('Run Command Integration', () => {
  const originalConsoleLog = console.log;
  const originalConsoleError = console.error;
  const originalProcessExit = process.exit;
  const originalProcessArgv = process.argv;
  const originalNodeEnv = process.env.NODE_ENV;
  
  beforeEach(() => {
    jest.clearAllMocks();
    jest.resetModules();
    
    // Set NODE_ENV for testing
    process.env.NODE_ENV = 'test';
    
    // Mock console methods
    console.log = jest.fn();
    console.error = jest.fn();
    
    // Mock process.exit with proper type
    process.exit = jest.fn() as unknown as (code?: number) => never;
    
    // Setup default mock behavior
    runThinktank.mockResolvedValue('Mock result content');
    fileExists.mockResolvedValue(true);
    
    // Mock configManager
    (configManager.loadConfig as jest.Mock).mockResolvedValue({
      models: [
        { provider: 'openai', modelId: 'gpt-4o', enabled: true },
        { provider: 'anthropic', modelId: 'claude-3', enabled: true }
      ],
      groups: {
        default: {
          systemPrompt: { text: 'You are a helpful assistant.' },
          models: [{ provider: 'openai', modelId: 'gpt-4o' }]
        },
        coding: {
          systemPrompt: { text: 'You are a coding assistant.' },
          models: [{ provider: 'anthropic', modelId: 'claude-3' }]
        }
      }
    });
    
    // Mock fs access to make the prompt file appear to exist
    (fs.access as jest.Mock).mockResolvedValue(undefined);
  });
  
  afterEach(() => {
    // Restore originals
    console.log = originalConsoleLog;
    console.error = originalConsoleError;
    process.exit = originalProcessExit;
    process.argv = originalProcessArgv;
    process.env.NODE_ENV = originalNodeEnv;
  });
  
  describe('CLI Interface', () => {
    it('should call runThinktank when a valid input file is specified', async () => {
      // Import the CLI module
      const { main } = await import('../cli');
      
      // Set command-line arguments programmatically for the test
      process.argv = ['node', 'thinktank', 'test-prompt.txt'];
      
      // Run the CLI main function
      await main();
      
      // Verify runThinktank was called with expected parameters
      expect(runThinktank).toHaveBeenCalledWith({
        input: 'test-prompt.txt'
      });
      
      // In test mode, process.exit should not be called
      expect(process.exit).not.toHaveBeenCalled();
    });
    
    it('should display help when no arguments are provided', async () => {
      // Import the CLI module
      const { main } = await import('../cli');
      
      // Set empty command-line arguments 
      process.argv = ['node', 'thinktank'];
      
      // Run the CLI main function
      await main();
      
      // Verify help was shown (help shows on console.error)
      expect(console.error).toHaveBeenCalled();
      expect(process.exit).toHaveBeenCalledWith(1); // Missing required args = error exit code
    });
    
    it('should handle file not found errors gracefully', async () => {
      // Setup fs access to simulate file not found
      (fs.access as jest.Mock).mockRejectedValue(new Error('ENOENT: File not found'));
      
      // Import the CLI module
      const { main } = await import('../cli');
      
      // Set command-line arguments for a non-existent file
      process.argv = ['node', 'thinktank', 'nonexistent.txt'];
      
      // The function should throw an error
      await expect(main()).rejects.toThrow();
      
      // Verify error handling was invoked
      expect(console.error).toHaveBeenCalled();
    });
  });
});