import {
  createVirtualFs,
  resetVirtualFs,
  getVirtualFs,
  mockFsModules,
  addVirtualGitignoreFile,
  normalizePathForMemfs,
  createFsError,
  createMockStats,
  createMockDirent,
} from '../virtualFsUtils';

// Set up mocks for fs modules outside of any test
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Import fs modules after mocking
import fs from 'fs';
import fsPromises from 'fs/promises';

describe('virtualFsUtils', () => {
  beforeEach(() => {
    // Reset the virtual filesystem before each test
    resetVirtualFs();
  });

  describe('normalizePathForMemfs', () => {
    it('adds leading slash to paths', () => {
      expect(normalizePathForMemfs('path/to/file')).toBe('/path/to/file');
      expect(normalizePathForMemfs('/path/to/file')).toBe('/path/to/file');
    });

    it('converts backslashes to forward slashes', () => {
      expect(normalizePathForMemfs('path\\to\\file')).toBe('/path/to/file');
    });

    it('handles Windows-style paths with drive letters', () => {
      expect(normalizePathForMemfs('C:\\path\\to\\file')).toBe('/C:/path/to/file');
      expect(normalizePathForMemfs('C:/path/to/file')).toBe('/C:/path/to/file');
    });

    it('normalizes directory traversal', () => {
      expect(normalizePathForMemfs('/path/to/../other/file')).toBe('/path/other/file');
      expect(normalizePathForMemfs('path/./to/file')).toBe('/path/to/file');
    });

    it('handles empty paths', () => {
      expect(normalizePathForMemfs('')).toBe('/');
      expect(normalizePathForMemfs(null as unknown as string)).toBe('/');
      expect(normalizePathForMemfs(undefined as unknown as string)).toBe('/');
    });

    it('handles dot paths', () => {
      expect(normalizePathForMemfs('.')).toBe('/.');
      expect(normalizePathForMemfs('./')).toBe('/');
    });
  });

  describe('createVirtualFs', () => {
    it('should create a virtual filesystem with specified structure', async () => {
      // Create a virtual filesystem structure
      createVirtualFs({
        '/test/file.txt': 'test content',
        '/test/dir/nested.txt': 'nested content',
      });

      // Test reading from the virtual filesystem
      expect(await fsPromises.readFile('/test/file.txt', 'utf-8')).toBe('test content');
      expect(await fsPromises.readFile('/test/dir/nested.txt', 'utf-8')).toBe('nested content');

      // Test directory structure
      const dirContents = await fsPromises.readdir('/test');
      expect(dirContents).toContain('file.txt');
      expect(dirContents).toContain('dir');

      // Test file stats
      const stats = await fsPromises.stat('/test/file.txt');
      expect(stats.isFile()).toBe(true);
      expect(stats.isDirectory()).toBe(false);
    });

    it('should reset the filesystem if called multiple times', async () => {
      // Create initial structure
      createVirtualFs({
        '/test/file1.txt': 'content 1',
      });

      // Create new structure (should replace the first one)
      createVirtualFs({
        '/test/file2.txt': 'content 2',
      });

      // First file should no longer exist
      await expect(fsPromises.access('/test/file1.txt')).rejects.toThrow();

      // Second file should exist
      expect(await fsPromises.readFile('/test/file2.txt', 'utf-8')).toBe('content 2');
    });
  });

  describe('resetVirtualFs', () => {
    it('should clear the virtual filesystem', async () => {
      // Create a virtual filesystem structure
      createVirtualFs({
        '/test/file.txt': 'test content',
      });

      // Verify file exists
      expect(await fsPromises.readFile('/test/file.txt', 'utf-8')).toBe('test content');

      // Reset filesystem
      resetVirtualFs();

      // File should no longer exist
      await expect(fsPromises.access('/test/file.txt')).rejects.toThrow();
    });
  });

  describe('getVirtualFs', () => {
    it('should return the current virtual filesystem instance', () => {
      // Get the virtual filesystem
      const virtualFs = getVirtualFs();

      // Should have expected methods
      expect(typeof virtualFs.readFileSync).toBe('function');
      expect(typeof virtualFs.writeFileSync).toBe('function');
      expect(typeof virtualFs.mkdirSync).toBe('function');
    });

    it('should allow direct manipulation of the filesystem', () => {
      // Get virtual filesystem
      const virtualFs = getVirtualFs();

      // Create file directly
      virtualFs.mkdirSync('/direct', { recursive: true });
      virtualFs.writeFileSync('/direct/test.txt', 'direct content');

      // Read using fs module (should be mocked to use virtual fs)
      expect(fs.readFileSync('/direct/test.txt', 'utf-8')).toBe('direct content');
    });
  });

  describe('error simulation', () => {
    it('should correctly simulate ENOENT errors', async () => {
      // Try to access a file that doesn't exist
      await expect(fsPromises.access('/nonexistent.txt')).rejects.toThrow();
      await expect(fsPromises.access('/nonexistent.txt')).rejects.toHaveProperty('code', 'ENOENT');
    });

    it('should handle standard fs errors properly', async () => {
      // Create minimal structure
      createVirtualFs({
        '/test/file.txt': 'test content',
      });

      // Create a more complex error by trying to read a directory as a file
      await expect(fsPromises.readFile('/test', 'utf-8')).rejects.toThrow();
    });
  });

  describe('hidden file support', () => {
    it('should properly create and handle hidden files', async () => {
      // Create a virtual filesystem with hidden files and directories
      createVirtualFs({
        '/project/.gitignore': '*.log\n/dist/',
        '/project/.config/settings.json': '{"debug": true}',
        '/project/src/app.ts': 'console.log("Hello");',
        '/project/app.log': 'Error log content',
      });

      // Test that hidden files are created correctly
      expect(await fsPromises.readFile('/project/.gitignore', 'utf-8')).toBe('*.log\n/dist/');

      // Test hidden directory structure
      const rootContents = await fsPromises.readdir('/project');
      expect(rootContents).toContain('.gitignore');
      expect(rootContents).toContain('.config');

      // Test nested files in hidden directories
      expect(await fsPromises.readFile('/project/.config/settings.json', 'utf-8')).toBe(
        '{"debug": true}'
      );

      // Test file stats for hidden files
      const stats = await fsPromises.stat('/project/.gitignore');
      expect(stats.isFile()).toBe(true);

      // Test that we can modify hidden files
      await fsPromises.writeFile('/project/.gitignore', '*.log\n/dist/\n/temp/');
      expect(await fsPromises.readFile('/project/.gitignore', 'utf-8')).toBe(
        '*.log\n/dist/\n/temp/'
      );
    });
  });

  describe('addVirtualGitignoreFile', () => {
    it('should create a .gitignore file with provided patterns', async () => {
      // Setup
      resetVirtualFs();
      createVirtualFs({
        '/project/src/index.ts': 'console.log("Hello");',
      });

      // Action
      await addVirtualGitignoreFile('/project/.gitignore', '*.log\n/dist/\nnode_modules/');

      // Assert - File exists with correct content
      expect(await fsPromises.readFile('/project/.gitignore', 'utf-8')).toBe(
        '*.log\n/dist/\nnode_modules/'
      );

      // Directory structure should contain the file
      const rootContents = await fsPromises.readdir('/project');
      expect(rootContents).toContain('.gitignore');
    });

    it('should create parent directories if they do not exist', async () => {
      // Setup
      resetVirtualFs();

      // Action - Create file in a directory that doesn't exist yet
      await addVirtualGitignoreFile('/new-project/subdir/.gitignore', '*.log');

      // Assert - Both directories and file should be created
      expect(await fsPromises.readFile('/new-project/subdir/.gitignore', 'utf-8')).toBe('*.log');

      // Verify directory structure
      const rootContents = await fsPromises.readdir('/');
      expect(rootContents).toContain('new-project');

      const projectContents = await fsPromises.readdir('/new-project');
      expect(projectContents).toContain('subdir');
    });

    it('should overwrite existing file if it already exists', async () => {
      // Setup
      resetVirtualFs();
      createVirtualFs({
        '/project/.gitignore': 'old-pattern\n*.bak',
      });

      // Action
      await addVirtualGitignoreFile('/project/.gitignore', 'new-pattern\n*.log');

      // Assert - Content should be replaced
      expect(await fsPromises.readFile('/project/.gitignore', 'utf-8')).toBe('new-pattern\n*.log');
    });
  });

  describe('createFsError', () => {
    it('should create an error with correct properties', () => {
      // Create a file not found error
      const error = createFsError('ENOENT', 'File not found', 'open', '/path/to/missing.txt');

      // Check error properties
      expect(error).toBeInstanceOf(Error);
      expect(error.code).toBe('ENOENT');
      expect(error.message).toBe('File not found');
      expect(error.syscall).toBe('open');
      expect(error.path).toBe('/path/to/missing.txt');
      expect(error.errno).toBe(-2); // ENOENT errno value
    });

    it('should normalize the file path', () => {
      // Create an error with a Windows-style path
      const error = createFsError('ENOENT', 'File not found', 'open', 'C:\\path\\to\\file.txt');

      // Path should be normalized
      expect(error.path).toBe('/C:/path/to/file.txt');
    });

    it('should map common error codes to errno values', () => {
      // Test different error codes and check their errno values
      const enoentError = createFsError('ENOENT', 'Not found', 'stat', '/file');
      expect(enoentError.errno).toBe(-2);

      const eaccessError = createFsError('EACCES', 'Permission denied', 'access', '/file');
      expect(eaccessError.errno).toBe(-13);

      const epermError = createFsError('EPERM', 'Operation not permitted', 'open', '/file');
      expect(epermError.errno).toBe(-1);

      const erofsError = createFsError('EROFS', 'Read-only filesystem', 'write', '/file');
      expect(erofsError.errno).toBe(-30);

      const ebusyError = createFsError('EBUSY', 'Resource busy', 'unlink', '/file');
      expect(ebusyError.errno).toBe(-16);

      const emfileError = createFsError('EMFILE', 'Too many open files', 'open', '/file');
      expect(emfileError.errno).toBe(-24);
    });
  });

  describe('createMockStats', () => {
    it('should create a file stats object', () => {
      const stats = createMockStats(true); // isFile=true
      
      // Check methods
      expect(stats.isFile()).toBe(true);
      expect(stats.isDirectory()).toBe(false);
      expect(stats.isBlockDevice()).toBe(false);
      expect(stats.isCharacterDevice()).toBe(false);
      expect(stats.isFIFO()).toBe(false);
      expect(stats.isSocket()).toBe(false);
      expect(stats.isSymbolicLink()).toBe(false);
      
      // Check default size for a file (1024)
      expect(stats.size).toBe(1024);
      
      // Check time properties exist
      expect(stats.atime).toBeInstanceOf(Date);
      expect(stats.mtime).toBeInstanceOf(Date);
      expect(stats.ctime).toBeInstanceOf(Date);
      expect(stats.birthtime).toBeInstanceOf(Date);
    });

    it('should create a directory stats object', () => {
      const stats = createMockStats(false); // isFile=false (directory)
      
      expect(stats.isFile()).toBe(false);
      expect(stats.isDirectory()).toBe(true);
      
      // Check default size for a directory (4096)
      expect(stats.size).toBe(4096);
    });

    it('should allow custom size', () => {
      const stats = createMockStats(true, 2048); // isFile=true, size=2048
      
      expect(stats.isFile()).toBe(true);
      expect(stats.size).toBe(2048);
    });
  });

  describe('createMockDirent', () => {
    it('should create a file dirent object', () => {
      const dirent = createMockDirent('file.txt', true); // isFile=true
      
      // Check name property
      expect(dirent.name).toBe('file.txt');
      
      // Check methods
      expect(dirent.isFile()).toBe(true);
      expect(dirent.isDirectory()).toBe(false);
      expect(dirent.isBlockDevice()).toBe(false);
      expect(dirent.isCharacterDevice()).toBe(false);
      expect(dirent.isFIFO()).toBe(false);
      expect(dirent.isSocket()).toBe(false);
      expect(dirent.isSymbolicLink()).toBe(false);
    });

    it('should create a directory dirent object', () => {
      const dirent = createMockDirent('folder', false); // isFile=false (directory)
      
      expect(dirent.name).toBe('folder');
      expect(dirent.isFile()).toBe(false);
      expect(dirent.isDirectory()).toBe(true);
    });
  });
});
