/**
 * Tests for ConcreteFileSystem implementation
 * 
 * These tests verify the behavior of the ConcreteFileSystem class using a
 * virtual file system (memfs) rather than mocking internal dependencies.
 */
import { Stats } from 'fs';
import { setupTestHooks, setupBasicFs, getFs } from '../../../test/setup';
import { ConcreteFileSystem } from '../FileSystem';
import { FileSystemError } from '../errors/types/filesystem';

describe('ConcreteFileSystem', () => {
  setupTestHooks(); // Handles resetting the virtual FS and other test setup
  let fileSystem: ConcreteFileSystem;

  beforeEach(() => {
    fileSystem = new ConcreteFileSystem();
  });

  describe('readFileContent', () => {
    const testFile = '/path/to/file.txt';
    const testContent = 'file content';

    it('should read content from a file', async () => {
      // Set up test file in virtual FS
      setupBasicFs({
        [testFile]: testContent
      });

      // Act
      const result = await fileSystem.readFileContent(testFile);

      // Assert
      expect(result).toBe(testContent);
    });

    it('should pass options to readFileContent', async () => {
      const rawContent = 'content with line endings';
      
      // Set up test file in virtual FS
      setupBasicFs({
        [testFile]: rawContent
      });
      
      // Test with different normalization options
      const normalizedResult = await fileSystem.readFileContent(testFile);
      const rawResult = await fileSystem.readFileContent(testFile, { normalize: false });
      
      // Basic checks
      expect(normalizedResult).toBe(rawContent);
      expect(rawResult).toBe(rawContent);
    });

    it('should throw FileSystemError when file not found', async () => {
      const nonExistentFile = '/nonexistent/file.txt';
      
      // Act & Assert
      await expect(fileSystem.readFileContent(nonExistentFile))
        .rejects.toThrow(FileSystemError);
      
      await expect(fileSystem.readFileContent(nonExistentFile))
        .rejects.toThrow(/not found/);
    });
  });

  describe('writeFile', () => {
    const testFile = '/path/to/output.txt';
    const testContent = 'new content';

    it('should write content to a file', async () => {
      // Act
      await fileSystem.writeFile(testFile, testContent);
      
      // Assert
      const vfs = getFs();
      expect(vfs.existsSync(testFile)).toBe(true);
      expect(vfs.readFileSync(testFile, 'utf8')).toBe(testContent);
    });

    it('should create parent directories if needed', async () => {
      const nestedFile = '/path/to/nested/dir/output.txt';
      
      // Act
      await fileSystem.writeFile(nestedFile, testContent);
      
      // Assert
      const vfs = getFs();
      expect(vfs.existsSync(nestedFile)).toBe(true);
      expect(vfs.existsSync('/path/to/nested/dir')).toBe(true);
      expect(vfs.readFileSync(nestedFile, 'utf8')).toBe(testContent);
    });
  });

  describe('fileExists', () => {
    it('should return true for existing file', async () => {
      const testFile = '/path/to/file.txt';
      
      // Set up test file
      setupBasicFs({
        [testFile]: 'content'
      });
      
      // Act
      const result = await fileSystem.fileExists(testFile);
      
      // Assert
      expect(result).toBe(true);
    });

    it('should return false for non-existent file', async () => {
      const nonExistentFile = '/nonexistent/file.txt';
      
      // Act
      const result = await fileSystem.fileExists(nonExistentFile);
      
      // Assert
      expect(result).toBe(false);
    });
  });

  describe('mkdir', () => {
    const testDir = '/path/to/dir';

    it('should create a directory', async () => {
      // Create parent directories first
      setupBasicFs({
        '/path/to/': ''
      });
      
      // Act
      await fileSystem.mkdir(testDir);
      
      // Assert
      const vfs = getFs();
      expect(vfs.existsSync(testDir)).toBe(true);
      expect(vfs.statSync(testDir).isDirectory()).toBe(true);
    });

    it('should create nested directories with recursive flag', async () => {
      const nestedDir = '/path/to/nested/dir';
      
      // Act
      await fileSystem.mkdir(nestedDir, { recursive: true });
      
      // Assert
      const vfs = getFs();
      expect(vfs.existsSync(nestedDir)).toBe(true);
      expect(vfs.statSync(nestedDir).isDirectory()).toBe(true);
      expect(vfs.existsSync('/path/to/nested')).toBe(true);
    });

    it('should handle directory already exists', async () => {
      // Set up existing directory
      setupBasicFs({
        [testDir]: ''
      });
      
      // Act & Assert - should throw without recursive
      await expect(fileSystem.mkdir(testDir))
        .rejects.toThrow(FileSystemError);
      
      await expect(fileSystem.mkdir(testDir))
        .rejects.toThrow(/already exists/);
      
      // Should NOT throw with recursive
      await expect(fileSystem.mkdir(testDir, { recursive: true }))
        .resolves.not.toThrow();
    });

    it('should throw FileSystemError when parent directory missing', async () => {
      const nestedDir = '/nonexistent/parent/dir';
      
      // Act & Assert
      await expect(fileSystem.mkdir(nestedDir))
        .rejects.toThrow(FileSystemError);
      
      await expect(fileSystem.mkdir(nestedDir))
        .rejects.toThrow(/parent directory does not exist|Failed to create directory/);
    });
  });

  describe('readdir', () => {
    const testDir = '/path/to/dir';
    const testFiles = ['file1.txt', 'file2.txt', 'file3.txt'];

    beforeEach(() => {
      // Set up test directory with files
      const fsSetup: Record<string, string> = {
        [testDir]: ''
      };
      
      testFiles.forEach(file => {
        fsSetup[`${testDir}/${file}`] = `content of ${file}`;
      });
      
      setupBasicFs(fsSetup);
    });

    it('should list files in a directory', async () => {
      // Act
      const files = await fileSystem.readdir(testDir);
      
      // Assert
      expect(files).toHaveLength(testFiles.length);
      testFiles.forEach(file => {
        expect(files).toContain(file);
      });
    });

    it('should throw FileSystemError when directory not found', async () => {
      const nonExistentDir = '/nonexistent/dir';
      
      // Act & Assert
      await expect(fileSystem.readdir(nonExistentDir))
        .rejects.toThrow(FileSystemError);
      
      await expect(fileSystem.readdir(nonExistentDir))
        .rejects.toThrow(/not found|directory/i);
    });
  });

  describe('stat', () => {
    const testFile = '/path/to/file.txt';
    const testDir = '/path/to/dir';

    beforeEach(() => {
      // Set up test file and directory
      setupBasicFs({
        [testFile]: 'content',
        [testDir]: ''
      });
    });

    it('should return stats for a file', async () => {
      // Act
      const stats = await fileSystem.stat(testFile);
      
      // Assert
      expect(stats).toBeInstanceOf(Stats);
      expect(stats.isFile()).toBe(true);
      expect(stats.isDirectory()).toBe(false);
    });

    it('should return stats for a directory', async () => {
      // Act
      const stats = await fileSystem.stat(testDir);
      
      // Assert
      expect(stats).toBeInstanceOf(Stats);
      expect(stats.isDirectory()).toBe(true);
      expect(stats.isFile()).toBe(false);
    });

    it('should throw FileSystemError when path not found', async () => {
      const nonExistentPath = '/nonexistent/path';
      
      // Act & Assert
      await expect(fileSystem.stat(nonExistentPath))
        .rejects.toThrow(FileSystemError);
      
      await expect(fileSystem.stat(nonExistentPath))
        .rejects.toThrow(/not found/);
    });
  });

  describe('access', () => {
    const testFile = '/path/to/file.txt';

    beforeEach(() => {
      // Set up test file
      setupBasicFs({
        [testFile]: 'content'
      });
    });

    it('should resolve for existing file with access', async () => {
      // Act & Assert
      await expect(fileSystem.access(testFile))
        .resolves.not.toThrow();
    });

    it('should throw FileSystemError when file not found', async () => {
      const nonExistentFile = '/nonexistent/file.txt';
      
      // Act & Assert
      await expect(fileSystem.access(nonExistentFile))
        .rejects.toThrow(FileSystemError);
      
      await expect(fileSystem.access(nonExistentFile))
        .rejects.toThrow(/not found/);
    });
  });

  describe('getConfigDir', () => {
    it('should return a valid config directory path', async () => {
      // Act
      const result = await fileSystem.getConfigDir();
      
      // Assert - just check it returns a string ending with thinktank
      expect(typeof result).toBe('string');
      expect(result.endsWith('thinktank')).toBe(true);
    });
  });

  describe('getConfigFilePath', () => {
    it('should return a valid config file path', async () => {
      // Act
      const result = await fileSystem.getConfigFilePath();
      
      // Assert - check it returns a string ending with config.json
      expect(typeof result).toBe('string');
      expect(result.endsWith('config.json')).toBe(true);
    });
  });
});
