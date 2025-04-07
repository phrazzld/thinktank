import { ThrottledSpinner } from '../throttledSpinner';
import ora from 'ora';

// Mock ora spinner with a more robust implementation
const mockOraInstance = {
  start: jest.fn().mockReturnThis(),
  stop: jest.fn().mockReturnThis(),
  succeed: jest.fn().mockReturnThis(),
  fail: jest.fn().mockReturnThis(),
  warn: jest.fn().mockReturnThis(),
  info: jest.fn().mockReturnThis(),
  text: 'Initial text',
  isSpinning: true
};

jest.mock('ora', () => {
  return jest.fn().mockImplementation(() => mockOraInstance);
});

describe('ThrottledSpinner', () => {
  let throttledSpinner: ThrottledSpinner;
  
  beforeEach(() => {
    // Setup fresh mocks and timers for each test
    jest.useFakeTimers();
    jest.clearAllMocks();
    // Reset the mock ora instance's text
    mockOraInstance.text = 'Initial text';
    mockOraInstance.isSpinning = true;
    // Create a new spinner instance
    throttledSpinner = new ThrottledSpinner({ initialText: 'Initial text' });
  });
  
  afterEach(() => {
    // Clean up timers
    jest.runOnlyPendingTimers();
    jest.useRealTimers();
  });
  
  test('should create a throttled spinner with initial text', () => {
    expect(throttledSpinner).toBeDefined();
    expect(ora).toHaveBeenCalledWith('Initial text');
  });
  
  test('should update internal text immediately but throttle visible updates', () => {
    // Start the spinner
    throttledSpinner.start();
    
    // Update text multiple times in quick succession
    throttledSpinner.setText('Update 1');
    throttledSpinner.setText('Update 2');
    throttledSpinner.setText('Update 3');
    
    // Internal text should be the latest update
    expect(throttledSpinner.getCurrentText()).toBe('Update 3');
    
    // But the visible ora spinner text shouldn't have changed yet due to throttling
    expect(mockOraInstance.text).toBe('Initial text');
    
    // Fast-forward past throttle interval
    jest.advanceTimersByTime(200); // Default throttle interval
    
    // After advancing time, the ora spinner text should be updated
    expect(mockOraInstance.text).toBe('Update 3');
    
    // Properly clean up
    throttledSpinner.stop();
  });
  
  test('should allow critical updates to bypass throttling', () => {
    // Start the spinner
    throttledSpinner.start();
    
    // Make a normal update
    throttledSpinner.setText('Normal update');
    
    // This update should be throttled, so ora text shouldn't change immediately
    expect(mockOraInstance.text).toBe('Initial text');
    
    // Make a critical update
    throttledSpinner.setText('Critical update', true);
    
    // Critical updates should bypass throttling and update immediately
    expect(mockOraInstance.text).toBe('Critical update');
    
    // Properly clean up
    throttledSpinner.stop();
  });
  
  test('should handle methods like start and stop', () => {
    // Check start
    throttledSpinner.start();
    expect(mockOraInstance.start).toHaveBeenCalled();
    
    // Check stop
    throttledSpinner.stop();
    expect(mockOraInstance.stop).toHaveBeenCalled();
    
    // Check info
    throttledSpinner.info('Info');
    expect(mockOraInstance.info).toHaveBeenCalledWith('Info');
    
    // Check warn
    throttledSpinner.warn('Warning');
    expect(mockOraInstance.warn).toHaveBeenCalledWith('Warning');
    
    // Check succeed
    throttledSpinner.succeed('Success');
    expect(mockOraInstance.succeed).toHaveBeenCalledWith('Success');
    
    // Check fail
    throttledSpinner.fail('Failure');
    expect(mockOraInstance.fail).toHaveBeenCalledWith('Failure');
  });
  
  test('should clear pending updates when stop is called', () => {
    // Start the spinner
    throttledSpinner.start();
    
    // Update text which should schedule a throttled update
    throttledSpinner.setText('Update before stop');
    
    // There should be a pending update
    expect(throttledSpinner.getCurrentText()).toBe('Update before stop');
    expect(mockOraInstance.text).toBe('Initial text'); // Not updated yet due to throttling
    
    // Stop the spinner, which should clear pending updates
    throttledSpinner.stop();
    
    // Run pending timers to verify no updates occur
    jest.runAllTimers();
    
    // ora.text should still be the initial value
    expect(mockOraInstance.text).toBe('Initial text');
    
    // Start spinner again
    throttledSpinner.start();
    
    // Set new text
    throttledSpinner.setText('New update');
    
    // Fast-forward timer
    jest.advanceTimersByTime(200);
    
    // Now the text should be updated
    expect(mockOraInstance.text).toBe('New update');
    
    // Clean up
    throttledSpinner.stop();
  });
});
