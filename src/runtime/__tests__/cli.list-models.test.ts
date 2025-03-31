/**
 * Tests for the list-models CLI command
 */
import * as listModelsWorkflowModule from '../../templates/listModelsWorkflow';
// fs is not needed for list-models tests
import dotenv from 'dotenv';
import yargs from 'yargs';

// Mock dependencies
jest.mock('../../templates/listModelsWorkflow');
jest.mock('dotenv');
jest.mock('yargs');

// Access the mock
const listAvailableModels = listModelsWorkflowModule.listAvailableModels as jest.MockedFunction<typeof listModelsWorkflowModule.listAvailableModels>;

describe('CLI List-Models Command', () => {
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
    listAvailableModels.mockResolvedValue('Mock models list');
    
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
  
  it('should handle list-models command with no options', async () => {
    // Setup yargs parsing result for list-models command
    const yargsInstance = yargs([] as any);
    (yargsInstance.parseAsync as jest.Mock).mockResolvedValue({
      _: ['list-models'],
      '$0': 'thinktank',
    });
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Verify execution
    expect(listAvailableModels).toHaveBeenCalledWith({});
    expect(console.log).toHaveBeenCalledWith('Mock models list');
    expect(process.exit).toHaveBeenCalledWith(0);
  });
  
  it('should handle provider filter option correctly', async () => {
    // Setup yargs parsing result with provider filter
    const yargsInstance = yargs([] as any);
    (yargsInstance.parseAsync as jest.Mock).mockResolvedValue({
      _: ['list-models'],
      '$0': 'thinktank',
      provider: 'anthropic',
    });
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Verify provider is passed correctly
    expect(listAvailableModels).toHaveBeenCalledWith({
      provider: 'anthropic',
    });
    expect(console.log).toHaveBeenCalledWith('Mock models list');
  });
  
  it('should handle config option correctly', async () => {
    // Setup yargs parsing result with config path
    const yargsInstance = yargs([] as any);
    (yargsInstance.parseAsync as jest.Mock).mockResolvedValue({
      _: ['list-models'],
      '$0': 'thinktank',
      config: 'custom-config.json',
    });
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Verify config path is passed correctly
    expect(listAvailableModels).toHaveBeenCalledWith({
      config: 'custom-config.json',
    });
    expect(console.log).toHaveBeenCalledWith('Mock models list');
  });
  
  it('should handle errors correctly', async () => {
    // Setup listAvailableModels to throw an error
    listAvailableModels.mockRejectedValue(new Error('List models error'));
    
    // Setup yargs parsing result
    const yargsInstance = yargs([] as any);
    (yargsInstance.parseAsync as jest.Mock).mockResolvedValue({
      _: ['list-models'],
      '$0': 'thinktank',
    });
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Verify error handling
    expect(console.error).toHaveBeenCalledWith(expect.stringContaining('Unexpected error:'));
    expect(process.exit).toHaveBeenCalledWith(1);
  });
});