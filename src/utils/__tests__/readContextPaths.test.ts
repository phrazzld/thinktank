/**
 * Tests for the master readContextPaths function
 */
import { 
  createVirtualFs, 
  resetVirtualFs, 
  mockFsModules, 
  createFsError, 
  getVirtualFs,
  addVirtualGitignoreFile
} from '../../__tests__/utils/virtualFsUtils';

// Setup mocks for fs modules
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Import modules after mocking
import fs from 'fs';
import fsPromises from 'fs/promises';
import path from 'path';
import { ContextFileResult } from '../fileReaderTypes';
import { readContextPaths } from '../fileReader';
import * as gitignoreUtils from '../gitignoreUtils';

describe('readContextPaths function', () => {
  const testFile = '/path/to/file.txt';
  const testDir = '/path/to/directory';
  const testSubfile1 = '/path/to/directory/subfile1.txt';
  const testSubfile2 = '/path/to/directory/subfile2.md';
  const nonexistentFile = '/path/to/nonexistent-file.txt';

  beforeEach(() => {
    // Reset virtual filesystem and mocks before each test
    resetVirtualFs();
    jest.clearAllMocks();
    
    // Clear gitignore cache
    gitignoreUtils.clearIgnoreCache();
  });
  
  it('should process a mix of files and directories', async () => {
    // Test paths to process
    const testPaths = [
      testFile,
      testDir
    ];
    
    // Create virtual filesystem with test structure
    createVirtualFs({
      [testFile]: 'Content of file.txt',
      [testSubfile1]: 'Content of subfile1.txt', 
      [testSubfile2]: 'Content of subfile2.md'
    });
    
    // Create the directory structure
    const virtualFs = getVirtualFs();
    virtualFs.mkdirSync(testDir, { recursive: true });
    
    // Call the function
    const results = await readContextPaths(testPaths);
    
    // Should have 3 results (1 file + 2 directory files)
    expect(results.length).toBe(3);
    
    // Verify individual file was processed
    expect(results.some((r: ContextFileResult) => r.path === testFile)).toBe(true);
    
    // Verify directory files were processed
    expect(results.some((r: ContextFileResult) => r.path.includes('subfile1.txt'))).toBe(true);
    expect(results.some((r: ContextFileResult) => r.path.includes('subfile2.md'))).toBe(true);
    
    // Verify content was read
    expect(results.find((r: ContextFileResult) => r.path === testFile)?.content).toBe('Content of file.txt');
  });
  
  it('should handle empty paths array', async () => {
    // Create a spy on stat to verify it's not called
    const statSpy = jest.spyOn(fsPromises, 'stat');
    
    const results = await readContextPaths([]);
    
    expect(results).toEqual([]);
    expect(statSpy).not.toHaveBeenCalled();
    
    statSpy.mockRestore();
  });
  
  it('should handle errors for individual paths', async () => {
    // Test paths with one valid and one non-existent file
    const testPaths = [
      testFile,
      nonexistentFile
    ];
    
    // Create only the valid file in the virtual filesystem
    createVirtualFs({
      [testFile]: 'Content of valid-file.txt'
    });
    
    const results = await readContextPaths(testPaths);
    
    // Should still have 2 results, but one with an error
    expect(results.length).toBe(2);
    
    // Verify valid file was processed correctly
    const validResult = results.find((r: ContextFileResult) => r.path === testFile);
    expect(validResult?.content).toBe('Content of valid-file.txt');
    expect(validResult?.error).toBeNull();
    
    // Verify error file has appropriate error info
    const errorResult = results.find((r: ContextFileResult) => r.path === nonexistentFile);
    expect(errorResult?.content).toBeNull();
    expect(errorResult?.error?.code).toBe('ENOENT');
  });
  
  it('should handle permission denied errors', async () => {
    // Setup the filesystem with a file that we'll make inaccessible
    createVirtualFs({
      [testFile]: 'Content of file.txt'
    });
    
    // Mock access to throw permission denied error
    const accessSpy = jest.spyOn(fsPromises, 'access');
    accessSpy.mockImplementation((path) => {
      if (path === testFile) {
        throw createFsError('EACCES', 'Permission denied', 'access', testFile);
      }
      // Let other access calls proceed normally
      return Promise.resolve();
    });
    
    const results = await readContextPaths([testFile]);
    
    // Verify error message for permission denied
    expect(results.length).toBe(1);
    expect(results[0].error?.code).toBe('EACCES');
    expect(results[0].error?.message).toContain('Unable to access path');
    
    accessSpy.mockRestore();
  });
  
  it('should handle paths that are neither files nor directories', async () => {
    // First create a file in the virtual filesystem
    createVirtualFs({
      [testFile]: 'File content'
    });
    
    // Mock stat to make the path appear as neither a file nor directory
    const statSpy = jest.spyOn(fsPromises, 'stat');
    statSpy.mockResolvedValue({
      isFile: () => false,
      isDirectory: () => false,
      isBlockDevice: () => false,
      isCharacterDevice: () => false,
      isFIFO: () => false,
      isSocket: () => false,
      isSymbolicLink: () => true, // Make it a symlink instead
      dev: 0,
      ino: 0,
      mode: 0,
      nlink: 0,
      uid: 0,
      gid: 0,
      rdev: 0,
      size: 0,
      blksize: 0,
      blocks: 0,
      atimeMs: 0,
      mtimeMs: 0,
      ctimeMs: 0,
      birthtimeMs: 0,
      atime: new Date(),
      mtime: new Date(),
      ctime: new Date(),
      birthtime: new Date()
    } as fs.Stats);
    
    const results = await readContextPaths([testFile]);
    
    // Verify error message for invalid path type
    expect(results.length).toBe(1);
    expect(results[0].error?.code).toBe('INVALID_PATH_TYPE');
    expect(results[0].error?.message).toContain('Path is neither a file nor a directory');
    
    statSpy.mockRestore();
  });
  
  it('should handle errors during directory reading', async () => {
    // Create a directory structure in the virtual filesystem
    createVirtualFs({
      [testSubfile1]: 'Subfile 1 content'
    });
    
    // Create the directory
    const virtualFs = getVirtualFs();
    virtualFs.mkdirSync(testDir, { recursive: true });
    
    // Mock readdir to throw an error
    const readdirSpy = jest.spyOn(fsPromises, 'readdir');
    readdirSpy.mockRejectedValue(createFsError('EMFILE', 'Too many open files', 'readdir', testDir));
    
    const results = await readContextPaths([testDir]);
    
    // Should return an error for the directory
    expect(results.length).toBe(1);
    expect(results[0].error?.code).toBe('READ_ERROR');
    expect(results[0].error?.message).toContain('Error reading directory');
    
    readdirSpy.mockRestore();
  });
  
  it('should handle relative paths', async () => {
    // Define a relative path
    const relativePath = 'relative/path.txt';
    const absolutePath = path.resolve(process.cwd(), relativePath);
    
    // Create the file in the virtual filesystem
    createVirtualFs({
      [absolutePath]: 'Content of relative file'
    });
    
    const results = await readContextPaths([relativePath]);
    
    // Should resolve the relative path and process the file
    expect(results.length).toBe(1);
    expect(results[0].path).toBe(relativePath); // Should preserve the original path
    expect(results[0].content).toBe('Content of relative file');
    expect(results[0].error).toBeNull();
  });
  
  it('should respect simple gitignore patterns', async () => {
    // Test directory with various files
    const gitignoreDir = '/test-gitignore';
    const gitignorePath = path.join(gitignoreDir, '.gitignore');
    
    // Create files, including some that should be ignored
    createVirtualFs({
      [path.join(gitignoreDir, 'regular-file.txt')]: 'Regular file content',
      [path.join(gitignoreDir, 'ignored-file.log')]: 'Log file that should be ignored',
      [path.join(gitignoreDir, 'another-file.js')]: 'JavaScript file content',
      [path.join(gitignoreDir, 'node_modules', 'package.json')]: '{"name": "test"}'
    });
    
    // Create the test directory explicitly to ensure it exists
    const virtualFs = getVirtualFs();
    virtualFs.mkdirSync(gitignoreDir, { recursive: true });
    virtualFs.mkdirSync(path.join(gitignoreDir, 'node_modules'), { recursive: true });
    
    // Add .gitignore file to ignore *.log files
    await addVirtualGitignoreFile(gitignorePath, '*.log');
    
    // Read the directory - using the real gitignoreUtils functionality
    const results = await readContextPaths([gitignoreDir]);
    
    // Should include the regular files
    expect(results.some(r => r.path.includes('regular-file.txt'))).toBe(true);
    expect(results.some(r => r.path.includes('another-file.js'))).toBe(true);
    
    // Should not include ignored files
    expect(results.some(r => r.path.includes('ignored-file.log'))).toBe(false);
    
    // Default patterns should also work (node_modules should be ignored)
    expect(results.some(r => r.path.includes('node_modules'))).toBe(false);
  });
  
  it('should respect specific gitignore patterns directly', async () => {
    // Test if the gitignoreUtils.shouldIgnorePath function works properly
    // with the virtual filesystem
    
    const testDir = '/test-gitignore-direct';
    const gitignorePath = path.join(testDir, '.gitignore');
    
    // Create directory and .gitignore file
    const virtualFs = getVirtualFs();
    virtualFs.mkdirSync(testDir, { recursive: true });
    
    // Add .gitignore file with various patterns
    await addVirtualGitignoreFile(gitignorePath, 
      '# Environment variables\n' +
      '.env\n' +
      '\n' +
      '# Logs\n' +
      'logs/\n' +
      '*.log\n' +
      '\n' +
      '# Build directory\n' +
      'build/\n' +
      '\n' +
      '# Test artifacts\n' +
      'tests/artifacts/\n' +
      '\n' +
      '# Temporary files\n' +
      '*.tmp\n' +
      '\n' +
      '# Dependencies\n' +
      'node_modules/'
    );
    
    // Clear cache to ensure our .gitignore file is read fresh
    gitignoreUtils.clearIgnoreCache();
    
    // Now test individual paths directly against shouldIgnorePath
    const filesToTest = [
      { path: path.join(testDir, 'app.js'), shouldBeIgnored: false },
      { path: path.join(testDir, 'src', 'utils.js'), shouldBeIgnored: false },
      { path: path.join(testDir, '.env'), shouldBeIgnored: true },
      { path: path.join(testDir, 'logs', 'error.log'), shouldBeIgnored: true },
      { path: path.join(testDir, 'build', 'bundle.js'), shouldBeIgnored: true },
      { path: path.join(testDir, 'tests', 'artifacts', 'report.html'), shouldBeIgnored: true },
      { path: path.join(testDir, 'node_modules', 'lodash', 'index.js'), shouldBeIgnored: true },
      { path: path.join(testDir, 'src', 'temp.tmp'), shouldBeIgnored: true }
    ];
    
    // Test each file path directly
    for (const fileTest of filesToTest) {
      const isIgnored = await gitignoreUtils.shouldIgnorePath(testDir, fileTest.path);
      expect(isIgnored).toBe(fileTest.shouldBeIgnored);
    }
  });
});
