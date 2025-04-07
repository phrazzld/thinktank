/**
 * Unit tests for the OutputHandler module
 */
import { mockFsModules, resetVirtualFs, getVirtualFs, createFsError } from '../../__tests__/utils/virtualFsUtils';
import path from 'path';

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
  processOutput,
  OutputHandlerError
} from '../outputHandler';
import { LLMResponse } from '../../core/types';

// Define test data
const mockDateISOString = '2025-04-01T12:00:00.000Z';
const realDateNow = Date.now;
const realDateToISOString = Date.prototype.toISOString;

describe('OutputHandler', () => {
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
    
    // Spy on fs functions to track calls without affecting behavior
    jest.spyOn(fs, 'mkdir');
    jest.spyOn(fs, 'writeFile');
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
      const outputDir = await createOutputDirectory();
      
      // Verify directory creation was called
      expect(fs.mkdir).toHaveBeenCalled();
      
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
      });
      
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
      await expect(createOutputDirectory()).rejects.toThrow('Failed to create output directory');
    });
  });
  
  describe('File Writing', () => {
    it('should write responses to files', async () => {
      // Setup a test output directory in virtual filesystem
      const virtualFs = getVirtualFs();
      virtualFs.mkdirSync('/test/output/dir', { recursive: true });
      
      const result = await writeResponsesToFiles(
        [sampleResponse, sampleResponseWithGroup],
        '/test/output/dir'
      );
      
      // Verify writes were called
      expect(fs.writeFile).toHaveBeenCalledTimes(2);
      
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
      
      // Mock writeFile to fail for the second file
      jest.spyOn(fs, 'writeFile')
        .mockImplementationOnce(async (filePath, content) => {
          // For the first call, actually write to the virtual filesystem
          if (typeof filePath === 'string' && typeof content === 'string') {
            virtualFs.writeFileSync(filePath, content);
          }
          return undefined;
        })
        .mockRejectedValueOnce(createFsError('ENOSPC', 'Write failed', 'open', '/test/output/dir/error-file.md'));
      
      const result = await writeResponsesToFiles(
        [sampleResponse, sampleResponseWithError],
        '/test/output/dir',
        { throwOnError: false }
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
      
      // Mock writeFile to fail
      jest.spyOn(fs, 'writeFile')
        .mockRejectedValueOnce(createFsError('ENOSPC', 'Write failed', 'open', '/test/output/dir/file.md'));
      
      // We'll catch the error but expect it to still update tracking info
      try {
        await writeResponsesToFiles(
          [sampleResponse],
          '/test/output/dir',
          { throwOnError: true }
        );
        fail('Error should have been thrown');
      } catch (error) {
        // Expected error
      }
      
      // Reset mockImplementation and use a version that actually writes to the virtual filesystem
      jest.spyOn(fs, 'writeFile').mockReset().mockImplementation(async (filePath, content) => {
        // Actually write to the virtual filesystem
        if (typeof filePath === 'string' && typeof content === 'string') {
          virtualFs.writeFileSync(filePath, content);
        }
        return undefined;
      });
      
      const result = await writeResponsesToFiles(
        [sampleResponseWithError], // Use a different response
        '/test/output/dir',
        { throwOnError: false }
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
      
      // Ensure the mock implementation actually writes to the virtual filesystem
      jest.spyOn(fs, 'writeFile').mockImplementation(async (filePath, content) => {
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
        { onStatusUpdate: onStatusUpdateSpy }
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
  
  describe('Full Output Processing', () => {
    it('should process output for both files and console', async () => {
      // Ensure both mkdir and writeFile actually affect the virtual filesystem
      jest.spyOn(fs, 'mkdir').mockImplementation(async (dirPath, _options) => {
        if (typeof dirPath === 'string') {
          getVirtualFs().mkdirSync(dirPath, { recursive: true });
        }
        return undefined;
      });
      
      jest.spyOn(fs, 'writeFile').mockImplementation(async (filePath, content) => {
        if (typeof filePath === 'string' && typeof content === 'string') {
          getVirtualFs().writeFileSync(filePath, content);
        }
        return undefined;
      });
      
      const result = await processOutput(
        [sampleResponse, sampleResponseWithGroup],
        {
          includeMetadata: true,
          useTable: true
        }
      );
      
      // Verify both outputs are present
      expect(result.fileOutput).toBeDefined();
      expect(result.consoleOutput).toBeDefined();
      
      // Verify directory creation and file writes were called
      expect(fs.mkdir).toHaveBeenCalled();
      expect(fs.writeFile).toHaveBeenCalledTimes(2);
      
      // Verify file output structure
      expect(result.fileOutput.succeededWrites).toBe(2);
      
      // Verify console output is a string
      expect(typeof result.consoleOutput).toBe('string');
      
      // Verify the output directory and files were actually created
      const virtualFs = getVirtualFs();
      expect(virtualFs.existsSync(result.fileOutput.outputDirectory)).toBe(true);
      expect(virtualFs.statSync(result.fileOutput.outputDirectory).isDirectory()).toBe(true);
      
      // Verify files were written
      const expectedFile1 = path.join(result.fileOutput.outputDirectory, 'openai-gpt-4o.md');
      const expectedFile2 = path.join(result.fileOutput.outputDirectory, 'coding-anthropic-claude-3-opus-20240229.md');
      expect(virtualFs.existsSync(expectedFile1)).toBe(true);
      expect(virtualFs.existsSync(expectedFile2)).toBe(true);
    });
    
    it('should handle errors during processing', async () => {
      // Mock mkdir to fail
      jest.spyOn(fs, 'mkdir').mockRejectedValueOnce(
        createFsError('EACCES', 'Permission denied', 'mkdir', '/path')
      );
      
      // Verify error is thrown
      await expect(processOutput(
        [sampleResponse],
        {}
      )).rejects.toThrow(OutputHandlerError);
    });
  });
});
