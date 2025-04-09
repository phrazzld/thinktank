/**
 * End-to-End tests for CLI functionality.
 * These tests use execa to simulate actual CLI usage with command-line arguments
 * and interact with the real filesystem using temporary directories.
 */
import execa from 'execa';
import fs from 'fs/promises';
import path from 'path';
import {
  createTempTestDir,
  createTestFile,
  createTestDir,
  createTestConfig,
  cleanupTestDir,
  shouldSkipFsE2ETests,
} from '../../__tests__/utils/e2eTestUtils';

// Note: Execa error type can be inferred from execa's return type

// Skip tests if environment indicates they should be skipped
const skipTests = shouldSkipFsE2ETests();

// We need to verify the CLI is built before running E2E tests
async function isCliBuilt(): Promise<boolean> {
  const cliPath = path.resolve(__dirname, '../../../dist/cli.js');
  try {
    await fs.access(cliPath);
    return true;
  } catch {
    return false;
  }
}

describe('CLI E2E Tests', () => {
  // Set timeout to 30 seconds for long-running tests
  jest.setTimeout(30000);

  // Variables for test resources
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
  let shouldSkipCliTests: boolean;

  // Setup test environment before all tests
  beforeAll(async () => {
    // Check if CLI is built
    const cliBuilt = await isCliBuilt();
    shouldSkipCliTests = skipTests || !cliBuilt;

    if (shouldSkipCliTests) {
      return;
    }

    // Create temp directory for test files
    tempDir = await createTempTestDir('cli-e2e-test-');

    // Create config file
    configPath = await createTestConfig(tempDir);

    // Create prompt file
    promptPath = await createTestFile(
      tempDir,
      'test-prompt.txt',
      'Tell me about the code in the context.'
    );

    // Create output directory
    outputDir = await createTestDir(tempDir, 'output');

    // Create various context files
    contextFile = await createTestFile(
      tempDir,
      'context.txt',
      'This is a simple text file for testing context.'
    );

    codeContextFile = await createTestFile(
      tempDir,
      'example.js',
      'function add(a, b) {\n  return a + b;\n}\n\nmodule.exports = { add };'
    );

    fileWithSpaces = await createTestFile(
      tempDir,
      'file with spaces.txt',
      'This is a test file with spaces in its name.'
    );

    // Create multiple context files
    multipleContextFiles = [
      await createTestFile(
        tempDir,
        'utils.ts',
        'export function multiply(a: number, b: number): number {\n  return a * b;\n}'
      ),
      await createTestFile(tempDir, 'config.json', '{\n  "maxItems": 100,\n  "debug": true\n}'),
    ];

    // Create a directory with multiple files as context
    contextDir = await createTestDir(tempDir, 'context-dir', {
      'index.js': 'const utils = require("./utils");\nconsole.log(utils.sum(5, 10));',
      'utils.js': 'function sum(a, b) {\n  return a + b;\n}\n\nmodule.exports = { sum };',
      'README.md': '# Test Context Directory\n\nThis directory contains test files for context.',
    });

    // Path to the CLI script
    cliPath = path.resolve(__dirname, '../../../dist/cli.js');

    // Set environment variable for mock
    process.env.MOCK_API_KEY = 'mock-api-key';
  });

  // Clean up after tests
  afterAll(async () => {
    if (shouldSkipCliTests) {
      return;
    }

    // Clean up test directory
    await cleanupTestDir(tempDir);

    // Remove environment variable
    delete process.env.MOCK_API_KEY;
  });

  // Individual test cases
  it('should be able to execute system commands', async () => {
    // Skip if necessary
    if (shouldSkipCliTests) {
      return;
    }

    // Simple echo command to verify execa works
    const { stdout } = await execa('echo', ['Execa test successful']);
    expect(stdout).toBe('Execa test successful');
  });

  // Test the CLI binary with the help flag
  it('should show help message with --help flag', async () => {
    // Skip if necessary
    if (shouldSkipCliTests) {
      return;
    }

    const { stdout } = await execa('node', [cliPath, '--help']);
    expect(stdout).toContain('Usage:');
    expect(stdout).toContain('Examples:');
  });

  // Test version flag outputs version information
  it('should show version with --version flag', async () => {
    // Skip if necessary
    if (shouldSkipCliTests) {
      return;
    }

    const { stdout } = await execa('node', [cliPath, '--version']);
    // Version should match semver pattern
    expect(stdout).toMatch(/\d+\.\d+\.\d+/);
  });

  // Tests for context files functionality
  describe('Context Files E2E Tests', () => {
    it('should show context paths in help text', async () => {
      if (shouldSkipCliTests) {
        return;
      }

      const { stdout } = await execa('node', [cliPath, 'run', '--help']);

      // Check that the help text mentions context paths
      expect(stdout).toContain('contextPaths');
      expect(stdout).toContain('files or directories');
    });

    it('should accept a single context file', async () => {
      if (shouldSkipCliTests) {
        return;
      }

      // Create a specific output directory for this test
      const singleOutputDir = path.join(outputDir, 'single-file');
      await fs.mkdir(singleOutputDir, { recursive: true });

      // Run the CLI with a single context file
      // Using --dry-run to prevent actual API calls
      const { stdout, stderr } = await execa('node', [
        cliPath,
        'run',
        promptPath,
        contextFile,
        '--config',
        configPath,
        '--verbose',
        '--dry-run',
        '--output',
        singleOutputDir,
      ]);

      // Check that the output mentions the context file
      expect(stdout).toContain('context.txt');
      expect(stdout).toContain('context file');

      // Verify no errors
      expect(stderr).toBe('');
    });

    it('should handle code files as context', async () => {
      if (shouldSkipCliTests) {
        return;
      }

      // Create a specific output directory for this test
      const codeOutputDir = path.join(outputDir, 'code-file');
      await fs.mkdir(codeOutputDir, { recursive: true });

      // Run the CLI with a code file as context
      const { stdout, stderr } = await execa('node', [
        cliPath,
        'run',
        promptPath,
        codeContextFile,
        '--config',
        configPath,
        '--verbose',
        '--dry-run',
        '--output',
        codeOutputDir,
      ]);

      // Check that the output mentions the context file
      expect(stdout).toContain('example.js');
      expect(stdout).toContain('context file');

      // Verify no errors
      expect(stderr).toBe('');
    });

    it('should handle multiple context files', async () => {
      if (shouldSkipCliTests) {
        return;
      }

      // Create a specific output directory for this test
      const multiOutputDir = path.join(outputDir, 'multiple-files');
      await fs.mkdir(multiOutputDir, { recursive: true });

      // Run the CLI with multiple context files
      const { stdout, stderr } = await execa('node', [
        cliPath,
        'run',
        promptPath,
        ...multipleContextFiles,
        '--config',
        configPath,
        '--verbose',
        '--dry-run',
        '--output',
        multiOutputDir,
      ]);

      // Check that the output mentions all context files
      expect(stdout).toContain('utils.ts');
      expect(stdout).toContain('config.json');
      expect(stdout).toContain('context files');

      // Verify no errors
      expect(stderr).toBe('');
    });

    it('should handle directory as context', async () => {
      if (shouldSkipCliTests) {
        return;
      }

      // Create a specific output directory for this test
      const dirOutputDir = path.join(outputDir, 'directory');
      await fs.mkdir(dirOutputDir, { recursive: true });

      // Run the CLI with a directory as context
      const { stdout, stderr } = await execa('node', [
        cliPath,
        'run',
        promptPath,
        contextDir,
        '--config',
        configPath,
        '--verbose',
        '--dry-run',
        '--output',
        dirOutputDir,
      ]);

      // Check that the output mentions the context directory
      expect(stdout).toContain('context files');

      // Verify no errors
      expect(stderr).toBe('');
    });

    it('should handle file paths with spaces', async () => {
      if (shouldSkipCliTests) {
        return;
      }

      // Create a specific output directory for this test
      const spacesOutputDir = path.join(outputDir, 'spaces');
      await fs.mkdir(spacesOutputDir, { recursive: true });

      // Run the CLI with a file that has spaces in its name
      const { stdout, stderr } = await execa('node', [
        cliPath,
        'run',
        promptPath,
        fileWithSpaces,
        '--config',
        configPath,
        '--verbose',
        '--dry-run',
        '--output',
        spacesOutputDir,
      ]);

      // Check that the output mentions the context file
      expect(stdout).toContain('context file');

      // Verify no errors
      expect(stderr).toBe('');
    });

    it('should handle mixed files and directories as context', async () => {
      if (shouldSkipCliTests) {
        return;
      }

      // Create a specific output directory for this test
      const mixedOutputDir = path.join(outputDir, 'mixed');
      await fs.mkdir(mixedOutputDir, { recursive: true });

      // Run the CLI with a mix of files and directories
      const { stdout, stderr } = await execa('node', [
        cliPath,
        'run',
        promptPath,
        contextFile,
        contextDir,
        '--config',
        configPath,
        '--verbose',
        '--dry-run',
        '--output',
        mixedOutputDir,
      ]);

      // Check that the output mentions the context
      expect(stdout).toContain('context files');

      // Verify no errors
      expect(stderr).toBe('');
    });

    it('should handle non-existent context paths gracefully', async () => {
      if (shouldSkipCliTests) {
        return;
      }

      const nonExistentPath = path.join(tempDir, 'non-existent-file.txt');

      // Create a specific output directory for this test
      const nonExistentOutputDir = path.join(outputDir, 'non-existent');
      await fs.mkdir(nonExistentOutputDir, { recursive: true });

      // Run the CLI with a non-existent file and expect it to fail
      await expect(
        execa('node', [
          cliPath,
          'run',
          promptPath,
          nonExistentPath,
          '--config',
          configPath,
          '--verbose',
          '--dry-run',
          '--output',
          nonExistentOutputDir,
        ])
      ).rejects.toMatchObject({
        stderr: expect.stringContaining('not found'),
        exitCode: expect.anything(),
      });
    });

    it('should handle a valid context file along with a non-existent one', async () => {
      if (shouldSkipCliTests) {
        return;
      }

      const nonExistentPath = path.join(tempDir, 'non-existent-file.txt');

      // Create a specific output directory for this test
      const mixedValidInvalidDir = path.join(outputDir, 'mixed-valid-invalid');
      await fs.mkdir(mixedValidInvalidDir, { recursive: true });

      // Run the CLI with one valid and one non-existent file
      const { stdout } = await execa('node', [
        cliPath,
        'run',
        promptPath,
        contextFile,
        nonExistentPath,
        '--config',
        configPath,
        '--verbose',
        '--dry-run',
        '--output',
        mixedValidInvalidDir,
      ]);

      // When one file exists and one doesn't, the CLI should continue with a warning
      expect(stdout).toContain('context.txt');
      expect(stdout).toContain('could not be read');
    });
  });
});
