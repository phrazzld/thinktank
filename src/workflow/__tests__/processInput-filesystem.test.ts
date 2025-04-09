/**
 * Unit tests for the refactored _processInput helper function with FileSystem interface
 */
import { _processInput } from '../runThinktankHelpers';
import * as inputHandler from '../inputHandler';
import { FileSystemError } from '../../core/errors';
import { InputSourceType } from '../inputHandler';
import * as fileReader from '../../utils/fileReader';
import { ContextFileResult } from '../../utils/fileReaderTypes';
import {
  setupTestHooks,
  setupWorkflowTestEnvironment,
  WorkflowTestMocks,
} from '../../../test/setup';
import { createMockSpinner } from './oraTestHelper';

// Mock dependencies
jest.mock('../inputHandler');
jest.mock('../../utils/fileReader');

describe('_processInput Helper with FileSystem Interface', () => {
  // Set up test hooks
  setupTestHooks();

  // Create mocks for each test
  let mocks: WorkflowTestMocks;
  // Create a mock spinner that implements Ora interface
  const mockSpinner = createMockSpinner();

  // Reset all mocks before each test
  beforeEach(() => {
    // Set up workflow test environment with all required mocks
    mocks = setupWorkflowTestEnvironment();

    // Reset mockSpinner state
    jest.resetAllMocks();
  });

  it('should successfully process input using the FileSystem interface', async () => {
    // Setup mocks
    (inputHandler.processInput as jest.Mock).mockImplementation(async options => {
      // Verify fileSystem is passed through to processInput
      expect(options.fileSystem).toBe(mocks.mockFileSystem);

      return {
        content: 'Test prompt content',
        sourceType: InputSourceType.FILE,
        sourcePath: '/path/to/file.txt',
        metadata: {
          processingTimeMs: 5,
          originalLength: 20,
          finalLength: 20,
          normalized: true,
        },
      };
    });

    // Call the function with fileSystem
    const result = await _processInput({
      spinner: mockSpinner,
      input: 'file.txt',
      fileSystem: mocks.mockFileSystem,
    });

    // Verify the important properties
    expect(result.inputResult.content).toBe('Test prompt content');
    expect(result.inputResult.sourceType).toBe(InputSourceType.FILE);
    expect(result.inputResult.sourcePath).toBe('/path/to/file.txt');

    // Verify inputHandler.processInput was called with fileSystem
    expect(inputHandler.processInput).toHaveBeenCalledWith({
      input: 'file.txt',
      fileSystem: mocks.mockFileSystem,
    });
  });

  it('should process context paths using the FileSystem interface', async () => {
    // Setup mocks for input
    (inputHandler.processInput as jest.Mock).mockResolvedValue({
      content: 'Main prompt content',
      sourceType: InputSourceType.FILE,
      sourcePath: '/path/to/prompt.txt',
      metadata: {
        processingTimeMs: 5,
        originalLength: 20,
        finalLength: 20,
        normalized: true,
      },
    });

    // Mock context files
    const mockContextFiles: ContextFileResult[] = [
      {
        path: '/path/to/context1.js',
        content: 'Context file 1 content',
        error: null,
      },
      {
        path: '/path/to/context2.md',
        content: 'Context file 2 content',
        error: null,
      },
    ];

    // Mock readContextPaths to verify fileSystem is passed through
    (fileReader.readContextPaths as jest.Mock).mockImplementation(async (_paths, fs) => {
      // Verify fileSystem is passed to readContextPaths
      expect(fs).toBe(mocks.mockFileSystem);
      return mockContextFiles;
    });

    // Mock formatCombinedInput
    const formattedContent =
      '# CONTEXT DOCUMENTS\n\n## File: /path/to/context1.js\n```javascript\nContext file 1 content\n```\n\n## File: /path/to/context2.md\n```markdown\nContext file 2 content\n```\n\n# USER PROMPT\n\nMain prompt content';
    (fileReader.formatCombinedInput as jest.Mock).mockReturnValue(formattedContent);

    // Call function with context paths and fileSystem
    const result = await _processInput({
      spinner: mockSpinner,
      input: 'prompt.txt',
      contextPaths: ['context1.js', 'context2.md'],
      fileSystem: mocks.mockFileSystem,
    });

    // Verify the result
    expect(result.inputResult.content).toBe(formattedContent);
    expect(result.contextFiles).toBeDefined();
    expect(result.contextFiles?.length).toBe(2);

    // Verify fileSystem was passed to readContextPaths
    expect(fileReader.readContextPaths).toHaveBeenCalledWith(
      ['context1.js', 'context2.md'],
      mocks.mockFileSystem
    );
  });

  it('should handle file system errors when processing context paths', async () => {
    // Setup mocks for input handler
    (inputHandler.processInput as jest.Mock).mockResolvedValue({
      content: 'Main prompt content',
      sourceType: InputSourceType.FILE,
      sourcePath: '/path/to/prompt.txt',
      metadata: {
        processingTimeMs: 5,
        originalLength: 20,
        finalLength: 20,
        normalized: true,
      },
    });

    // Mock readContextPaths to throw error
    const fsError = new FileSystemError('Failed to read context files');
    (fileReader.readContextPaths as jest.Mock).mockRejectedValue(fsError);

    // Function should throw FileSystemError
    await expect(
      _processInput({
        spinner: mockSpinner,
        input: 'prompt.txt',
        contextPaths: ['context1.js'],
        fileSystem: mocks.mockFileSystem,
      })
    ).rejects.toThrow(FileSystemError);

    // Verify fileSystem was passed to readContextPaths before error
    expect(fileReader.readContextPaths).toHaveBeenCalledWith(['context1.js'], mocks.mockFileSystem);
  });
});
