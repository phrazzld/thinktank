/**
 * Tests for ConcreteFileSystem implementation
 */
import { ConcreteFileSystem } from '../FileSystem';
import * as fileReader from '../../utils/fileReader';
import { FileSystemError } from '../errors/types/filesystem';
import { setupBasicFs, resetFs } from '../../../jest/setupFiles/fs';

// Mock fileReader module
jest.mock('../../utils/fileReader');

describe('ConcreteFileSystem', () => {
  let fileSystem: ConcreteFileSystem;
  const mockedFileReader = fileReader as jest.Mocked<typeof fileReader>;

  beforeEach(() => {
    // Reset mocks and virtual filesystem
    jest.clearAllMocks();
    resetFs();
    
    // Set up test files in virtual filesystem
    setupBasicFs({
      '/path/to/file.txt': 'file content',
      '/path/to/dir/': '',
      '/path/to/dir/file1.txt': 'nested file content',
      '/path/to/dir/file2.txt': 'another nested file'
    });
    
    fileSystem = new ConcreteFileSystem();
  });

  describe('readFileContent', () => {
    it('should delegate to fileReader.readFileContent and return the result', async () => {
      // Arrange
      const filePath = '/path/to/file.txt';
      const expectedContent = 'file content';
      mockedFileReader.readFileContent.mockResolvedValue(expectedContent);

      // Act
      const result = await fileSystem.readFileContent(filePath);

      // Assert
      expect(mockedFileReader.readFileContent).toHaveBeenCalledWith(filePath, undefined);
      expect(result).toBe(expectedContent);
    });

    it('should pass options to fileReader.readFileContent', async () => {
      // Arrange
      const filePath = '/path/to/file.txt';
      const options = { normalize: false };
      mockedFileReader.readFileContent.mockResolvedValue('raw content');

      // Act
      await fileSystem.readFileContent(filePath, options);

      // Assert
      expect(mockedFileReader.readFileContent).toHaveBeenCalledWith(filePath, options);
    });

    it('should wrap FileReadError in FileSystemError for file not found', async () => {
      // Arrange
      const filePath = '/nonexistent/file.txt';
      const originalError = new Error('File not found: /nonexistent/file.txt');
      mockedFileReader.readFileContent.mockRejectedValue(originalError);

      // Act & Assert
      await expect(fileSystem.readFileContent(filePath)).rejects.toThrow(FileSystemError);
      await expect(fileSystem.readFileContent(filePath)).rejects.toThrow(/not found/);
    });

    it('should wrap FileReadError in FileSystemError for permission denied', async () => {
      // Arrange
      const filePath = '/protected/file.txt';
      const originalError = new Error('Permission denied to read file: /protected/file.txt');
      mockedFileReader.readFileContent.mockRejectedValue(originalError);

      // Act & Assert
      await expect(fileSystem.readFileContent(filePath)).rejects.toThrow(FileSystemError);
      await expect(fileSystem.readFileContent(filePath)).rejects.toThrow(/Permission denied/);
    });
  });

  describe('writeFile', () => {
    it('should delegate to fileReader.writeFile', async () => {
      // Arrange
      const filePath = '/path/to/output.txt';
      const content = 'new content';
      mockedFileReader.writeFile.mockResolvedValue(undefined);

      // Act
      await fileSystem.writeFile(filePath, content);

      // Assert
      expect(mockedFileReader.writeFile).toHaveBeenCalledWith(filePath, content);
    });

    it('should wrap errors in FileSystemError for permission denied', async () => {
      // Arrange
      const filePath = '/protected/file.txt';
      const content = 'new content';
      const originalError = new Error('Permission denied writing to file: /protected/file.txt');
      mockedFileReader.writeFile.mockRejectedValue(originalError);

      // Act & Assert
      await expect(fileSystem.writeFile(filePath, content)).rejects.toThrow(FileSystemError);
      await expect(fileSystem.writeFile(filePath, content)).rejects.toThrow(/Permission denied/);
    });
  });

  describe('fileExists', () => {
    it('should delegate to fileReader.fileExists and return the result', async () => {
      // Arrange
      const filePath = '/path/to/file.txt';
      mockedFileReader.fileExists.mockResolvedValue(true);

      // Act
      const result = await fileSystem.fileExists(filePath);

      // Assert
      expect(mockedFileReader.fileExists).toHaveBeenCalledWith(filePath);
      expect(result).toBe(true);
    });
  });

  describe('getConfigDir', () => {
    it('should delegate to fileReader.getConfigDir and return the result', async () => {
      // Arrange
      const configDir = '/home/user/.config/thinktank';
      mockedFileReader.getConfigDir.mockResolvedValue(configDir);

      // Act
      const result = await fileSystem.getConfigDir();

      // Assert
      expect(mockedFileReader.getConfigDir).toHaveBeenCalled();
      expect(result).toBe(configDir);
    });

    it('should wrap errors in FileSystemError', async () => {
      // Arrange
      const originalError = new Error('Failed to access or create config directory');
      mockedFileReader.getConfigDir.mockRejectedValue(originalError);

      // Act & Assert
      await expect(fileSystem.getConfigDir()).rejects.toThrow(FileSystemError);
      await expect(fileSystem.getConfigDir()).rejects.toThrow(/Failed to access or create config directory/);
    });
  });

  describe('getConfigFilePath', () => {
    it('should delegate to fileReader.getConfigFilePath and return the result', async () => {
      // Arrange
      const configPath = '/home/user/.config/thinktank/config.json';
      mockedFileReader.getConfigFilePath.mockResolvedValue(configPath);

      // Act
      const result = await fileSystem.getConfigFilePath();

      // Assert
      expect(mockedFileReader.getConfigFilePath).toHaveBeenCalled();
      expect(result).toBe(configPath);
    });

    it('should wrap errors in FileSystemError', async () => {
      // Arrange
      const originalError = new Error('Failed to determine config file path');
      mockedFileReader.getConfigFilePath.mockRejectedValue(originalError);

      // Act & Assert
      await expect(fileSystem.getConfigFilePath()).rejects.toThrow(FileSystemError);
      await expect(fileSystem.getConfigFilePath()).rejects.toThrow(/Failed to determine config file path/);
    });
  });
});
