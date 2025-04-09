/**
 * Tests for gitignore test helpers
 */
// path module is used internally by the helpers we're testing
import { setupTestHooks } from '../common';
import { getFs } from '../fs';
import {
  addGitignoreFile,
  setupBasicGitignore,
  setupWithGitignore,
  setupMultiGitignore,
  createIgnoreChecker
} from '../gitignore';
import { clearIgnoreCache } from '../../../src/utils/gitignoreUtils';
import { normalizePathForMemfs } from '../../../src/__tests__/utils/virtualFsUtils';

describe('gitignore test helpers', () => {
  // Set up hooks to reset virtual filesystem and mocks for each test
  setupTestHooks();

  // Helper to read a file from the virtual filesystem
  const readVirtualFile = (filePath: string): string => {
    const normalizedPath = normalizePathForMemfs(filePath);
    return getFs().readFileSync(normalizedPath, 'utf8') as string;
  };

  // Helper to check if a file/directory exists in the virtual filesystem
  const fileExists = (filePath: string): boolean => {
    const normalizedPath = normalizePathForMemfs(filePath);
    return getFs().existsSync(normalizedPath);
  };
  
  describe('addGitignoreFile', () => {
    it('should create a .gitignore file with content', async () => {
      // Action
      const gitignorePath = '/test-project/.gitignore';
      const content = '*.log\nnode_modules/';
      await addGitignoreFile(gitignorePath, content);
      
      // Assert
      expect(fileExists(gitignorePath)).toBe(true);
      expect(readVirtualFile(gitignorePath)).toBe(content);
    });
    
    it('should overwrite an existing .gitignore file', async () => {
      // Setup
      const gitignorePath = '/test-project/.gitignore';
      const initialContent = 'initial-content';
      await addGitignoreFile(gitignorePath, initialContent);
      
      // Action
      const newContent = '*.log\nnode_modules/';
      await addGitignoreFile(gitignorePath, newContent);
      
      // Assert
      expect(readVirtualFile(gitignorePath)).toBe(newContent);
    });
    
    it('should handle nested paths correctly', async () => {
      // Action
      const nestedPath = '/test-project/nested/deep/.gitignore';
      const content = '*.tmp';
      await addGitignoreFile(nestedPath, content);
      
      // Assert
      expect(fileExists('/test-project/nested/deep')).toBe(true);
      expect(fileExists(nestedPath)).toBe(true);
      expect(readVirtualFile(nestedPath)).toBe(content);
    });
  });
  
  describe('setupBasicGitignore', () => {
    it('should create default .gitignore at /project/.gitignore', async () => {
      // Action
      await setupBasicGitignore();
      
      // Assert
      expect(fileExists('/project/.gitignore')).toBe(true);
      const content = readVirtualFile('/project/.gitignore');
      expect(content).toContain('node_modules/');
      expect(content).toContain('*.log');
      expect(content).toContain('.DS_Store');
      expect(content).toContain('/dist/');
    });
    
    it('should create .gitignore at custom path', async () => {
      // Action
      await setupBasicGitignore('/custom-project');
      
      // Assert
      expect(fileExists('/custom-project/.gitignore')).toBe(true);
      const content = readVirtualFile('/custom-project/.gitignore');
      expect(content).toContain('node_modules/');
    });
    
    it('should use custom patterns', async () => {
      // Action
      const customPatterns = '*.txt\n/build/';
      await setupBasicGitignore('/project', customPatterns);
      
      // Assert
      expect(readVirtualFile('/project/.gitignore')).toBe(customPatterns);
    });
  });
  
  describe('setupWithGitignore', () => {
    it('should create directory structure and .gitignore file', async () => {
      // Action
      await setupWithGitignore('/test-project', '*.log', {
        'src/index.ts': 'console.log("Hello");',
        'debug.log': 'Log content'
      });
      
      // Assert
      // Verify directory structure
      expect(fileExists('/test-project/src')).toBe(true);
      expect(fileExists('/test-project/src/index.ts')).toBe(true);
      expect(fileExists('/test-project/debug.log')).toBe(true);
      
      // Verify file contents
      expect(readVirtualFile('/test-project/src/index.ts')).toBe('console.log("Hello");');
      expect(readVirtualFile('/test-project/debug.log')).toBe('Log content');
      
      // Verify .gitignore file
      expect(fileExists('/test-project/.gitignore')).toBe(true);
      expect(readVirtualFile('/test-project/.gitignore')).toBe('*.log');
    });
    
    it('should respect reset: false option', async () => {
      // Setup - Create initial files
      await setupWithGitignore('/test-project', '*.log', {
        'first.txt': 'First file'
      });
      
      // Action - Add more files without resetting
      await setupWithGitignore('/test-project', '*.txt', {
        'second.txt': 'Second file'
      }, { reset: false });
      
      // Assert
      // Both files should exist
      expect(fileExists('/test-project/first.txt')).toBe(true);
      expect(fileExists('/test-project/second.txt')).toBe(true);
      
      // .gitignore content should be updated
      expect(readVirtualFile('/test-project/.gitignore')).toBe('*.txt');
    });
    
    it('should handle nested project files', async () => {
      // Action
      await setupWithGitignore('/test-project', '*.log', {
        'src/components/button.tsx': 'export const Button = () => <button>Click</button>;',
        'src/utils/helpers.ts': 'export const add = (a, b) => a + b;',
        'logs/debug.log': 'Debug log content'
      });
      
      // Assert
      expect(fileExists('/test-project/src/components')).toBe(true);
      expect(fileExists('/test-project/src/utils')).toBe(true);
      expect(fileExists('/test-project/logs')).toBe(true);
      
      expect(readVirtualFile('/test-project/src/components/button.tsx'))
        .toBe('export const Button = () => <button>Click</button>;');
      expect(readVirtualFile('/test-project/src/utils/helpers.ts'))
        .toBe('export const add = (a, b) => a + b;');
      expect(readVirtualFile('/test-project/logs/debug.log'))
        .toBe('Debug log content');
    });
  });
  
  describe('setupMultiGitignore', () => {
    it('should create complex structure with multiple gitignore files', async () => {
      // Action
      await setupMultiGitignore('/test-project', {
        '.gitignore': '*.log\n/dist/',
        'src/.gitignore': '*.tmp\n*.cache',
        'src/components/.gitignore': '*.bak'
      }, {
        'src/index.ts': 'console.log("Hello");',
        'src/temp.tmp': 'Temporary file',
        'src/components/button.tsx': 'Button component',
        'src/components/button.bak': 'Old button',
        'dist/bundle.js': 'Bundle content',
        'debug.log': 'Debug log'
      });
      
      // Assert
      // Verify gitignore files
      expect(readVirtualFile('/test-project/.gitignore')).toBe('*.log\n/dist/');
      expect(readVirtualFile('/test-project/src/.gitignore')).toBe('*.tmp\n*.cache');
      expect(readVirtualFile('/test-project/src/components/.gitignore')).toBe('*.bak');
      
      // Verify project files
      expect(readVirtualFile('/test-project/src/index.ts')).toBe('console.log("Hello");');
      expect(readVirtualFile('/test-project/src/temp.tmp')).toBe('Temporary file');
      expect(readVirtualFile('/test-project/src/components/button.tsx')).toBe('Button component');
      expect(readVirtualFile('/test-project/src/components/button.bak')).toBe('Old button');
      expect(readVirtualFile('/test-project/dist/bundle.js')).toBe('Bundle content');
      expect(readVirtualFile('/test-project/debug.log')).toBe('Debug log');
    });
    
    it('should create snapshot of the directory structure', async () => {
      // Action
      await setupMultiGitignore('/project', {
        '.gitignore': '*.log',
        'src/.gitignore': '*.tmp'
      }, {
        'README.md': '# Project',
        'src/index.js': 'console.log("Hello");'
      });
      
      // Assert using snapshot
      // Get a representation of the filesystem structure for snapshot testing
      const vfs = getFs();
      
      // This creates a simpler representation for snapshot testing
      const fsStructure = {
        files: Object.keys(vfs.toJSON()).sort(),
        gitignores: {
          root: readVirtualFile('/project/.gitignore'),
          src: readVirtualFile('/project/src/.gitignore')
        }
      };
      
      // Use a snapshot to quickly verify the whole structure
      expect(fsStructure).toMatchSnapshot();
    });
  });
  
  describe('createIgnoreChecker', () => {
    beforeEach(() => {
      // Clear the gitignore cache before each test
      clearIgnoreCache();
    });
    
    it('should return a function that respects root .gitignore', async () => {
      // Setup
      await setupWithGitignore('/project', '*.log\nnode_modules/', {
        'app.js': 'console.log("Hello");',
        'error.log': 'Error log content',
        'node_modules/package/index.js': 'module code'
      });
      
      // Action
      const checker = createIgnoreChecker('/project');
      
      // Assert
      expect(await checker('app.js')).toBe(false); // Not ignored
      expect(await checker('error.log')).toBe(true); // Ignored by *.log
      expect(await checker('node_modules/package/index.js')).toBe(true); // Ignored by node_modules/
    });
    
    it('should check directly in the base directory for .gitignore', async () => {
      // Setup
      await setupMultiGitignore('/project', {
        '.gitignore': '*.log\n*.tmp', // Include *.tmp in root gitignore
        'src/.gitignore': '*.js' // This won't be checked by the root checker
      }, {
        'debug.log': 'Should be ignored by root',
        'src/index.js': 'Would be ignored by src/.gitignore but not by root',
        'src/cache.tmp': 'Should be ignored by root .gitignore'
      });
      
      // Action
      const rootChecker = createIgnoreChecker('/project');
      
      // Assert - rootChecker only checks /project/.gitignore
      expect(await rootChecker('debug.log')).toBe(true); // Ignored by root .gitignore
      expect(await rootChecker('src/index.js')).toBe(false); // Not ignored by root
      expect(await rootChecker('src/cache.tmp')).toBe(true); // Ignored by root .gitignore
      
      // Create a separate checker for the src directory
      const srcChecker = createIgnoreChecker('/project/src');
      
      // This checker should respect src/.gitignore
      expect(await srcChecker('index.js')).toBe(true); // Ignored by src/.gitignore
    });
    
    it('should handle paths relative to the checker baseDir', async () => {
      // Setup
      await setupWithGitignore('/deep/nested/project', '*.log', {
        'app.js': 'console.log("Hello");',
        'error.log': 'Error log content'
      });
      
      // Action - Create a checker with the deep path as baseDir
      const checker = createIgnoreChecker('/deep/nested/project');
      
      // Assert - Paths should be relative to the baseDir
      expect(await checker('app.js')).toBe(false); // Not ignored
      expect(await checker('error.log')).toBe(true); // Ignored by *.log
    });
    
    it('should handle negated patterns in .gitignore', async () => {
      // Setup - Use negated pattern to NOT ignore a specific log file
      await setupWithGitignore('/project', '*.log\n!important.log', {
        'debug.log': 'Should be ignored',
        'important.log': 'Should NOT be ignored despite *.log'
      });
      
      // Action
      const checker = createIgnoreChecker('/project');
      
      // Assert
      expect(await checker('debug.log')).toBe(true); // Ignored by *.log
      expect(await checker('important.log')).toBe(false); // Not ignored due to !important.log
    });
  });
});
