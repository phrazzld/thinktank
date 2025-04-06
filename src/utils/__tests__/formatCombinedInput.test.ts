/**
 * Tests for the formatCombinedInput function
 */
import path from 'path';
import { formatCombinedInput } from '../fileReader';
import { ContextFileResult } from '../fileReader';

describe('formatCombinedInput', () => {
  // Set up test data
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
});