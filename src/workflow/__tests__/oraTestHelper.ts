/**
 * Ora spinner mock for tests
 */
import { Ora } from 'ora';

/**
 * Creates a mock Ora spinner for testing
 */
export function createMockSpinner(): jest.Mocked<Ora> {
  return {
    start: jest.fn().mockReturnThis(),
    stop: jest.fn().mockReturnThis(),
    succeed: jest.fn().mockReturnThis(),
    fail: jest.fn().mockReturnThis(),
    warn: jest.fn().mockReturnThis(),
    info: jest.fn().mockReturnThis(),
    stopAndPersist: jest.fn().mockReturnThis(),
    clear: jest.fn().mockReturnThis(),
    render: jest.fn().mockReturnThis(),
    frame: jest.fn().mockReturnThis(),
    text: '',
    isSpinning: true,
    prefixText: '',
    color: 'white',
    indent: 0,
    spinner: {
      interval: 80,
      frames: ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏']
    }
  } as unknown as jest.Mocked<Ora>;
}
