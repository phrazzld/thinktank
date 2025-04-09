/**
 * Tests for the audit-fs-mocks script
 * 
 * These tests verify that the script correctly identifies mocking patterns
 * in test files and generates appropriate reports.
 */

const fs = require('fs').promises;
const path = require('path');
const os = require('os');

// Mock modules before requiring the script
const mockReadFile = jest.fn();
const mockWriteFile = jest.fn();
const mockGlob = jest.fn();

jest.mock('fs', () => ({
  promises: {
    readFile: mockReadFile,
    writeFile: mockWriteFile
  }
}));

jest.mock('fast-glob', () => mockGlob);

// Import the script functions
const { 
  patterns, 
  analyzeFile, 
  generateMarkdownReport,
  auditFsMocks 
} = require('../audit-fs-mocks');

describe('Audit FS Mocks Script', () => {
  // Mock logger for testing
  const mockLogger = jest.fn();
  
  // Reset mocks before each test
  beforeEach(() => {
    jest.clearAllMocks();
    mockReadFile.mockResolvedValue('');
    mockWriteFile.mockResolvedValue(undefined);
    mockGlob.mockResolvedValue([]);
  });
  
  describe('Pattern Matching', () => {
    it('identifies direct fs mocking', () => {
      const contentWithFsMock = `
        jest.mock('fs', () => ({
          readFileSync: jest.fn()
        }));
      `;
      
      expect(patterns.directMock.fs.test(contentWithFsMock)).toBe(true);
      
      const contentWithoutFsMock = `
        jest.mock('path');
        jest.mock('./someModule');
      `;
      
      expect(patterns.directMock.fs.test(contentWithoutFsMock)).toBe(false);
    });
    
    it('identifies direct fs/promises mocking', () => {
      const contentWithFsPromisesMock = `
        jest.mock('fs/promises', () => ({
          readFile: jest.fn()
        }));
      `;
      
      expect(patterns.directMock.fsPromises.test(contentWithFsPromisesMock)).toBe(true);
      
      const contentWithoutFsPromisesMock = `
        jest.mock('path');
        jest.mock('./someModule');
      `;
      
      expect(patterns.directMock.fsPromises.test(contentWithoutFsPromisesMock)).toBe(false);
    });
    
    it('identifies legacy utils imports', () => {
      const contentWithLegacyImport = `
        import { mockFs, setupMockFs } from '../../__tests__/utils/mockFsUtils';
      `;
      
      expect(patterns.legacyUtil.import.test(contentWithLegacyImport)).toBe(true);
      
      const contentWithoutLegacyImport = `
        import { something } from '../../utils/helpers';
      `;
      
      expect(patterns.legacyUtil.import.test(contentWithoutLegacyImport)).toBe(false);
    });
    
    it('identifies legacy utils function usage', () => {
      const contentWithLegacyUsage = `
        test('something', () => {
          mockFs({ '/path/to/file.txt': 'content' });
          setupMockFs({ files: {} });
          expect(mockReadFile).toHaveBeenCalled();
          mockWriteFile('/path/to/file.txt', 'new content');
        });
      `;
      
      expect(patterns.legacyUtil.usage.test(contentWithLegacyUsage)).toBe(true);
      
      const contentWithoutLegacyUsage = `
        test('something', () => {
          fs.readFileSync('/path/to/file.txt');
          expect(something).toBe(true);
        });
      `;
      
      expect(patterns.legacyUtil.usage.test(contentWithoutLegacyUsage)).toBe(false);
    });
    
    it('identifies virtual filesystem imports', () => {
      const contentWithVirtualFsImport = `
        import { setupBasicFs } from '../../../test/setup/fs';
        import { normalizePathForMemfs } from '../../__tests__/utils/virtualFsUtils';
      `;
      
      expect(patterns.virtualFs.import.test(contentWithVirtualFsImport)).toBe(true);
      
      const contentWithoutVirtualFsImport = `
        import { something } from '../../utils/helpers';
      `;
      
      expect(patterns.virtualFs.import.test(contentWithoutVirtualFsImport)).toBe(false);
    });
    
    it('identifies virtual filesystem function usage', () => {
      const contentWithVirtualFsUsage = `
        test('something', () => {
          setupBasicFs({ '/path/to/file.txt': 'content' });
          createVirtualFs({ files: {} });
          addVirtualGitignoreFile('/path', 'pattern');
          const path = normalizePathForMemfs('/some/path');
        });
      `;
      
      expect(patterns.virtualFs.usage.test(contentWithVirtualFsUsage)).toBe(true);
      
      const contentWithoutVirtualFsUsage = `
        test('something', () => {
          fs.readFileSync('/path/to/file.txt');
          expect(something).toBe(true);
        });
      `;
      
      expect(patterns.virtualFs.usage.test(contentWithoutVirtualFsUsage)).toBe(false);
    });
  });
  
  describe('analyzeFile function', () => {
    it('categorizes direct fs mocking correctly', () => {
      const content = `
        jest.mock('fs', () => ({
          readFileSync: jest.fn()
        }));
      `;
      
      const result = analyzeFile('test.ts', content);
      
      expect(result.category).toBe('Direct Mock');
      expect(result.hasDirectFsMock).toBe(true);
      expect(result.hasLegacyUtil).toBe(false);
      expect(result.hasVirtualFs).toBe(false);
    });
    
    it('categorizes legacy util usage correctly', () => {
      const content = `
        import { mockFs } from '../../__tests__/utils/mockFsUtils';
        
        test('something', () => {
          mockFs({ '/path/to/file.txt': 'content' });
        });
      `;
      
      const result = analyzeFile('test.ts', content);
      
      expect(result.category).toBe('Legacy Util');
      expect(result.hasDirectFsMock).toBe(false);
      expect(result.hasLegacyUtil).toBe(true);
      expect(result.hasVirtualFs).toBe(false);
    });
    
    it('categorizes mixed usage correctly', () => {
      const content = `
        jest.mock('fs', () => ({ readFileSync: jest.fn() }));
        import { setupBasicFs } from '../../../test/setup/fs';
        
        test('something', () => {
          setupBasicFs({ '/path/to/file.txt': 'content' });
        });
      `;
      
      const result = analyzeFile('test.ts', content);
      
      expect(result.category).toBe('Mixed');
      expect(result.hasDirectFsMock).toBe(true);
      expect(result.hasLegacyUtil).toBe(false);
      expect(result.hasVirtualFs).toBe(true);
    });
    
    it('categorizes virtual fs usage correctly', () => {
      const content = `
        import { setupBasicFs } from '../../../test/setup/fs';
        
        test('something', () => {
          setupBasicFs({ '/path/to/file.txt': 'content' });
        });
      `;
      
      const result = analyzeFile('test.ts', content);
      
      expect(result.category).toBe('Virtual FS');
      expect(result.hasDirectFsMock).toBe(false);
      expect(result.hasLegacyUtil).toBe(false);
      expect(result.hasVirtualFs).toBe(true);
    });
    
    it('categorizes non-fs-mocking files correctly', () => {
      const content = `
        import { something } from '../../utils/helpers';
        
        test('something', () => {
          expect(something).toBe(true);
        });
      `;
      
      const result = analyzeFile('test.ts', content);
      
      expect(result.category).toBe('None');
      expect(result.hasDirectFsMock).toBe(false);
      expect(result.hasLegacyUtil).toBe(false);
      expect(result.hasVirtualFs).toBe(false);
    });
  });
  
  describe('generateMarkdownReport function', () => {
    it('generates correct markdown with results', () => {
      const results = [
        {
          filePath: 'src/test1.test.ts',
          category: 'Direct Mock',
          hasDirectFsMock: true,
          hasLegacyUtil: false,
          hasVirtualFs: false
        },
        {
          filePath: 'src/test2.test.ts',
          category: 'Legacy Util',
          hasDirectFsMock: false,
          hasLegacyUtil: true,
          hasVirtualFs: false
        }
      ];
      
      const markdown = generateMarkdownReport(results);
      
      expect(markdown).toContain('# Filesystem Mocking Audit Results');
      expect(markdown).toContain('| src/test1.test.ts | Direct Mock |');
      expect(markdown).toContain('| src/test2.test.ts | Legacy Util |');
      expect(markdown).toContain('## Categories:');
      expect(markdown).toContain('## Complexity Guidelines');
      expect(markdown).toContain('## Priority Guidelines');
    });
    
    it('generates correct markdown with empty results', () => {
      const results = [];
      
      const markdown = generateMarkdownReport(results);
      
      expect(markdown).toContain('# Filesystem Mocking Audit Results');
      expect(markdown).toContain('## Files Needing Migration');
      expect(markdown).toContain('| File Path | Category | Complexity | Notes | Priority |');
      expect(markdown).not.toContain('| src/');
      expect(markdown).toContain('## Categories:');
    });
  });
  
  describe('auditFsMocks function', () => {
    it('processes files and returns results and report', async () => {
      // Set up mocks
      mockGlob.mockResolvedValueOnce(['src/test1.test.ts', 'src/test2.test.ts']);
      
      // Define a getContent function to avoid filesystem access
      const getContent = (filePath) => {
        if (filePath === 'src/test1.test.ts') {
          return `
            jest.mock('fs', () => ({ readFileSync: jest.fn() }));
            test('something', () => {});
          `;
        } else if (filePath === 'src/test2.test.ts') {
          return `
            import { setupBasicFs } from '../../../test/setup/fs';
            test('something', () => {});
          `;
        }
        return '';
      };
      
      const { results, markdown } = await auditFsMocks({
        // Don't write to file during test
        outputPath: null
      }, mockLogger, {
        glob: mockGlob,
        getContent
      });
      
      // There should be one file with deprecated fs mocking
      expect(results.length).toBe(1);
      expect(results[0].filePath).toBe('src/test1.test.ts');
      expect(results[0].category).toBe('Direct Mock');
      
      // The report should be generated
      expect(markdown).toContain('# Filesystem Mocking Audit Results');
      expect(markdown).toContain('| src/test1.test.ts | Direct Mock |');
      
      // The logger should be called
      expect(mockLogger).toHaveBeenCalledWith('Scanning for test files...');
      expect(mockLogger).toHaveBeenCalledWith('Found 2 test files. Analyzing...');
    });
    
    it('writes report to file when outputPath is provided', async () => {
      // Set up mocks
      mockGlob.mockResolvedValueOnce(['src/test1.test.ts']);
      
      // Define getContent function
      const getContent = (filePath) => {
        if (filePath === 'src/test1.test.ts') {
          return `
            jest.mock('fs', () => ({ readFileSync: jest.fn() }));
            test('something', () => {});
          `;
        }
        return '';
      };
      
      await auditFsMocks({
        outputPath: 'output.md'
      }, mockLogger, {
        glob: mockGlob,
        getContent,
        writeFile: mockWriteFile
      });
      
      // The writeFile method should be called
      expect(mockWriteFile).toHaveBeenCalledWith('output.md', expect.any(String));
      
      // The logger should log the write operation
      expect(mockLogger).toHaveBeenCalledWith('Results written to output.md');
    });
    
    it('handles errors gracefully', async () => {
      // Mock glob to throw an error
      const mockErrorGlob = jest.fn().mockRejectedValue(new Error('Test error'));
      
      // Mock console.error to capture the error
      const originalConsoleError = console.error;
      console.error = jest.fn();
      
      // Execute the function
      await expect(auditFsMocks({}, mockLogger, {
        glob: mockErrorGlob
      })).rejects.toThrow('Test error');
      
      // The error should be logged
      expect(console.error).toHaveBeenCalledWith('Error during audit:', expect.any(Error));
      
      // Restore console.error
      console.error = originalConsoleError;
    });
  });
});
