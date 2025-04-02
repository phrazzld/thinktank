/**
 * Test for empty configuration handling
 */
import { loadConfig } from '../configManager';

// Mock the file system functions
jest.mock('../../utils/fileReader', () => ({
  fileExists: jest.fn().mockResolvedValue(false),
  readFileContent: jest.fn(),
  writeFile: jest.fn().mockResolvedValue(undefined),
  getConfigFilePath: jest.fn().mockResolvedValue('/test/xdg/config.json'),
}));

// Mock constants for a clean default config
jest.mock('../constants', () => ({
  DEFAULT_CONFIG: {
    models: [],
    groups: {
      default: {
        name: 'default',
        systemPrompt: { text: 'Test prompt' },
        models: [],
        description: 'Test group'
      }
    }
  },
  DEFAULT_CONFIG_TEMPLATE_PATH: '/test/template.json',
}));

describe('Empty configuration handling', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should create default config when no config exists', async () => {
    // Load config without specifying a path
    await loadConfig();

    // Verify that the config file would have been created
    const writeFile = jest.requireMock('../../utils/fileReader').writeFile;
    expect(writeFile).toHaveBeenCalledWith(
      '/test/xdg/config.json',
      expect.any(String)
    );

    // Verify content of written file includes our mocked data
    const writtenContent = writeFile.mock.calls[0][1];
    expect(writtenContent).toContain('Test prompt');
    expect(writtenContent).toContain('Test group');
  });
});
