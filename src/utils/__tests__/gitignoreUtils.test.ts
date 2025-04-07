/**
 * Tests for gitignore utilities
 */
import path from 'path';
import * as gitignoreUtils from '../gitignoreUtils';
import * as fileReader from '../fileReader';
import { 
  resetMockFs, 
  setupMockFs, 
  mockReadFile,
  mockedFs
} from '../../__tests__/utils/mockFsUtils';

// Mock dependencies
jest.mock('fs/promises');
jest.mock('../fileReader');

// Access mocked fileReader function
const mockedFileExists = jest.mocked(fileReader.fileExists);

describe('gitignoreUtils', () => {
  const testDirPath = '/test/dir';
  const gitignorePath = path.join(testDirPath, '.gitignore');
  
  beforeEach(() => {
    // Reset and setup mocks
    resetMockFs();
    setupMockFs();
    gitignoreUtils.clearIgnoreCache();
    
    // Default mocks
    mockedFileExists.mockResolvedValue(true);
    mockReadFile(gitignorePath, '*.log\ntmp/\n.DS_Store');
  });
  
  describe('createIgnoreFilter', () => {
    it('should create an ignore filter with patterns from .gitignore file', async () => {
      const ignoreFilter = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Verify .gitignore was checked and read
      expect(mockedFileExists).toHaveBeenCalledWith(gitignorePath);
      expect(mockedFs.readFile).toHaveBeenCalledWith(gitignorePath, 'utf-8');
      
      // Test the ignore filter
      expect(ignoreFilter.ignores('file.log')).toBe(true);
      expect(ignoreFilter.ignores('tmp/file.txt')).toBe(true);
      expect(ignoreFilter.ignores('.DS_Store')).toBe(true);
      expect(ignoreFilter.ignores('file.txt')).toBe(false);
    });
    
    it('should use default patterns when .gitignore does not exist', async () => {
      // Mock that .gitignore doesn't exist
      mockedFileExists.mockResolvedValue(false);
      
      const ignoreFilter = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Verify .gitignore was checked but not read
      expect(mockedFileExists).toHaveBeenCalledWith(gitignorePath);
      expect(mockedFs.readFile).not.toHaveBeenCalled();
      
      // Test default patterns
      expect(ignoreFilter.ignores('node_modules/package.json')).toBe(true);
      expect(ignoreFilter.ignores('.git/config')).toBe(true);
      expect(ignoreFilter.ignores('dist/index.js')).toBe(true);
      expect(ignoreFilter.ignores('file.txt')).toBe(false);
    });
    
    it('should handle errors when reading .gitignore file', async () => {
      // Mock error when reading .gitignore
      mockReadFile(gitignorePath, new Error('Read error'));
      
      // Capture console warnings
      const consoleWarnSpy = jest.spyOn(console, 'warn').mockImplementation();
      
      const ignoreFilter = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Verify warning was logged
      expect(consoleWarnSpy).toHaveBeenCalled();
      expect(consoleWarnSpy.mock.calls[0][0]).toContain('Could not read .gitignore file');
      
      // Should still use default patterns
      expect(ignoreFilter.ignores('node_modules/package.json')).toBe(true);
      expect(ignoreFilter.ignores('file.txt')).toBe(false);
      
      consoleWarnSpy.mockRestore();
    });
    
    it('should cache ignore filters for the same directory', async () => {
      // Call twice for the same directory
      const ignoreFilter1 = await gitignoreUtils.createIgnoreFilter(testDirPath);
      const ignoreFilter2 = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Verify .gitignore was read only once
      expect(mockedFileExists).toHaveBeenCalledTimes(1);
      expect(mockedFs.readFile).toHaveBeenCalledTimes(1);
      
      // Both filters should be the same instance
      expect(ignoreFilter1).toBe(ignoreFilter2);
    });
    
    it('should create different ignore filters for different directories', async () => {
      const otherDirPath = '/other/dir';
      const otherGitignorePath = path.join(otherDirPath, '.gitignore');
      
      // Set up mock for the second directory
      mockedFileExists.mockImplementation(async (path) => {
        return path === gitignorePath; // Only the first path has a .gitignore
      });
      
      // Call for two different directories
      const ignoreFilter1 = await gitignoreUtils.createIgnoreFilter(testDirPath);
      const ignoreFilter2 = await gitignoreUtils.createIgnoreFilter(otherDirPath);
      
      // Verify .gitignore existence was checked for both directories
      expect(mockedFileExists).toHaveBeenCalledTimes(2);
      expect(mockedFileExists).toHaveBeenCalledWith(gitignorePath);
      expect(mockedFileExists).toHaveBeenCalledWith(otherGitignorePath);
      
      // But file was read only for the first one
      expect(mockedFs.readFile).toHaveBeenCalledTimes(1);
      expect(mockedFs.readFile).toHaveBeenCalledWith(gitignorePath, 'utf-8');
      expect(mockedFs.readFile).not.toHaveBeenCalledWith(otherGitignorePath, 'utf-8');
      
      // Filters should be different instances
      expect(ignoreFilter1).not.toBe(ignoreFilter2);
    });
  });
  
  describe('shouldIgnorePath', () => {
    it('should return true for paths that match ignore patterns', async () => {
      const result = await gitignoreUtils.shouldIgnorePath(testDirPath, 'logs/test.log');
      expect(result).toBe(true);
    });
    
    it('should return false for paths that do not match ignore patterns', async () => {
      const result = await gitignoreUtils.shouldIgnorePath(testDirPath, 'src/index.ts');
      expect(result).toBe(false);
    });
    
    it('should handle absolute paths', async () => {
      const absolutePath = path.join(testDirPath, 'logs/test.log');
      const result = await gitignoreUtils.shouldIgnorePath(testDirPath, absolutePath);
      expect(result).toBe(true);
    });
  });
  
  describe('clearIgnoreCache', () => {
    it('should clear the ignore filter cache', async () => {
      // Populate cache
      await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Clear mocks to verify second call behavior
      jest.clearAllMocks();
      
      // This should use cached version
      await gitignoreUtils.createIgnoreFilter(testDirPath);
      expect(mockedFileExists).not.toHaveBeenCalled();
      expect(mockedFs.readFile).not.toHaveBeenCalled();
      
      // Clear cache
      gitignoreUtils.clearIgnoreCache();
      
      // This should need to read the file again
      await gitignoreUtils.createIgnoreFilter(testDirPath);
      expect(mockedFileExists).toHaveBeenCalled();
      expect(mockedFs.readFile).toHaveBeenCalled();
    });
  });
});