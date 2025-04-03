/**
 * Integration tests for CLI interface
 */
import * as runThinktankModule from '../../workflow/runThinktank';
import fs from 'fs/promises';
import dotenv from 'dotenv';

// Mock dependencies
jest.mock('../../workflow/runThinktank');
jest.mock('fs/promises');
jest.mock('dotenv');

// Access the mock
const runThinktank = runThinktankModule.runThinktank as jest.MockedFunction<typeof runThinktankModule.runThinktank>;

describe('CLI Interface', () => {
  // Store original implementations
  const originalConsoleLog = console.log;
  const originalConsoleError = console.error;
  const originalProcessExit = process.exit;
  const originalProcessArgv = process.argv;
  const originalNodeEnv = process.env.NODE_ENV;
  
  beforeEach(() => {
    jest.clearAllMocks();
    
    // Set NODE_ENV for testing
    process.env.NODE_ENV = 'test';
    
    // Mock methods
    console.log = jest.fn();
    console.error = jest.fn();
    process.exit = jest.fn() as any;
    
    // Reset mocked implementations
    (dotenv.config as jest.Mock).mockReturnValue({});
    (fs.access as jest.Mock).mockResolvedValue(undefined);
    runThinktank.mockResolvedValue('Mock result');
  });
  
  afterEach(() => {
    // Restore methods
    console.log = originalConsoleLog;
    console.error = originalConsoleError;
    process.exit = originalProcessExit;
    process.argv = originalProcessArgv;
    process.env.NODE_ENV = originalNodeEnv;
  });
  
  function setMockArgs(args: string[]): void {
    process.argv = ['node', 'thinktank', ...args];
  }
  
  it('should handle prompt file with no group specified (default group)', async () => {
    // Set up test arguments
    setMockArgs(['test-prompt.txt']);
    
    // Import and execute the module
    const { main } = await import('../cli');
    try {
      await main();
    
      // Verify file was checked
      expect(fs.access).toHaveBeenCalledWith('test-prompt.txt');
      
      // Verify runThinktank was called with correct parameters
      expect(runThinktank).toHaveBeenCalledWith({
        input: 'test-prompt.txt',
      });
      
      // In test mode, process.exit should not be called
      expect(process.exit).not.toHaveBeenCalled();
    } catch (error) {
      fail('Should not throw an error: ' + String(error));
    }
  });
  
  it('should handle file not found error correctly with enhanced messages', async () => {
    // Setup mock to throw file not found
    (fs.access as jest.Mock).mockRejectedValue(new Error('File not found'));
    
    // Set up test arguments
    setMockArgs(['nonexistent.txt']);
    
    // Import and execute the module
    const { main } = await import('../cli');
    
    try {
      await main();
      // The function should not reach this point
      fail('Should have thrown an error but did not');
    } catch (error) {
      // Expected to catch an error
      
      // Verify that error handling occurred with the enhanced error message
      expect(console.error).toHaveBeenCalledWith(expect.stringContaining('Error'));
      expect(console.error).toHaveBeenCalledWith(expect.stringContaining('File System'));
      
      // Verify suggestions were displayed
      const calls = (console.error as jest.Mock).mock.calls;
      let suggestionsHeaderShown = false;
      let examplesHeaderShown = false;
      let correctUsageShown = false;
      
      for (const call of calls) {
        const message = call[0];
        if (typeof message === 'string') {
          if (message.includes('Suggestions:')) suggestionsHeaderShown = true;
          if (message.includes('Example commands:')) examplesHeaderShown = true;
          if (message.includes('Correct usage:')) correctUsageShown = true;
        }
      }
      
      expect(suggestionsHeaderShown).toBe(true);
      expect(examplesHeaderShown).toBe(true);
      expect(correctUsageShown).toBe(true);
    }
  });
  
  it('should handle errors from runThinktank correctly', async () => {
    // Setup runThinktank to throw an error
    runThinktank.mockRejectedValue(new Error('Some error'));
    
    // Set up test arguments
    setMockArgs(['test-prompt.txt']);
    
    // Import and execute the module
    const { main } = await import('../cli');
    
    // The function should throw an error
    await expect(main()).rejects.toThrow();
    
    // Verify error handling was invoked
    expect(console.error).toHaveBeenCalledWith(expect.stringContaining('Unexpected error: Some error'));
  });
  
  it('should handle prompt file with specific group', async () => {
    // Set up test arguments - run with 'coding' group
    setMockArgs(['test-prompt.txt', 'coding']);
    
    // Import and execute the module
    const { main } = await import('../cli');
    try {
      await main();
      
      // Verify runThinktank was called with correct parameters
      expect(runThinktank).toHaveBeenCalledWith({
        input: 'test-prompt.txt',
        groupName: 'coding'
      });
      
      // In test mode, process.exit should not be called
      expect(process.exit).not.toHaveBeenCalled();
    } catch (error) {
      fail('Should not throw an error: ' + String(error));
    }
  });
  
  it('should handle prompt file with specific model', async () => {
    // Set up test arguments - run with specific model
    setMockArgs(['test-prompt.txt', 'openai:gpt-4o']);
    
    // Import and execute the module
    const { main } = await import('../cli');
    try {
      await main();
      
      // Verify runThinktank was called with correct parameters
      expect(runThinktank).toHaveBeenCalledWith({
        input: 'test-prompt.txt',
        specificModel: 'openai:gpt-4o'
      });
      
      // In test mode, process.exit should not be called
      expect(process.exit).not.toHaveBeenCalled();
    } catch (error) {
      fail('Should not throw an error: ' + String(error));
    }
  });
  
  it('should validate the provider:model format with helpful errors', async () => {
    // Set up test arguments - invalid model format (missing model ID)
    setMockArgs(['test-prompt.txt', 'openai:']);
    
    // Import and execute the module
    const { main } = await import('../cli');
    
    try {
      await main();
      // The function should not reach this point
      fail('Should have thrown an error but did not');
    } catch (error) {
      // Expected to catch an error
      
      // Verify that enhanced error handling occurred 
      expect(console.error).toHaveBeenCalledWith(expect.stringContaining('Error'));
      
      // Verify runThinktank was not called
      expect(runThinktank).not.toHaveBeenCalled();
    }
  });
  
  it('should handle no arguments by showing help', async () => {
    // Set up test arguments - no arguments
    setMockArgs([]);
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Verify help was shown
    expect(console.error).toHaveBeenCalledWith('Usage:');
    expect(process.exit).toHaveBeenCalledWith(1);
  });

  it('should display help with --help flag and exit with success code', async () => {
    // Clear all mocks before this test 
    jest.resetModules();
    runThinktank.mockClear();
    
    // Set up test arguments - just the help flag
    setMockArgs(['--help']);
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Verify help was shown and runThinktank not called
    expect(console.error).toHaveBeenCalledWith('Usage:');
    expect(runThinktank).not.toHaveBeenCalled();
    expect(process.exit).toHaveBeenCalledWith(0); // Success code when help is explicitly requested
  });

  it('should prioritize --help flag over other arguments', async () => {
    // Clear all mocks before this test
    jest.resetModules();
    runThinktank.mockClear();
    
    // Set up test arguments - help with other valid arguments
    // Test with --help not at the beginning
    setMockArgs(['test-prompt.txt', 'openai:gpt-4o', '--help']);
    
    // Import a fresh copy of the module
    const { main } = await import('../cli');
    await main();
    
    // Verify help was shown and other operations weren't performed
    expect(console.error).toHaveBeenCalledWith('Usage:');
    expect(runThinktank).not.toHaveBeenCalled();
    expect(process.exit).toHaveBeenCalledWith(0);
  });
});