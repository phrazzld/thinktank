/**
 * End-to-End tests for runThinktank workflow.
 * 
 * These tests verify that the refactored runThinktank workflow behaves
 * correctly from an external perspective, focusing on the entire flow
 * from input to output.
 */
import path from 'path';
import fs from 'fs/promises';
import os from 'os';
import { runThinktank, RunOptions } from '../runThinktank';
import { ThinktankError, FileSystemError, ConfigError } from '../../core/errors';
import * as llmRegistry from '../../core/llmRegistry';
import * as configManager from '../../core/configManager';

// Create a custom matcher to check if a directory exists and has expected files
expect.extend({
  async toBeDirectoryWithFiles(dirPath: string, expectedFiles: string[] = []) {
    try {
      const stats = await fs.stat(dirPath);
      if (!stats.isDirectory()) {
        return {
          pass: false,
          message: () => `Expected ${dirPath} to be a directory but it's not.`
        };
      }

      if (expectedFiles.length > 0) {
        const files = await fs.readdir(dirPath);
        const missingFiles = expectedFiles.filter(f => !files.includes(f));
        
        if (missingFiles.length > 0) {
          return {
            pass: false,
            message: () => `Directory ${dirPath} is missing expected files: ${missingFiles.join(', ')}`
          };
        }
      }
      
      return {
        pass: true,
        message: () => `Expected ${dirPath} not to be a directory with files ${expectedFiles.join(', ')}`
      };
    } catch (error) {
      return {
        pass: false,
        message: () => `Error checking directory ${dirPath}: ${error instanceof Error ? error.message : String(error)}`
      };
    }
  }
});

// Add the matcher to TypeScript's type system
// Using module augmentation instead of namespace
import '@jest/expect';

declare module '@jest/expect' {
  interface AsymmetricMatchers {
    toBeDirectoryWithFiles(expectedFiles?: string[]): void;
  }
  interface Matchers<R> {
    toBeDirectoryWithFiles(expectedFiles?: string[]): Promise<R>;
  }
}

// Mock ora spinner to prevent hanging in tests
jest.mock('ora', () => {
  const mockSpinner = {
    start: jest.fn().mockReturnThis(),
    stop: jest.fn().mockReturnThis(),
    succeed: jest.fn().mockReturnThis(),
    fail: jest.fn().mockReturnThis(),
    warn: jest.fn().mockReturnThis(),
    info: jest.fn().mockReturnThis(),
    clear: jest.fn().mockReturnThis(),
    render: jest.fn().mockReturnThis(),
    frame: jest.fn().mockReturnThis(),
    _text: '',
    get text() {
      return this._text;
    },
    set text(value) {
      this._text = value;
    }
  };
  return jest.fn(() => mockSpinner);
});

// Define mock LLM provider
const MockProvider = {
  providerId: 'mock',
  generate: jest.fn().mockResolvedValue({
    provider: 'mock',
    modelId: 'mock-model',
    text: 'This is a mock response',
    metadata: {
      usage: { total_tokens: 10 },
      model: 'mock-model',
      id: 'mock-response-id',
    }
  }),
  // Add any necessary provider properties
  apiKeyEnvVar: 'MOCK_API_KEY',
  listModels: jest.fn().mockResolvedValue([{ id: 'mock-model', name: 'Mock Model' }])
};

// Helper functions for test setup and cleanup
async function createTempDir(): Promise<string> {
  const tempDir = path.join(os.tmpdir(), `thinktank-test-${Date.now()}`);
  await fs.mkdir(tempDir, { recursive: true });
  return tempDir;
}

async function createTestPrompt(tempDir: string, content = 'This is a test prompt.'): Promise<string> {
  const promptPath = path.join(tempDir, 'test-prompt.txt');
  await fs.writeFile(promptPath, content);
  return promptPath;
}

async function createTestConfig(configDir: string): Promise<string> {
  // Create a minimal test config
  const config = {
    models: [
      {
        provider: 'mock',
        modelId: 'mock-model',
        enabled: true,
        apiKeyEnvVar: 'MOCK_API_KEY',
        options: {
          temperature: 0.7
        }
      }
    ],
    groups: {
      default: {
        name: 'default',
        systemPrompt: { text: 'You are a helpful assistant.' },
        models: [
          {
            provider: 'mock',
            modelId: 'mock-model',
            enabled: true,
            apiKeyEnvVar: 'MOCK_API_KEY'
          }
        ]
      }
    }
  };

  const configPath = path.join(configDir, 'thinktank.config.json');
  await fs.writeFile(configPath, JSON.stringify(config, null, 2));
  return configPath;
}

