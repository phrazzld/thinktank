/**
 * End-to-End tests for CLI functionality.
 * These tests use execa to simulate actual CLI usage with command-line arguments.
 */
import execa from 'execa';
import path from 'path';
import fs from 'fs/promises';
import os from 'os';

// Define a proper type for Execa errors
interface ExecaError extends Error {
  stderr?: string;
  stdout?: string;
  exitCode?: number;
  command?: string;
}

// Helper to create a temp directory for tests
async function createTempDir(): Promise<string> {
  const tempDir = path.join(os.tmpdir(), `thinktank-test-${Date.now()}`);
  await fs.mkdir(tempDir, { recursive: true });
  return tempDir;
}

// Helper to create a minimal test config file in a temp directory
async function createTestConfig(tempDir: string): Promise<string> {
  const configPath = path.join(tempDir, 'thinktank.config.json');
  const mockConfig = {
    models: [
      {
        provider: 'mock',
        modelId: 'mock-model',
        enabled: true,
        apiKeyEnvVar: 'MOCK_API_KEY'
      }
    ],
    groups: {
      'default': {
        name: 'default',
        systemPrompt: { text: 'You are a helpful assistant.' },
        models: [
          {
            provider: 'mock',
            modelId: 'mock-model',
            enabled: true
          }
        ]
      }
    }
  };
  await fs.writeFile(configPath, JSON.stringify(mockConfig, null, 2));
  return configPath;
}

// Create a temporary prompt file for testing
async function createTestPrompt(tempDir: string, content = 'This is a test prompt.'): Promise<string> {
  const promptPath = path.join(tempDir, 'test-prompt.txt');
  await fs.writeFile(promptPath, content);
  return promptPath;
}

// Create a test context file
async function createTestContextFile(dir: string, filename: string, content: string): Promise<string> {
  const filePath = path.join(dir, filename);
  await fs.writeFile(filePath, content);
  return filePath;
}

// Create a test directory with multiple files
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

// Helper to clean up temporary directory
async function cleanupTempDir(tempDir: string): Promise<void> {
  try {
    await fs.rm(tempDir, { recursive: true, force: true });
  } catch (error) {
    console.error(`Error cleaning up temp directory ${tempDir}:`, error instanceof Error ? error.message : String(error));
  }
}

// Skip full E2E tests if:
// 1. SKIP_E2E_TESTS environment variable is set, OR
// 2. dist/cli.js doesn't exist (not built yet)
const skipE2ETests = async (): Promise<boolean> => {
  if (process.env.SKIP_E2E_TESTS) {
    return true;
  }
  
  const cliPath = path.resolve(__dirname, '../../../dist/cli.js');
  try {
    await fs.access(cliPath);
    return false; // File exists, don't skip
  } catch {
    return true; // File doesn't exist, skip tests
  }
};

// We need to determine skipping before the tests run
let shouldSkipTests = false;

// Using a global beforeAll hook to determine if tests should be skipped
beforeAll(async () => {
  shouldSkipTests = await skipE2ETests();
  if (shouldSkipTests) {
    console.log('Skipping CLI E2E tests - CLI not built or skipped via env var');
  }
});

