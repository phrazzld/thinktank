/**
 * ThrottledSpinner
 * 
 * A wrapper around the ora spinner that throttles text updates 
 * to reduce terminal flicker while maintaining a responsive UI.
 */

import ora, { Ora } from 'ora';

/**
 * Options for creating a ThrottledSpinner
 */
export interface ThrottledSpinnerOptions {
  /** Initial text to display in the spinner */
  initialText?: string;
  /** Throttle interval in milliseconds (default: 200ms) */
  throttleInterval?: number;
}

/**
 * A wrapper for the ora spinner that provides throttled text updates
 * to reduce terminal flicker while maintaining responsiveness.
 */
export class ThrottledSpinner {
  /** The underlying ora spinner instance */
  private spinner: Ora;
  /** The current text that should be displayed (may differ from visible text due to throttling) */
  private currentText: string;
  /** Throttle interval in milliseconds */
  private throttleInterval: number;
  /** Timer ID for the next scheduled update */
  private updateTimer: NodeJS.Timeout | null = null;
  /** Flag indicating if an update is scheduled */
  private updatePending: boolean = false;
  /** Last time the visible spinner text was updated */
  private lastUpdateTime: number = 0;

  /**
   * Creates a new ThrottledSpinner
   * 
   * @param options Configuration options for the throttled spinner
   */
  constructor(options: ThrottledSpinnerOptions = {}) {
    const { initialText = '', throttleInterval = 200 } = options;
    this.spinner = ora(initialText);
    this.currentText = initialText;
    this.throttleInterval = throttleInterval;
    this.lastUpdateTime = Date.now();
  }

  /**
   * Get the current text (the text that should be displayed,
   * not necessarily what is currently visible)
   * 
   * @returns The current spinner text
   */
  public getCurrentText(): string {
    return this.currentText;
  }

  /**
   * Updates the spinner text with throttling to reduce flicker
   * 
   * @param text The new text to display
   * @param critical Whether this update should bypass throttling (for important messages)
   */
  public setText(text: string, critical: boolean = false): ThrottledSpinner {
    // Always update the internal state
    this.currentText = text;

    // For critical updates, bypass throttling and update immediately
    if (critical) {
      this.spinner.text = text;
      this.lastUpdateTime = Date.now();
      // Clear any pending updates as we've just updated
      if (this.updateTimer) {
        clearTimeout(this.updateTimer);
        this.updateTimer = null;
        this.updatePending = false;
      }
      return this;
    }

    // If no update is pending, schedule one
    if (!this.updatePending) {
      const timeSinceLastUpdate = Date.now() - this.lastUpdateTime;
      const timeToNextUpdate = Math.max(0, this.throttleInterval - timeSinceLastUpdate);

      this.updatePending = true;
      this.updateTimer = setTimeout(() => {
        this.spinner.text = this.currentText;
        this.lastUpdateTime = Date.now();
        this.updatePending = false;
        this.updateTimer = null;
      }, timeToNextUpdate);
    }

    return this;
  }

  /**
   * Starts the spinner
   */
  public start(): ThrottledSpinner {
    this.spinner.start();
    return this;
  }

  /**
   * Stops the spinner
   */
  public stop(): ThrottledSpinner {
    // Clear any pending updates
    if (this.updateTimer) {
      clearTimeout(this.updateTimer);
      this.updateTimer = null;
      this.updatePending = false;
    }
    this.spinner.stop();
    return this;
  }

  /**
   * Shows a success message
   */
  public succeed(text?: string): ThrottledSpinner {
    // Clear any pending updates
    if (this.updateTimer) {
      clearTimeout(this.updateTimer);
      this.updateTimer = null;
      this.updatePending = false;
    }
    this.spinner.succeed(text);
    return this;
  }

  /**
   * Shows a failure message
   */
  public fail(text?: string): ThrottledSpinner {
    // Clear any pending updates
    if (this.updateTimer) {
      clearTimeout(this.updateTimer);
      this.updateTimer = null;
      this.updatePending = false;
    }
    this.spinner.fail(text);
    return this;
  }

  /**
   * Shows an info message
   */
  public info(text?: string): ThrottledSpinner {
    // Clear any pending updates
    if (this.updateTimer) {
      clearTimeout(this.updateTimer);
      this.updateTimer = null;
      this.updatePending = false;
    }
    this.spinner.info(text);
    return this;
  }

  /**
   * Shows a warning message
   */
  public warn(text?: string): ThrottledSpinner {
    // Clear any pending updates
    if (this.updateTimer) {
      clearTimeout(this.updateTimer);
      this.updateTimer = null;
      this.updatePending = false;
    }
    this.spinner.warn(text);
    return this;
  }

  /**
   * Sets the spinner text directly (property setter)
   */
  set text(value: string) {
    this.setText(value);
  }

  /**
   * Gets the spinner text directly (property getter)
   */
  get text(): string {
    return this.spinner.text;
  }

  /**
   * Gets whether the spinner is currently spinning
   */
  get isSpinning(): boolean {
    return this.spinner.isSpinning;
  }
}