/**
 * Workflow test environment utilities
 * 
 * This module provides a centralized setup helper for workflow integration tests,
 * creating and configuring all common mocks needed for testing workflow components.
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
import { AppConfig } from '../../src/core/types';

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