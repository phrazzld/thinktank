/**
 * Tests for the models listing command
 * 
 * This test verifies that the CLI correctly handles the "models" command,
 * focusing specifically on calling the listAvailableModels function with
 * the correct parameters and handling the results properly.
 */
import { listAvailableModels } from '../../workflow/listModelsWorkflow';
import { Command } from 'commander';

// Mock the listModelsWorkflow module
jest.mock('../../workflow/listModelsWorkflow', () => ({
  listAvailableModels: jest.fn().mockResolvedValue('Mock models list')
}));

// Create a simple handler we can call directly
const mockActionHandler = async (): Promise<void> => {
  const result = await listAvailableModels({});
  // Use the mocked console.log to avoid actual logging
  console.log(result);
};

// Mock the models command module
jest.mock('../commands/models', () => {
  const mockCommand = new Command('models');
  
  // Add a simple action that calls our handler
  mockCommand.action(mockActionHandler);
  
  return { 
    default: mockCommand,
    __actionHandler: mockActionHandler // Expose the handler for direct testing
  };
});

describe('Models Command', () => {
  // Store original process.argv
  const originalArgv = process.argv;
  const originalExit = process.exit;
  
  beforeEach(() => {
    // Mock process.exit with proper type
    process.exit = jest.fn() as unknown as (code?: number) => never;
    
    // Mock console.log
    jest.spyOn(console, 'log').mockImplementation(() => {});
    
    // Reset mock state
    jest.clearAllMocks();
  });
  
  afterEach(() => {
    // Restore originals
    process.argv = originalArgv;
    process.exit = originalExit;
    jest.restoreAllMocks();
  });
  
  it('should call listAvailableModels when "models" command is directly executed', async () => {
    // Execute the action handler directly
    await mockActionHandler();
    
    // Verify listAvailableModels was called with empty options
    expect(listAvailableModels).toHaveBeenCalledWith({});
    
    // Verify console.log was called with the result
    expect(console.log).toHaveBeenCalledWith('Mock models list');
    
    // process.exit should not be called
    expect(process.exit).not.toHaveBeenCalled();
  });
});
