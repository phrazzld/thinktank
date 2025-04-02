/**
 * End-to-End tests for CLI functionality.
 * These tests use execa to simulate actual CLI usage with command-line arguments.
 */
import execa from 'execa';
import path from 'path';
// Import fs and os only when the commented-out functions are used
// import fs from 'fs/promises';
// import os from 'os';

// Define a proper type for Execa errors
interface ExecaError extends Error {
  stderr?: string;
  stdout?: string;
  exitCode?: number;
  command?: string;
}

// These functions are commented out to avoid TypeScript errors until they're used in actual tests
/*
// Helper to create a temp directory for tests
async function createTempDir() {
  const tempDir = path.join(os.tmpdir(), `thinktank-test-${Date.now()}`);
  await fs.mkdir(tempDir, { recursive: true });
  return tempDir;
}

// Helper to create a minimal test config file in a temp directory
async function createTestConfig(tempDir: string) {
  const configPath = path.join(tempDir, 'thinktank.config.json');
  const mockConfig = {
    models: [
      {
        provider: 'openai',
        modelId: 'gpt-4o',
        enabled: true,
        apiKeyEnvVar: 'OPENAI_API_KEY'
      },
      {
        provider: 'anthropic',
        modelId: 'claude-3-sonnet',
        enabled: true,
        apiKeyEnvVar: 'ANTHROPIC_API_KEY'
      }
    ],
    groups: {
      'default': {
        systemPrompt: { text: 'You are a helpful assistant.' },
        models: [
          {
            provider: 'openai',
            modelId: 'gpt-4o'
          }
        ]
      },
      'coding': {
        systemPrompt: { text: 'You are a coding assistant.' },
        models: [
          {
            provider: 'anthropic',
            modelId: 'claude-3-sonnet'
          }
        ],
        description: 'Models optimized for programming tasks'
      }
    }
  };
  await fs.writeFile(configPath, JSON.stringify(mockConfig, null, 2));
  return configPath;
}

// Create a temporary prompt file for testing
async function createTestPrompt(tempDir: string, content = 'This is a test prompt.') {
  const promptPath = path.join(tempDir, 'test-prompt.txt');
  await fs.writeFile(promptPath, content);
  return promptPath;
}
*/

import fs from 'fs/promises';

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
});