// Create a describe function that conditionally skips tests
describe('CLI E2E Tests', () => {
  // Set timeout to 30 seconds for long-running tests
  jest.setTimeout(30000); 

  it('should be able to execute system commands', async () => {
    // Skip if necessary
    if (shouldSkipTests) {
      return;
    }
    
    // Simple echo command to verify execa works
    const { stdout } = await execa('echo', ['Execa test successful']);
    expect(stdout).toBe('Execa test successful');
  });
  
  // Test the CLI binary with the help flag
  it('should show help message with --help flag', async () => {
    // Skip if necessary
    if (shouldSkipTests) {
      return;
    }
    
    const cliPath = path.resolve(__dirname, '../../../dist/cli.js');
    
    try {
      const { stdout } = await execa('node', [cliPath, '--help']);
      expect(stdout).toContain('Usage:');
      expect(stdout).toContain('Examples:');
    } catch (error) {
      console.error('Error output:', (error as ExecaError).stderr);
      throw error;
    }
  });
  
  // Test version flag outputs version information
  it('should show version with --version flag', async () => {
    // Skip if necessary
    if (shouldSkipTests) {
      return;
    }
    
    const cliPath = path.resolve(__dirname, '../../../dist/cli.js');
    
    try {
      const { stdout } = await execa('node', [cliPath, '--version']);
      // Version should match semver pattern
      expect(stdout).toMatch(/\d+\.\d+\.\d+/);
    } catch (error) {
      console.error('Error output:', (error as ExecaError).stderr);
      throw error;
    }
  });
  
  // Tests for context files functionality
  describe('Context Files E2E Tests', () => {
    let tempDir: string;
    let configPath: string;
    let promptPath: string;
    let contextFile: string;
    let codeContextFile: string;
    let multipleContextFiles: string[];
    let contextDir: string;
    let fileWithSpaces: string;
    let cliPath: string;
    let outputDir: string;

    // Set up test environment before tests
    beforeAll(async () => {
      if (shouldSkipTests) {
        return;
      }
      
      // Create temp directory for test files
      tempDir = await createTempDir();
      
      // Create config file
      configPath = await createTestConfig(tempDir);
      
      // Create prompt file
      promptPath = await createTestPrompt(tempDir, 'Tell me about the code in the context.');
      
      // Create output directory
      outputDir = path.join(tempDir, 'output');
      await fs.mkdir(outputDir, { recursive: true });
      
      // Create various context files
      contextFile = await createTestContextFile(
        tempDir,
        'context.txt',
        'This is a simple text file for testing context.'
      );
      
      codeContextFile = await createTestContextFile(
        tempDir,
        'example.js',
        'function add(a, b) {\n  return a + b;\n}\n\nmodule.exports = { add };'
      );
      
      fileWithSpaces = await createTestContextFile(
        tempDir,
        'file with spaces.txt',
        'This is a test file with spaces in its name.'
      );
      
      // Create multiple context files
      multipleContextFiles = [
        await createTestContextFile(
          tempDir,
          'utils.ts',
          'export function multiply(a: number, b: number): number {\n  return a * b;\n}'
        ),
        await createTestContextFile(
          tempDir,
          'config.json',
          '{\n  "maxItems": 100,\n  "debug": true\n}'
        )
      ];
      
      // Create a directory with multiple files as context
      contextDir = await createTestContextDirectory(
        tempDir,
        'context-dir',
        {
          'index.js': 'const utils = require("./utils");\nconsole.log(utils.sum(5, 10));',
          'utils.js': 'function sum(a, b) {\n  return a + b;\n}\n\nmodule.exports = { sum };',
          'README.md': '# Test Context Directory\n\nThis directory contains test files for context.'
        }
      );
      
      // Path to the CLI script
      cliPath = path.resolve(__dirname, '../../../dist/cli.js');
      
      // Set environment variable for mock
      process.env.MOCK_API_KEY = 'mock-api-key';
    });
    
    // Clean up after tests
    afterAll(async () => {
      if (shouldSkipTests) {
        return;
      }
      
      // Clean up test directory
      await cleanupTempDir(tempDir);
      
      // Remove environment variable
      delete process.env.MOCK_API_KEY;
    });
    
    // Skip individual tests if necessary
    beforeEach(() => {
      if (shouldSkipTests) {
        return;
      }
    });
    
    it('should show context paths in help text', async () => {
      if (shouldSkipTests) {
        return;
      }
      
      try {
        const { stdout } = await execa('node', [cliPath, 'run', '--help']);
        
        // Check that the help text mentions context paths
        expect(stdout).toContain('contextPaths');
        expect(stdout).toContain('files or directories');
      } catch (error) {
        console.error('Error output:', (error as ExecaError).stderr);
        throw error;
      }
    });
    
    it('should accept a single context file', async () => {
      if (shouldSkipTests) {
        return;
      }
      
      try {
        // Run the CLI with a single context file
        // Using --dry-run to prevent actual API calls
        const { stdout, stderr } = await execa('node', [
          cliPath,
          'run',
          promptPath,
          contextFile,
          '--config', configPath,
          '--verbose',
          '--dry-run',
          '--output', path.join(outputDir, 'single-file')
        ]);
        
        // Check that the output mentions the context file
        expect(stdout).toContain('context.txt');
        expect(stdout).toContain('context file');
        
        // Verify no errors
        expect(stderr).toBe('');
      } catch (error) {
        console.error('Error output:', (error as ExecaError).stderr);
        throw error;
      }
    });
    
    it('should handle code files as context', async () => {
      if (shouldSkipTests) {
        return;
      }
      
      try {
        // Run the CLI with a code file as context
        const { stdout, stderr } = await execa('node', [
          cliPath,
          'run',
          promptPath,
          codeContextFile,
          '--config', configPath,
          '--verbose',
          '--dry-run',
          '--output', path.join(outputDir, 'code-file')
        ]);
        
        // Check that the output mentions the context file
        expect(stdout).toContain('example.js');
        expect(stdout).toContain('context file');
        
        // Verify no errors
        expect(stderr).toBe('');
      } catch (error) {
        console.error('Error output:', (error as ExecaError).stderr);
        throw error;
      }
    });
    
    it('should handle multiple context files', async () => {
      if (shouldSkipTests) {
        return;
      }
      
      try {
        // Run the CLI with multiple context files
        const { stdout, stderr } = await execa('node', [
          cliPath,
          'run',
          promptPath,
          ...multipleContextFiles,
          '--config', configPath,
          '--verbose',
          '--dry-run',
          '--output', path.join(outputDir, 'multiple-files')
        ]);
        
        // Check that the output mentions all context files
        expect(stdout).toContain('utils.ts');
        expect(stdout).toContain('config.json');
        expect(stdout).toContain('context files');
        
        // Verify no errors
        expect(stderr).toBe('');
      } catch (error) {
        console.error('Error output:', (error as ExecaError).stderr);
        throw error;
      }
    });
    
    it('should handle directory as context', async () => {
      if (shouldSkipTests) {
        return;
      }
      
      try {
        // Run the CLI with a directory as context
        const { stdout, stderr } = await execa('node', [
          cliPath,
          'run',
          promptPath,
          contextDir,
          '--config', configPath,
          '--verbose',
          '--dry-run',
          '--output', path.join(outputDir, 'directory')
        ]);
        
        // Check that the output mentions the context directory
        expect(stdout).toContain('context files');
        
        // Verify no errors
        expect(stderr).toBe('');
      } catch (error) {
        console.error('Error output:', (error as ExecaError).stderr);
        throw error;
      }
    });
    
    it('should handle file paths with spaces', async () => {
      if (shouldSkipTests) {
        return;
      }
      
      try {
        // Run the CLI with a file that has spaces in its name
        const { stdout, stderr } = await execa('node', [
          cliPath,
          'run',
          promptPath,
          fileWithSpaces,
          '--config', configPath,
          '--verbose',
          '--dry-run',
          '--output', path.join(outputDir, 'spaces')
        ]);
        
        // Check that the output mentions the context file
        expect(stdout).toContain('context file');
        
        // Verify no errors
        expect(stderr).toBe('');
      } catch (error) {
        console.error('Error output:', (error as ExecaError).stderr);
        throw error;
      }
    });
    
    it('should handle mixed files and directories as context', async () => {
      if (shouldSkipTests) {
        return;
      }
      
      try {
        // Run the CLI with a mix of files and directories
        const { stdout, stderr } = await execa('node', [
          cliPath,
          'run',
          promptPath,
          contextFile,
          contextDir,
          '--config', configPath,
          '--verbose',
          '--dry-run',
          '--output', path.join(outputDir, 'mixed')
        ]);
        
        // Check that the output mentions the context
        expect(stdout).toContain('context files');
        
        // Verify no errors
        expect(stderr).toBe('');
      } catch (error) {
        console.error('Error output:', (error as ExecaError).stderr);
        throw error;
      }
    });
    
    it('should handle non-existent context paths gracefully', async () => {
      if (shouldSkipTests) {
        return;
      }
      
      const nonExistentPath = path.join(tempDir, 'non-existent-file.txt');
      
      try {
        // Run the CLI with a non-existent file
        await execa('node', [
          cliPath,
          'run',
          promptPath,
          nonExistentPath,
          '--config', configPath,
          '--verbose',
          '--dry-run',
          '--output', path.join(outputDir, 'non-existent')
        ]);
        
        // We expect this to fail, so if we get here, something is wrong
        fail('Command should have failed with non-existent path');
      } catch (error) {
        // This is expected - verify the error message
        const err = error as ExecaError;
        expect(err.stderr).toContain('not found');
        expect(err.exitCode).not.toBe(0);
      }
    });
    
    it('should handle a valid context file along with a non-existent one', async () => {
      if (shouldSkipTests) {
        return;
      }
      
      const nonExistentPath = path.join(tempDir, 'non-existent-file.txt');
      
      try {
        // Run the CLI with one valid and one non-existent file
        const { stdout } = await execa('node', [
          cliPath,
          'run',
          promptPath,
          contextFile,
          nonExistentPath,
          '--config', configPath,
          '--verbose',
          '--dry-run',
          '--output', path.join(outputDir, 'mixed-valid-invalid')
        ]);
        
        // When one file exists and one doesn't, the CLI should continue with a warning
        expect(stdout).toContain('context.txt');
        expect(stdout).toContain('could not be read');
        
      } catch (error) {
        console.error('Error output:', (error as ExecaError).stderr);
        throw error;
      }
    });
  });
});