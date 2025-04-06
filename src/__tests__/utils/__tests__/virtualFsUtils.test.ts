import { createVirtualFs, resetVirtualFs, getVirtualFs, mockFsModules } from '../virtualFsUtils';

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
  
  describe('createVirtualFs', () => {
    it('should create a virtual filesystem with specified structure', async () => {
      // Create a virtual filesystem structure
      createVirtualFs({
        '/test/file.txt': 'test content',
        '/test/dir/nested.txt': 'nested content'
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
        '/test/file1.txt': 'content 1'
      });
      
      // Create new structure (should replace the first one)
      createVirtualFs({
        '/test/file2.txt': 'content 2'
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
        '/test/file.txt': 'test content'
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
        '/test/file.txt': 'test content'
      });
      
      // Create a more complex error by trying to read a directory as a file
      await expect(fsPromises.readFile('/test', 'utf-8')).rejects.toThrow();
    });
  });
});