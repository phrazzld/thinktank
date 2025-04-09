/**
 * Example of using I/O mock utilities for testing
 * 
 * This example demonstrates how to use the standardized mock implementations
 * for FileSystem, ConsoleLogger, and UISpinner in tests.
 */
/* eslint-disable @typescript-eslint/unbound-method */
import path from 'path';
import { 
  setupTestHooks, 
  createMockFileSystem, 
  createMockConsoleLogger, 
  createMockUISpinner,
  setupBasicFs
} from '../../test/setup';
import { FileSystem, ConsoleLogger, UISpinner } from '../../src/core/interfaces';

/**
 * Example function that uses FileSystem, ConsoleLogger, and UISpinner interfaces.
 * This would typically be in your actual codebase, not in a test file.
 */
async function processFiles(
  directory: string,
  outputFile: string,
  fileSystem: FileSystem,
  logger: ConsoleLogger,
  spinner?: UISpinner
): Promise<boolean> {
  // Use the spinner if provided
  if (spinner) {
    spinner.start(`Processing files in ${directory}...`);
  }

  try {
    // Check if directory exists
    if (!await fileSystem.fileExists(directory)) {
      logger.error(`Directory not found: ${directory}`);
      if (spinner) spinner.fail(`Failed: directory not found`);
      return false;
    }

    // List files in directory
    logger.info(`Reading files from ${directory}`);
    const files = await fileSystem.readdir(directory);
    
    // Process the files
    let combinedContent = '';
    for (const file of files) {
      const filePath = path.join(directory, file);
      const stats = await fileSystem.stat(filePath);
      
      if (stats.isFile()) {
        logger.debug(`Reading file: ${file}`);
        const content = await fileSystem.readFileContent(filePath);
        combinedContent += content + '\n';
      }
    }

    // Save the result
    await fileSystem.writeFile(outputFile, combinedContent);
    logger.success(`Saved combined content to ${outputFile}`);
    
    if (spinner) {
      spinner.succeed(`Processed ${files.length} files`);
    }
    
    return true;
  } catch (error) {
    logger.error(`Failed to process files: ${error instanceof Error ? error.message : String(error)}`);
    if (spinner) {
      spinner.fail(`Processing failed: ${error instanceof Error ? error.message : String(error)}`);
    }
    return false;
  }
}

// Test suite using mock utilities
describe('processFiles function', () => {
  // Setup and cleanup for each test
  setupTestHooks();
  
  // Declare mock objects
  let mockFileSystem: jest.Mocked<FileSystem>;
  let mockLogger: jest.Mocked<ConsoleLogger>;
  let mockSpinner: jest.Mocked<UISpinner>;
  
  beforeEach(() => {
    // Create mock objects for each test
    mockFileSystem = createMockFileSystem();
    mockLogger = createMockConsoleLogger();
    mockSpinner = createMockUISpinner();
    
    // Set up some files in the virtual filesystem
    setupBasicFs({
      '/test/file1.txt': 'Content of file 1',
      '/test/file2.txt': 'Content of file 2',
      '/test/subdir/file3.txt': 'Content of file 3'
    });
  });
  
  it('should successfully process files when directory exists', async () => {
    // Call the function being tested with our mocks
    const result = await processFiles(
      '/test',
      '/test/output.txt',
      mockFileSystem,
      mockLogger,
      mockSpinner
    );
    
    // Verify the result
    expect(result).toBe(true);
    
    // Verify the mock interactions
    expect(mockLogger.info).toHaveBeenCalledWith('Reading files from /test');
    expect(mockLogger.success).toHaveBeenCalledWith('Saved combined content to /test/output.txt');
    
    // Verify spinner interactions
    expect(mockSpinner.start).toHaveBeenCalledWith('Processing files in /test...');
    expect(mockSpinner.succeed).toHaveBeenCalled();
    
    // Verify filesystem operations
    expect(mockFileSystem.readdir).toHaveBeenCalledWith('/test');
    expect(mockFileSystem.writeFile).toHaveBeenCalledWith(
      '/test/output.txt',
      expect.stringContaining('Content of file 1')
    );
    
    // Verify the actual file was written to the virtual filesystem
    const fileContent = await mockFileSystem.readFileContent('/test/output.txt');
    expect(fileContent).toContain('Content of file 1');
    expect(fileContent).toContain('Content of file 2');
  });
  
  it('should handle non-existent directories', async () => {
    // Call with a non-existent directory
    const result = await processFiles(
      '/nonexistent',
      '/test/output.txt',
      mockFileSystem,
      mockLogger,
      mockSpinner
    );
    
    // Verify the result is false
    expect(result).toBe(false);
    
    // Verify error logging
    expect(mockLogger.error).toHaveBeenCalledWith('Directory not found: /nonexistent');
    
    // Verify spinner failure
    expect(mockSpinner.fail).toHaveBeenCalledWith('Failed: directory not found');
    
    // Verify file was not written
    expect(mockFileSystem.writeFile).not.toHaveBeenCalled();
  });
  
  it('should work without a spinner', async () => {
    // Call without passing a spinner
    const result = await processFiles(
      '/test',
      '/test/output.txt',
      mockFileSystem,
      mockLogger
    );
    
    // Verify the result
    expect(result).toBe(true);
    
    // Verify logger still worked
    expect(mockLogger.info).toHaveBeenCalled();
    expect(mockLogger.success).toHaveBeenCalled();
  });
});
