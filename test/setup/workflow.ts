/**
 * Workflow test environment utilities
 * 
 * This module provides a centralized setup helper for workflow integration tests,
 * creating and configuring all common mocks needed for testing workflow components.
 * 
 * It includes:
 * - Basic mock creation via setupWorkflowTestEnvironment()
 * - Higher-level scenario mock configuration helpers for common test cases
 */
import { 
  FileSystem, 
  ConsoleLogger, 
  UISpinner, 
  ConfigManagerInterface,
  LLMClient 
} from '../../src/core/interfaces';
import { createMockFileSystem } from './fs';
import { createMockConsoleLogger, createMockUISpinner } from './io';
import { AppConfig, LLMResponse } from '../../src/core/types';
import { 
  createAppConfig, 
  createLlmResponse 
} from '../factories';

/**
 * Creates a mock ConfigManager implementation for testing
 * 
 * @param defaultConfig - Optional default configuration to return from loadConfig
 * @returns A Jest-mocked ConfigManagerInterface
 */
export function createMockConfigManager(
  defaultConfig: Partial<AppConfig> = { models: [], groups: {} }
): jest.Mocked<ConfigManagerInterface> {
  return {
    loadConfig: jest.fn().mockResolvedValue(defaultConfig as AppConfig),
    saveConfig: jest.fn().mockResolvedValue(undefined),
    mergeConfig: jest.fn().mockImplementation((base, override) => ({
      ...base,
      ...override
    })),
    getConfigValue: jest.fn().mockImplementation((config, key) => {
      // Safely navigate the object using the dot-notation key
      return key.split('.').reduce((obj: any, k: string) => (obj && obj[k] !== undefined) ? obj[k] : undefined, config);
    }),
    getDefaultConfig: jest.fn().mockResolvedValue(defaultConfig as AppConfig)
  } as unknown as jest.Mocked<ConfigManagerInterface>;
}

/**
 * Creates a mock LLMClient implementation for testing
 * 
 * @returns A Jest-mocked LLMClient
 */
export function createMockLLMClient(): jest.Mocked<LLMClient> {
  return {
    generate: jest.fn().mockResolvedValue({
      provider: 'mock',
      modelId: 'mock-model',
      text: 'This is a mock response',
      metadata: {
        responseTime: 1000,
        tokenUsage: {
          input: 100,
          output: 200,
          total: 300
        }
      }
    }),
    isProviderAvailable: jest.fn().mockReturnValue(true),
    getAvailableProviders: jest.fn().mockReturnValue(['mock'])
  } as unknown as jest.Mocked<LLMClient>;
}

/**
 * Interface for all workflow test mocks
 */
export interface WorkflowTestMocks {
  mockFileSystem: jest.Mocked<FileSystem>;
  mockLogger: jest.Mocked<ConsoleLogger>;
  mockSpinner: jest.Mocked<UISpinner>;
  mockConfigManager: jest.Mocked<ConfigManagerInterface>;
  mockLLMClient: jest.Mocked<LLMClient>;
}

/**
 * Sets up a standardized test environment for workflow integration tests
 * 
 * This helper creates and configures all commonly needed mocks for testing
 * workflow components, reducing duplication across test files and ensuring
 * consistent testing patterns.
 * 
 * @returns Object containing all mocked dependencies
 * 
 * Usage:
 * ```typescript
 * describe('Workflow Test', () => {
 *   setupTestHooks(); // For basic setup/teardown
 *   
 *   let mocks: WorkflowTestMocks;
 *   
 *   beforeEach(() => {
 *     mocks = setupWorkflowTestEnvironment();
 *     // Configure specific mock behaviors for this test:
 *     mocks.mockConfigManager.loadConfig.mockResolvedValue({ ... });
 *   });
 *   
 *   it('should process workflow correctly', async () => {
 *     // Call function under test, passing mocks:
 *     await runThinktank({ input: 'test.txt' }, mocks.mockFileSystem, ...);
 *     
 *     // Assert on interactions with mocks:
 *     expect(mocks.mockFileSystem.writeFile).toHaveBeenCalledWith(
 *       expect.stringContaining('.md'),
 *       expect.stringContaining('mock response')
 *     );
 *   });
 * });
 * ```
 */
export function setupWorkflowTestEnvironment(): WorkflowTestMocks {
  // Create mock instances
  const mockFileSystem = createMockFileSystem();
  const mockLogger = createMockConsoleLogger();
  const mockSpinner = createMockUISpinner();
  const mockConfigManager = createMockConfigManager();
  const mockLLMClient = createMockLLMClient();
  
  // Configure common default behaviors
  mockFileSystem.fileExists.mockResolvedValue(true);
  
  return {
    mockFileSystem,
    mockLogger,
    mockSpinner,
    mockConfigManager,
    mockLLMClient
  };
}

/**
 * Options for configuring the successful run scenario
 */
export interface SuccessfulRunOptions {
  /**
   * App configuration to use for the test
   * If not provided, a default will be used
   */
  config?: AppConfig;
  
  /**
   * Content to return when reading the prompt file
   * If not provided, a default will be used
   */
  promptContent?: string;
  
  /**
   * LLM response to return from the generate method
   * If not provided, a default will be used
   */
  llmResponse?: LLMResponse;
}

