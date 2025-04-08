/**
 * Tests for gitignore utilities
 */
import { 
  mockFsModules, 
  addVirtualGitignoreFile,
  createFsError
} from '../../__tests__/utils/virtualFsUtils';
import {
  setupBasicFiles,
  setupGitignoreMocking
} from '../../__tests__/utils/fsTestSetup';

// Setup mocks (must be before importing fs modules)
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Now import fs and other modules
import path from 'path';
import fs from 'fs/promises';
import * as gitignoreUtils from '../gitignoreUtils';
import * as fileReader from '../fileReader';

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
    // Set up gitignore mocking with our reusable helper
    setupGitignoreMocking(gitignoreUtils, mockedFileExists);
    
    // Setup virtual filesystem with base directory structure and gitignore
    setupBasicFiles({
      // Create the directory by creating a file in it
      [testDirPath + '/placeholder.txt']: 'This is just to ensure directory exists'
    });
    
    // Add .gitignore file using the dedicated function
    await addVirtualGitignoreFile(gitignorePath, '*.log\ntmp/\n.DS_Store');
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
      // Create a fresh setup with no gitignore file
      setupGitignoreMocking(gitignoreUtils, mockedFileExists);
      
      // Create directory but NO .gitignore file
      setupBasicFiles({
        [testDirPath + '/placeholder.txt']: 'This is just to ensure directory exists'
      });
      
      const ignoreFilter = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Test default patterns are applied when no .gitignore exists
      expect(ignoreFilter.ignores('node_modules/package.json')).toBe(true);
      expect(ignoreFilter.ignores('.git/config')).toBe(true);
      expect(ignoreFilter.ignores('dist/index.js')).toBe(true);
      expect(ignoreFilter.ignores('file.txt')).toBe(false);
    });
    
    it('should handle errors when reading .gitignore file', async () => {
      // Create a special scenario where we'll mock a read error
      setupGitignoreMocking(gitignoreUtils, mockedFileExists);
      
      // Create test directory
      setupBasicFiles({
        [testDirPath + '/placeholder.txt']: 'This is just to ensure directory exists'
      });
      
      // Mock console warnings without actually outputting to console
      const consoleWarnSpy = jest.spyOn(console, 'warn').mockImplementation(() => {});
      
      // Mock fs.readFile to simulate an error when reading a gitignore file
      // Using createFsError helper to create a standardized error object
      const readFileSpy = jest.spyOn(fs, 'readFile');
      readFileSpy.mockRejectedValue(
        createFsError('EACCES', 'Simulated read error', 'readFile', gitignorePath)
      );
      
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
      // Create a fresh setup
      setupGitignoreMocking(gitignoreUtils, mockedFileExists);
      
      // Create test directory and file with gitignore
      setupBasicFiles({
        [testDirPath + '/placeholder.txt']: 'This is just to ensure directory exists'
      });
      
      await addVirtualGitignoreFile(gitignorePath, '*.log\ntmp/\n.DS_Store');
      
      // First call to populate cache
      const ignoreFilter1 = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Clear virtual filesystem to ensure we're testing cache
      setupBasicFiles({});

      // Second call should use cached version despite filesystem reset
      const ignoreFilter2 = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Both filters should be the same instance
      expect(ignoreFilter1).toBe(ignoreFilter2);
    });
    
    it('should create different ignore filters for different directories', async () => {
      // Set up a fresh environment
      setupGitignoreMocking(gitignoreUtils, mockedFileExists);
      
      const otherDirPath = '/other/dir';
      const otherGitignorePath = path.join(otherDirPath, '.gitignore');
      
      // Create test directories
      setupBasicFiles({
        [testDirPath + '/placeholder.txt']: 'This is just to ensure directory exists',
        [otherDirPath + '/placeholder.txt']: 'This is just to ensure directory exists'
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
      // Using a path that is within the test directory
      const testLogPath = path.join(testDirPath, 'test.log');
      const result = await gitignoreUtils.shouldIgnorePath(testDirPath, testLogPath);
      expect(result).toBe(true);
    });
    
    it('should return false for paths that do not match ignore patterns', async () => {
      // Test a normal file that shouldn't be ignored
      // Using a path that is within the test directory
      const srcPath = path.join(testDirPath, 'src/index.ts');
      const result = await gitignoreUtils.shouldIgnorePath(testDirPath, srcPath);
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
      // Set up a fresh environment
      setupGitignoreMocking(gitignoreUtils, mockedFileExists);
      
      // Create test directory and file
      setupBasicFiles({
        [testDirPath + '/placeholder.txt']: 'This is just to ensure directory exists'
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

      // Setup a new environment with different gitignore content
      setupBasicFiles({
        [testDirPath + '/placeholder.txt']: 'This is just to ensure directory exists'
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
