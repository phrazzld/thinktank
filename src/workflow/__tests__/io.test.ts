/**
 * Tests for the I/O module
 * 
 * These tests verify the functionality of the centralized I/O operations
 * provided by the io.ts module, including:
 * - Directory creation
 * - File writing
 * - Console output formatting
 * - Spinner updates
 */
import path from 'path';
import { 
  createDirectory, 
  writeFiles,
  logFileOutputResult,
  updateSpinnerWithFileOutput,
  FileIOOptions
} from '../io';
import { 
  FileSystem, 
  ConsoleLogger, 
  UISpinner 
} from '../../core/interfaces';
import { FileSystemError } from '../../core/errors/types/filesystem';
import {
  setupBasicFs,
  createMockFileSystem
} from '../../../test/setup/fs';
import { FileData, FileOutputResult } from '../outputHandler';

describe('I/O Module', () => {
  describe('createDirectory', () => {
    let fileSystem: jest.Mocked<FileSystem>;
    
    beforeEach(() => {
      // Reset the virtual filesystem and create a mock FileSystem
      setupBasicFs({});
      fileSystem = createMockFileSystem();
    });
    
    it('should create a directory successfully', async () => {
      // Arrange
      const testDirPath = '/test/output/dir';
      
      // Act
      const result = await createDirectory(testDirPath, fileSystem);
      
      // Assert
      expect(result).toEqual({
        directoryPath: testDirPath,
        success: true
      });
      
      // Verify the directory was created
      expect(await fileSystem.fileExists(testDirPath)).toBe(true);
    });
    
    it('should handle errors without throwing when throwOnError is false', async () => {
      // Arrange
      const testDirPath = '/test/output/dir';
      const mockError = new Error('Permission denied');
      
      // Mock the mkdir method to throw
      fileSystem.mkdir.mockRejectedValueOnce(mockError);
      
      // Act
      const result = await createDirectory(testDirPath, fileSystem);
      
      // Assert
      expect(result).toEqual({
        directoryPath: testDirPath,
        success: false,
        error: 'Permission denied'
      });
    });
    
    it('should throw FileSystemError when throwOnError is true', async () => {
      // Arrange
      const testDirPath = '/test/output/dir';
      const mockError = new Error('Permission denied');
      const options: FileIOOptions = { throwOnError: true };
      
      // Reset mock counters
      fileSystem.mkdir.mockReset();
      
      // Mock the mkdir method to throw for both invocations
      fileSystem.mkdir.mockRejectedValue(mockError);
      
      // Act & Assert
      await expect(createDirectory(testDirPath, fileSystem, options))
        .rejects
        .toThrow(FileSystemError);
      
      // Reset the mock again for the next assertion
      fileSystem.mkdir.mockReset();
      fileSystem.mkdir.mockRejectedValue(mockError);
      
      await expect(createDirectory(testDirPath, fileSystem, options))
        .rejects
        .toThrow(`Failed to create directory: ${testDirPath}`);
    });
  });
  
  describe('writeFiles', () => {
    let fileSystem: jest.Mocked<FileSystem>;
    let files: FileData[];
    const outputDir = '/test/output';
    
    beforeEach(() => {
      // Reset the virtual filesystem and create a mock FileSystem
      setupBasicFs({});
      fileSystem = createMockFileSystem();
      
      // Set up test files
      files = [
        {
          filename: 'file1.md',
          content: '# Content for file 1',
          modelKey: 'openai:gpt-4'
        },
        {
          filename: 'file2.md',
          content: '# Content for file 2',
          modelKey: 'anthropic:claude-3'
        }
      ];
    });
    
    it('should write all files successfully', async () => {
      // Act
      const result = await writeFiles(files, outputDir, fileSystem);
      
      // Assert
      expect(result.succeededWrites).toBe(2);
      expect(result.failedWrites).toBe(0);
      expect(result.files).toHaveLength(2);
      expect(result.files[0].status).toBe('success');
      expect(result.files[1].status).toBe('success');
      
      // Check that files exist
      expect(await fileSystem.fileExists(path.join(outputDir, 'file1.md'))).toBe(true);
      expect(await fileSystem.fileExists(path.join(outputDir, 'file2.md'))).toBe(true);
      
      // Check file contents
      expect(await fileSystem.readFileContent(path.join(outputDir, 'file1.md')))
        .toBe('# Content for file 1');
      expect(await fileSystem.readFileContent(path.join(outputDir, 'file2.md')))
        .toBe('# Content for file 2');
    });
    
    it('should handle output directory creation failure', async () => {
      // Arrange - make mkdir fail for the output directory
      fileSystem.mkdir.mockImplementationOnce(() => Promise.reject(new Error('Permission denied')));
      
      // Act
      const result = await writeFiles(files, outputDir, fileSystem);
      
      // Assert
      expect(result.succeededWrites).toBe(0);
      expect(result.failedWrites).toBe(2);
      expect(result.files).toHaveLength(0);
    });
    
    it('should throw error when output directory creation fails and throwOnError is true', async () => {
      // Arrange
      const options: FileIOOptions = { throwOnError: true };
      
      // Reset mock counters
      fileSystem.mkdir.mockReset();
      
      // Make mkdir fail for the output directory - persistent to affect both calls
      fileSystem.mkdir.mockRejectedValue(new Error('Permission denied'));
      
      // Act & Assert
      await expect(writeFiles(files, outputDir, fileSystem, options))
        .rejects
        .toThrow(FileSystemError);
      
      // Reset mock for second assertion
      fileSystem.mkdir.mockReset();
      fileSystem.mkdir.mockRejectedValue(new Error('Permission denied'));
      
      await expect(writeFiles(files, outputDir, fileSystem, options))
        .rejects
        .toThrow(`Failed to create output directory: ${outputDir}`);
    });
    
    it('should handle individual file write failures', async () => {
      // Arrange - make the second file fail to write
      fileSystem.writeFile.mockImplementationOnce(() => Promise.resolve()); // First file succeeds
      fileSystem.writeFile.mockImplementationOnce(() => Promise.reject(new Error('Disk full'))); // Second file fails
      
      // Act
      const result = await writeFiles(files, outputDir, fileSystem);
      
      // Assert
      expect(result.succeededWrites).toBe(1);
      expect(result.failedWrites).toBe(1);
      expect(result.files).toHaveLength(2);
      expect(result.files[0].status).toBe('success');
      expect(result.files[1].status).toBe('error');
      expect(result.files[1].error).toBe('Disk full');
    });
    
    it('should throw error on file write failure when throwOnError is true', async () => {
      // Arrange
      const options: FileIOOptions = { throwOnError: true };
      
      // First test with first assertion
      fileSystem.mkdir.mockReset();
      fileSystem.writeFile.mockReset();
      
      // We need to allow the first mkdir to succeed (for the output directory)
      // but make the second writeFile fail
      fileSystem.mkdir.mockResolvedValue();
      
      // Make first write succeed, second fail
      let callCount = 0;
      fileSystem.writeFile.mockImplementation(() => {
        callCount++;
        if (callCount === 1) {
          return Promise.resolve();
        } else {
          return Promise.reject(new Error('Disk full'));
        }
      });
      
      // Act & Assert
      await expect(writeFiles(files, outputDir, fileSystem, options))
        .rejects
        .toThrow(FileSystemError);
        
      // Reset for second assertion
      fileSystem.mkdir.mockReset();
      fileSystem.writeFile.mockReset();
      callCount = 0;
      
      // Configure mocks again
      fileSystem.mkdir.mockResolvedValue();
      fileSystem.writeFile.mockImplementation(() => {
        callCount++;
        if (callCount === 1) {
          return Promise.resolve();
        } else {
          return Promise.reject(new Error('Disk full'));
        }
      });
      
      await expect(writeFiles(files, outputDir, fileSystem, options))
        .rejects
        .toThrow(`Failed to write file: ${path.join(outputDir, 'file2.md')}`);
    });
    
    it('should create parent directories for each file if needed', async () => {
      // Arrange - Test nested paths
      files = [
        {
          filename: 'nested/file3.md',
          content: '# Content for nested file',
          modelKey: 'google:gemini-pro'
        }
      ];
      
      // Act
      const result = await writeFiles(files, outputDir, fileSystem);
      
      // Assert
      expect(result.succeededWrites).toBe(1);
      expect(result.failedWrites).toBe(0);
      
      const nestedFilePath = path.join(outputDir, 'nested/file3.md');
      expect(await fileSystem.fileExists(nestedFilePath)).toBe(true);
      expect(await fileSystem.readFileContent(nestedFilePath))
        .toBe('# Content for nested file');
    });
    
    it('should track timing information', async () => {
      // Act
      const result = await writeFiles(files, outputDir, fileSystem);
      
      // Assert
      expect(result.timing).toBeDefined();
      expect(result.timing.startTime).toBeGreaterThan(0);
      expect(result.timing.endTime).toBeGreaterThan(0);
      expect(result.timing.durationMs).toBeGreaterThanOrEqual(0);
    });
  });
  
  describe('logFileOutputResult', () => {
    let mockLogger: jest.Mocked<ConsoleLogger>;
    let fileResult: FileOutputResult;
    
    beforeEach(() => {
      // Create a mock console logger
      mockLogger = {
        error: jest.fn(),
        warn: jest.fn(),
        info: jest.fn(),
        success: jest.fn(),
        debug: jest.fn(),
        plain: jest.fn()
      } as jest.Mocked<ConsoleLogger>;
      
      // Set up a sample file output result
      fileResult = {
        outputDirectory: '/test/output',
        files: [
          {
            modelKey: 'openai:gpt-4',
            filename: 'file1.md',
            filePath: '/test/output/file1.md',
            status: 'success'
          }
        ],
        succeededWrites: 1,
        failedWrites: 0,
        timing: {
          startTime: Date.now() - 100,
          endTime: Date.now(),
          durationMs: 100
        }
      };
    });
    
    it('should log success when all files succeed', () => {
      // Act
      logFileOutputResult(fileResult, mockLogger);
      
      // Assert
      expect(mockLogger.success).toHaveBeenCalledWith(
        expect.stringContaining('Wrote 1 file to /test/output')
      );
      expect(mockLogger.warn).not.toHaveBeenCalled();
      expect(mockLogger.error).not.toHaveBeenCalled();
      expect(mockLogger.info).not.toHaveBeenCalled();
    });
    
    it('should log warning when some files fail', () => {
      // Arrange
      fileResult.files.push({
        modelKey: 'anthropic:claude-3',
        filename: 'file2.md',
        filePath: '/test/output/file2.md',
        status: 'error',
        error: 'Disk full'
      });
      fileResult.succeededWrites = 1;
      fileResult.failedWrites = 1;
      
      // Act
      logFileOutputResult(fileResult, mockLogger);
      
      // Assert
      expect(mockLogger.warn).toHaveBeenCalledWith(
        expect.stringContaining('Wrote 1 file to /test/output (1 failed)')
      );
      expect(mockLogger.success).not.toHaveBeenCalled();
    });
    
    it('should log error when all files fail', () => {
      // Arrange
      fileResult.files = [
        {
          modelKey: 'anthropic:claude-3',
          filename: 'file2.md',
          filePath: '/test/output/file2.md',
          status: 'error',
          error: 'Disk full'
        }
      ];
      fileResult.succeededWrites = 0;
      fileResult.failedWrites = 1;
      
      // Act
      logFileOutputResult(fileResult, mockLogger);
      
      // Assert
      expect(mockLogger.error).toHaveBeenCalledWith(
        expect.stringContaining('Failed to write any files to /test/output')
      );
      expect(mockLogger.success).not.toHaveBeenCalled();
      expect(mockLogger.warn).not.toHaveBeenCalled();
    });
    
    it('should log detailed file information when verbose is true', () => {
      // Arrange
      fileResult.files = [
        {
          modelKey: 'openai:gpt-4',
          filename: 'file1.md',
          filePath: '/test/output/file1.md',
          status: 'success'
        },
        {
          modelKey: 'anthropic:claude-3',
          filename: 'file2.md',
          filePath: '/test/output/file2.md',
          status: 'error',
          error: 'Disk full'
        }
      ];
      fileResult.succeededWrites = 1;
      fileResult.failedWrites = 1;
      
      // Act
      logFileOutputResult(fileResult, mockLogger, { verbose: true });
      
      // Assert
      expect(mockLogger.info).toHaveBeenCalledWith(expect.stringContaining('File details:'));
      expect(mockLogger.info).toHaveBeenCalledWith(expect.stringContaining('file1.md'));
      expect(mockLogger.error).toHaveBeenCalledWith(expect.stringContaining('file2.md: Disk full'));
    });
    
    it('should not log details for a single successful file when verbose is true', () => {
      // Arrange
      fileResult.files = [
        {
          modelKey: 'openai:gpt-4',
          filename: 'file1.md',
          filePath: '/test/output/file1.md',
          status: 'success'
        }
      ];
      fileResult.succeededWrites = 1;
      fileResult.failedWrites = 0;
      
      // Act
      logFileOutputResult(fileResult, mockLogger, { verbose: true });
      
      // Assert - Should not show details for a single successful file
      expect(mockLogger.info).not.toHaveBeenCalled();
    });
  });
  
  describe('updateSpinnerWithFileOutput', () => {
    let mockSpinner: UISpinner & { text: string };
    let fileResult: FileOutputResult;
    
    beforeEach(() => {
      // Create a mock spinner
      mockSpinner = {
        text: '',
        start: jest.fn().mockReturnThis(),
        stop: jest.fn().mockReturnThis(),
        succeed: jest.fn().mockReturnThis(),
        fail: jest.fn().mockReturnThis(),
        warn: jest.fn().mockReturnThis(),
        info: jest.fn().mockReturnThis(),
        setText: jest.fn().mockReturnThis(),
        isSpinning: true
      };
      
      // Set up a sample file output result
      fileResult = {
        outputDirectory: '/test/output',
        files: [
          {
            modelKey: 'openai:gpt-4',
            filename: 'file1.md',
            filePath: '/test/output/file1.md',
            status: 'success'
          }
        ],
        succeededWrites: 1,
        failedWrites: 0,
        timing: {
          startTime: Date.now() - 100,
          endTime: Date.now(),
          durationMs: 100
        }
      };
    });
    
    it('should update spinner text for successful writes', () => {
      // Act
      updateSpinnerWithFileOutput(fileResult, mockSpinner);
      
      // Assert
      expect(mockSpinner.text).toBe('Files written: 1 succeeded');
    });
    
    it('should update spinner text for mixed success/failed writes', () => {
      // Arrange
      fileResult.succeededWrites = 2;
      fileResult.failedWrites = 1;
      
      // Act
      updateSpinnerWithFileOutput(fileResult, mockSpinner);
      
      // Assert
      expect(mockSpinner.text).toBe('Files written: 2 succeeded, 1 failed');
    });
    
    it('should update spinner text for completely failed writes', () => {
      // Arrange
      fileResult.succeededWrites = 0;
      fileResult.failedWrites = 3;
      
      // Act
      updateSpinnerWithFileOutput(fileResult, mockSpinner);
      
      // Assert
      expect(mockSpinner.text).toBe('Files written: 0 succeeded, 3 failed');
    });
    
    it('should work with plain objects that have a text property', () => {
      // Arrange
      const plainObject = { text: 'Original text' };
      
      // Act
      updateSpinnerWithFileOutput(fileResult, plainObject);
      
      // Assert
      expect(plainObject.text).toBe('Files written: 1 succeeded');
    });
  });
});
