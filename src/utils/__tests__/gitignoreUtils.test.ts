/**
 * Tests for gitignore utilities
 */
import { 
  mockFsModules, 
  resetVirtualFs, 
  createVirtualFs, 
  addVirtualGitignoreFile
} from '../../__tests__/utils/virtualFsUtils';

// No longer needed - we'll use addVirtualGitignoreFile in the tests below

// Setup mocks (must be before importing fs modules)
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Now import fs and other modules
import path from 'path';
import fs from 'fs/promises';
import * as gitignoreUtils from '../gitignoreUtils';
import * as fileReader from '../fileReader';
import { getVirtualFs } from '../../__tests__/utils/virtualFsUtils';

// Mock dependencies - but allow the real implementation in the module
jest.mock('../fileReader', () => {
  const originalModule = jest.requireActual('../fileReader');
  return {
    ...originalModule,
    fileExists: jest.fn()
  };
});

// Access mocked fileReader function
const mockedFileExists = jest.mocked(fileReader.fileExists);

// Tests have been updated to use actual implementation with virtual filesystem
describe('gitignoreUtils', () => {
  const testDirPath = '/test/dir';
  const gitignorePath = path.join(testDirPath, '.gitignore');
  
  beforeEach(async () => {
    // Reset virtual filesystem
    resetVirtualFs();
    
    // Clear gitignore cache
    gitignoreUtils.clearIgnoreCache();
    
    // Setup virtual filesystem with base directory structure but no .gitignore yet
    createVirtualFs({
      [testDirPath + '/']: '' // Just create the directory
    });
    
    // Add .gitignore file using the dedicated function
    await addVirtualGitignoreFile(gitignorePath, '*.log\ntmp/\n.DS_Store');
    
    // Set up fileExists mock to use the virtual filesystem
    // This ensures that calls to fileExists will return true for the .gitignore file
    mockedFileExists.mockImplementation(async (filePath) => {
      // Using imported getVirtualFs from the top of the file
      const virtualFs = getVirtualFs();
      const normalizedPath = filePath.startsWith('/') ? filePath.substring(1) : filePath;
      try {
        // Check if file exists in the virtual filesystem
        virtualFs.statSync(normalizedPath);
        return true;
      } catch (error) {
        console.log(`File doesn't exist in the virtual filesystem: ${normalizedPath}`);
        return false;
      }
    });
    
    // Spy on fs.readFile to track calls and log calls
    jest.spyOn(fs, 'readFile').mockImplementation(async (filePath, options) => {
      const normalizedPath = typeof filePath === 'string' && filePath.startsWith('/') 
        ? filePath.substring(1) 
        : String(filePath);
      
      // Default encoding is utf8 if not specified
      const encoding = typeof options === 'string' 
        ? options 
        : (options && typeof options === 'object' && 'encoding' in options && typeof options.encoding === 'string')
          ? options.encoding
          : 'utf8';
        
      try {
        console.log(`Reading file: ${normalizedPath}`);
        const virtualFs = getVirtualFs();
        // Use virtual fs readFileSync and convert to a promise for fs.readFile
        const content = virtualFs.readFileSync(normalizedPath, encoding as BufferEncoding);
        console.log(`Successfully read file: ${normalizedPath}, content length: ${content.length}`);
        return content;
      } catch (error) {
        console.error(`Error reading file: ${normalizedPath}`, error);
        // Re-throw the error to maintain the original behavior
        throw error;
      }
    });
  });
  
  afterEach(() => {
    jest.restoreAllMocks();
  });
  
  describe('createIgnoreFilter', () => {
    it('should create an ignore filter with patterns from .gitignore file', async () => {
      const ignoreFilter = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Test the ignore filter behavior directly
      // This tests that the .gitignore file we created is being read properly
      expect(ignoreFilter.ignores('file.log')).toBe(true);
      expect(ignoreFilter.ignores('tmp/file.txt')).toBe(true);
      expect(ignoreFilter.ignores('.DS_Store')).toBe(true);
      expect(ignoreFilter.ignores('file.txt')).toBe(false);
    });
    
    it('should use default patterns when .gitignore does not exist', async () => {
      // Reset mocks and file system for this test
      jest.clearAllMocks();
      resetVirtualFs();
      
      // Create directory but NO .gitignore file
      createVirtualFs({
        [testDirPath + '/']: '' // Create just the directory
      });
      
      // Don't add a .gitignore file for this test
      
      const ignoreFilter = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Test default patterns are applied when no .gitignore exists
      expect(ignoreFilter.ignores('node_modules/package.json')).toBe(true);
      expect(ignoreFilter.ignores('.git/config')).toBe(true);
      expect(ignoreFilter.ignores('dist/index.js')).toBe(true);
      expect(ignoreFilter.ignores('file.txt')).toBe(false);
    });
    
    it('should handle errors when reading .gitignore file', async () => {
      // Create a special scenario where we'll mock a read error
      jest.clearAllMocks();
      resetVirtualFs();
      
      // Create test directory
      createVirtualFs({
        [testDirPath + '/']: ''
      });
      
      // Capture console warnings
      const consoleWarnSpy = jest.spyOn(console, 'warn').mockImplementation();
      
      // Mock fs.readFile to simulate an error when reading a gitignore file
      const readFileSpy = jest.spyOn(fs, 'readFile');
      readFileSpy.mockRejectedValue(new Error('Simulated read error'));
      
      // Add a real .gitignore file, but our mock will prevent it from being read
      await addVirtualGitignoreFile(gitignorePath, '*.log');
      
      const ignoreFilter = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Verify default patterns were used despite the error
      expect(ignoreFilter.ignores('node_modules/package.json')).toBe(true);
      expect(ignoreFilter.ignores('file.txt')).toBe(false);
      
      // Since we've simulated an error, we should not be respecting the .gitignore patterns
      expect(ignoreFilter.ignores('test.log')).toBe(false);
      
      // Restore mocks
      consoleWarnSpy.mockRestore();
      readFileSpy.mockRestore();
    });
    
    it('should cache ignore filters for the same directory', async () => {
      // Reset and create fresh environment
      jest.clearAllMocks();
      resetVirtualFs();
      gitignoreUtils.clearIgnoreCache();
      
      // Create test directory and file
      createVirtualFs({
        [testDirPath + '/']: ''
      });
      
      // Add .gitignore file
      await addVirtualGitignoreFile(gitignorePath, '*.log\ntmp/\n.DS_Store');
      
      // First call to populate cache
      const ignoreFilter1 = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Clear caches to ensure we're actually testing the gitignoreUtils cache
      resetVirtualFs();

      // This would make the test fail if the cache wasn't working
      // because the file no longer exists but should be cached
      
      // Second call should use cached version
      const ignoreFilter2 = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Both filters should be the same instance
      expect(ignoreFilter1).toBe(ignoreFilter2);
    });
    
    it('should create different ignore filters for different directories', async () => {
      // Reset environment
      jest.clearAllMocks();
      resetVirtualFs();
      gitignoreUtils.clearIgnoreCache();
      
      const otherDirPath = '/other/dir';
      const otherGitignorePath = path.join(otherDirPath, '.gitignore');
      
      // Create test directories
      createVirtualFs({
        [testDirPath + '/']: '',
        [otherDirPath + '/']: ''
      });
      
      // Add different .gitignore files to the two directories
      await addVirtualGitignoreFile(gitignorePath, '*.log\ntmp/');
      await addVirtualGitignoreFile(otherGitignorePath, '*.tmp\n*.bak');
      
      // Call for first directory
      const ignoreFilter1 = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Call for second directory
      const ignoreFilter2 = await gitignoreUtils.createIgnoreFilter(otherDirPath);
      
      // Filters should be different instances
      expect(ignoreFilter1).not.toBe(ignoreFilter2);
      
      // Verify they have different patterns
      expect(ignoreFilter1.ignores('test.log')).toBe(true);
      expect(ignoreFilter1.ignores('test.tmp')).toBe(false);
      
      expect(ignoreFilter2.ignores('test.log')).toBe(false);
      expect(ignoreFilter2.ignores('test.tmp')).toBe(true);
    });
  });
  
  describe('shouldIgnorePath', () => {
    it('should return true for paths that match ignore patterns', async () => {
      // Test a file with .log extension which should be ignored per our gitignore content
      const result = await gitignoreUtils.shouldIgnorePath(testDirPath, 'test.log');
      expect(result).toBe(true);
    });
    
    it('should return false for paths that do not match ignore patterns', async () => {
      // Test a normal file that shouldn't be ignored
      const result = await gitignoreUtils.shouldIgnorePath(testDirPath, 'src/index.ts');
      expect(result).toBe(false);
    });
    
    it('should handle absolute paths', async () => {
      // Test with an absolute path that should be ignored
      const absolutePath = path.join(testDirPath, 'some/path/file.log');
      const result = await gitignoreUtils.shouldIgnorePath(testDirPath, absolutePath);
      expect(result).toBe(true);
    });
  });
  
  describe('clearIgnoreCache', () => {
    it('should clear the ignore filter cache', async () => {
      // Reset environment
      jest.clearAllMocks();
      resetVirtualFs();
      gitignoreUtils.clearIgnoreCache();
      
      // Create test directory and file
      createVirtualFs({
        [testDirPath + '/']: ''
      });
      
      // Add .gitignore file
      await addVirtualGitignoreFile(gitignorePath, '*.log\ntmp/\n.DS_Store');
      
      // First call to populate cache
      const ignoreFilter1 = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Second call should use cached version
      const ignoreFilter2 = await gitignoreUtils.createIgnoreFilter(testDirPath);
      expect(ignoreFilter1).toBe(ignoreFilter2); // Same instance from cache
      
      // Clear cache
      gitignoreUtils.clearIgnoreCache();

      // Remove the .gitignore file to confirm we're getting a new instance
      resetVirtualFs();
      createVirtualFs({
        [testDirPath + '/']: ''
      });
      await addVirtualGitignoreFile(gitignorePath, '*.txt'); // Different content
      
      // This should create a new filter
      const ignoreFilter3 = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Should be a different instance with different behavior
      expect(ignoreFilter1).not.toBe(ignoreFilter3);
      expect(ignoreFilter1.ignores('test.log')).toBe(true);
      expect(ignoreFilter3.ignores('test.log')).toBe(false);
      expect(ignoreFilter3.ignores('test.txt')).toBe(true);
    });
  });
});
