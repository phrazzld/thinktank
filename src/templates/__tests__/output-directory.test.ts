/**
 * Integration tests for the output directory feature.
 * 
 * These tests verify that the output directory functionality works correctly
 * including directory creation, file writing, and error handling.
 */
import { runthinktank, RunOptions } from '../runthinktank';
import * as fileReader from '../../molecules/fileReader';
import * as configManager from '../../organisms/configManager';
import * as llmRegistry from '../../organisms/llmRegistry';
import * as helpers from '../../atoms/helpers';

// Mock dependencies
jest.mock('../../molecules/fileReader');
jest.mock('../../organisms/configManager');
jest.mock('../../organisms/llmRegistry');
jest.mock('fs/promises');
jest.mock('path');
jest.mock('ora', () => {
  return jest.fn().mockImplementation(() => {
    return {
      start: jest.fn().mockReturnThis(),
      stop: jest.fn().mockReturnThis(),
      succeed: jest.fn().mockReturnThis(),
      fail: jest.fn().mockReturnThis(),
      warn: jest.fn().mockReturnThis(),
      info: jest.fn().mockReturnThis(),
      text: '',
    };
  });
});

// Import mocked modules to access their functions
import fs from 'fs/promises';
import path from 'path';

// Mock console to prevent test output pollution
const originalConsoleLog = console.log;
const originalConsoleError = console.error;

describe('Output Directory Feature', () => {
  // Setup constants for testing
  const mockRunDirectoryName = 'thinktank_run_20230515_143000_000';
  const mockOutputDir = '/mock/output/path';
  const mockFullOutputDir = `${mockOutputDir}/${mockRunDirectoryName}`;
  
  // Access the mocked modules directly
  const mockFs = fs as jest.Mocked<typeof fs>;
  const mockPath = path as jest.Mocked<typeof path>;
  
  beforeEach(() => {
    jest.clearAllMocks();
    
    // Silence console during tests
    console.log = jest.fn();
    console.error = jest.fn();
    
    // Configure mocked fs functions
    mockFs.mkdir.mockResolvedValue(undefined);
    mockFs.writeFile.mockResolvedValue(undefined);
    
    // Configure path functions
    mockPath.join.mockImplementation((...args) => args.join('/'));
    mockPath.resolve.mockImplementation((...args) => args.join('/'));
    mockPath.basename.mockImplementation(filePath => {
      const parts = filePath.split('/');
      return parts[parts.length - 1];
    });
    
    // Mock helper functions
    jest.spyOn(helpers, 'generateRunDirectoryName').mockReturnValue(mockRunDirectoryName);
    jest.spyOn(helpers, 'resolveOutputDirectory').mockReturnValue(mockFullOutputDir);
    jest.spyOn(helpers, 'sanitizeFilename').mockImplementation((input) => {
      // Simple sanitize implementation
      return input.replace(/[/\\:*?"<>|]/g, '_').replace(/\s+/g, '_');
    });
    
    // Setup successful API responses
    (fileReader.readFileContent as jest.Mock).mockResolvedValue('Test prompt for output directory');
    
    // Configure LLM provider behavior
    (llmRegistry.getProvider as jest.Mock).mockImplementation((providerId) => {
      if (providerId === 'provider-a') {
        return {
          providerId: 'provider-a',
          generate: jest.fn().mockResolvedValue({
            provider: 'provider-a',
            modelId: 'model-1',
            text: 'Provider A test response',
            metadata: { usage: { total_tokens: 10 } }
          })
        };
      } else if (providerId === 'error-provider') {
        return {
          providerId: 'error-provider',
          generate: jest.fn().mockRejectedValue(new Error('API error for testing'))
        };
      }
      return null;
    });
    
    // Setup config with models
    (configManager.loadConfig as jest.Mock).mockResolvedValue({
      models: [
        {
          provider: 'provider-a',
          modelId: 'model-1',
          enabled: true
        },
        {
          provider: 'error-provider',
          modelId: 'error-model',
          enabled: true
        }
      ]
    });
    
    // Default enabled models implementation
    (configManager.getEnabledModels as jest.Mock).mockImplementation((config) => {
      return config.models.filter((model: any) => model.enabled);
    });
    
    // No missing API keys by default
    (configManager.validateModelApiKeys as jest.Mock).mockReturnValue({
      missingKeyModels: []
    });
  });
  
  afterEach(() => {
    // Restore console functions
    console.log = originalConsoleLog;
    console.error = originalConsoleError;
  });
  
  it('should create the output directory with the correct structure', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      output: mockOutputDir,
      includeMetadata: false,
      useColors: false,
    };
    
    await runthinktank(options);
    
    // Verify mkdir was called
    expect(mockFs.mkdir).toHaveBeenCalled();
    
    // Check that the options include recursive
    const mkdirCalls = mockFs.mkdir.mock.calls;
    expect(mkdirCalls.length).toBeGreaterThan(0);
    expect(mkdirCalls[0][1]).toEqual(expect.objectContaining({ recursive: true }));
  });
  
  it('should write files for each model response', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      output: mockOutputDir,
      includeMetadata: true,
      useColors: false,
    };
    
    await runthinktank(options);
    
    // Verify writeFile was called
    expect(mockFs.writeFile).toHaveBeenCalled();
    
    // Check the content format
    const writeFileCalls = mockFs.writeFile.mock.calls;
    expect(writeFileCalls.length).toBeGreaterThan(0);
    
    // Verify content format (markdown)
    for (const [_, content] of writeFileCalls) {
      const contentStr = content as string;
      expect(typeof contentStr).toBe('string');
      expect(contentStr).toContain('# '); // Title
      // Every response file should have either a Response section or an Error section
      const hasResponseSection = contentStr.indexOf('## Response') !== -1;
      const hasErrorSection = contentStr.indexOf('## Error') !== -1;
      expect(hasResponseSection || hasErrorSection).toBe(true);
    }
  });
  
  it('should still write files for models that return errors', async () => {
    // Use a model that will error
    (configManager.getEnabledModels as jest.Mock).mockReturnValue([
      {
        provider: 'error-provider',
        modelId: 'error-model',
        enabled: true
      }
    ]);
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
      output: mockOutputDir,
      includeMetadata: false,
      useColors: false,
    };
    
    await runthinktank(options);
    
    // Verify writeFile was still called
    expect(mockFs.writeFile).toHaveBeenCalled();
    
    // Verify error content
    const writeFileCalls = mockFs.writeFile.mock.calls;
    const content = writeFileCalls[0][1] as string;
    expect(typeof content).toBe('string');
    expect(content.indexOf('## Error')).not.toBe(-1); // Contains error section
    expect(content.indexOf('API error for testing')).not.toBe(-1); // Contains error message
  });
  
  it('should handle errors during directory creation', async () => {
    // Make mkdir fail
    mockFs.mkdir.mockRejectedValue(new Error('Permission denied'));
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
      output: mockOutputDir,
      includeMetadata: false,
      useColors: false,
    };
    
    await expect(runthinktank(options)).rejects.toThrow('Failed to create output directory');
  });
  
  it('should handle errors during file writing without crashing', async () => {
    // Make writeFile fail
    mockFs.writeFile.mockRejectedValue(new Error('Disk full'));
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
      output: mockOutputDir,
      includeMetadata: false,
      useColors: false,
    };
    
    // Should not throw
    await runthinktank(options);
    
    // Test passes if no exception
    expect(true).toBe(true);
  });
});