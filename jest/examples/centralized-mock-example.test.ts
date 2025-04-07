/**
 * Example test file demonstrating the centralized mock setup
 * 
 * This file shows how to write tests using the centralized mock configuration,
 * reducing boilerplate and standardizing the test setup process.
 */

// Import the fs module after mocking is set up globally
import fsPromises from 'fs/promises';
import path from 'path';

// Import test helpers (these would normally be imported from a utility module)
const { readFile } = async (filePath: string): Promise<string> => {
  try {
    return await fsPromises.readFile(filePath, 'utf-8');
  } catch (error) {
    throw new Error(`Failed to read file: ${filePath}`);
  }
};

// Import the fs setup helper
import { setupBasicFs } from '../setupFiles/fs';
// Import the gitignore setup helper
import { setupBasicGitignore, addGitignoreFile } from '../setupFiles/gitignore';
// Import directly from virtualFsUtils still works
import { resetVirtualFs, createVirtualFs } from '../../src/__tests__/utils/virtualFsUtils';

describe('Centralized Mock Example', () => {
  // Setup test files
  const testDir = '/test/project';
  const configFile = path.join(testDir, 'config.json');
  const logFile = path.join(testDir, 'app.log');
  const srcFile = path.join(testDir, 'src/app.js');

  describe('Using direct virtualFsUtils API', () => {
    beforeEach(() => {
      // Reset the virtual filesystem directly
      resetVirtualFs();
      
      // Create test files using direct API
      createVirtualFs({
        [configFile]: JSON.stringify({ name: 'test-project' }),
        [logFile]: 'Log content',
        [srcFile]: 'console.log("Hello");'
      });
    });

    it('should read configuration file', async () => {
      const content = await readFile(configFile);
      expect(content).toBe(JSON.stringify({ name: 'test-project' }));
    });

    it('should access nested files', async () => {
      const content = await readFile(srcFile);
      expect(content).toBe('console.log("Hello");');
    });
  });

  describe('Using centralized setup helpers', () => {
    beforeEach(() => {
      // Use the helper to set up the filesystem
      setupBasicFs({
        [configFile]: JSON.stringify({ name: 'test-project' }),
        [logFile]: 'Log content',
        [srcFile]: 'console.log("Hello");'
      });
      
      // Set up gitignore mocking
      setupBasicGitignore();
      
      // Add a gitignore file to ignore logs
      addGitignoreFile(path.join(testDir, '.gitignore'), '*.log');
    });

    it('should read configuration file with centralized setup', async () => {
      const content = await readFile(configFile);
      expect(content).toBe(JSON.stringify({ name: 'test-project' }));
    });
  });

  describe('Error Handling', () => {
    beforeEach(() => {
      // Just set up a minimal fs
      setupBasicFs({
        '/existing.txt': 'I exist'
      });
    });

    it('should handle file not found', async () => {
      // The file was not created in the virtual fs
      await expect(readFile('/missing.txt')).rejects.toThrow('Failed to read file');
    });
  });
});