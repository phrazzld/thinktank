/**
 * True End-to-End tests for runThinktank workflow.
 * 
 * These tests verify that the runThinktank workflow behaves correctly when invoked through the CLI, 
 * treating the application as a black box and controlling it only through external interfaces
 * (CLI arguments and configuration files).
 */
import path from 'path';
import fs from 'fs/promises';
import execa from 'execa';
// No longer need to import these error types when using CLI as a black box
import {
  createTempTestDir,
  createTestFile,
  createTestDir,
  createTestConfig,
  cleanupTestDir,
  listFilesRecursive,
  shouldSkipFsE2ETests
} from '../../__tests__/utils/e2eTestUtils';

// Skip all tests if environment indicates E2E tests should be skipped
const skipTests = shouldSkipFsE2ETests();

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
// Using module augmentation
import '@jest/expect';

declare module '@jest/expect' {
  interface AsymmetricMatchers {
    toBeDirectoryWithFiles(expectedFiles?: string[]): void;
  }
  interface Matchers<R> {
    toBeDirectoryWithFiles(expectedFiles?: string[]): Promise<R>;
  }
}

describe('runThinktank End-to-End Tests', () => {
  // Test variables
  let tempDir: string;
  let promptPath: string;
  let configPath: string;
  let outputDir: string;
  let cliPath: string;
  
  beforeAll(async () => {
    // Skip setup if tests will be skipped
    if (skipTests) return;
    
    // Setup test environment with real filesystem
    tempDir = await createTempTestDir();
    promptPath = await createTestFile(tempDir, 'test-prompt.txt', 'This is a test prompt.');
    outputDir = await createTestDir(tempDir, 'output');
    
    // Create a standard test config with mock models and groups
    const baseTestConfig = {
      models: [
        { provider: 'mock', modelId: 'test-model-a', enabled: true, apiKeyEnvVar: 'MOCK_API_KEY' },
        { provider: 'mock', modelId: 'test-model-b', enabled: true, apiKeyEnvVar: 'MOCK_API_KEY' },
        { provider: 'mock', modelId: 'test-model-c', enabled: false, apiKeyEnvVar: 'MOCK_API_KEY' },
      ],
      groups: {
        'test-group-a': { 
          name: 'test-group-a', 
          models: [{ provider: 'mock', modelId: 'test-model-a', enabled: true }], 
          systemPrompt: { text: 'Test group A prompt' } 
        },
        'test-group-b': { 
          name: 'test-group-b', 
          models: [{ provider: 'mock', modelId: 'test-model-b', enabled: true }], 
          systemPrompt: { text: 'Test group B prompt' } 
        }
      }
    };
    
    configPath = await createTestConfig(tempDir, baseTestConfig);
    
    // Path to the CLI script
    cliPath = path.resolve(__dirname, '../../../dist/cli/index.js');
    
    // Set the mock environment variable
    process.env.MOCK_API_KEY = 'mock-api-key-value';
  });
  
  afterAll(async () => {
    // Skip cleanup if tests were skipped
    if (skipTests) return;
    
    // Clean up test directory
    await cleanupTestDir(tempDir);
    
    // Clean up environment variables
    delete process.env.MOCK_API_KEY;
  });
  
  // Skip tests conditionally
  beforeEach(() => {
    // This is intentionally empty to skip tests when needed
  });
  
  it('should process a prompt file and generate output files', async () => {
    if (skipTests) return;
    
    // Create a dedicated output directory for this test
    const singleModelOutput = path.join(outputDir, 'single-model');
    await fs.mkdir(singleModelOutput, { recursive: true });
    
    // Run the CLI using execa with the base config
    const { stdout, stderr } = await execa('node', [
      cliPath,
      'run',
      promptPath,
      '--config', configPath,
      '--models', 'mock:test-model-a',
      '--output', singleModelOutput,
      '--verbose'
    ]);
    
    // Verify the CLI output contains expected information
    expect(stdout).toContain('mock:test-model-a'); // Output should mention the model
    expect(stderr).toBe(''); // No errors should be reported
    
    // Get a directory listing recursively to see all files
    const allFiles = await listFilesRecursive(singleModelOutput);
    
    // Find any file containing 'mock' in its name
    const mockFiles = allFiles.filter(f => path.basename(f).includes('mock'));
    
    // Check we have at least one file with 'mock' in its name
    expect(mockFiles.length).toBeGreaterThan(0);
    
    // Read the content of the first mock file
    const fileContent = await fs.readFile(mockFiles[0], 'utf-8');
    expect(fileContent).toContain('This is a mock response');
  });

  it('should handle multiple models when specified', async () => {
    if (skipTests) return;
    
    // Create a dedicated output directory for this test
    const multiModelOutput = path.join(outputDir, 'multi-model');
    await fs.mkdir(multiModelOutput, { recursive: true });
    
    // Run the CLI using execa with multiple models
    const { stdout, stderr } = await execa('node', [
      cliPath,
      'run',
      promptPath,
      '--config', configPath,
      '--models', 'mock:test-model-a,mock:test-model-b',
      '--output', multiModelOutput,
      '--verbose'
    ]);
    
    // Verify the CLI output contains expected information
    expect(stdout).toContain('mock:test-model-a'); 
    expect(stdout).toContain('mock:test-model-b');
    expect(stderr).toBe(''); // No errors should be reported
    
    // Get a directory listing recursively to see all files
    const allFiles = await listFilesRecursive(multiModelOutput);
    
    // Find files for each model
    const modelAFiles = allFiles.filter(f => path.basename(f).includes('test-model-a'));
    const modelBFiles = allFiles.filter(f => path.basename(f).includes('test-model-b'));
    
    // Verify we have files for each model
    expect(modelAFiles.length).toBe(1);
    expect(modelBFiles.length).toBe(1);
    
    // Verify each file has the correct model-specific content
    for (const modelFile of [...modelAFiles, ...modelBFiles]) {
      const content = await fs.readFile(modelFile, 'utf-8');
      expect(content).toContain('This is a mock response');
    }
  });

  it('should handle model groups when specified', async () => {
    if (skipTests) return;
    
    // Create a dedicated output directory for this test
    const groupOutput = path.join(outputDir, 'group');
    await fs.mkdir(groupOutput, { recursive: true });
    
    // Run the CLI using execa with a group
    const { stdout, stderr } = await execa('node', [
      cliPath,
      'run',
      promptPath,
      '--config', configPath,
      '--group', 'test-group-a',
      '--output', groupOutput,
      '--verbose'
    ]);
    
    // Verify the CLI output contains expected information
    expect(stdout).toContain('test-group-a'); 
    expect(stderr).toBe(''); // No errors should be reported
    
    // Get a directory listing recursively to see all files
    const allFiles = await listFilesRecursive(groupOutput);
    
    // Find files for test-model-a (the only model in test-group-a)
    const modelAFiles = allFiles.filter(f => path.basename(f).includes('test-model-a'));
    
    // Verify we have files for the model in the group
    expect(modelAFiles.length).toBe(1);
    
    // Verify each file has the correct model-specific content
    for (const modelFile of modelAFiles) {
      const content = await fs.readFile(modelFile, 'utf-8');
      expect(content).toContain('This is a mock response');
    }
  });

  it('should handle errors when input file does not exist', async () => {
    if (skipTests) return;
    
    // Define a non-existent input file
    const nonExistentFile = path.join(tempDir, 'nonexistent.txt');
    
    // Expect the CLI to fail with an error message about the input file
    await expect(execa('node', [
      cliPath,
      'run',
      nonExistentFile,
      '--config', configPath,
      '--model', 'mock:test-model-a'
    ])).rejects.toMatchObject({
      stderr: expect.stringContaining('not found'),
      exitCode: 1
    });
  });

  it('should handle errors when config is invalid', async () => {
    if (skipTests) return;
    
    // Create an invalid config file
    const invalidConfigPath = await createTestFile(tempDir, 'invalid-config.json', '{ invalid json }');
    
    // Expect the CLI to fail with an error message about the config file
    await expect(execa('node', [
      cliPath,
      'run',
      promptPath,
      '--config', invalidConfigPath,
      '--model', 'mock:test-model-a'
    ])).rejects.toMatchObject({
      stderr: expect.stringContaining('Configuration error'),
      exitCode: 1
    });
  });

  it('should handle scenario-specific configurations', async () => {
    if (skipTests) return;
    
    // Create a dedicated output directory for this test
    const scenarioOutput = path.join(outputDir, 'scenario');
    await fs.mkdir(scenarioOutput, { recursive: true });
    
    // Create a scenario-specific config that only enables test-model-b
    const scenarioConfig = {
      models: [
        { provider: 'mock', modelId: 'test-model-a', enabled: false, apiKeyEnvVar: 'MOCK_API_KEY' },
        { provider: 'mock', modelId: 'test-model-b', enabled: true, apiKeyEnvVar: 'MOCK_API_KEY' },
      ],
      groups: {
        'default': { 
          name: 'default', 
          models: [{ provider: 'mock', modelId: 'test-model-b', enabled: true }], 
          systemPrompt: { text: 'Default scenario prompt' } 
        }
      }
    };
    
    const scenarioConfigPath = await createTestConfig(
      tempDir, 
      scenarioConfig, 
      'scenario-config.json'
    );
    
    // Run the CLI using execa with the scenario-specific config
    const { stdout, stderr } = await execa('node', [
      cliPath,
      'run',
      promptPath,
      '--config', scenarioConfigPath,
      '--output', scenarioOutput,
      '--verbose'
    ]);
    
    // Verify the CLI output contains expected information
    expect(stdout).toContain('mock:test-model-b'); // Only model-b should be used
    expect(stdout).not.toContain('mock:test-model-a'); // model-a should not be mentioned
    expect(stderr).toBe(''); // No errors should be reported
    
    // Get a directory listing recursively to see all files
    const allFiles = await listFilesRecursive(scenarioOutput);
    
    // Find files for each model
    const modelBFiles = allFiles.filter(f => path.basename(f).includes('test-model-b'));
    const modelAFiles = allFiles.filter(f => path.basename(f).includes('test-model-a'));
    
    // Verify we have files only for model B
    expect(modelBFiles.length).toBe(1);
    expect(modelAFiles.length).toBe(0);
  });

  /**
   * Integration tests for context paths
   */
  describe('Context Path Integration Tests', () => {
    // Variables for context files and directories
    let testContextFile: string;
    let testMultipleFiles: string[];
    let testContextDir: string;
    let fileWithSpace: string;
    let contextOutputDir: string;
    
    // Setup context test files and directories before all tests
    beforeAll(async () => {
      if (skipTests) return;
      
      // Create a separate output directory for context tests
      contextOutputDir = await createTestDir(tempDir, 'context-output');
      
      // Create a context file
      testContextFile = await createTestFile(
        tempDir, 
        'context-file.js', 
        'function add(a, b) {\n  return a + b;\n}\n\nmodule.exports = { add };'
      );
      
      // Create multiple context files
      testMultipleFiles = await Promise.all([
        createTestFile(
          tempDir,
          'utils.ts',
          'export function multiply(a: number, b: number): number {\n  return a * b;\n}'
        ),
        createTestFile(
          tempDir,
          'config.json',
          '{\n  "maxItems": 100,\n  "debug": true\n}'
        )
      ]);
      
      // Create a context directory with multiple files
      testContextDir = await createTestDir(
        tempDir,
        'context-dir',
        {
          'index.js': 'const utils = require("./utils");\nconsole.log(utils.sum(5, 10));',
          'utils.js': 'function sum(a, b) {\n  return a + b;\n}\n\nmodule.exports = { sum };',
          'README.md': '# Test Context Directory\n\nThis directory contains test files for context.'
        }
      );
      
      // Create a file with space in name
      fileWithSpace = await createTestFile(
        tempDir,
        'file with spaces.txt',
        'This is a test file with spaces in its name.'
      );
    });
    
    it('should process a single context file and include it in the prompt', async () => {
      if (skipTests) return;
      
      // Create a dedicated output directory for this test
      const singleFileOutput = path.join(contextOutputDir, 'single-file');
      await fs.mkdir(singleFileOutput, { recursive: true });
      
      // Run the CLI using execa with a single context file
      const { stdout } = await execa('node', [
        cliPath,
        'run',
        promptPath,
        testContextFile,
        '--config', configPath,
        '--models', 'mock:test-model-a',
        '--output', singleFileOutput,
        '--verbose'
      ]);
      
      // Verify the CLI output contains expected information
      expect(stdout).toContain('context-file.js');
      
      // Get a directory listing recursively to see all files
      const allFiles = await listFilesRecursive(singleFileOutput);
      const mockFiles = allFiles.filter(f => path.basename(f).includes('mock'));
      
      // Check we have at least one output file
      expect(mockFiles.length).toBeGreaterThan(0);
    });
    
    it('should process multiple context files and include them in the prompt', async () => {
      if (skipTests) return;
      
      // Create a dedicated output directory for this test
      const multiFileOutput = path.join(contextOutputDir, 'multi-file');
      await fs.mkdir(multiFileOutput, { recursive: true });
      
      // Run the CLI using execa with multiple context files
      const { stdout } = await execa('node', [
        cliPath,
        'run',
        promptPath,
        ...testMultipleFiles,
        '--config', configPath,
        '--models', 'mock:test-model-a',
        '--output', multiFileOutput,
        '--verbose'
      ]);
      
      // Verify the CLI output contains expected information
      expect(stdout).toContain('utils.ts');
      expect(stdout).toContain('config.json');
      
      // Get a directory listing recursively to see all files
      const allFiles = await listFilesRecursive(multiFileOutput);
      const mockFiles = allFiles.filter(f => path.basename(f).includes('mock'));
      
      // Check we have at least one output file
      expect(mockFiles.length).toBeGreaterThan(0);
    });
    
    it('should process a directory as context and include its files in the prompt', async () => {
      if (skipTests) return;
      
      // Create a dedicated output directory for this test
      const dirOutput = path.join(contextOutputDir, 'directory');
      await fs.mkdir(dirOutput, { recursive: true });
      
      // Run the CLI using execa with a directory as context
      const { stdout } = await execa('node', [
        cliPath,
        'run',
        promptPath,
        testContextDir,
        '--config', configPath,
        '--models', 'mock:test-model-a',
        '--output', dirOutput,
        '--verbose'
      ]);
      
      // Verify the CLI output contains expected information
      expect(stdout).toContain('context-dir');
      
      // Get a directory listing recursively to see all files
      const allFiles = await listFilesRecursive(dirOutput);
      const mockFiles = allFiles.filter(f => path.basename(f).includes('mock'));
      
      // Check we have at least one output file
      expect(mockFiles.length).toBeGreaterThan(0);
    });
    
    it('should process mixed context paths (files and directories)', async () => {
      if (skipTests) return;
      
      // Create a dedicated output directory for this test
      const mixedOutput = path.join(contextOutputDir, 'mixed');
      await fs.mkdir(mixedOutput, { recursive: true });
      
      // Run the CLI using execa with mixed context paths
      const { stdout } = await execa('node', [
        cliPath,
        'run',
        promptPath,
        testContextFile,
        testContextDir,
        '--config', configPath,
        '--models', 'mock:test-model-a',
        '--output', mixedOutput,
        '--verbose'
      ]);
      
      // Verify the CLI output contains expected information
      expect(stdout).toContain('context-file.js');
      expect(stdout).toContain('context-dir');
      
      // Get a directory listing recursively to see all files
      const allFiles = await listFilesRecursive(mixedOutput);
      const mockFiles = allFiles.filter(f => path.basename(f).includes('mock'));
      
      // Check we have at least one output file
      expect(mockFiles.length).toBeGreaterThan(0);
    });
    
    it('should handle files with spaces in their names', async () => {
      if (skipTests) return;
      
      // Create a dedicated output directory for this test
      const spacesOutput = path.join(contextOutputDir, 'spaces');
      await fs.mkdir(spacesOutput, { recursive: true });
      
      // Run the CLI using execa with a file that has spaces in its name
      const { stdout } = await execa('node', [
        cliPath,
        'run',
        promptPath,
        fileWithSpace,
        '--config', configPath,
        '--models', 'mock:test-model-a',
        '--output', spacesOutput,
        '--verbose'
      ]);
      
      // Verify the CLI output contains expected information
      expect(stdout).toContain('file with spaces.txt');
      
      // Get a directory listing recursively to see all files
      const allFiles = await listFilesRecursive(spacesOutput);
      const mockFiles = allFiles.filter(f => path.basename(f).includes('mock'));
      
      // Check we have at least one output file
      expect(mockFiles.length).toBeGreaterThan(0);
    });
    
    it('should handle a mix of valid and invalid context paths', async () => {
      if (skipTests) return;
      
      // Define a non-existent file path
      const nonExistentPath = path.join(tempDir, 'non-existent-file.txt');
      
      // Create a dedicated output directory for this test
      const mixedValidInvalidOutput = path.join(contextOutputDir, 'mixed-valid-invalid');
      await fs.mkdir(mixedValidInvalidOutput, { recursive: true });
      
      // Run the CLI using execa with mixed valid/invalid paths
      // Note: This may output warnings but should not fail completely
      try {
        const { stdout, stderr } = await execa('node', [
          cliPath,
          'run',
          promptPath,
          testContextFile,
          nonExistentPath,
          '--config', configPath,
          '--models', 'mock:test-model-a',
          '--output', mixedValidInvalidOutput,
          '--verbose'
        ]);
        
        // Verify the CLI output contains expected information
        expect(stdout).toContain('context-file.js');
        expect(stderr).toContain('not found'); // Warning about non-existent file
        
      } catch (error) {
        // If this fails completely, we'll check if it includes the valid file
        // It may be stricter in error handling
        const execaError = error as { stderr?: string };
        expect(execaError.stderr).toContain('context-file.js');
      }
    });
  });
});
