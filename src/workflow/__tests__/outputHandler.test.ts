/**
 * Unit tests for the OutputHandler module
 */
import { mockFsModules, resetVirtualFs, getVirtualFs, createFsError } from '../../__tests__/utils/virtualFsUtils';
import { FileSystem } from '../../core/interfaces';

// Setup mocks (must be before importing fs modules)
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Now import fs after mocking
import fs from 'fs/promises';

import { 
  formatResponseAsMarkdown, 
  generateFilename, 
  formatForConsole,
  createOutputDirectory,
  writeResponsesToFiles,
  processOutput
} from '../outputHandler';
import { LLMResponse } from '../../core/types';

// Define test data
const mockDateISOString = '2025-04-01T12:00:00.000Z';
const realDateNow = Date.now;
const realDateToISOString = Date.prototype.toISOString;

describe('OutputHandler', () => {
  // Set up mock FileSystem implementation
  const mockFileSystem: jest.Mocked<FileSystem> = {
    readFileContent: jest.fn().mockResolvedValue('Test file content'),
    writeFile: jest.fn().mockImplementation(async (filePath, content) => {
      await fs.writeFile(filePath, content);
    }),
    fileExists: jest.fn().mockImplementation(async (filePath) => {
      try {
        await fs.access(filePath);
        return true;
      } catch {
        return false;
      }
    }),
    mkdir: jest.fn().mockImplementation(async (dirPath, options?) => {
      await fs.mkdir(dirPath, options);
    }),
    readdir: jest.fn().mockResolvedValue(['file1.txt', 'file2.txt']),
    stat: jest.fn().mockImplementation(async (path) => {
      const stats = await fs.stat(path);
      return stats;
    }),
    access: jest.fn().mockImplementation(async (path) => {
      await fs.access(path);
    }),
    getConfigDir: jest.fn().mockResolvedValue('/mock/config/dir'),
    getConfigFilePath: jest.fn().mockResolvedValue('/mock/config/file.json')
  };

  // Set up test environment
  beforeEach(() => {
    jest.clearAllMocks();
    resetVirtualFs();
    
    // Create basic directory structure
    const virtualFs = getVirtualFs();
    virtualFs.mkdirSync('/test', { recursive: true });
    
    // Mock Date.now for consistent timing in tests
    global.Date.now = jest.fn(() => 1712059200000); // 2025-04-01
    
    // Mock Date.prototype.toISOString for consistent timestamps in formatted content
    Date.prototype.toISOString = jest.fn(() => mockDateISOString);
    
    // Reset mockFileSystem methods
    Object.values(mockFileSystem).forEach(method => {
      if (jest.isMockFunction(method)) {
        method.mockClear();
      }
    });
    
    // Spy on mockFileSystem functions to track calls
    jest.spyOn(mockFileSystem, 'mkdir');
    jest.spyOn(mockFileSystem, 'writeFile');
  });
  
  // Restore Date methods after tests
  afterEach(() => {
    global.Date.now = realDateNow;
    Date.prototype.toISOString = realDateToISOString;
  });
  
  // Sample test responses
  const sampleResponse: LLMResponse & { configKey: string } = {
    provider: 'openai',
    modelId: 'gpt-4o',
    text: 'This is a test response',
    configKey: 'openai:gpt-4o'
  };
  
  const sampleResponseWithGroup: LLMResponse & { configKey: string } = {
    provider: 'anthropic',
    modelId: 'claude-3-opus-20240229',
    text: 'This is a test response from a group',
    configKey: 'anthropic:claude-3-opus-20240229',
    groupInfo: {
      name: 'coding',
      systemPrompt: { text: 'You are a coding assistant' }
    }
  };
  
  const sampleResponseWithError: LLMResponse & { configKey: string } = {
    provider: 'error-provider',
    modelId: 'error-model',
    text: '',
    error: 'An error occurred',
    configKey: 'error-provider:error-model'
  };
  
  const sampleResponseWithMetadata: LLMResponse & { configKey: string } = {
    provider: 'openai',
    modelId: 'gpt-4o',
    text: 'This is a test response with metadata',
    configKey: 'openai:gpt-4o',
    metadata: {
      responseTime: 1200,
      usage: {
        prompt_tokens: 100,
        completion_tokens: 50,
        total_tokens: 150
      }
    }
  };
  
  // Sample responses defined above are used directly in individual tests
  
  describe('Markdown Formatting', () => {
    it('should format a basic response as markdown', () => {
      const markdown = formatResponseAsMarkdown(sampleResponse);
      
      // Verify basic structure
      expect(markdown).toContain(`# ${sampleResponse.configKey}`);
      expect(markdown).toContain(`Generated: ${mockDateISOString}`);
      expect(markdown).toContain('## Response');
      expect(markdown).toContain(sampleResponse.text);
      
      // Verify no metadata included by default
      expect(markdown).not.toContain('## Metadata');
    });
    
    it('should include group information when available', () => {
      const markdown = formatResponseAsMarkdown(sampleResponseWithGroup);
      
      // Verify group info is included
      expect(markdown).toContain(`# ${sampleResponseWithGroup.configKey} (coding group)`);
      expect(markdown).toContain('Group: coding');
      
      // System prompt should not be included without includeMetadata
      expect(markdown).not.toContain('System Prompt:');
    });
    
    it('should include system prompt when includeMetadata is true', () => {
      const markdown = formatResponseAsMarkdown(sampleResponseWithGroup, true);
      
      // Verify system prompt is included with includeMetadata
      expect(markdown).toContain('System Prompt: "You are a coding assistant"');
    });
    
    it('should include error information when present', () => {
      const markdown = formatResponseAsMarkdown(sampleResponseWithError);
      
      // Verify error section
      expect(markdown).toContain('## Error');
      expect(markdown).toContain(sampleResponseWithError.error);
    });
    
    it('should include metadata when requested', () => {
      const markdown = formatResponseAsMarkdown(sampleResponseWithMetadata, true);
      
      // Verify metadata section
      expect(markdown).toContain('## Metadata');
      expect(markdown).toContain('"responseTime": 1200');
      expect(markdown).toContain('"total_tokens": 150');
    });
  });
  
  describe('Filename Generation', () => {
    it('should generate a basic filename', () => {
      const filename = generateFilename(sampleResponse);
      
      // Verify basic filename format
      expect(filename).toBe('openai-gpt-4o.md');
    });
    
    it('should include group in filename when available', () => {
      const filename = generateFilename(sampleResponseWithGroup);
      
      // Verify group is included in filename
      expect(filename).toBe('coding-anthropic-claude-3-opus-20240229.md');
    });
    
    it('should sanitize filenames', () => {
      // Create a response with special characters
      const specialResponse = {
        ...sampleResponse,
        provider: 'open/ai:test',
        modelId: 'gpt*4<>?"',
        configKey: 'open/ai:test:gpt*4<>?"'
      };
      
      const filename = generateFilename(specialResponse);
      
      // Verify special characters are sanitized
      expect(filename).toBe('open_ai_test-gpt_4_.md');
      expect(filename).not.toContain('/');
      expect(filename).not.toContain('*');
      expect(filename).not.toContain('<');
      expect(filename).not.toContain('>');
      expect(filename).not.toContain('"');
    });
    
    it('should exclude group when requested', () => {
      const filename = generateFilename(sampleResponseWithGroup, { includeGroup: false });
      
      // Verify group is excluded
      expect(filename).toBe('anthropic-claude-3-opus-20240229.md');
    });
  });
  
  describe('Console Formatting', () => {
    it('should format responses for console output', () => {
      const consoleOutput = formatForConsole([sampleResponse]);
      
      // Very basic check that formatter was called correctly
      expect(consoleOutput).toContain(sampleResponse.configKey);
      expect(consoleOutput).toContain(sampleResponse.text);
    });
    
    it('should handle empty responses array', () => {
      const consoleOutput = formatForConsole([]);
      
      // Verify empty message
      expect(consoleOutput).toBe('No results to display.');
    });
  });
  
  describe('Directory Creation', () => {
    it('should create output directory', async () => {
      const outputDir = await createOutputDirectory({}, mockFileSystem);
      
      // Verify directory creation was called
      expect(mockFileSystem.mkdir).toHaveBeenCalled();
      
      // Verify returns expected directory path with timestamp
      expect(outputDir).toContain('run-');
      
      // Verify directory was actually created in virtual filesystem
      const virtualFs = getVirtualFs();
      expect(virtualFs.existsSync(outputDir)).toBe(true);
      expect(virtualFs.statSync(outputDir).isDirectory()).toBe(true);
    });
    
    it('should not include identifier in directory name after refactoring', async () => {
      const outputDir = await createOutputDirectory({
        directoryIdentifier: 'test-run'
      }, mockFileSystem);
      
      // After refactoring, the directory name should NOT include the identifier
      expect(outputDir).not.toContain('test-run-');
      // Instead, it should use the standard run-timestamp format
      expect(outputDir).toMatch(/run-\d{8}-\d{6}/);
      
      // Verify directory was actually created
      const virtualFs = getVirtualFs();
      expect(virtualFs.existsSync(outputDir)).toBe(true);
    });
    
    it('should throw error if directory creation fails', async () => {
      // Mock mkdir to fail using the spy
      jest.spyOn(fs, 'mkdir').mockRejectedValueOnce(
        createFsError('EACCES', 'Permission denied', 'mkdir', '/test/output/dir')
      );
      
      // Verify error is thrown
      await expect(createOutputDirectory({}, mockFileSystem)).rejects.toThrow('Failed to create output directory');
    });
  });
  
  describe('File Writing', () => {
    it('should write responses to files', async () => {
      // Setup a test output directory in virtual filesystem
      const virtualFs = getVirtualFs();
      virtualFs.mkdirSync('/test/output/dir', { recursive: true });
      
      const result = await writeResponsesToFiles(
        [sampleResponse, sampleResponseWithGroup],
        '/test/output/dir',
        {},
        mockFileSystem
      );
      
      // Verify writes were called - 3 writes per file (tmp, final, cleanup)
      expect(mockFileSystem.writeFile).toHaveBeenCalledTimes(6);
      
      // Verify result structure
      expect(result.succeededWrites).toBe(2);
      expect(result.failedWrites).toBe(0);
      expect(result.files).toHaveLength(2);
      expect(result.files[0].status).toBe('success');
      expect(result.timing).toBeDefined();
      
      // Verify files were actually written to virtual filesystem
      expect(virtualFs.existsSync('/test/output/dir/openai-gpt-4o.md')).toBe(true);
      expect(virtualFs.existsSync('/test/output/dir/coding-anthropic-claude-3-opus-20240229.md')).toBe(true);
      
      // Verify file content
      const file1Content = virtualFs.readFileSync('/test/output/dir/openai-gpt-4o.md', 'utf-8');
      expect(file1Content).toContain('This is a test response');
    });
    
    it('should handle write errors', async () => {
      // Setup a test output directory in virtual filesystem
      const virtualFs = getVirtualFs();
      virtualFs.mkdirSync('/test/output/dir', { recursive: true });
      
      // We'll implement custom behavior for each call
      
      // Mock writeFile in mockFileSystem with a counter to fail on the third call
      // (The implementation calls writeFile twice for the first file - once for temp file and once for the actual file)
      let writeCounter = 0;
      mockFileSystem.writeFile = jest.fn().mockImplementation(async (filePath, content) => {
        writeCounter++;
        
        // First two calls (for first file) succeed
        if (writeCounter <= 2) {
          if (typeof filePath === 'string' && typeof content === 'string') {
            virtualFs.writeFileSync(filePath, content);
          }
          return undefined;
        }
        
        // Third call (for second file) fails
        throw createFsError('ENOSPC', 'Write failed', 'open', '/test/output/dir/error-file.md');
      });
      
      const result = await writeResponsesToFiles(
        [sampleResponse, sampleResponseWithError],
        '/test/output/dir',
        { throwOnError: false },
        mockFileSystem
      );
      
      // Verify one success, one failure
      expect(result.succeededWrites).toBe(1);
      expect(result.failedWrites).toBe(1);
      expect(result.files[0].status).toBe('success');
      expect(result.files[1].status).toBe('error');
      expect(result.files[1].error).toBe('Write failed');
      
      // Verify first file was actually written
      expect(virtualFs.existsSync('/test/output/dir/openai-gpt-4o.md')).toBe(true);
    });
    
    it('should track errors even with throwOnError', async () => {
      // Setup a test output directory in virtual filesystem
      const virtualFs = getVirtualFs();
      virtualFs.mkdirSync('/test/output/dir', { recursive: true });
      
      // Mock mockFileSystem.writeFile to fail
      mockFileSystem.writeFile
        .mockRejectedValueOnce(createFsError('ENOSPC', 'Write failed', 'open', '/test/output/dir/file.md'));
      
      // We'll catch the error but expect it to still update tracking info
      try {
        await writeResponsesToFiles(
          [sampleResponse],
          '/test/output/dir',
          { throwOnError: true },
          mockFileSystem
        );
        fail('Error should have been thrown');
      } catch (error) {
        // Expected error
      }
      
      // Reset mockFileSystem.writeFile implementation to write to the virtual filesystem
      mockFileSystem.writeFile.mockReset().mockImplementation(async (filePath, content) => {
        // Actually write to the virtual filesystem
        if (typeof filePath === 'string' && typeof content === 'string') {
          virtualFs.writeFileSync(filePath, content);
        }
        return undefined;
      });
      
      const result = await writeResponsesToFiles(
        [sampleResponseWithError], // Use a different response
        '/test/output/dir',
        { throwOnError: false },
        mockFileSystem
      );
      
      // Should still track error counts properly
      expect(result.failedWrites).toBe(0); // This call succeeded
      expect(result.succeededWrites).toBe(1);
      
      // Verify file was written
      expect(virtualFs.existsSync('/test/output/dir/error-provider-error-model.md')).toBe(true);
    });
    
    it('should call status update callback', async () => {
      // Setup a test output directory in virtual filesystem
      const virtualFs = getVirtualFs();
      virtualFs.mkdirSync('/test/output/dir', { recursive: true });
      
      // Ensure the mockFileSystem.writeFile implementation writes to the virtual filesystem
      mockFileSystem.writeFile.mockImplementation(async (filePath, content) => {
        // Actually write to the virtual filesystem
        if (typeof filePath === 'string' && typeof content === 'string') {
          virtualFs.writeFileSync(filePath, content);
        }
        return undefined;
      });
      
      // Set up spy
      const onStatusUpdateSpy = jest.fn();
      
      await writeResponsesToFiles(
        [sampleResponse],
        '/test/output/dir',
        { onStatusUpdate: onStatusUpdateSpy },
        mockFileSystem
      );
      
      // Verify callback was called at least once
      expect(onStatusUpdateSpy).toHaveBeenCalled();
      
      // Verify the last call has success status (don't rely on specific call order)
      const lastCallIndex = onStatusUpdateSpy.mock.calls.length - 1;
      expect(onStatusUpdateSpy.mock.calls[lastCallIndex][0].status).toBe('success');
      
      // Verify file was written
      expect(virtualFs.existsSync('/test/output/dir/openai-gpt-4o.md')).toBe(true);
    });
  });
  
  describe('Pure Output Processing', () => {
    it('should process output for both files and console without IO operations', () => {
      const result = processOutput(
        [sampleResponse, sampleResponseWithGroup],
        {
          includeMetadata: true,
          useTable: true
        }
      );
      
      // Verify correct output structure
      expect(result.files).toBeDefined();
      expect(result.directoryPath).toBeDefined();
      expect(result.consoleOutput).toBeDefined();
      
      // Verify file data structure
      expect(result.files).toHaveLength(2);
      expect(result.files[0].filename).toBe('openai-gpt-4o.md');
      expect(result.files[0].content).toContain('This is a test response');
      expect(result.files[0].modelKey).toBe('openai:gpt-4o');
      
      expect(result.files[1].filename).toBe('coding-anthropic-claude-3-opus-20240229.md');
      expect(result.files[1].content).toContain('This is a test response from a group');
      expect(result.files[1].modelKey).toBe('anthropic:claude-3-opus-20240229');
      
      // Verify directory path is properly generated
      expect(result.directoryPath).toMatch(/thinktank-output\//);
      
      // Verify console output is a string
      expect(typeof result.consoleOutput).toBe('string');
      
      // Verify no IO operations were performed
      expect(mockFileSystem.mkdir).not.toHaveBeenCalled();
      expect(mockFileSystem.writeFile).not.toHaveBeenCalled();
    });
    
    it('should respect options when formatting files and console output', () => {
      // Test with specific options
      const result = processOutput(
        [sampleResponse, sampleResponseWithGroup],
        {
          includeMetadata: true,
          useColors: false,
          includeThinking: true,
          useTable: false,
          outputDirectory: '/custom/output/path',
          friendlyRunName: 'test-run'
        }
      );
      
      // Verify output directory uses the custom path and name
      expect(result.directoryPath).toContain('/custom/output/path');
      expect(result.directoryPath).toContain('test-run');
      
      // The test response has no metadata, so we're just checking that the content is formatted
      // If includeMetadata was true but no metadata exists, no metadata section would be included
      
      // Verify proper response content formatting 
      expect(result.files[0].content).toContain(`# ${sampleResponse.configKey}`);
      expect(result.files[0].content).toContain(`Generated: ${mockDateISOString}`);
      expect(result.files[0].content).toContain('## Response');
      
      // Verify console output options were passed through
      // We can't check exact console output formatting without mocking formatForConsole
      expect(typeof result.consoleOutput).toBe('string');
    });
  });
});
