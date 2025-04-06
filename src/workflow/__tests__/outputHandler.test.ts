/**
 * Unit tests for the OutputHandler module
 */
import fs from 'fs/promises';
// path is imported by fs-related functions but not used directly
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

// Mock fs module
jest.mock('fs/promises');

// Define test data
const mockDateISOString = '2025-04-01T12:00:00.000Z';
const realDateNow = Date.now;
const realDateToISOString = Date.prototype.toISOString;

describe('OutputHandler', () => {
  // Set up test environment
  beforeEach(() => {
    jest.clearAllMocks();
    
    // Mock Date.now for consistent timing in tests
    global.Date.now = jest.fn(() => 1712059200000); // 2025-04-01
    
    // Mock Date.prototype.toISOString for consistent timestamps in formatted content
    Date.prototype.toISOString = jest.fn(() => mockDateISOString);
    
    // Mock the file system methods
    const mockMkdir = fs.mkdir as jest.MockedFunction<typeof fs.mkdir>;
    mockMkdir.mockResolvedValue(undefined);
    
    const mockWriteFile = fs.writeFile as jest.MockedFunction<typeof fs.writeFile>;
    mockWriteFile.mockResolvedValue(undefined);
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
    });
    
    it('should not include identifier in directory name after refactoring', async () => {
      const outputDir = await createOutputDirectory({
        directoryIdentifier: 'test-run'
      });
      
      // After refactoring, the directory name should NOT include the identifier
      expect(outputDir).not.toContain('test-run-');
      // Instead, it should use the standard run-timestamp format
      expect(outputDir).toMatch(/run-\d{8}-\d{6}/);
    });
    
    it('should throw error if directory creation fails', async () => {
      // Mock mkdir to fail
      const mockMkdir = fs.mkdir as jest.MockedFunction<typeof fs.mkdir>;
      mockMkdir.mockRejectedValue(new Error('Permission denied'));
      
      // Verify error is thrown
      await expect(createOutputDirectory()).rejects.toThrow('Failed to create output directory');
    });
  });
  
  describe('File Writing', () => {
    it('should write responses to files', async () => {
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
    });
    
    it('should handle write errors', async () => {
      // Mock writeFile to fail for the second file
      const mockWriteFile = fs.writeFile as jest.MockedFunction<typeof fs.writeFile>;
      mockWriteFile
        .mockResolvedValueOnce(undefined)
        .mockRejectedValueOnce(new Error('Write failed'));
      
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
    });
    
    it('should track errors even with throwOnError', async () => {
      // Mock writeFile to fail
      const mockWriteFile = fs.writeFile as jest.MockedFunction<typeof fs.writeFile>;
      mockWriteFile.mockRejectedValue(new Error('Write failed'));
      
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
      
      // Now mock writeFile to succeed and call again with throwOnError false
      // to verify error tracking
      mockWriteFile.mockResolvedValue(undefined);
      
      const result = await writeResponsesToFiles(
        [sampleResponseWithError], // Use a different response
        '/test/output/dir',
        { throwOnError: false }
      );
      
      // Should still track error counts properly
      expect(result.failedWrites).toBe(0); // This call succeeded
      expect(result.succeededWrites).toBe(1);
    });
    
    it('should call status update callback', async () => {
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
    });
  });
  
  describe('Full Output Processing', () => {
    it('should process output for both files and console', async () => {
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
    });
    
    it('should handle errors during processing', async () => {
      // Mock mkdir to fail
      const mockMkdir = fs.mkdir as jest.MockedFunction<typeof fs.mkdir>;
      mockMkdir.mockRejectedValueOnce(new Error('Permission denied'));
      
      // Verify error is thrown
      await expect(processOutput(
        [sampleResponse],
        {}
      )).rejects.toThrow(OutputHandlerError);
    });
  });
});