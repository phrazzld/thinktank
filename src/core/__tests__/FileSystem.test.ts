/**
 * Tests for FileSystemAdapter implementation
 */
import { FileSystemAdapter } from '../FileSystemAdapter';
import * as fileReader from '../../utils/fileReader';
import { FileReadError } from '../../utils/fileReaderTypes';
import { setupBasicFs, resetFs } from '../../../jest/setupFiles/fs';

// Mock fileReader module
jest.mock('../../utils/fileReader');

describe('FileSystemAdapter', () => {
  let fileSystem: FileSystemAdapter;
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
    
    fileSystem = new FileSystemAdapter();
  });

  describe('readFileContent', () => {
    it('should delegate to fileReader.readFileContent', async () => {
      // Arrange
      const filePath = '/path/to/file.txt';
      const options = { normalize: true };
      const expectedContent = 'file content';
      mockedFileReader.readFileContent.mockResolvedValue(expectedContent);

      // Act
      const result = await fileSystem.readFileContent(filePath, options);

      // Assert
      expect(result).toBe(expectedContent);
      expect(mockedFileReader.readFileContent).toHaveBeenCalledWith(filePath, options);
    });

    it('should propagate errors from fileReader.readFileContent', async () => {
      // Arrange
      const filePath = '/path/to/file.txt';
      const error = new FileReadError('File not found');
      mockedFileReader.readFileContent.mockRejectedValue(error);

      // Act & Assert
      await expect(fileSystem.readFileContent(filePath)).rejects.toThrow(error);
    });
  });

  describe('writeFile', () => {
    it('should delegate to fileReader.writeFile', async () => {
      // Arrange
      const filePath = '/path/to/file.txt';
      const content = 'file content';
      mockedFileReader.writeFile.mockResolvedValue();

      // Act
      await fileSystem.writeFile(filePath, content);

      // Assert
      expect(mockedFileReader.writeFile).toHaveBeenCalledWith(filePath, content);
    });

    it('should propagate errors from fileReader.writeFile', async () => {
      // Arrange
      const filePath = '/path/to/file.txt';
      const content = 'file content';
      const error = new FileReadError('Permission denied');
      mockedFileReader.writeFile.mockRejectedValue(error);

      // Act & Assert
      await expect(fileSystem.writeFile(filePath, content)).rejects.toThrow(error);
    });
  });

  describe('fileExists', () => {
    it('should delegate to fileReader.fileExists', async () => {
      // Arrange
      const filePath = '/path/to/file.txt';
      mockedFileReader.fileExists.mockResolvedValue(true);

      // Act
      const result = await fileSystem.fileExists(filePath);

      // Assert
      expect(result).toBe(true);
      expect(mockedFileReader.fileExists).toHaveBeenCalledWith(filePath);
    });
  });

  describe('mkdir', () => {
    it('should create directory with fs.mkdir', async () => {
      // Arrange
      const dirPath = '/path/to/new-dir';

      // Act
      await fileSystem.mkdir(dirPath, { recursive: true });

      // Assert - verify directory exists by attempting to readdir
      const result = await fileSystem.readdir(dirPath);
      expect(result).toEqual([]);
    });

    it('should handle EEXIST error when recursive is true', async () => {
      // Arrange - '/path/to/dir' already exists in our setup
      const dirPath = '/path/to/dir';
      const options = { recursive: true };

      // Act & Assert
      await expect(fileSystem.mkdir(dirPath, options)).resolves.not.toThrow();
    });

    // Note: We can't easily test permission errors with memfs
  });

  describe('readdir', () => {
    it('should read directory with fs.readdir', async () => {
      // Arrange
      const dirPath = '/path/to/dir';

      // Act
      const result = await fileSystem.readdir(dirPath);

      // Assert - should contain the two files we created
      expect(result).toContain('file1.txt');
      expect(result).toContain('file2.txt');
      expect(result.length).toBe(2);
    });

    it('should throw FileReadError when directory not found', async () => {
      // Arrange
      const dirPath = '/nonexistent/dir';

      // Act & Assert
      await expect(fileSystem.readdir(dirPath)).rejects.toThrow(FileReadError);
      await expect(fileSystem.readdir(dirPath)).rejects.toThrow(/directory/i);
    });
  });

  describe('stat', () => {
    it('should get stats with fs.stat', async () => {
      // Arrange
      const filePath = '/path/to/file.txt';

      // Act
      const result = await fileSystem.stat(filePath);

      // Assert
      expect(result.isFile()).toBe(true);
      expect(result.isDirectory()).toBe(false);
    });

    it('should throw FileReadError when path not found', async () => {
      // Arrange
      const filePath = '/nonexistent/file.txt';

      // Act & Assert
      await expect(fileSystem.stat(filePath)).rejects.toThrow(FileReadError);
      await expect(fileSystem.stat(filePath)).rejects.toThrow(/Path not found/);
    });
  });

  describe('access', () => {
    it('should check access with fs.access', async () => {
      // Arrange
      const filePath = '/path/to/file.txt';

      // Act & Assert - if it doesn't throw, access is allowed
      await expect(fileSystem.access(filePath)).resolves.not.toThrow();
    });

    it('should throw FileReadError when path not found', async () => {
      // Arrange
      const filePath = '/nonexistent/file.txt';

      // Act & Assert
      await expect(fileSystem.access(filePath)).rejects.toThrow(FileReadError);
      await expect(fileSystem.access(filePath)).rejects.toThrow(/Path not found/);
    });

    // Note: We can't easily test permission denied with memfs
  });
  describe('getConfigDir', () => {
    it('should delegate to fileReader.getConfigDir', async () => {
      // Arrange
      const expectedDir = '/home/user/.config/thinktank';
      mockedFileReader.getConfigDir.mockResolvedValue(expectedDir);

      // Act
      const result = await fileSystem.getConfigDir();

      // Assert
      expect(result).toBe(expectedDir);
      expect(mockedFileReader.getConfigDir).toHaveBeenCalled();
    });
  });

  describe('getConfigFilePath', () => {
    it('should delegate to fileReader.getConfigFilePath', async () => {
      // Arrange
      const expectedPath = '/home/user/.config/thinktank/config.json';
      mockedFileReader.getConfigFilePath.mockResolvedValue(expectedPath);

      // Act
      const result = await fileSystem.getConfigFilePath();

      // Assert
      expect(result).toBe(expectedPath);
      expect(mockedFileReader.getConfigFilePath).toHaveBeenCalled();
    });
  });
});