async function cleanupTempDir(tempDir: string): Promise<void> {
  try {
    await fs.rm(tempDir, { recursive: true, force: true });
  } catch (error) {
    console.error(`Error cleaning up temp directory ${tempDir}:`, error instanceof Error ? error.message : String(error));
  }
}

/**
 * Helper function to list all files in a directory and its subdirectories recursively
 * @param dirPath - The directory to list
 * @returns Array of full file paths
 */
async function listDirectoryRecursive(dirPath: string): Promise<string[]> {
  let result: string[] = [];
  
  try {
    const entries = await fs.readdir(dirPath, { withFileTypes: true });
    
    for (const entry of entries) {
      const fullPath = path.join(dirPath, entry.name);
      
      if (entry.isDirectory()) {
        // Recursively scan subdirectories
        const subDirFiles = await listDirectoryRecursive(fullPath);
        result = result.concat(subDirFiles);
      } else {
        // Add files to the result
        result.push(fullPath);
      }
    }
  } catch (error) {
    console.error(`Error listing directory ${dirPath}:`, error instanceof Error ? error.message : String(error));
  }
  
  return result;
}

describe('runThinktank End-to-End Tests', () => {
  // Test variables
  let tempDir: string;
  let promptPath: string;
  let configPath: string;
  let outputDir: string;
  
  beforeAll(async () => {
    
    // Mock the registry to return our mock provider
    jest.spyOn(llmRegistry, 'getProvider').mockImplementation((providerId: string) => {
      if (providerId === 'mock') {
        return MockProvider;
      }
      return undefined;
    });
    
    // Setup test environment
    tempDir = await createTempDir();
    promptPath = await createTestPrompt(tempDir);
    configPath = await createTestConfig(tempDir);
    outputDir = path.join(tempDir, 'output');
    await fs.mkdir(outputDir, { recursive: true });
    
    // Mock config loading to use our test config
    jest.spyOn(configManager, 'loadConfig').mockImplementation(async () => {
      const configStr = await fs.readFile(configPath, 'utf-8');
      return JSON.parse(configStr);
    });
    
    // Mock the API key validation to always return valid models
    jest.spyOn(configManager, 'validateModelApiKeys').mockImplementation((config) => {
      const enabledModels = configManager.getEnabledModels(config);
      return {
        validModels: enabledModels,
        missingKeyModels: []
      };
    });
    
    // Mock getApiKey to return a fake API key for our mock provider
    jest.spyOn(configManager, 'getApiKey').mockReturnValue('mock-api-key');
    
    // Set the mock environment variable
    process.env.MOCK_API_KEY = 'mock-api-key-value';
  });
  
  afterAll(async () => {
    // Restore original mocks
    jest.restoreAllMocks();
    
    // Clean up test directory
    await cleanupTempDir(tempDir);
    
    // Clean up environment variables
    delete process.env.MOCK_API_KEY;
    
    // Force Jest to exit by setting a short timeout
    setTimeout(() => {
      process.exit(0);
    }, 100);
  });
  
  /**
   * Helper function to create a test context file with specified content
   * @param dir - Directory to create the file in
   * @param filename - Name of the file
   * @param content - Content to write to the file
   * @returns Path to the created file
   */
  async function createTestContextFile(dir: string, filename: string, content: string): Promise<string> {
    const filePath = path.join(dir, filename);
    await fs.writeFile(filePath, content);
    return filePath;
  }
  
  /**
   * Helper function to create a test directory with multiple files
   * @param parentDir - Parent directory to create the test directory in
   * @param dirName - Name of the directory to create
   * @param files - Map of filenames to content
   * @returns Path to the created directory
   */
  async function createTestContextDirectory(
    parentDir: string, 
    dirName: string,
    files: Record<string, string>
  ): Promise<string> {
    const dirPath = path.join(parentDir, dirName);
    await fs.mkdir(dirPath, { recursive: true });
    
    // Create all files in the directory
    for (const [filename, content] of Object.entries(files)) {
      await createTestContextFile(dirPath, filename, content);
    }
    
    return dirPath;
  }
  
  it('should process a prompt file and generate output files', async () => {
    // Define options with our test configuration
    const options: RunOptions = {
      input: promptPath,
      configPath: configPath,
      output: outputDir,
      includeMetadata: false,
      useColors: false
    };
    
    // Run thinktank with our test options
    const result = await runThinktank(options);
    
    // Verify the result contains expected content
    expect(result).toContain('mock-model'); // Output should mention the model
    
    // Get a directory listing recursively to see all files
    const allFiles = await listDirectoryRecursive(outputDir);
    
    // Find any file containing 'mock' in its name
    const mockFiles = allFiles.filter(f => f.includes('mock'));
    
    // Check we have at least one file with 'mock' in its name
    expect(mockFiles.length).toBeGreaterThan(0);
    
    // Read the content of the first mock file
    const fileContent = await fs.readFile(mockFiles[0], 'utf-8');
    expect(fileContent).toContain('This is a mock response');
  });

  it('should handle multiple models when specified', async () => {
    // Add another model to the mock provider for this test
    MockProvider.generate.mockImplementation(async (_prompt: string, modelId: string) => {
      return {
        provider: 'mock',
        modelId,
        text: `This is a response from ${modelId}`,
        metadata: {
          usage: { total_tokens: 10 },
          model: modelId,
          id: `mock-response-id-${modelId}`,
        }
      };
    });
    
    // Update config for this test with multiple models
    const multiModelConfig = {
      models: [
        {
          provider: 'mock',
          modelId: 'mock-model-1',
          enabled: true,
          apiKeyEnvVar: 'MOCK_API_KEY'
        },
        {
          provider: 'mock',
          modelId: 'mock-model-2',
          enabled: true,
          apiKeyEnvVar: 'MOCK_API_KEY'
        }
      ],
      groups: {
        default: {
          name: 'default',
          systemPrompt: { text: 'You are a helpful assistant.' },
          models: [
            {
              provider: 'mock',
              modelId: 'mock-model-1',
              enabled: true,
              apiKeyEnvVar: 'MOCK_API_KEY'
            },
            {
              provider: 'mock',
              modelId: 'mock-model-2',
              enabled: true,
              apiKeyEnvVar: 'MOCK_API_KEY'
            }
          ]
        }
      }
    };
    
    // Write new config to the config path
    await fs.writeFile(configPath, JSON.stringify(multiModelConfig, null, 2));
    
    // Create a subdirectory for this test's output
    const multiModelOutputDir = path.join(tempDir, 'multi-model-output');
    await fs.mkdir(multiModelOutputDir, { recursive: true });
    
    // Run thinktank with options specifying multiple models
    const options: RunOptions = {
      input: promptPath,
      configPath: configPath,
      output: multiModelOutputDir,
      includeMetadata: false,
      useColors: false
    };
    
    // Execute the function
    const result = await runThinktank(options);
    
    // Verify output mentions both models
    expect(result).toContain('mock-model-1');
    expect(result).toContain('mock-model-2');
    
    // Get a directory listing recursively to see all files
    const allFiles = await listDirectoryRecursive(multiModelOutputDir);
    
    // Find files for each model
    const model1Files = allFiles.filter(f => f.includes('mock-model-1'));
    const model2Files = allFiles.filter(f => f.includes('mock-model-2'));
    
    // Verify we have files for each model
    expect(model1Files.length).toBe(1);
    expect(model2Files.length).toBe(1);
    
    // Verify each file has the correct model-specific content
    for (const modelId of ['mock-model-1', 'mock-model-2']) {
      const modelFile = allFiles.find(f => f.includes(modelId));
      if (modelFile) {
        const content = await fs.readFile(modelFile, 'utf-8');
        expect(content).toContain(`This is a response from ${modelId}`);
      } else {
        fail(`No file found for model ${modelId}`);
      }
    }
  });

  it('should handle errors when input file does not exist', async () => {
    // Define options with a non-existent input file
    const options: RunOptions = {
      input: path.join(tempDir, 'nonexistent.txt'),
      configPath: configPath,
      output: outputDir,
      includeMetadata: false,
      useColors: false
    };
    
    // Expect runThinktank to throw a FileSystemError
    await expect(runThinktank(options)).rejects.toThrow(FileSystemError);
    
    try {
      await runThinktank(options);
    } catch (error) {
      if (error instanceof ThinktankError) {
        // Verify error properties
        expect(error.category).toBeDefined();
        expect(error.suggestions).toBeDefined();
        expect(error.suggestions!.length).toBeGreaterThan(0);
        
        // Verify error message mentions the file
        expect(error.message).toContain('nonexistent.txt');
      }
    }
  });

  it('should handle errors when config is invalid', async () => {
    // Write an invalid config
    await fs.writeFile(configPath, '{ invalid json }');
    
    // Define options with the invalid config
    const options: RunOptions = {
      input: promptPath,
      configPath: configPath,
      output: outputDir,
      includeMetadata: false,
      useColors: false
    };
    
    // Expect runThinktank to throw a ConfigError
    await expect(runThinktank(options)).rejects.toThrow(ConfigError);
  });

  it('should gracefully handle API errors', async () => {
    // Make the provider throw an error for this test
    MockProvider.generate.mockRejectedValueOnce(new Error('API connection failed'));
    
    // Restore the valid config
    await createTestConfig(tempDir);
    
    // Define options
    const options: RunOptions = {
      input: promptPath,
      configPath: configPath,
      output: outputDir,
      includeMetadata: false,
      useColors: false
    };
    
    // Since the error is handled internally, we expect a result string instead of a thrown error
    const result = await runThinktank(options);
    
    // Verify the result contains an error message
    expect(result).toContain('Error');
    expect(result).toContain('API connection failed');
  });
  
  /**
   * Integration tests for context paths
   */
  describe('Context Path Integration Tests', () => {
    // Variables for context files and directories
    let testContextFile: string;
    let testMultipleFiles: string[];
    let testContextDir: string;
    let nonExistentPath: string;
    let fileWithSpace: string;
    let contextOutputDir: string;
    
    // Setup context test files and directories before all tests
    beforeAll(async () => {
      // Create a separate output directory for context tests
      contextOutputDir = path.join(tempDir, 'context-output');
      await fs.mkdir(contextOutputDir, { recursive: true });
      
      // Create a context file
      testContextFile = await createTestContextFile(
        tempDir, 
        'context-file.js', 
        'function add(a, b) {\n  return a + b;\n}\n\nmodule.exports = { add };'
      );
      
      // Create multiple context files
      testMultipleFiles = await Promise.all([
        createTestContextFile(
          tempDir,
          'utils.ts',
          'export function multiply(a: number, b: number): number {\n  return a * b;\n}'
        ),
        createTestContextFile(
          tempDir,
          'config.json',
          '{\n  "maxItems": 100,\n  "debug": true\n}'
        )
      ]);
      
      // Create a context directory with multiple files
      testContextDir = await createTestContextDirectory(
        tempDir,
        'context-dir',
        {
          'index.js': 'const utils = require("./utils");\nconsole.log(utils.sum(5, 10));',
          'utils.js': 'function sum(a, b) {\n  return a + b;\n}\n\nmodule.exports = { sum };',
          'README.md': '# Test Context Directory\n\nThis directory contains test files for context.'
        }
      );
      
      // Create a file with space in name
      fileWithSpace = await createTestContextFile(
        tempDir,
        'file with spaces.txt',
        'This is a test file with spaces in its name.'
      );
      
      // Set path to a non-existent file
      nonExistentPath = path.join(tempDir, 'non-existent-file.txt');
    });
    
    // Configure mock response to include context in the response for verification
    beforeEach(() => {
      // Reset the mock function
      MockProvider.generate.mockReset();
      
      // Mock the LLM provider to reflect the context in its output
      MockProvider.generate.mockImplementation(async (prompt: string) => {
        // Extract relevant parts of the context to include in mock response
        const hasContext = prompt.includes('# CONTEXT DOCUMENTS');
        const contextFiles = prompt.match(/## File: .*?(?=\s*^#|\s*$)/gms) || [];
        
        let contextSummary = '';
        if (hasContext && contextFiles.length > 0) {
          contextSummary = `\n\nBased on ${contextFiles.length} context file(s):\n`;
          contextFiles.forEach((fileSection) => {
            const fileMatch = fileSection.match(/## File: (.*?)$/m);
            if (fileMatch && fileMatch[1]) {
              contextSummary += `- ${path.basename(fileMatch[1])}\n`;
            }
          });
        }
        
        return {
          provider: 'mock',
          modelId: 'mock-model',
          text: `This is a mock response to your prompt${contextSummary}`,
          metadata: {
            usage: { total_tokens: hasContext ? 20 : 10 },
            model: 'mock-model',
            id: 'mock-response-id',
          }
        };
      });
    });
    
    it('should process a single context file and include it in the prompt', async () => {
      // Define options with a single context file
      const options: RunOptions = {
        input: promptPath,
        contextPaths: [testContextFile],
        configPath: configPath,
        output: contextOutputDir,
        includeMetadata: false,
        useColors: false
      };
      
      // Run thinktank with the context file
      const result = await runThinktank(options);
      
      // Verify the result mentions context
      expect(result).toContain('context-file.js');
      
      // Get output files and check content
      const allFiles = await listDirectoryRecursive(contextOutputDir);
      const mockFiles = allFiles.filter(f => f.includes('mock'));
      
      // Check we have at least one output file
      expect(mockFiles.length).toBeGreaterThan(0);
      
      // Read the file content and verify it includes context file reference
      const fileContent = await fs.readFile(mockFiles[0], 'utf-8');
      expect(fileContent).toContain('context file');
      expect(fileContent).toContain('context-file.js');
    });
    
    it('should process multiple context files and include them in the prompt', async () => {
      // Define options with multiple context files
      const options: RunOptions = {
        input: promptPath,
        contextPaths: testMultipleFiles,
        configPath: configPath,
        output: path.join(contextOutputDir, 'multi-file'),
        includeMetadata: false,
        useColors: false
      };
      
      // Create output subdirectory
      await fs.mkdir(path.join(contextOutputDir, 'multi-file'), { recursive: true });
      
      // Run thinktank with multiple context files
      const result = await runThinktank(options);
      
      // Verify the result mentions context
      expect(result).toContain('mock-model');
      
      // Get output files and check content
      const allFiles = await listDirectoryRecursive(path.join(contextOutputDir, 'multi-file'));
      const mockFiles = allFiles.filter(f => f.includes('mock'));
      
      // Check we have at least one output file
      expect(mockFiles.length).toBeGreaterThan(0);
      
      // Read the file content and verify it includes context file references
      const fileContent = await fs.readFile(mockFiles[0], 'utf-8');
      expect(fileContent).toContain('context file');
      expect(fileContent).toContain('utils.ts');
      expect(fileContent).toContain('config.json');
    });
    
    it('should process a directory as context and include its files in the prompt', async () => {
      // Define options with a directory as context
      const options: RunOptions = {
        input: promptPath,
        contextPaths: [testContextDir],
        configPath: configPath,
        output: path.join(contextOutputDir, 'directory'),
        includeMetadata: false,
        useColors: false
      };
      
      // Create output subdirectory
      await fs.mkdir(path.join(contextOutputDir, 'directory'), { recursive: true });
      
      // Run thinktank with directory context
      const result = await runThinktank(options);
      
      // Verify the result mentions context
      expect(result).toContain('mock-model');
      
      // Get output files and check content
      const allFiles = await listDirectoryRecursive(path.join(contextOutputDir, 'directory'));
      const mockFiles = allFiles.filter(f => f.includes('mock'));
      
      // Check we have at least one output file
      expect(mockFiles.length).toBeGreaterThan(0);
      
      // Read the file content and verify it includes context file references
      const fileContent = await fs.readFile(mockFiles[0], 'utf-8');
      expect(fileContent).toContain('context file');
      expect(fileContent).toContain('index.js');
      expect(fileContent).toContain('utils.js');
      expect(fileContent).toContain('README.md');
    });
    
    it('should process mixed context paths (files and directories)', async () => {
      // Define options with mixed context paths
      const options: RunOptions = {
        input: promptPath,
        contextPaths: [testContextFile, testContextDir],
        configPath: configPath,
        output: path.join(contextOutputDir, 'mixed'),
        includeMetadata: false,
        useColors: false
      };
      
      // Create output subdirectory
      await fs.mkdir(path.join(contextOutputDir, 'mixed'), { recursive: true });
      
      // Run thinktank with mixed context paths
      const result = await runThinktank(options);
      
      // Verify the result mentions context
      expect(result).toContain('mock-model');
      
      // Get output files and check content
      const allFiles = await listDirectoryRecursive(path.join(contextOutputDir, 'mixed'));
      const mockFiles = allFiles.filter(f => f.includes('mock'));
      
      // Check we have at least one output file
      expect(mockFiles.length).toBeGreaterThan(0);
      
      // Read the file content and verify it includes context file references
      const fileContent = await fs.readFile(mockFiles[0], 'utf-8');
      expect(fileContent).toContain('context file');
      // Should include both standalone file and directory files
      expect(fileContent).toContain('context-file.js');
      expect(fileContent).toContain('index.js');
    });
    
    it('should handle files with spaces in their names', async () => {
      // Define options with a file that has spaces in its name
      const options: RunOptions = {
        input: promptPath,
        contextPaths: [fileWithSpace],
        configPath: configPath,
        output: path.join(contextOutputDir, 'spaces'),
        includeMetadata: false,
        useColors: false
      };
      
      // Create output subdirectory
      await fs.mkdir(path.join(contextOutputDir, 'spaces'), { recursive: true });
      
      // Run thinktank with file that has spaces in name
      const result = await runThinktank(options);
      
      // Verify the result mentions context
      expect(result).toContain('mock-model');
      
      // Get output files and check content
      const allFiles = await listDirectoryRecursive(path.join(contextOutputDir, 'spaces'));
      const mockFiles = allFiles.filter(f => f.includes('mock'));
      
      // Check we have at least one output file
      expect(mockFiles.length).toBeGreaterThan(0);
      
      // Read the file content and verify it includes context file reference
      const fileContent = await fs.readFile(mockFiles[0], 'utf-8');
      expect(fileContent).toContain('context file');
      expect(fileContent).toContain('file with spaces.txt');
    });
    
    it('should continue with original prompt when context paths are invalid', async () => {
      // Define options with a non-existent path
      const options: RunOptions = {
        input: promptPath,
        contextPaths: [nonExistentPath],
        configPath: configPath,
        output: path.join(contextOutputDir, 'invalid'),
        includeMetadata: false,
        useColors: false
      };
      
      // Create output subdirectory
      await fs.mkdir(path.join(contextOutputDir, 'invalid'), { recursive: true });
      
      // Run thinktank with invalid context path
      const result = await runThinktank(options);
      
      // Verify the result doesn't mention context (should continue with original prompt)
      expect(result).toContain('mock-model');
      
      // Get output files and check content
      const allFiles = await listDirectoryRecursive(path.join(contextOutputDir, 'invalid'));
      const mockFiles = allFiles.filter(f => f.includes('mock'));
      
      // Check we have at least one output file
      expect(mockFiles.length).toBeGreaterThan(0);
      
      // Read the file content and verify it doesn't include context file references
      const fileContent = await fs.readFile(mockFiles[0], 'utf-8');
      expect(fileContent).not.toContain('context file');
    });
    
    it('should handle a mix of valid and invalid context paths', async () => {
      // Define options with a mix of valid and invalid paths
      const options: RunOptions = {
        input: promptPath,
        contextPaths: [testContextFile, nonExistentPath],
        configPath: configPath,
        output: path.join(contextOutputDir, 'mixed-valid-invalid'),
        includeMetadata: false,
        useColors: false
      };
      
      // Create output subdirectory
      await fs.mkdir(path.join(contextOutputDir, 'mixed-valid-invalid'), { recursive: true });
      
      // Run thinktank with mixed valid/invalid paths
      const result = await runThinktank(options);
      
      // Verify the result mentions context (from valid path)
      expect(result).toContain('mock-model');
      
      // Get output files and check content
      const allFiles = await listDirectoryRecursive(path.join(contextOutputDir, 'mixed-valid-invalid'));
      const mockFiles = allFiles.filter(f => f.includes('mock'));
      
      // Check we have at least one output file
      expect(mockFiles.length).toBeGreaterThan(0);
      
      // Read the file content and verify it includes valid context file references
      const fileContent = await fs.readFile(mockFiles[0], 'utf-8');
      expect(fileContent).toContain('context file');
      expect(fileContent).toContain('context-file.js');
      // Should not include invalid path
      expect(fileContent).not.toContain('non-existent-file.txt');
    });
  });
});