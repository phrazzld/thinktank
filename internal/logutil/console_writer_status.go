package logutil

import (
	"time"
)

// Status Tracking Implementation
// This file contains the status tracking methods for ConsoleWriter

// StartStatusTracking initializes status tracking for the given models
func (c *consoleWriter) StartStatusTracking(modelNames []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.quiet || c.noProgress {
		return
	}

	// Initialize tracking components
	c.statusTracker = NewModelStatusTracker(modelNames)
	c.statusDisplay = NewStatusDisplay(c.isInteractive)
	c.usingStatus = true

	// Display header
	c.statusDisplay.RenderSummaryHeader(len(modelNames))

	// Initialize all models and render initial state
	states := c.statusTracker.GetAllStates()
	c.statusDisplay.RenderStatus(states, true)
}

// UpdateModelStatus updates the status of a specific model in-place
func (c *consoleWriter) UpdateModelStatus(modelName string, status ModelStatus, duration time.Duration, errorMsg string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.usingStatus || c.quiet {
		return
	}

	// Update the tracker
	c.statusTracker.UpdateStatus(modelName, status, duration, errorMsg)

	// Render updated status
	states := c.statusTracker.GetAllStates()
	c.statusDisplay.RenderStatus(states, false)
}

// UpdateModelRateLimited updates a model's status to show rate limiting
func (c *consoleWriter) UpdateModelRateLimited(modelName string, retryAfter time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.usingStatus || c.quiet {
		return
	}

	// Update the tracker
	c.statusTracker.UpdateRateLimited(modelName, retryAfter)

	// Render updated status
	states := c.statusTracker.GetAllStates()
	c.statusDisplay.RenderStatus(states, false)
}

// RefreshStatusDisplay forces a refresh of the status display
func (c *consoleWriter) RefreshStatusDisplay() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.usingStatus || c.quiet {
		return
	}

	states := c.statusTracker.GetAllStates()
	summary := c.statusTracker.GetSummary()

	if c.isInteractive {
		c.statusDisplay.RenderStatus(states, true)
	} else {
		c.statusDisplay.RenderPeriodicUpdate(states, summary)
	}
}

// FinishStatusTracking completes status tracking and cleans up the display
func (c *consoleWriter) FinishStatusTracking() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.usingStatus {
		return
	}

	// Clean up the display
	if c.statusDisplay != nil {
		c.statusDisplay.RenderCompletion()
	}

	// Reset tracking state
	c.statusTracker = nil
	c.statusDisplay = nil
	c.usingStatus = false
}
