/**
 * Tests for the formatCombinedInput function
 * 
 * The formatCombinedInput function combines prompt content with context files
 * in a way that clearly separates the context from the prompt for LLM input.
 * These tests ensure that:
 * 
 * 1. Context files are properly formatted with appropriate syntax highlighting
 * 2. Files with errors are excluded from the context
 * 3. File paths are normalized and displayed correctly
 * 4. Section boundaries are clearly marked in markdown
 * 5. The function handles edge cases like empty arrays and all-error files
 * 6. Various file types are properly formatted with appropriate language tags
 */
import { createVirtualFs, resetVirtualFs, mockFsModules } from '../../__tests__/utils/virtualFsUtils';

// Setup mocks for fs modules
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Import modules after mocking
import path from 'path';
import { ContextFileResult } from '../fileReaderTypes';
import { formatCombinedInput, readContextFile } from '../fileReader';

describe('formatCombinedInput', () => {
  // Set up common test data
  const promptContent = 'This is the main prompt content.\nIt has multiple lines.';
  
  const mockContextFiles: ContextFileResult[] = [
    {
      path: '/path/to/file1.js',
      content: '// This is a JavaScript file\nconst hello = "world";\nconsole.log(hello);',
      error: null
    },
    {
      path: '/path/to/file2.md',
      content: '# Markdown File\n\nThis is a markdown file with **formatting**.',
      error: null
    },
    {
      path: '/path/to/error-file.txt',
      content: null,
      error: {
        code: 'ENOENT',
        message: 'File not found: /path/to/error-file.txt'
      }
    }
  ];
  
  beforeEach(() => {
    // Reset virtual filesystem and mocks before each test
    resetVirtualFs();
    jest.clearAllMocks();
  });
  
  describe('Basic Formatting', () => {
    it('should format context files and prompt with clear boundaries', () => {
      const result = formatCombinedInput(promptContent, mockContextFiles);
      
      // Should include the main headers
      expect(result).toContain('# CONTEXT DOCUMENTS');
      expect(result).toContain('# USER PROMPT');
      
      // Should include file paths as headers
      expect(result).toContain(`## File: ${path.normalize('/path/to/file1.js')}`);
      expect(result).toContain(`## File: ${path.normalize('/path/to/file2.md')}`);
      
      // Should include content of files
      expect(result).toContain('// This is a JavaScript file');
      expect(result).toContain('# Markdown File');
      
      // Should include the prompt content
      expect(result).toContain('This is the main prompt content.');
      
      // Verify code block formatting
      expect(result).toContain('```javascript');
      expect(result).toContain('```markdown');
      
      // Error files should be excluded
      expect(result).not.toContain('error-file.txt');
    });
    
    it('should maintain the exact order of files as provided in the input array', () => {
      const result = formatCombinedInput(promptContent, mockContextFiles);
      
      // Check that first file appears before second file in output
      const file1Index = result.indexOf('/path/to/file1.js');
      const file2Index = result.indexOf('/path/to/file2.md');
      
      expect(file1Index).toBeLessThan(file2Index);
    });
    
    it('should create valid markdown structure with proper nesting', () => {
      const result = formatCombinedInput(promptContent, mockContextFiles);
      
      // Check markdown heading hierarchy
      const lines = result.split('\n');
      
      // Find heading levels and ensure proper nesting
      const headings = lines
        .filter(line => line.startsWith('#'))
        .map(line => {
          const level = line.indexOf(' ');
          const text = line.substring(level + 1);
          return { level, text };
        });
      
      // Main sections should be level 1 headings (# )
      expect(headings[0].level).toBe(1);  // # CONTEXT DOCUMENTS
      
      // File headings should be level 2 (## )
      expect(headings[1].level).toBe(2);  // ## File: path/to/file1.js
      expect(headings[2].level).toBe(2);  // ## File: path/to/file2.md
      
      // USER PROMPT should be level 1 heading
      expect(headings[3].level).toBe(1);  // # USER PROMPT
    });
  });
  
  describe('Edge Cases and Special Scenarios', () => {
    it('should handle empty context files array', () => {
      const result = formatCombinedInput(promptContent, []);
      
      // Should only include the prompt section when no context files
      expect(result).not.toContain('# CONTEXT DOCUMENTS');
      expect(result).toContain('# USER PROMPT');
      expect(result).toContain(promptContent);
    });
    
    it('should handle all error context files', () => {
      const allErrorFiles: ContextFileResult[] = [
        {
          path: '/path/to/error1.txt',
          content: null,
          error: {
            code: 'ENOENT',
            message: 'File not found: /path/to/error1.txt'
          }
        },
        {
          path: '/path/to/error2.txt',
          content: null,
          error: {
            code: 'EACCES',
            message: 'Permission denied: /path/to/error2.txt'
          }
        }
      ];
      
      const result = formatCombinedInput(promptContent, allErrorFiles);
      
      // Should only include the prompt section when all context files have errors
      expect(result).not.toContain('# CONTEXT DOCUMENTS');
      expect(result).toContain('# USER PROMPT');
      expect(result).toContain(promptContent);
    });
    
    it('should handle mixed valid and null content files correctly', () => {
      const mixedFiles: ContextFileResult[] = [
        {
          path: '/path/to/valid.js',
          content: 'console.log("valid")',
          error: null
        },
        {
          path: '/path/to/null-content.txt',
          content: null,
          error: null  // Edge case: null content but no error
        }
      ];
      
      const result = formatCombinedInput(promptContent, mixedFiles);
      
      // Should include only the valid file
      expect(result).toContain('/path/to/valid.js');
      expect(result).not.toContain('/path/to/null-content.txt');
    });
    
    it('should properly handle empty but non-null content files', () => {
      const emptyFiles: ContextFileResult[] = [
        {
          path: '/path/to/empty.txt',
          content: '',  // Empty string, not null
          error: null
        }
      ];
      
      const result = formatCombinedInput(promptContent, emptyFiles);
      
      // Empty file should be included, but with empty content
      expect(result).toContain('# CONTEXT DOCUMENTS');
      expect(result).toContain('/path/to/empty.txt');
      // There should be a code block with nothing between the opening and closing ticks
      expect(result).toContain('```text\n\n```');
    });
    
    it('should handle empty prompt content', () => {
      const emptyPrompt = '';
      const result = formatCombinedInput(emptyPrompt, mockContextFiles);
      
      // Should still format context files correctly
      expect(result).toContain('# CONTEXT DOCUMENTS');
      expect(result).toContain('# USER PROMPT');
      
      // USER PROMPT section should be followed by empty content
      const userPromptIndex = result.indexOf('# USER PROMPT');
      const contentAfterPrompt = result.substring(userPromptIndex + '# USER PROMPT'.length);
      
      // Content after user prompt header should just be the newlines, no actual content
      expect(contentAfterPrompt.trim()).toBe('');
    });
  });
  
  describe('File Type Handling', () => {
    it('should detect and format various file types with correct language markers', () => {
      const multipleFileTypes: ContextFileResult[] = [
        {
          path: '/path/to/script.py',
          content: 'def hello():\n    print("Hello, World!")',
          error: null
        },
        {
          path: '/path/to/styles.css',
          content: 'body { font-family: sans-serif; }',
          error: null
        },
        {
          path: '/path/to/config.json',
          content: '{ "name": "thinktank", "version": "1.0.0" }',
          error: null
        },
        {
          path: '/path/to/script.sh',
          content: '#!/bin/bash\necho "Hello"',
          error: null
        },
        {
          path: '/path/to/config.yml',
          content: 'name: thinktank\nversion: 1.0.0',
          error: null
        },
        {
          path: '/path/to/unknown.xyz',
          content: 'This is a file with an unknown extension',
          error: null
        }
      ];
      
      const result = formatCombinedInput(promptContent, multipleFileTypes);
      
      // Check for correct language markers in code blocks
      expect(result).toContain('```python');
      expect(result).toContain('```css');
      expect(result).toContain('```json');
      expect(result).toContain('```bash');
      expect(result).toContain('```yaml');
      
      // Unknown extension should default to text
      expect(result).toContain('```text');
    });
    
    it('should handle file paths with special characters correctly', () => {
      const specialPathFiles: ContextFileResult[] = [
        {
          path: '/path/with spaces/file.js',
          content: 'console.log("spaces in path")',
          error: null
        },
        {
          path: '/path/with-hyphens/file.js',
          content: 'console.log("hyphens in path")',
          error: null
        },
        {
          path: '/path/with_underscores/file.js',
          content: 'console.log("underscores in path")',
          error: null
        }
      ];
      
      const result = formatCombinedInput(promptContent, specialPathFiles);
      
      // All files should be included with their paths properly escaped/normalized
      expect(result).toContain(path.normalize('/path/with spaces/file.js'));
      expect(result).toContain(path.normalize('/path/with-hyphens/file.js'));
      expect(result).toContain(path.normalize('/path/with_underscores/file.js'));
    });
  });
  
  describe('Output Structure', () => {
    it('should produce output with the expected overall structure', () => {
      // Create test files with simple content
      const simpleFiles: ContextFileResult[] = [
        {
          path: '/path/to/file1.js',
          content: 'const x = 1;',
          error: null
        },
        {
          path: '/path/to/file2.txt',
          content: 'Plain text content',
          error: null
        }
      ];
      
      const result = formatCombinedInput('Simple prompt', simpleFiles);
      
      // Check the overall structure of the output
      const expectedStructure = [
        '# CONTEXT DOCUMENTS',
        '',
        `## File: ${path.normalize('/path/to/file1.js')}`,
        '```javascript',
        'const x = 1;',
        '```',
        '',
        `## File: ${path.normalize('/path/to/file2.txt')}`,
        '```text',
        'Plain text content',
        '```',
        '',
        '# USER PROMPT',
        '',
        'Simple prompt'
      ].join('\n');
      
      // Since path.normalize might use different separators on different OSes,
      // we normalize both the expected and actual results for comparison
      const normalizedResult = result.replace(/\r\n/g, '\n');
      const normalizedExpected = expectedStructure.replace(/\r\n/g, '\n');
      
      expect(normalizedResult).toBe(normalizedExpected);
    });
  });
  
  describe('Integration with Virtual Filesystem', () => {
    it('should work correctly with files from virtual filesystem', async () => {
      // Create test files in virtual filesystem
      createVirtualFs({
        '/virtual/file1.js': '// Virtual JavaScript file\nconst test = true;',
        '/virtual/file2.md': '# Virtual Markdown\n\nThis is content from a virtual file.'
      });
      
      // Read files using readContextFile
      const file1Result = await readContextFile('/virtual/file1.js');
      const file2Result = await readContextFile('/virtual/file2.md');
      
      // Format the results
      const result = formatCombinedInput('Test prompt', [file1Result, file2Result]);
      
      // Check for expected content
      expect(result).toContain('# CONTEXT DOCUMENTS');
      expect(result).toContain('# USER PROMPT');
      expect(result).toContain('// Virtual JavaScript file');
      expect(result).toContain('# Virtual Markdown');
      expect(result).toContain('```javascript');
      expect(result).toContain('```markdown');
      expect(result).toContain('Test prompt');
    });
    
    it('should handle errors from virtual filesystem correctly', async () => {
      // Create only one file in virtual filesystem
      createVirtualFs({
        '/virtual/real-file.txt': 'This file exists'
      });
      
      // Read both existing and non-existent files
      const realFileResult = await readContextFile('/virtual/real-file.txt');
      const missingFileResult = await readContextFile('/virtual/missing-file.txt');
      
      // Format the results
      const result = formatCombinedInput('Test prompt', [realFileResult, missingFileResult]);
      
      // Check that only the real file is included
      expect(result).toContain('/virtual/real-file.txt');
      expect(result).toContain('This file exists');
      expect(result).not.toContain('/virtual/missing-file.txt');
    });
  });
});
