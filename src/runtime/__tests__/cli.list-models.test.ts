/**
 * Tests for the models listing command
 */
import * as listModelsWorkflowModule from '../../templates/listModelsWorkflow';
import dotenv from 'dotenv';

// Mock dependencies
jest.mock('../../templates/listModelsWorkflow');
jest.mock('dotenv');

// Access the mock
const listAvailableModels = listModelsWorkflowModule.listAvailableModels as jest.MockedFunction<typeof listModelsWorkflowModule.listAvailableModels>;

describe('CLI Models Command', () => {
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
  
  it('should handle "models" command', async () => {
    // Set up test arguments
    setMockArgs(['models']);
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Verify execution
    expect(listAvailableModels).toHaveBeenCalledWith({});
    expect(console.log).toHaveBeenCalledWith('Mock models list');
    expect(process.exit).toHaveBeenCalledWith(0);
  });
  
  it('should handle errors from listAvailableModels', async () => {
    // Setup listAvailableModels to throw an error
    listAvailableModels.mockRejectedValue(new Error('List models error'));
    
    // Set up test arguments
    setMockArgs(['models']);
    
    // Import and execute the module
    const { main } = await import('../cli');
    await main();
    
    // Verify error handling
    expect(console.error).toHaveBeenCalledWith(expect.stringContaining('Unexpected error:'));
    expect(process.exit).toHaveBeenCalledWith(1);
  });
});