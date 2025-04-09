import {
  configureSpinnerFactory,
  createSpinner,
  resetSpinnerFactoryConfig,
  getSpinnerFactoryConfig,
} from '../spinnerFactory';
import { ThrottledSpinner } from '../throttledSpinner';
import originalOra from 'ora';

// Mock original ora
jest.mock('ora', () => {
  return jest.fn().mockImplementation(() => ({
    start: jest.fn().mockReturnThis(),
    stop: jest.fn().mockReturnThis(),
    succeed: jest.fn().mockReturnThis(),
    fail: jest.fn().mockReturnThis(),
    warn: jest.fn().mockReturnThis(),
    info: jest.fn().mockReturnThis(),
    text: '',
    isSpinning: true,
  }));
});

describe('spinnerFactory', () => {
  beforeEach(() => {
    resetSpinnerFactoryConfig();
    jest.clearAllMocks();
  });

  test('should create ThrottledSpinner by default', () => {
    const spinner = createSpinner('Test spinner');
    expect(spinner).toBeInstanceOf(ThrottledSpinner);
  });

  test('should create original Ora spinner when configured', () => {
    configureSpinnerFactory({ useThrottledSpinner: false });
    createSpinner('Test spinner');
    expect(originalOra).toHaveBeenCalledWith('Test spinner');
  });

  test('should handle object options', () => {
    const spinner = createSpinner({ initialText: 'Test with options' });
    expect(spinner).toBeInstanceOf(ThrottledSpinner);
  });

  test('should reset configuration to defaults', () => {
    configureSpinnerFactory({ useThrottledSpinner: false, defaultThrottleInterval: 500 });
    resetSpinnerFactoryConfig();

    const config = getSpinnerFactoryConfig();
    expect(config.useThrottledSpinner).toBe(true);
    expect(config.defaultThrottleInterval).toBe(200);
  });
});
