import { ThrottledSpinner } from '../throttledSpinner';
import ora from 'ora';

// Mock ora
jest.mock('ora', () => {
  const mockSpinner = {
    start: jest.fn().mockReturnThis(),
    stop: jest.fn().mockReturnThis(),
    succeed: jest.fn().mockReturnThis(),
    fail: jest.fn().mockReturnThis(),
    warn: jest.fn().mockReturnThis(),
    info: jest.fn().mockReturnThis(),
    text: '',
    isSpinning: true
  };
  
  return jest.fn().mockImplementation(() => mockSpinner);
});

describe('ThrottledSpinner', () => {
  let throttledSpinner: ThrottledSpinner;
  
  beforeEach(() => {
    jest.useFakeTimers();
    throttledSpinner = new ThrottledSpinner({ initialText: 'Initial text' });
  });
  
  afterEach(() => {
    jest.useRealTimers();
    jest.clearAllMocks();
  });
  
  test('should create a throttled spinner with initial text', () => {
    expect(throttledSpinner).toBeDefined();
    expect(ora).toHaveBeenCalledWith('Initial text');
  });
  
  test('should update internal text immediately but throttle visible updates', () => {
    // Update text multiple times in quick succession
    throttledSpinner.setText('Update 1');
    throttledSpinner.setText('Update 2');
    throttledSpinner.setText('Update 3');
    
    // Internal text should be the latest update
    expect(throttledSpinner.getCurrentText()).toBe('Update 3');
    
    // Fast-forward past throttle interval
    jest.advanceTimersByTime(200); // Default throttle interval
  });
  
  test('should allow critical updates to bypass throttling', () => {
    throttledSpinner.setText('Normal update');
    throttledSpinner.setText('Critical update', true);
  });
  
  test('should handle methods like start and stop', () => {
    throttledSpinner.start();
    throttledSpinner.stop();
    throttledSpinner.info('Info');
    throttledSpinner.warn('Warning');
    throttledSpinner.succeed('Success');
    throttledSpinner.fail('Failure');
  });
  
  test('should clear pending updates when stop is called', () => {
    throttledSpinner.setText('Update before stop');
    throttledSpinner.stop();
    
    // Clear any pending timeouts
    jest.runAllTimers();
    
    // Start spinner again
    throttledSpinner.start();
    throttledSpinner.setText('New update');
    
    // Fast-forward timer
    jest.advanceTimersByTime(200);
  });
});