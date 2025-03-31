/**
 * Integration tests for CLI interface
 */
import * as runthinktankModule from '../../templates/runthinktank';
import fs from 'fs/promises';
import dotenv from 'dotenv';
import yargs from 'yargs';

// Mock dependencies
jest.mock('../../templates/runthinktank');
jest.mock('fs/promises');
jest.mock('dotenv');
jest.mock('yargs');

// Access the mock
const runthinktank = runthinktankModule.runthinktank as jest.MockedFunction<typeof runthinktankModule.runthinktank>;

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
    runthinktank.mockResolvedValue('Mock result');
    
    // Setup yargs mock with command
    const mockYargs = {
      usage: jest.fn().mockReturnThis(),
      option: jest.fn().mockReturnThis(),
      command: jest.fn().mockImplementation((_cmd, _desc, builder, _handler) => {
        // Call the builder with a mock yargs instance that returns itself
        if (builder) {
          const mockBuilder = {
            option: jest.fn().mockReturnThis(),
            example: jest.fn().mockReturnThis(),
          };
          builder(mockBuilder as any);
        }
        return mockYargs;
      }),
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
    // Setup yargs parsing result for default command
    const yargsInstance = yargs([] as any);
    (yargsInstance.parseAsync as jest.Mock).mockResolvedValue({
      _: [], // Default command (empty array)
      '$0': 'thinktank',
      input: 'test-prompt.txt',
      metadata: false,
      'no-color': false,
      // These need to be included to match the structure expected by the CLI
      command: '$0', // Default command identifier
    });
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Verify execution - note that command structure means it won't call fs.access directly
    expect(runthinktank).toHaveBeenCalledWith({
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
    // Setup yargs parsing result for default command with output
    const yargsInstance = yargs([] as any);
    (yargsInstance.parseAsync as jest.Mock).mockResolvedValue({
      _: [],
      '$0': 'thinktank',
      input: 'test-prompt.txt',
      output: 'result.txt',
      metadata: false,
      'no-color': false,
      command: '$0', // Default command identifier
    });
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Verify execution
    expect(runthinktank).toHaveBeenCalledWith(expect.objectContaining({
      input: 'test-prompt.txt',
      output: 'result.txt',
    }));
    expect(console.log).not.toHaveBeenCalled();
  });
  
  it('should handle file not found error correctly', async () => {
    // Setup mock to throw file not found
    (fs.access as jest.Mock).mockRejectedValue(new Error('File not found'));
    
    // Setup yargs parsing result for default command with nonexistent file
    const yargsInstance = yargs([] as any);
    (yargsInstance.parseAsync as jest.Mock).mockResolvedValue({
      _: [],
      '$0': 'thinktank',
      input: 'nonexistent.txt',
      metadata: false,
      'no-color': false,
      command: '$0', // Default command identifier
    });
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Verify error handling - just check that an error was shown and exit was called
    expect(console.error).toHaveBeenCalled();
    expect(process.exit).toHaveBeenCalledWith(1);
  });
  
  it('should handle errors correctly', async () => {
    // Setup runthinktank to throw an error
    runthinktank.mockRejectedValue(new Error('Some error'));
    
    // Setup yargs parsing result for default command with error in runthinktank
    const yargsInstance = yargs([] as any);
    (yargsInstance.parseAsync as jest.Mock).mockResolvedValue({
      _: [],
      '$0': 'thinktank',
      input: 'test-prompt.txt',
      metadata: false,
      'no-color': false,
      command: '$0', // Default command identifier
    });
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Just verify that error handling was invoked
    expect(console.error).toHaveBeenCalled();
    expect(process.exit).toHaveBeenCalledWith(1);
  });
  
  
  it('should handle models filter correctly', async () => {
    // Setup yargs parsing result for default command with model filter
    const yargsInstance = yargs([] as any);
    (yargsInstance.parseAsync as jest.Mock).mockResolvedValue({
      _: [],
      '$0': 'thinktank',
      input: 'test-prompt.txt',
      model: ['openai:gpt-4o'],
      metadata: false,
      'no-color': false,
      command: '$0', // Default command identifier
    });
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Verify models are passed correctly
    expect(runthinktank).toHaveBeenCalledWith(
      expect.objectContaining({
        models: ['openai:gpt-4o'],
      })
    );
  });
});