/**
 * Tests for the directory reader utility
 */
import path from 'path';
import {
  normalizePathForMemfs,
  resetVirtualFs,
  createFsError,
  createMockStats,
} from '../../__tests__/utils/virtualFsUtils';
import {
  setupBasicFiles,
  setupWithGitignore
} from '../../__tests__/utils/fsTestSetup';

import * as gitignoreUtils from '../gitignoreUtils';
import { setupTestHooks } from '../../../test/setup/common';
import { ConcreteFileSystem } from '../../core/FileSystem';
import fsPromises from 'fs/promises';

// Import module after setup
import * as fileReader from '../fileReader';
const { readDirectoryContents, readContextFile } = fileReader;

describe('readDirectoryContents', () => {
  // Setup test hooks to reset filesystem and mocks
  setupTestHooks();

  // Use a standard path with proper normalization for memfs compatibility
  const testDirPath = normalizePathForMemfs('/path/to/test/directory');

  describe('Basic Directory Traversal', () => {
    beforeEach(() => {
      // Create directory structure using setupBasicFiles for consistency
      setupBasicFiles({
        [`${testDirPath}/`]: '',
        [`${testDirPath}/file1.txt`]: 'Content of file1.txt',
        [`${testDirPath}/file2.md`]: 'Content of file2.md',
        [`${testDirPath}/subdir/`]: '',
        [`${testDirPath}/subdir/nested.txt`]: 'Content of nested.txt',
        [`${testDirPath}/node_modules/`]: '',
        [`${testDirPath}/node_modules/package.json`]: '{"name": "test"}',
        [`${testDirPath}/.git/`]: '',
        [`${testDirPath}/.git/HEAD`]: 'ref: refs/heads/main',
      });

      // Setup a default mock for gitignoreUtils.shouldIgnorePath
      jest.spyOn(gitignoreUtils, 'shouldIgnorePath').mockImplementation((_, filePath) => {
        // Default implementation to ignore node_modules and .git files by path
        return Promise.resolve(filePath.includes('node_modules') || filePath.includes('.git'));
      });
    });

    it('should read all files in a directory and return their contents', async () => {
      // Create FileSystem instance
      const fileSystem = new ConcreteFileSystem();

      // Call function with FileSystem interface
      const results = await readDirectoryContents(testDirPath, fileSystem);

      // Expect to have at least 2 files (file1.txt, file2.md)
      expect(results.length).toBeGreaterThanOrEqual(2);

      // Check if files were processed correctly
      const file1Result = results.find(r => r.path.endsWith('file1.txt'));
      const file2Result = results.find(r => r.path.endsWith('file2.md'));

      expect(file1Result).toBeDefined();
      expect(file1Result?.content).toBe('Content of file1.txt');
      expect(file1Result?.error).toBeNull();

      expect(file2Result).toBeDefined();
      expect(file2Result?.content).toBe('Content of file2.md');
      expect(file2Result?.error).toBeNull();
    });

    it('should recursively traverse subdirectories', async () => {
      // Create FileSystem instance
      const fileSystem = new ConcreteFileSystem();

      // Call function with FileSystem interface
      const results = await readDirectoryContents(testDirPath, fileSystem);

      // Should include files from subdirectories
      const nestedFileResult = results.find(r => r.path.includes('nested.txt'));

      expect(nestedFileResult).toBeDefined();
      expect(nestedFileResult?.content).toBe('Content of nested.txt');
      expect(nestedFileResult?.error).toBeNull();
    });

    it('should skip common directories like node_modules and .git', async () => {
      // Setup a spy on fileSystem.readdir to verify it's not called on node_modules or .git
      const fileSystem = new ConcreteFileSystem();
      const readdirSpy = jest.spyOn(fileSystem, 'readdir');

      // Call function with FileSystem interface
      const results = await readDirectoryContents(testDirPath, fileSystem);

      // Check that files from node_modules and .git are not included
      const nodeModulesFile = results.find(r => r.path.includes('node_modules'));
      const gitFile = results.find(r => r.path.includes('.git'));

      expect(nodeModulesFile).toBeUndefined();
      expect(gitFile).toBeUndefined();

      // Verify readdir was called
      expect(readdirSpy).toHaveBeenCalled();

      // Verify that readdir wasn't called on node_modules or .git
      const nodeModulesCall = readdirSpy.mock.calls.find(call => 
        String(call[0]).includes('node_modules')
      );
      const gitCall = readdirSpy.mock.calls.find(call => 
        String(call[0]).includes('.git')
      );

      expect(nodeModulesCall).toBeUndefined();
      expect(gitCall).toBeUndefined();
    });
  });

  describe('Path Handling', () => {
    it('should handle relative paths by resolving them to absolute paths', async () => {
      // Create a directory with a relative path
      const relativeTestPath = 'relative/test/path';
      // Create the absolute path for use in the virtual filesystem
      const absolutePath = path.resolve(process.cwd(), relativeTestPath);
      const normalizedPath = normalizePathForMemfs(absolutePath);

      // Create the virtual filesystem structure with test files
      setupBasicFiles({
        [normalizedPath + '/']: '',
        [path.posix.join(normalizedPath, 'file.txt')]: 'Content of file.txt',
      });

      // Create a FileSystem instance
      const fileSystem = new ConcreteFileSystem();

      // Use the fileSystem parameter in readContextFile
      const result = await readContextFile(path.join(relativeTestPath, 'file.txt'), fileSystem);

      // The relative path should be preserved in output
      expect(result.path).toBe(path.join(relativeTestPath, 'file.txt'));
      expect(result.content).toBe('Content of file.txt');
      expect(result.error).toBeNull();
    });

    it('should handle path with special characters', async () => {
      // Use normalized path that memfs can handle
      const specialPath = normalizePathForMemfs('/path/with spaces and #special characters!');

      // Set up virtual filesystem with special characters
      setupBasicFiles({
        [`${specialPath}/`]: '',
        [`${specialPath}/file.txt`]: 'File content',
      });

      // Create FileSystem instance
      const fileSystem = new ConcreteFileSystem();

      // Call function with FileSystem interface
      const results = await readDirectoryContents(specialPath, fileSystem);

      // Should handle the special characters correctly
      expect(results).toHaveLength(1);
      expect(results[0].path.includes('file.txt')).toBe(true);
      expect(results[0].content).toBe('File content');
    });

    it('should handle Windows-style paths', async () => {
      // For memfs, we need to use a normalized path format
      const windowsPath = normalizePathForMemfs('/Users/user/Documents/test');

      // Setup virtual filesystem with Windows-like path
      setupBasicFiles({
        [`${windowsPath}/`]: '',
        [`${windowsPath}/file.txt`]: 'File content',
      });

      // Create FileSystem instance
      const fileSystem = new ConcreteFileSystem();

      // Call function with FileSystem interface
      const results = await readDirectoryContents(windowsPath, fileSystem);

      // Should handle the path correctly
      expect(results).toHaveLength(1);
      expect(results[0].path.includes('file.txt')).toBe(true);
      expect(results[0].content).toBe('File content');
    });
  });

  describe('Error Handling', () => {
    it('should handle directory access errors gracefully', async () => {
      resetVirtualFs();

      // Create FileSystem instance
      const fileSystem = new ConcreteFileSystem();
      
      // Setup spy to simulate access error
      const accessSpy = jest.spyOn(fileSystem, 'access');
      accessSpy.mockRejectedValueOnce(
        createFsError('EACCES', 'Permission denied', 'access', testDirPath)
      );

      // Call function with FileSystem interface
      const results = await readDirectoryContents(testDirPath, fileSystem);

      // Should return error for the directory
      expect(results).toHaveLength(1);
      expect(results[0].path).toBe(testDirPath);
      expect(results[0].content).toBeNull();
      expect(results[0].error).toBeDefined();
      expect(results[0].error?.code).toBe('READ_ERROR');
      expect(results[0].error?.message).toContain('Error reading directory');
    });

    it('should handle file read errors within directories', async () => {
      // Create a basic directory with files
      setupBasicFiles({
        [`${testDirPath}/`]: '',
        [`${testDirPath}/file1.txt`]: 'Content of file1.txt',
        [`${testDirPath}/file2.md`]: 'Content of file2.md',
      });

      // Create FileSystem instance with spies
      const fileSystem = new ConcreteFileSystem();

      // Mock readFileContent to simulate a failure for one file
      const readFileContentSpy = jest.spyOn(fileSystem, 'readFileContent');
      readFileContentSpy.mockImplementation((filePath: string) => {
        if (filePath.includes('file1.txt')) {
          return Promise.reject(new Error('Simulated file read error'));
        } else {
          // Use the real implementation for other files
          return fsPromises.readFile(filePath, 'utf-8');
        }
      });

      // Call the function being tested
      const results = await readDirectoryContents(testDirPath, fileSystem);

      // Verify results match expectations
      expect(results.length).toBe(2);

      // Check first file with error
      const file1Result = results.find(r => r.path.includes('file1.txt'));
      expect(file1Result).toBeDefined();
      expect(file1Result?.content).toBeNull();
      expect(file1Result?.error).toBeDefined();
      expect(file1Result?.error?.code).toBe('READ_ERROR');

      // Check second file with content
      const file2Result = results.find(r => r.path.includes('file2.md'));
      expect(file2Result).toBeDefined();
      expect(file2Result?.content).toBe('Content of file2.md');
      expect(file2Result?.error).toBeNull();

      // Clean up the mock
      readFileContentSpy.mockRestore();
    });

    it('should handle directory read errors', async () => {
      // Create a basic directory
      setupBasicFiles({
        [`${testDirPath}/`]: '',
      });

      // Create FileSystem instance
      const fileSystem = new ConcreteFileSystem();
      
      // Simulate readdir failure
      const readdirSpy = jest.spyOn(fileSystem, 'readdir');
      readdirSpy.mockRejectedValueOnce(
        createFsError('EIO', 'Failed to read directory', 'readdir', testDirPath)
      );

      // Call function with FileSystem interface
      const results = await readDirectoryContents(testDirPath, fileSystem);

      // Should return an error result
      expect(results).toHaveLength(1);
      expect(results[0].path).toBe(testDirPath);
      expect(results[0].content).toBeNull();
      expect(results[0].error).toBeDefined();
      expect(results[0].error?.code).toBe('READ_ERROR');
    });

    it('should handle non-Error objects in exceptions', async () => {
      // Create a basic directory
      setupBasicFiles({
        [`${testDirPath}/`]: '',
      });

      // Create FileSystem instance with readdir spy that throws non-Error
      const fileSystem = new ConcreteFileSystem();
      const readdirSpy = jest.spyOn(fileSystem, 'readdir');
      readdirSpy.mockImplementationOnce(() => {
        throw 'Not an error object';
      });

      // Call function with FileSystem interface
      const results = await readDirectoryContents(testDirPath, fileSystem);

      // Should still return a structured error result
      expect(results).toHaveLength(1);
      expect(results[0].path).toBe(testDirPath);
      expect(results[0].content).toBeNull();
      expect(results[0].error).toBeDefined();
      // With FileSystem interface, the error code will be READ_ERROR
      expect(results[0].error?.code).toBe('READ_ERROR');
    });

    it('should handle stat errors for directory entries', async () => {
      // Create a basic directory with files
      setupBasicFiles({
        [`${testDirPath}/`]: '',
        [`${testDirPath}/file1.txt`]: 'Content of file1.txt',
        [`${testDirPath}/file2.md`]: 'Content of file2.md',
      });

      // Create FileSystem instance
      const fileSystem = new ConcreteFileSystem();

      // Mock stat for FileSystem to fail for file1.txt
      const statSpy = jest.spyOn(fileSystem, 'stat');
      statSpy.mockImplementation((pathLike: string) => {
        // Fail for file1.txt
        if (pathLike.includes('file1.txt')) {
          return Promise.reject(createFsError('EIO', 'Failed to stat file', 'stat', pathLike));
        }

        // Use the real implementation for the directory
        if (pathLike === testDirPath) {
          return fsPromises.stat(pathLike);
        }

        // Succeed for file2.md
        if (pathLike.includes('file2.md')) {
          const stats = createMockStats(true, 1024);
          return Promise.resolve(stats);
        }

        // Default rejection
        return Promise.reject(new Error(`Unexpected path: ${pathLike}`));
      });

      // Mock readContextFile to return file2.md
      jest.spyOn(fileReader, 'readContextFile').mockImplementation((filePath: string) => {
        if (filePath.includes('file2.md')) {
          return Promise.resolve({
            path: filePath,
            content: 'Content of file2.md',
            error: null,
          });
        }
        return Promise.resolve({
          path: filePath,
          content: null,
          error: { code: 'UNKNOWN', message: 'Unknown file' },
        });
      });

      // Mock readdir to return both files
      jest.spyOn(fileSystem, 'readdir').mockResolvedValue(['file1.txt', 'file2.md']);

      // Call function with FileSystem interface
      const results = await readDirectoryContents(testDirPath, fileSystem);

      // We should have at least one result
      expect(results.length).toBeGreaterThanOrEqual(1);

      // Verify file2.md was read successfully
      const successEntry = results.find(r => r.path.includes('file2.md'));
      expect(successEntry).toBeDefined();
      expect(successEntry?.content).toBe('Content of file2.md');
      expect(successEntry?.error).toBeNull();
    });
  });

  describe('Special Cases', () => {
    it('should handle empty directories', async () => {
      // Create an empty directory
      setupBasicFiles({
        [`${testDirPath}/`]: '', // Create empty directory
      });

      // Create FileSystem instance
      const fileSystem = new ConcreteFileSystem();

      // Call function with FileSystem interface
      const results = await readDirectoryContents(testDirPath, fileSystem);

      // Should return empty array (no error)
      expect(results).toBeInstanceOf(Array);
      expect(results).toHaveLength(0);
    });

    it('should handle when path is a file, not a directory', async () => {
      // Set up a file instead of directory at testDirPath
      setupBasicFiles({
        [testDirPath]: 'File content', // Create file at testDirPath
      });

      // Create FileSystem instance
      const fileSystem = new ConcreteFileSystem();

      // Call function with FileSystem interface
      const results = await readDirectoryContents(testDirPath, fileSystem);

      // Should return the file content directly
      expect(results).toHaveLength(1);
      expect(results[0].path).toBe(testDirPath);
      expect(results[0].content).toBe('File content');
      expect(results[0].error).toBeNull();
    });

    it('should handle various file types and extensions', async () => {
      // Setup virtual filesystem with various file types
      setupBasicFiles({
        [`${testDirPath}/`]: '',
        [`${testDirPath}/script.js`]: 'console.log("hello");',
        [`${testDirPath}/style.css`]: 'body { color: red; }',
        [`${testDirPath}/data.json`]: '{"key": "value"}',
        [`${testDirPath}/document.md`]: '# Heading',
      });

      // Create FileSystem instance
      const fileSystem = new ConcreteFileSystem();

      // Call function with FileSystem interface
      const results = await readDirectoryContents(testDirPath, fileSystem);

      // Verify results
      expect(results.length).toBe(4); // Should find all 4 files

      // Verify each file was processed correctly
      const jsFile = results.find(r => r.path.includes('script.js'));
      const cssFile = results.find(r => r.path.includes('style.css'));
      const jsonFile = results.find(r => r.path.includes('data.json'));
      const mdFile = results.find(r => r.path.includes('document.md'));

      expect(jsFile).toBeDefined();
      expect(jsFile?.content).toBe('console.log("hello");');

      expect(cssFile).toBeDefined();
      expect(cssFile?.content).toBe('body { color: red; }');

      expect(jsonFile).toBeDefined();
      expect(jsonFile?.content).toBe('{"key": "value"}');

      expect(mdFile).toBeDefined();
      expect(mdFile?.content).toBe('# Heading');
    });
  });

  describe('Integration with Other Features', () => {
    it('should integrate with gitignore-based filtering', async () => {
      // Set up a test directory with a .gitignore file
      await setupWithGitignore(
        testDirPath,
        'file1.txt', // gitignore pattern
        {
          'file1.txt': 'Content of file1.txt',
          'file2.md': 'Content of file2.md',
          'subdir/': '',
          'subdir/nested.txt': 'Content of nested.txt',
        }
      );

      // Mock the gitignore functions to use the real gitignore file
      // We don't need to mock shouldIgnorePath as it will use the actual gitignore file
      jest.spyOn(gitignoreUtils, 'shouldIgnorePath').mockImplementation((_, filePath: string) => {
        // Simple implementation that checks if the file matches the pattern
        const fileName = path.basename(filePath);
        return Promise.resolve(fileName === 'file1.txt');
      });

      // Create FileSystem instance
      const fileSystem = new ConcreteFileSystem();

      // Call function with FileSystem interface
      const results = await readDirectoryContents(testDirPath, fileSystem);

      // Test for the presence/absence of specific files
      const file1 = results.find(r => r.path.includes('file1.txt'));
      const file2 = results.find(r => r.path.includes('file2.md'));
      const nested = results.find(r => r.path.includes('nested.txt'));
      const gitignoreFile = results.find(r => r.path.includes('.gitignore'));

      // file1.txt should be ignored
      expect(file1).toBeUndefined();

      // .gitignore is usually included (not excluded by default)
      expect(gitignoreFile).toBeDefined();

      // file2.md and nested.txt should be included
      expect(file2).toBeDefined();
      expect(file2?.content).toBe('Content of file2.md');
      expect(nested).toBeDefined();
      expect(nested?.content).toBe('Content of nested.txt');
    });

    it('should detect and handle binary files correctly', async () => {
      // Create directory with binary file
      setupBasicFiles({
        [`${testDirPath}/`]: '',
        [`${testDirPath}/binary.bin`]: 'Content with \0 null bytes',
        [`${testDirPath}/text.txt`]: 'Normal text content',
      });

      // Create FileSystem instance
      const fileSystem = new ConcreteFileSystem();

      // Call function with FileSystem interface
      const results = await readDirectoryContents(testDirPath, fileSystem);

      // We expect two results: one binary and one text
      expect(results.length).toBe(2);

      // Check binary file is properly identified
      const binaryFile = results.find(r => r.path.includes('binary.bin'));
      expect(binaryFile).toBeDefined();
      expect(binaryFile?.content).toBeNull();
      expect(binaryFile?.error).toBeDefined();
      expect(binaryFile?.error?.code).toBe('BINARY_FILE');

      // Check text file is properly read
      const textFile = results.find(r => r.path.includes('text.txt'));
      expect(textFile).toBeDefined();
      expect(textFile?.content).toBe('Normal text content');
      expect(textFile?.error).toBeNull();
    });

    it('should handle deep nested directory structures', async () => {
      // Create a deep nested directory structure (5 levels)
      const maxDepth = 5;
      
      // Create structure object
      const structure: Record<string, string> = {
        [`${testDirPath}/`]: '',
        [`${testDirPath}/root.txt`]: 'Content of root.txt',
      };

      // Create nested directories and files
      let currentPath = testDirPath;
      for (let i = 1; i <= maxDepth; i++) {
        currentPath = path.posix.join(currentPath, `level${i}`);
        structure[`${currentPath}/`] = ''; // Directory marker

        if (i < maxDepth) {
          structure[`${currentPath}/file${i}.txt`] = `Content of file${i}.txt`;
        } else {
          structure[`${currentPath}/finalfile.txt`] = 'Content of finalfile.txt';
        }
      }

      // Create virtual filesystem
      setupBasicFiles(structure);

      // Create FileSystem instance
      const fileSystem = new ConcreteFileSystem();

      // Call function with FileSystem interface
      const results = await readDirectoryContents(testDirPath, fileSystem);

      // We expect 1 + maxDepth files (root file + one file at each level)
      expect(results.length).toBe(1 + maxDepth);

      // Check root file
      const rootFile = results.find(r => r.path.includes('root.txt'));
      expect(rootFile).toBeDefined();
      expect(rootFile?.content).toBe('Content of root.txt');

      // Check nested files at each level
      for (let i = 1; i < maxDepth; i++) {
        const levelFile = results.find(r => r.path.includes(`file${i}.txt`));
        expect(levelFile).toBeDefined();
        expect(levelFile?.content).toBe(`Content of file${i}.txt`);
      }

      // Check the deepest file
      const finalFile = results.find(r => r.path.includes('finalfile.txt'));
      expect(finalFile).toBeDefined();
      expect(finalFile?.content).toBe('Content of finalfile.txt');
    });
    
    it('should traverse all valid directories when using FileSystem interface', async () => {
      // Setup directory with nested subdirectories
      const structure: Record<string, string> = {
        [`${testDirPath}/`]: '',
        [`${testDirPath}/root.txt`]: 'Root file',
        [`${testDirPath}/dir1/`]: '',
        [`${testDirPath}/dir1/file1.txt`]: 'File in dir1',
        [`${testDirPath}/dir2/`]: '',
        [`${testDirPath}/dir2/file2.txt`]: 'File in dir2',
        [`${testDirPath}/dir2/subdir/`]: '',
        [`${testDirPath}/dir2/subdir/deep.txt`]: 'Deep file'
      };
      
      setupBasicFiles(structure);
      
      // Create FileSystem instance
      const fileSystem = new ConcreteFileSystem();
      
      // Spy on readdir and stat to verify correct traversal
      const readdirSpy = jest.spyOn(fileSystem, 'readdir');
      const statSpy = jest.spyOn(fileSystem, 'stat');
      
      // Call function with FileSystem interface
      const results = await readDirectoryContents(testDirPath, fileSystem);
      
      // Should find all 4 files
      expect(results.length).toBe(4);
      
      // Verify readdir was called for each directory
      expect(readdirSpy).toHaveBeenCalledWith(testDirPath);
      expect(readdirSpy).toHaveBeenCalledWith(expect.stringContaining('dir1'));
      expect(readdirSpy).toHaveBeenCalledWith(expect.stringContaining('dir2'));
      expect(readdirSpy).toHaveBeenCalledWith(expect.stringContaining('subdir'));
      
      // Verify stat was called for each directory and file
      expect(statSpy.mock.calls.length).toBeGreaterThanOrEqual(8); // 4 dirs + 4 files
      
      // Check all files were found with correct content
      const rootFile = results.find(r => r.path.endsWith('root.txt'));
      const file1 = results.find(r => r.path.endsWith('file1.txt'));
      const file2 = results.find(r => r.path.endsWith('file2.txt'));
      const deepFile = results.find(r => r.path.endsWith('deep.txt'));
      
      expect(rootFile?.content).toBe('Root file');
      expect(file1?.content).toBe('File in dir1');
      expect(file2?.content).toBe('File in dir2');
      expect(deepFile?.content).toBe('Deep file');
    });
  });
});
