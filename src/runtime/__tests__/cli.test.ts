/**
 * Integration tests for CLI interface
 */
import * as runThinktankModule from '../../templates/runThinktank';
import fs from 'fs/promises';
import dotenv from 'dotenv';
import yargs from 'yargs';

// Mock dependencies
jest.mock('../../templates/runThinktank');
jest.mock('fs/promises');
jest.mock('dotenv');
jest.mock('yargs');

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
    
    // Setup yargs mock
    const mockYargs = {
      usage: jest.fn().mockReturnThis(),
      option: jest.fn().mockReturnThis(),
      help: jest.fn().mockReturnThis(),
      alias: jest.fn().mockReturnThis(),
      version: jest.fn().mockReturnThis(),
      example: jest.fn().mockReturnThis(),
      epilogue: jest.fn().mockReturnThis(),
      parseAsync: jest.fn(),
    };
    (yargs as unknown as jest.Mock).mockReturnValue(mockYargs);
  });
  
  afterEach(() => {
    // Restore methods
    console.log = originalConsoleLog;
    console.error = originalConsoleError;
    process.exit = originalProcessExit;
    process.argv = originalProcessArgv;
  });
  
  it('should handle successful execution', async () => {
    // Setup yargs parsing result
    const yargsInstance = yargs([] as any);
    (yargsInstance.parseAsync as jest.Mock).mockResolvedValue({
      input: 'test-prompt.txt',
      metadata: false,
      'no-color': false,
    });
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Verify execution
    expect(fs.access).toHaveBeenCalledWith('test-prompt.txt');
    expect(runThinktank).toHaveBeenCalledWith({
      input: 'test-prompt.txt',
      configPath: undefined,
      output: undefined,
      models: undefined,
      includeMetadata: false,
      useColors: true,
    });
    
    // Only check that process.exit was called with success code
    expect(process.exit).toHaveBeenCalledWith(0);
  });
  
  it('should handle output file correctly', async () => {
    // Setup yargs parsing result
    const yargsInstance = yargs([] as any);
    (yargsInstance.parseAsync as jest.Mock).mockResolvedValue({
      input: 'test-prompt.txt',
      output: 'result.txt',
      metadata: false,
      'no-color': false,
    });
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Verify execution
    expect(runThinktank).toHaveBeenCalledWith(expect.objectContaining({
      input: 'test-prompt.txt',
      output: 'result.txt',
    }));
    expect(console.log).not.toHaveBeenCalled();
  });
  
  it('should handle file not found error correctly', async () => {
    // Setup mock to throw file not found
    (fs.access as jest.Mock).mockRejectedValue(new Error('File not found'));
    
    // Setup yargs parsing result
    const yargsInstance = yargs([] as any);
    (yargsInstance.parseAsync as jest.Mock).mockResolvedValue({
      input: 'nonexistent.txt',
      metadata: false,
      'no-color': false,
    });
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Verify error handling
    expect(console.error).toHaveBeenCalledWith(expect.stringContaining('Error:'));
    expect(process.exit).toHaveBeenCalledWith(1);
  });
  
  it('should handle errors correctly', async () => {
    // Setup runThinktank to throw an error
    runThinktank.mockRejectedValue(new Error('Some error'));
    
    // Setup yargs parsing result
    const yargsInstance = yargs([] as any);
    (yargsInstance.parseAsync as jest.Mock).mockResolvedValue({
      input: 'test-prompt.txt',
      metadata: false,
      'no-color': false,
    });
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Just verify that error handling was invoked
    expect(console.error).toHaveBeenCalled();
    expect(process.exit).toHaveBeenCalledWith(1);
  });
  
  
  it('should handle models filter correctly', async () => {
    // Setup yargs parsing result
    const yargsInstance = yargs([] as any);
    (yargsInstance.parseAsync as jest.Mock).mockResolvedValue({
      input: 'test-prompt.txt',
      model: ['openai:gpt-4o'],
      metadata: false,
      'no-color': false,
    });
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Verify models are passed correctly
    expect(runThinktank).toHaveBeenCalledWith(
      expect.objectContaining({
        models: ['openai:gpt-4o'],
      })
    );
  });
});