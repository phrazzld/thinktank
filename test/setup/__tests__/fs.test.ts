/**
 * Tests for the filesystem setup utilities
 */
import {
  setupWithFiles,
  setupBasicFs,
  setupWithSingleFile,
  setupWithNestedDirectories,
  getFs
} from '../fs';
import { resetVirtualFs } from '../../../src/__tests__/utils/virtualFsUtils';

describe('File system setup utilities', () => {
  // Reset the FS after each test to avoid cross-test interference
  afterEach(() => {
    // Use the resetVirtualFs function that's already imported via fs.ts
    resetVirtualFs();
  });

  describe('setupWithFiles', () => {
    it('should set up a virtual filesystem with multiple files', () => {
      // Arrange
      const files = {
        '/file1.txt': 'content 1',
        '/dir/file2.txt': 'content 2',
        '/empty-dir/': ''
      };

      // Act
      setupWithFiles(files);
      const vfs = getFs();

      // Assert
      expect(vfs.readFileSync('/file1.txt', 'utf8')).toBe('content 1');
      expect(vfs.readFileSync('/dir/file2.txt', 'utf8')).toBe('content 2');
      expect(vfs.existsSync('/empty-dir')).toBe(true);
      expect(vfs.statSync('/empty-dir').isDirectory()).toBe(true);
    });

    it('should reset the filesystem by default', () => {
      // Arrange - Setup initial state
      const vfs = getFs();
      vfs.writeFileSync('/existing.txt', 'existing content');
      
      // Act - Setup new files, which should clear existing ones
      setupWithFiles({
        '/new.txt': 'new content'
      });

      // Assert
      expect(vfs.existsSync('/existing.txt')).toBe(false);
      expect(vfs.existsSync('/new.txt')).toBe(true);
    });

    it('should preserve existing files when reset is false', () => {
      // Arrange - Setup initial state
      const vfs = getFs();
      vfs.writeFileSync('/existing.txt', 'existing content');
      
      // Act - Setup new files without resetting
      setupWithFiles({
        '/new.txt': 'new content'
      }, { reset: false });

      // Assert
      expect(vfs.existsSync('/existing.txt')).toBe(true);
      expect(vfs.existsSync('/new.txt')).toBe(true);
    });
  });

  describe('setupBasicFs', () => {
    it('should be an alias for setupWithFiles', () => {
      // Arrange
      const files = {
        '/file.txt': 'test content'
      };

      // Act
      setupBasicFs(files);
      const vfs = getFs();

      // Assert
      expect(vfs.readFileSync('/file.txt', 'utf8')).toBe('test content');
    });
  });

  describe('setupWithSingleFile', () => {
    it('should set up a single file in the virtual filesystem', () => {
      // Arrange
      const filePath = '/path/to/config.json';
      const content = '{ "setting": "value" }';

      // Act
      setupWithSingleFile(filePath, content);
      const vfs = getFs();

      // Assert
      expect(vfs.readFileSync(filePath, 'utf8')).toBe(content);
    });

    it('should create parent directories as needed', () => {
      // Arrange
      const filePath = '/deeply/nested/path/file.txt';
      const content = 'file content';

      // Act
      setupWithSingleFile(filePath, content);
      const vfs = getFs();

      // Assert
      expect(vfs.existsSync('/deeply/nested/path')).toBe(true);
      expect(vfs.statSync('/deeply/nested/path').isDirectory()).toBe(true);
      expect(vfs.readFileSync(filePath, 'utf8')).toBe(content);
    });

    it('should reset the filesystem before creating the file', () => {
      // Arrange - Setup initial state
      const vfs = getFs();
      vfs.writeFileSync('/existing.txt', 'existing content');
      
      // Act
      setupWithSingleFile('/new.txt', 'new content');

      // Assert
      expect(vfs.existsSync('/existing.txt')).toBe(false);
      expect(vfs.existsSync('/new.txt')).toBe(true);
    });
  });

  describe('setupWithNestedDirectories', () => {
    it('should set up a filesystem with nested directories and files', () => {
      // Arrange
      const structure = {
        '/root/dir1/file1.txt': 'content 1',
        '/root/dir2/subdir/file2.txt': 'content 2',
        '/root/dir3/empty/': ''
      };

      // Act
      setupWithNestedDirectories(structure);
      const vfs = getFs();

      // Assert
      expect(vfs.readFileSync('/root/dir1/file1.txt', 'utf8')).toBe('content 1');
      expect(vfs.readFileSync('/root/dir2/subdir/file2.txt', 'utf8')).toBe('content 2');
      expect(vfs.existsSync('/root/dir3/empty')).toBe(true);
      expect(vfs.statSync('/root/dir3/empty').isDirectory()).toBe(true);
    });

    it('should create empty directories with trailing slash paths', () => {
      // Arrange
      const structure = {
        '/empty-dir1/': '',
        '/path/to/empty-dir2/': ''
      };

      // Act
      setupWithNestedDirectories(structure);
      const vfs = getFs();

      // Assert
      expect(vfs.existsSync('/empty-dir1')).toBe(true);
      expect(vfs.statSync('/empty-dir1').isDirectory()).toBe(true);
      expect(vfs.existsSync('/path/to/empty-dir2')).toBe(true);
      expect(vfs.statSync('/path/to/empty-dir2').isDirectory()).toBe(true);
    });

    it('should reset the filesystem before setup', () => {
      // Arrange - Setup initial state
      const vfs = getFs();
      vfs.writeFileSync('/existing.txt', 'existing content');
      
      // Act
      setupWithNestedDirectories({
        '/nested/file.txt': 'nested content'
      });

      // Assert
      expect(vfs.existsSync('/existing.txt')).toBe(false);
      expect(vfs.existsSync('/nested/file.txt')).toBe(true);
    });
  });
});
