/**
 * Tests to verify the interface definitions can be implemented and used correctly
 */
import { FileSystem, LLMClient, ConsoleLogger, UISpinner, ConfigManagerInterface } from '../interfaces';
import { Stats } from 'fs';
import type { AppConfig, LLMResponse, ModelConfig } from '../types';

describe('Interface Definitions', () => {
  describe('FileSystem', () => {
    it('can be implemented as a mock', () => {
      // Create a simple mock implementation of the FileSystem interface
      const mockFileSystem: FileSystem = {
        readFileContent: jest.fn().mockResolvedValue('file content'),
        writeFile: jest.fn().mockResolvedValue(undefined),
        fileExists: jest.fn().mockResolvedValue(true),
        mkdir: jest.fn().mockResolvedValue(undefined),
        readdir: jest.fn().mockResolvedValue(['file1', 'file2']),
        stat: jest.fn().mockResolvedValue({ isFile: () => true } as Stats),
        access: jest.fn().mockResolvedValue(undefined),
        getConfigDir: jest.fn().mockResolvedValue('/home/user/.config/thinktank'),
        getConfigFilePath: jest.fn().mockResolvedValue('/home/user/.config/thinktank/config.json'),
      };

      // Use the mock implementation
      void expect(mockFileSystem.readFileContent('/test.txt')).resolves.toBe('file content');
      void expect(mockFileSystem.fileExists('/test.txt')).resolves.toBe(true);
    });
  });

  describe('LLMClient', () => {
    it('can be implemented as a mock', () => {
      // Mock response
      const mockResponse: LLMResponse = {
        provider: 'openai',
        modelId: 'gpt-4',
        text: 'Generated response',
        metadata: {
          usage: {
            promptTokens: 10,
            completionTokens: 20,
            totalTokens: 30,
          },
        },
      };

      // Create a mock implementation of the LLMClient interface
      const mockLLMClient: LLMClient = {
        generate: jest.fn().mockResolvedValue(mockResponse),
      };

      // Use the mock implementation
      void expect(mockLLMClient.generate('Test prompt', 'openai:gpt-4')).resolves.toEqual(mockResponse);
    });
  });

  describe('ConsoleLogger', () => {
    it('can be implemented as a mock', () => {
      // Create a simple mock implementation of the ConsoleLogger interface
      const mockLogger: ConsoleLogger = {
        error: jest.fn(),
        warn: jest.fn(),
        info: jest.fn(),
        success: jest.fn(),
        debug: jest.fn(),
        plain: jest.fn(),
      };

      // Use the mock implementation
      mockLogger.info('Test info message');
      mockLogger.error('Test error message', new Error('Test error'));
      mockLogger.warn('Test warning message');
      mockLogger.success('Test success message');
      mockLogger.debug('Test debug message');
      mockLogger.plain('Test plain message');

      // Verify the methods were called with correct arguments
      expect(mockLogger.info).toHaveBeenCalledWith('Test info message');
      expect(mockLogger.error).toHaveBeenCalledWith('Test error message', expect.any(Error));
      expect(mockLogger.warn).toHaveBeenCalledWith('Test warning message');
      expect(mockLogger.success).toHaveBeenCalledWith('Test success message');
      expect(mockLogger.debug).toHaveBeenCalledWith('Test debug message');
      expect(mockLogger.plain).toHaveBeenCalledWith('Test plain message');
    });

    it('can be used for dependency injection', () => {
      // Create a function that uses the ConsoleLogger interface
      function logMessage(logger: ConsoleLogger, level: string, message: string): void {
        switch (level) {
          case 'info':
            logger.info(message);
            break;
          case 'error':
            logger.error(message);
            break;
          case 'warn':
            logger.warn(message);
            break;
          case 'success':
            logger.success(message);
            break;
          case 'debug':
            logger.debug(message);
            break;
          default:
            logger.plain(message);
        }
      }

      // Create a mock logger
      const mockLogger: ConsoleLogger = {
        error: jest.fn(),
        warn: jest.fn(),
        info: jest.fn(),
        success: jest.fn(),
        debug: jest.fn(),
        plain: jest.fn(),
      };

      // Use the function with the mock logger
      logMessage(mockLogger, 'info', 'Test info message');

      // Verify the mock was called correctly
      expect(mockLogger.info).toHaveBeenCalledWith('Test info message');
    });
  });

  describe('UISpinner', () => {
    it('can be implemented as a mock with chaining methods', () => {
      // Create a mock implementation of the UISpinner interface
      const mockSpinner: UISpinner = {
        start: jest.fn().mockReturnThis(),
        stop: jest.fn().mockReturnThis(),
        succeed: jest.fn().mockReturnThis(),
        fail: jest.fn().mockReturnThis(),
        warn: jest.fn().mockReturnThis(),
        info: jest.fn().mockReturnThis(),
        setText: jest.fn().mockReturnThis(),
        text: '',
        isSpinning: false,
      };

      // Use the mock implementation with chaining
      mockSpinner
        .start('Starting process')
        .setText('Processing item 1')
        .setText('Processing item 2')
        .succeed('Process completed');

      // Verify the methods were called with correct arguments
      expect(mockSpinner.start).toHaveBeenCalledWith('Starting process');
      expect(mockSpinner.setText).toHaveBeenCalledWith('Processing item 1');
      expect(mockSpinner.setText).toHaveBeenCalledWith('Processing item 2');
      expect(mockSpinner.succeed).toHaveBeenCalledWith('Process completed');
    });

    it('can be used with optional methods', () => {
      // Create a mock implementation with optional methods
      const enhancedMockSpinner: UISpinner = {
        start: jest.fn().mockReturnThis(),
        stop: jest.fn().mockReturnThis(),
        succeed: jest.fn().mockReturnThis(),
        fail: jest.fn().mockReturnThis(),
        warn: jest.fn().mockReturnThis(),
        info: jest.fn().mockReturnThis(),
        setText: jest.fn().mockReturnThis(),
        text: '',
        isSpinning: false,
        updateForModelStatus: jest.fn().mockReturnThis(),
        updateForModelSummary: jest.fn().mockReturnThis(),
      };

      // Optional methods can be called safely
      if (enhancedMockSpinner.updateForModelStatus) {
        enhancedMockSpinner.updateForModelStatus('openai:gpt-4', { status: 'running' });
      }

      if (enhancedMockSpinner.updateForModelSummary) {
        enhancedMockSpinner.updateForModelSummary(2, 0);
      }

      // Verify methods were called
      expect(enhancedMockSpinner.updateForModelStatus).toHaveBeenCalledWith(
        'openai:gpt-4', 
        { status: 'running' }
      );
      
      expect(enhancedMockSpinner.updateForModelSummary).toHaveBeenCalledWith(2, 0);
    });

    it('can be used for dependency injection', () => {
      // Create a function that uses the UISpinner interface
      function runProcessWithProgress(spinner: UISpinner, items: string[]): void {
        spinner.start('Starting process');

        items.forEach(item => {
          spinner.setText(`Processing ${item}`);
          // Process the item...
        });

        spinner.succeed('Process completed');
      }

      // Create a mock spinner
      const mockSpinner: UISpinner = {
        start: jest.fn().mockReturnThis(),
        stop: jest.fn().mockReturnThis(),
        succeed: jest.fn().mockReturnThis(),
        fail: jest.fn().mockReturnThis(),
        warn: jest.fn().mockReturnThis(),
        info: jest.fn().mockReturnThis(),
        setText: jest.fn().mockReturnThis(),
        text: '',
        isSpinning: false,
      };

      // Use the function with the mock spinner
      runProcessWithProgress(mockSpinner, ['item1', 'item2']);

      // Verify the mock was called correctly
      expect(mockSpinner.start).toHaveBeenCalledWith('Starting process');
      expect(mockSpinner.setText).toHaveBeenCalledWith('Processing item1');
      expect(mockSpinner.setText).toHaveBeenCalledWith('Processing item2');
      expect(mockSpinner.succeed).toHaveBeenCalledWith('Process completed');
    });
  });

  describe('ConfigManagerInterface', () => {
    it('can be implemented as a mock', () => {
      // Mock config
      const mockConfig: AppConfig = {
        models: [],
        groups: {},
      };

      // Create a mock implementation of the ConfigManagerInterface
      const mockConfigManager: ConfigManagerInterface = {
        loadConfig: jest.fn().mockResolvedValue(mockConfig),
        saveConfig: jest.fn().mockResolvedValue(undefined),
        getActiveConfigPath: jest.fn().mockResolvedValue('/home/user/.config/thinktank/config.json'),
        getDefaultConfigPath: jest.fn().mockReturnValue('./thinktank.config.json'),
        addOrUpdateModel: jest.fn().mockReturnValue(mockConfig),
        removeModel: jest.fn().mockReturnValue(mockConfig),
        addOrUpdateGroup: jest.fn().mockReturnValue(mockConfig),
        removeGroup: jest.fn().mockReturnValue(mockConfig),
        addModelToGroup: jest.fn().mockReturnValue(mockConfig),
        removeModelFromGroup: jest.fn().mockReturnValue(mockConfig),
      };

      // Use the mock implementation
      void expect(mockConfigManager.loadConfig()).resolves.toEqual(mockConfig);
      expect(mockConfigManager.getDefaultConfigPath()).toBe('./thinktank.config.json');

      // Test newly added methods
      const newModel: ModelConfig = { 
        provider: 'openai',
        modelId: 'gpt-4',
        enabled: true
      };
      
      expect(mockConfigManager.addOrUpdateModel(mockConfig, newModel)).toEqual(mockConfig);
      expect(mockConfigManager.addOrUpdateGroup(mockConfig, 'test-group', { models: [] })).toEqual(mockConfig);
    });
  });
});
