package logutil

import (
	"testing"
	"time"
)

func TestModelStatusTracker(t *testing.T) {
	modelNames := []string{"model1", "model2", "model3"}
	tracker := NewModelStatusTracker(modelNames)

	// Test initial state
	if tracker.totalModels != 3 {
		t.Errorf("Expected 3 total models, got %d", tracker.totalModels)
	}

	states := tracker.GetAllStates()
	if len(states) != 3 {
		t.Errorf("Expected 3 states, got %d", len(states))
	}

	// Verify initial status is queued
	for i, state := range states {
		if state.Status != StatusQueued {
			t.Errorf("Expected status queued for model %d, got %s", i, state.Status)
		}
		if state.Name != modelNames[i] {
			t.Errorf("Expected model name %s, got %s", modelNames[i], state.Name)
		}
		if state.Index != i+1 {
			t.Errorf("Expected index %d, got %d", i+1, state.Index)
		}
	}

	// Test status updates
	tracker.UpdateStatus("model1", StatusProcessing, 0, "")
	tracker.UpdateStatus("model2", StatusCompleted, 100*time.Millisecond, "")
	tracker.UpdateStatus("model3", StatusFailed, 50*time.Millisecond, "test error")

	states = tracker.GetAllStates()
	if states[0].Status != StatusProcessing {
		t.Errorf("Expected model1 to be processing, got %s", states[0].Status)
	}
	if states[1].Status != StatusCompleted {
		t.Errorf("Expected model2 to be completed, got %s", states[1].Status)
	}
	if states[2].Status != StatusFailed || states[2].ErrorMsg != "test error" {
		t.Errorf("Expected model3 to be failed with error message, got %s with '%s'", states[2].Status, states[2].ErrorMsg)
	}

	// Test rate limited status
	tracker.UpdateRateLimited("model1", 2*time.Second)
	states = tracker.GetAllStates()
	if states[0].Status != StatusRateLimited || states[0].RetryAfter != 2*time.Second {
		t.Errorf("Expected model1 to be rate limited with 2s retry, got %s with %v", states[0].Status, states[0].RetryAfter)
	}

	// Test completion check
	if tracker.IsAllComplete() {
		t.Error("Expected IsAllComplete to be false with one model still processing")
	}

	tracker.UpdateStatus("model1", StatusCompleted, 200*time.Millisecond, "")
	if !tracker.IsAllComplete() {
		t.Error("Expected IsAllComplete to be true with all models completed or failed")
	}

	// Test summary
	summary := tracker.GetSummary()
	if summary.TotalModels != 3 {
		t.Errorf("Expected 3 total models in summary, got %d", summary.TotalModels)
	}
	if summary.CompletedCount != 2 {
		t.Errorf("Expected 2 completed models in summary, got %d", summary.CompletedCount)
	}
	if summary.FailedCount != 1 {
		t.Errorf("Expected 1 failed model in summary, got %d", summary.FailedCount)
	}

	// Test completion rate
	completionRate := summary.GetCompletionRate()
	if completionRate != 100.0 {
		t.Errorf("Expected 100%% completion rate, got %.1f%%", completionRate)
	}

	// Test success rate
	successRate := summary.GetSuccessRate()
	expectedSuccessRate := float64(2) / float64(3) * 100 // 2 completed out of 3 finished
	if successRate != expectedSuccessRate {
		t.Errorf("Expected %.1f%% success rate, got %.1f%%", expectedSuccessRate, successRate)
	}
}

func TestStatusSummary_EdgeCases(t *testing.T) {
	// Test empty summary
	summary := StatusSummary{}
	if summary.GetCompletionRate() != 0 {
		t.Error("Expected 0% completion rate for empty summary")
	}
	if summary.GetSuccessRate() != 0 {
		t.Error("Expected 0% success rate for empty summary")
	}

	// Test summary with no finished models
	summary = StatusSummary{
		TotalModels:     3,
		ProcessingCount: 2,
		QueuedCount:     1,
	}
	if summary.GetCompletionRate() != 0 {
		t.Error("Expected 0% completion rate with no finished models")
	}
	if summary.GetSuccessRate() != 0 {
		t.Error("Expected 0% success rate with no finished models")
	}
}
