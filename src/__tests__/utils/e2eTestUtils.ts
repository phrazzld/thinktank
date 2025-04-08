/**
 * E2E Test Utilities
 * 
 * Provides helper functions for setting up, managing, and cleaning up resources
 * for end-to-end tests that interact with the real filesystem.
 */
import fs from 'fs/promises';
import path from 'path';
import os from 'os';

/**
 * Creates a temporary directory for testing.
 * The directory is created with a unique name in the system's temp directory.
 * 
 * @param prefix - Optional prefix for the temp directory name
 * @returns Path to the created temporary directory
 */
export async function createTempTestDir(prefix = 'thinktank-test-'): Promise<string> {
  const tempDirBase = path.join(os.tmpdir(), `${prefix}${Date.now()}-${Math.floor(Math.random() * 10000)}`);
  await fs.mkdir(tempDirBase, { recursive: true });
  return tempDirBase;
}

/**
 * Creates a file with the specified content in the given directory.
 * 
 * @param dir - Directory to create the file in
 * @param filename - Name of the file to create
 * @param content - Content to write to the file
 * @returns Path to the created file
 */
export async function createTestFile(dir: string, filename: string, content: string): Promise<string> {
  const filePath = path.join(dir, filename);
  await fs.writeFile(filePath, content, 'utf-8');
  return filePath;
}

/**
 * Creates a directory with multiple files for testing.
 * 
 * @param parentDir - Parent directory where the new directory will be created
 * @param dirName - Name of the directory to create
 * @param files - Map of filenames to content
 * @returns Path to the created directory
 */
export async function createTestDir(
  parentDir: string, 
  dirName: string,
  files: Record<string, string> = {}
): Promise<string> {
  const dirPath = path.join(parentDir, dirName);
  await fs.mkdir(dirPath, { recursive: true });
  
  // Create all files in the directory
  for (const [filename, content] of Object.entries(files)) {
    await createTestFile(dirPath, filename, content);
  }
  
  return dirPath;
}

/**
 * Creates a minimal test config file for thinktank.
 * 
 * @param dir - Directory to create the config file in
 * @param customConfig - Optional custom configuration to use instead of default
 * @param filename - Optional custom filename for the config file (defaults to 'thinktank.config.json')
 * @returns Path to the created config file
 */
export async function createTestConfig(
  dir: string, 
  customConfig?: Record<string, any>,
  filename: string = 'thinktank.config.json'
): Promise<string> {
  const configPath = path.join(dir, filename);
  
  // Default minimal config if no custom config provided
  const config = customConfig || {
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
  
  await fs.writeFile(configPath, JSON.stringify(config, null, 2), 'utf-8');
  return configPath;
}

/**
 * Safely cleans up a temporary test directory.
 * 
 * @param dir - Path to the directory to clean up
 * @param options - Cleanup options
 */
export async function cleanupTestDir(
  dir: string, 
  options: { silent?: boolean; force?: boolean } = {}
): Promise<void> {
  try {
    await fs.rm(dir, { recursive: true, force: options.force !== false });
  } catch (error) {
    // Don't output errors during test cleanup to keep test output clean
    // If silent is false, we could use the test's logger or jest.fn() here
    // But we'll just quietly fail to avoid polluting test output
    // Don't throw - cleanup errors shouldn't fail tests
  }
}

/**
 * Lists all files in a directory and its subdirectories recursively.
 * 
 * @param dirPath - The directory to list files from
 * @returns Array of file paths (absolute)
 */
export async function listFilesRecursive(dirPath: string): Promise<string[]> {
  let result: string[] = [];
  
  try {
    const entries = await fs.readdir(dirPath, { withFileTypes: true });
    
    for (const entry of entries) {
      const fullPath = path.join(dirPath, entry.name);
      
      if (entry.isDirectory()) {
        // Recursively scan subdirectories
        const subDirFiles = await listFilesRecursive(fullPath);
        result = result.concat(subDirFiles);
      } else {
        // Add files to the result
        result.push(fullPath);
      }
    }
  } catch (error) {
    // Silently fail rather than polluting test output with console.error
    // In a real production case we might want to propagate this error
  }
  
  return result;
}

/**
 * Returns a normalized path appropriate for the current platform.
 * Useful for writing platform-agnostic tests.
 * 
 * @param inputPath - Path to normalize
 * @returns Normalized path for the current platform
 */
export function platformPath(inputPath: string): string {
  // Normalize path separators for the current platform
  return path.normalize(inputPath);
}

/**
 * Creates a self-cleanup function for tests to use in afterAll/afterEach.
 * 
 * @param dirsToClean - Array of directory paths to clean up
 * @returns Function that will clean up all specified directories
 */
export function createCleanupFunction(...dirsToClean: string[]): () => Promise<void> {
  return async () => {
    for (const dir of dirsToClean) {
      await cleanupTestDir(dir, { silent: true, force: true });
    }
  };
}

/**
 * Helper function to check if it's safe to run filesystem E2E tests.
 * This helps avoid running test that modify real files in certain environments.
 * 
 * @returns True if E2E tests should be skipped
 */
export function shouldSkipFsE2ETests(): boolean {
  // Skip if explicitly configured to skip
  if (process.env.SKIP_E2E_TESTS === 'true' || process.env.SKIP_FS_TESTS === 'true') {
    return true;
  }
  
  // Skip in CI environments unless explicitly enabled
  if (process.env.CI === 'true' && process.env.ENABLE_FS_TESTS !== 'true') {
    return true;
  }
  
  return false;
}