/**
 * Configures mocks for a successful workflow run scenario
 * 
 * Sets up all mocks to simulate a complete successful workflow run,
 * including successful config loading, file operations, and LLM API calls.
 * 
 * @param mocks - The WorkflowTestMocks object from setupWorkflowTestEnvironment
 * @param options - Optional configuration for the scenario
 * @returns The same mocks object, configured for this scenario
 * 
 * @example
 * ```typescript
 * const mocks = setupWorkflowTestEnvironment();
 * setupMocksForSuccessfulRun(mocks, {
 *   promptContent: 'Custom prompt'
 * });
 * ```
 */
export function setupMocksForSuccessfulRun(
  mocks: WorkflowTestMocks,
  options: SuccessfulRunOptions = {}
): WorkflowTestMocks {
  // Use defaults from test factories if not provided
  const config = options.config || createAppConfig();
  const promptContent = options.promptContent || 'Default test prompt content';
  const llmResponse = options.llmResponse || createLlmResponse();
  
  // Configure ConfigManager
  mocks.mockConfigManager.loadConfig.mockResolvedValue(config);
  
  // Configure FileSystem
  mocks.mockFileSystem.fileExists.mockResolvedValue(true);
  mocks.mockFileSystem.readFileContent.mockResolvedValue(promptContent);
  mocks.mockFileSystem.mkdir.mockResolvedValue(undefined);
  mocks.mockFileSystem.writeFile.mockResolvedValue(undefined);
  
  // Configure LLMClient
  mocks.mockLLMClient.generate.mockResolvedValue(llmResponse);
  
  // Configure UI elements
  mocks.mockSpinner.succeed.mockReturnThis();
  mocks.mockLogger.success.mockReturnValue(undefined);
  
  return mocks;
}

/**
 * Configures mocks for a workflow scenario where config loading fails
 * 
 * @param mocks - The WorkflowTestMocks object from setupWorkflowTestEnvironment
 * @param error - The error to throw during config loading
 * @returns The same mocks object, configured for this scenario
 * 
 * @example
 * ```typescript
 * const mocks = setupWorkflowTestEnvironment();
 * setupMocksForConfigError(mocks, new Error('Config not found'));
 * ```
 */
export function setupMocksForConfigError(
  mocks: WorkflowTestMocks,
  error: Error = new Error('Failed to load config')
): WorkflowTestMocks {
  mocks.mockConfigManager.loadConfig.mockRejectedValue(error);
  mocks.mockSpinner.fail.mockReturnThis();
  
  return mocks;
}

/**
 * Configures mocks for a workflow scenario where the LLM API call fails
 * 
 * Sets up a scenario where config loads and file operations succeed,
 * but the LLM API call fails with an error.
 * 
 * @param mocks - The WorkflowTestMocks object from setupWorkflowTestEnvironment
 * @param error - The error to throw during the LLM API call
 * @param options - Optional configuration for the successful parts
 * @returns The same mocks object, configured for this scenario
 * 
 * @example
 * ```typescript
 * const mocks = setupWorkflowTestEnvironment();
 * setupMocksForApiError(mocks, new Error('Rate limit exceeded'));
 * ```
 */
export function setupMocksForApiError(
  mocks: WorkflowTestMocks,
  error: Error = new Error('API error'),
  options: Omit<SuccessfulRunOptions, 'llmResponse'> = {}
): WorkflowTestMocks {
  // Configure for successful setup steps
  const config = options.config || createAppConfig();
  const promptContent = options.promptContent || 'Default test prompt content';
  
  // Configure ConfigManager
  mocks.mockConfigManager.loadConfig.mockResolvedValue(config);
  
  // Configure FileSystem for successful operations
  mocks.mockFileSystem.fileExists.mockResolvedValue(true);
  mocks.mockFileSystem.readFileContent.mockResolvedValue(promptContent);
  
  // Configure LLMClient to fail
  mocks.mockLLMClient.generate.mockRejectedValue(error);
  
  // Configure UI elements for failure
  mocks.mockSpinner.fail.mockReturnThis();
  mocks.mockLogger.error.mockReturnValue(undefined);
  
  return mocks;
}

/**
 * Configures mocks for a workflow scenario where file reading fails
 * 
 * Sets up a scenario where config loads successfully, but reading the
 * input file fails with an error.
 * 
 * @param mocks - The WorkflowTestMocks object from setupWorkflowTestEnvironment
 * @param error - The error to throw during file reading
 * @param options - Optional configuration for the successful parts
 * @returns The same mocks object, configured for this scenario
 * 
 * @example
 * ```typescript
 * const mocks = setupWorkflowTestEnvironment();
 * setupMocksForFileReadError(mocks, new Error('File not found'));
 * ```
 */
export function setupMocksForFileReadError(
  mocks: WorkflowTestMocks,
  error: Error = new Error('File read error'),
  options: Omit<SuccessfulRunOptions, 'promptContent'> = {}
): WorkflowTestMocks {
  // Configure for successful config loading
  const config = options.config || createAppConfig();
  
  // Configure ConfigManager
  mocks.mockConfigManager.loadConfig.mockResolvedValue(config);
  
  // Configure FileSystem to fail reading
  mocks.mockFileSystem.fileExists.mockResolvedValue(true); // File exists but can't be read
  mocks.mockFileSystem.readFileContent.mockRejectedValue(error);
  
  // Configure UI elements for failure
  mocks.mockSpinner.fail.mockReturnThis();
  mocks.mockLogger.error.mockReturnValue(undefined);
  
  return mocks;
}