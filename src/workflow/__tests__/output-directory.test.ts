/**
 * Integration tests for the output directory feature.
 * 
 * These tests verify that the output directory functionality works correctly
 * including directory creation, file writing, and error handling.
 */
import { mockFsModules, resetVirtualFs, getVirtualFs, createFsError } from '../../__tests__/utils/virtualFsUtils';

// Setup mocks (must be before importing fs modules)
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Now import fs after mocking
import fs from 'fs/promises';
import path from 'path';

import { runThinktank, RunOptions } from '../runThinktank';
import * as fileReader from '../../utils/fileReader';
import * as configManager from '../../core/configManager';
import * as llmRegistry from '../../core/llmRegistry';
import * as helpers from '../../utils/helpers';

// Mock non-fs dependencies
jest.mock('../../utils/fileReader');
jest.mock('../../core/configManager');
jest.mock('../../core/llmRegistry');
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

// Mock console to prevent test output pollution
const originalConsoleLog = console.log;
const originalConsoleError = console.error;

describe('Output Directory Feature', () => {
  // Setup constants for testing
  const mockRunDirectoryName = 'thinktank_run_20230515_143000_000';
  const mockOutputDir = '/mock/output/path';
  const mockFullOutputDir = `${mockOutputDir}/${mockRunDirectoryName}`;
  
  // Access the virtual filesystem
  const virtualFs = getVirtualFs();
  
  beforeEach(() => {
    jest.clearAllMocks();
    resetVirtualFs();
    
    // Silence console during tests
    console.log = jest.fn();
    console.error = jest.fn();
    
    // Configure path functions
    (path.join as jest.Mock).mockImplementation((...args) => args.join('/'));
    (path.resolve as jest.Mock).mockImplementation((...args) => args.join('/'));
    (path.basename as jest.Mock).mockImplementation(filePath => {
      const parts = filePath.split('/');
      return parts[parts.length - 1];
    });
    
    // Reset the volume
    resetVirtualFs();
    
    // Setup filesystem for tests - manually create directories and files
    const vol = getVirtualFs();
    // Create test prompt file
    vol.writeFileSync('test-prompt.txt', 'Test prompt for output directory');
    
    // Spy on fs methods for assertions
    jest.spyOn(fs, 'mkdir');
    jest.spyOn(fs, 'writeFile');
    
    // Mock helper functions
    jest.spyOn(helpers, 'generateOutputDirectoryPath').mockReturnValue(mockFullOutputDir);
    jest.spyOn(helpers, 'sanitizeFilename').mockImplementation((input) => {
      // Simple sanitize implementation
      return input.replace(/[/\\:*?"<>|]/g, '_').replace(/\s+/g, '_');
    });
    
    // Setup successful API responses
    (fileReader.readFileContent as jest.Mock).mockResolvedValue('Test prompt for output directory');
    
    // Mock inputHandler module's processInput function
    jest.mock('../inputHandler', () => ({
      InputSourceType: { FILE: 'file', STDIN: 'stdin', TEXT: 'text' },
      processInput: jest.fn().mockResolvedValue({
        content: 'Test prompt for output directory',
        sourceType: 'file',
        sourcePath: 'test-prompt.txt',
        metadata: {
          processingTimeMs: 5,
          originalLength: 29,
          finalLength: 29,
          normalized: true
        }
      })
    }));
    
    // Mock moduleSelector module
    jest.mock('../modelSelector', () => ({
      selectModels: jest.fn().mockReturnValue({
        models: [
          {
            provider: 'provider-a',
            modelId: 'model-1',
            enabled: true
          }
        ],
        missingApiKeyModels: [],
        disabledModels: [],
        warnings: []
      })
    }));
    
    // Mock queryExecutor module
    jest.mock('../queryExecutor', () => ({
      executeQueries: jest.fn().mockResolvedValue({
        responses: [
          {
            provider: 'provider-a',
            modelId: 'model-1',
            text: 'Provider A test response',
            configKey: 'provider-a:model-1',
            metadata: { usage: { total_tokens: 10 } }
          }
        ],
        statuses: {
          'provider-a:model-1': { 
            status: 'success',
            startTime: 1,
            endTime: 2,
            durationMs: 1
          }
        },
        timing: {
          startTime: 1,
          endTime: 2,
          durationMs: 1
        }
      })
    }));
    
    // Mock outputHandler module
    jest.mock('../outputHandler', () => ({
      createOutputDirectory: jest.fn().mockResolvedValue('/fake/output/dir'),
      writeResponsesToFiles: jest.fn().mockResolvedValue({
        outputDirectory: '/fake/output/dir',
        files: [{ status: 'success', filename: 'provider-a-model-1.md' }],
        succeededWrites: 1,
        failedWrites: 0,
        timing: { startTime: 1, endTime: 2, durationMs: 1 }
      }),
      formatForConsole: jest.fn().mockReturnValue('Mock console output')
    }));
    
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
    
    await runThinktank(options);
    
    // Verify the output directory was created in the virtual filesystem
    const pathExists = virtualFs.existsSync(mockOutputDir);
    expect(pathExists).toBe(true);
    
    // Verify mkdir was called with recursive option
    expect(fs.mkdir).toHaveBeenCalled();
    const mkdirCalls = (fs.mkdir as jest.Mock).mock.calls;
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
    
    await runThinktank(options);
    
    // Verify writeFile was called
    expect(fs.writeFile).toHaveBeenCalled();
    
    // Check the file content
    const writeFileCalls = (fs.writeFile as jest.Mock).mock.calls;
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
    
    await runThinktank(options);
    
    // Verify writeFile was still called
    expect(fs.writeFile).toHaveBeenCalled();
    
    // Verify error content
    const writeFileCalls = (fs.writeFile as jest.Mock).mock.calls;
    const content = writeFileCalls[0][1] as string;
    expect(typeof content).toBe('string');
    expect(content.indexOf('## Error')).not.toBe(-1); // Contains error section
    expect(content.indexOf('API error for testing')).not.toBe(-1); // Contains error message
  });
  
  it('should handle errors during directory creation', async () => {
    // Spy on fs.mkdir and make it reject with an error
    jest.spyOn(fs, 'mkdir').mockRejectedValueOnce(
      createFsError('EACCES', 'Permission denied', 'mkdir', mockFullOutputDir)
    );
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
      output: mockOutputDir,
      includeMetadata: false,
      useColors: false,
    };
    
    await expect(runThinktank(options)).rejects.toThrow('Failed to create output directory');
  });
  
  it('should handle errors during file writing without crashing', async () => {
    // Spy on fs.writeFile and make it reject with an error
    jest.spyOn(fs, 'writeFile').mockRejectedValueOnce(
      createFsError('ENOSPC', 'No space left on device', 'write', `${mockFullOutputDir}/output.md`)
    );
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
      output: mockOutputDir,
      includeMetadata: false,
      useColors: false,
    };
    
    // Should not throw
    await runThinktank(options);
    
    // Test passes if no exception
    expect(true).toBe(true);
  });
});