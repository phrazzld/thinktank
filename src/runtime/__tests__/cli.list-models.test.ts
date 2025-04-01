/**
 * Tests for the models listing command
 * 
 * This test verifies that the CLI correctly handles the "models" command,
 * focusing specifically on calling the listAvailableModels function with
 * the correct parameters and handling the results properly.
 */
import { listAvailableModels } from '../../templates/listModelsWorkflow';

// Mock the listModelsWorkflow module
jest.mock('../../templates/listModelsWorkflow', () => ({
  listAvailableModels: jest.fn().mockResolvedValue('Mock models list')
}));

describe('Models Command', () => {
  // Store original process.argv
  const originalArgv = process.argv;
  const originalExit = process.exit;
  const originalConsoleLog = console.log;
  
  beforeEach(() => {
    // Mock process.exit
    process.exit = jest.fn() as any;
    
    // Mock console.log
    console.log = jest.fn();
    
    // Reset mock state
    jest.clearAllMocks();
  });
  
  afterEach(() => {
    // Restore originals
    process.argv = originalArgv;
    process.exit = originalExit;
    console.log = originalConsoleLog;
  });
  
  it('should call listAvailableModels when "models" command is used', async () => {
    // Set CLI arguments to simulate "models" command
    process.argv = ['node', 'thinktank', 'models'];
    
    // Require the cli module
    // Using require instead of import to avoid caching issues
    const cli = require('../cli');
    await cli.main();
    
    // Verify listAvailableModels was called with empty options
    expect(listAvailableModels).toHaveBeenCalledWith({});
    
    // Verify console.log was called with the result
    expect(console.log).toHaveBeenCalledWith('Mock models list');
    
    // Verify exit with success code
    expect(process.exit).toHaveBeenCalledWith(0);
  });
});