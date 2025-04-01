/**
 * Integration tests for CLI interface
 */
import * as runThinktankModule from '../../templates/runThinktank';
import fs from 'fs/promises';
import dotenv from 'dotenv';

// Mock dependencies
jest.mock('../../templates/runThinktank');
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
  
  beforeEach(() => {
    jest.clearAllMocks();
    
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
  });
  
  function setMockArgs(args: string[]): void {
    process.argv = ['node', 'thinktank', ...args];
  }
  
  it('should handle prompt file with no group specified (default group)', async () => {
    // Set up test arguments
    setMockArgs(['test-prompt.txt']);
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Verify file was checked
    expect(fs.access).toHaveBeenCalledWith('test-prompt.txt');
    
    // Verify runThinktank was called with correct parameters
    expect(runThinktank).toHaveBeenCalledWith({
      input: 'test-prompt.txt',
    });
    
    // Check that process.exit was called with success code
    expect(process.exit).toHaveBeenCalledWith(0);
  });
  
  it('should handle file not found error correctly', async () => {
    // Setup mock to throw file not found
    (fs.access as jest.Mock).mockRejectedValue(new Error('File not found'));
    
    // Set up test arguments
    setMockArgs(['nonexistent.txt']);
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Just verify that error handling occurred
    expect(console.error).toHaveBeenCalled();
    expect(process.exit).toHaveBeenCalledWith(1);
  });
  
  it('should handle errors from runThinktank correctly', async () => {
    // Setup runThinktank to throw an error
    runThinktank.mockRejectedValue(new Error('Some error'));
    
    // Set up test arguments
    setMockArgs(['test-prompt.txt']);
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Verify error handling was invoked
    expect(console.error).toHaveBeenCalledWith(expect.stringContaining('Unexpected error: Some error'));
    expect(process.exit).toHaveBeenCalledWith(1);
  });
  
  it('should handle prompt file with specific group', async () => {
    // Set up test arguments - run with 'coding' group
    setMockArgs(['test-prompt.txt', 'coding']);
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Verify runThinktank was called with correct parameters
    expect(runThinktank).toHaveBeenCalledWith({
      input: 'test-prompt.txt',
      groupName: 'coding'
    });
    
    expect(process.exit).toHaveBeenCalledWith(0);
  });
  
  it('should handle prompt file with specific model', async () => {
    // Set up test arguments - run with specific model
    setMockArgs(['test-prompt.txt', 'openai:gpt-4o']);
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Verify runThinktank was called with correct parameters
    expect(runThinktank).toHaveBeenCalledWith({
      input: 'test-prompt.txt',
      specificModel: 'openai:gpt-4o'
    });
    
    expect(process.exit).toHaveBeenCalledWith(0);
  });
  
  it('should validate the provider:model format', async () => {
    // Set up test arguments - invalid model format (missing model ID)
    setMockArgs(['test-prompt.txt', 'openai:']);
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Just verify that error handling occurred
    expect(console.error).toHaveBeenCalled();
    expect(process.exit).toHaveBeenCalledWith(1);
    
    // Verify runThinktank was not called
    expect(runThinktank).not.toHaveBeenCalled();
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
});