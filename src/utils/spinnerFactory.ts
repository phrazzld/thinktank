/**
 * Spinner Factory
 * 
 * Provides a configurable factory for creating either regular or throttled spinners.
 */

import originalOra from 'ora';
import { ThrottledSpinner, ThrottledSpinnerOptions } from './throttledSpinner';

/**
 * Configuration for the spinner factory
 */
interface SpinnerFactoryConfig {
  /** Whether to use throttled spinners (true) or regular ora spinners (false) */
  useThrottledSpinner: boolean;
  /** Default throttle interval in milliseconds (when throttling is enabled) */
  defaultThrottleInterval: number;
}

/**
 * Default configuration
 */
const DEFAULT_CONFIG: SpinnerFactoryConfig = {
  useThrottledSpinner: true,
  defaultThrottleInterval: 200
};

/**
 * Current configuration
 */
let currentConfig: SpinnerFactoryConfig = { ...DEFAULT_CONFIG };

/**
 * Updates the spinner factory configuration
 * 
 * @param config - New configuration options
 */
export function configureSpinnerFactory(config: Partial<SpinnerFactoryConfig>): void {
  currentConfig = { ...currentConfig, ...config };
}

/**
 * Creates a spinner with the current factory configuration
 * 
 * @param options - Text or options for the spinner
 * @returns Either an Ora spinner or a ThrottledSpinner, depending on configuration
 */
export function createSpinner(
  options?: string | ThrottledSpinnerOptions | originalOra.Options
): originalOra.Ora | ThrottledSpinner {
  if (currentConfig.useThrottledSpinner) {
    // Create a throttled spinner
    if (typeof options === 'string') {
      return new ThrottledSpinner({
        initialText: options,
        throttleInterval: currentConfig.defaultThrottleInterval
      });
    } else if (options) {
      const throttledOptions: ThrottledSpinnerOptions = {
        initialText: (options as { text?: string }).text || '',
        throttleInterval: currentConfig.defaultThrottleInterval
      };
      return new ThrottledSpinner(throttledOptions);
    }
    return new ThrottledSpinner({ throttleInterval: currentConfig.defaultThrottleInterval });
  } else {
    // Use the original ora spinner
    return originalOra(options as string | originalOra.Options);
  }
}

/**
 * Resets the spinner factory configuration to defaults
 */
export function resetSpinnerFactoryConfig(): void {
  currentConfig = { ...DEFAULT_CONFIG };
}

/**
 * Get the current spinner factory configuration
 * 
 * @returns The current configuration
 */
export function getSpinnerFactoryConfig(): SpinnerFactoryConfig {
  return { ...currentConfig };
}

// Default export that works like ora but uses our factory
export default createSpinner;
