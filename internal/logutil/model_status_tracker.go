package logutil

import (
	"sync"
	"time"
)

// ModelStatus represents the current state of a model's processing
type ModelStatus string

const (
	StatusQueued      ModelStatus = "queued"
	StatusStarting    ModelStatus = "starting"
	StatusProcessing  ModelStatus = "processing"
	StatusRateLimited ModelStatus = "rate_limited"
	StatusCompleted   ModelStatus = "completed"
	StatusFailed      ModelStatus = "failed"
)

// ModelState holds the complete state information for a model
type ModelState struct {
	Name        string // Internal model key for lookups and operations
	DisplayName string // User-facing display name (e.g., "provider/model")
	Index       int
	Status      ModelStatus
	Duration    time.Duration
	RetryAfter  time.Duration
	ErrorMsg    string
	StartTime   time.Time
	UpdateTime  time.Time
}

// ModelStatusTracker provides thread-safe tracking of multiple model processing states
type ModelStatusTracker struct {
	mu          sync.RWMutex
	models      map[string]*ModelState
	modelOrder  []string
	totalModels int
	startTime   time.Time
}

// ModelDisplayInfo holds both internal name and display name for a model
type ModelDisplayInfo struct {
	InternalName string // Model key used for internal operations
	DisplayName  string // User-facing display name (e.g., "provider/model")
}

// NewModelStatusTracker creates a new status tracker for the given model information
func NewModelStatusTracker(modelInfos []ModelDisplayInfo) *ModelStatusTracker {
	tracker := &ModelStatusTracker{
		models:      make(map[string]*ModelState),
		modelOrder:  make([]string, len(modelInfos)),
		totalModels: len(modelInfos),
		startTime:   time.Now(),
	}

	// Initialize all models in queued state
	for i, modelInfo := range modelInfos {
		tracker.models[modelInfo.InternalName] = &ModelState{
			Name:        modelInfo.InternalName,
			DisplayName: modelInfo.DisplayName,
			Index:       i + 1, // 1-based indexing for display
			Status:      StatusQueued,
			StartTime:   time.Now(),
			UpdateTime:  time.Now(),
		}
		tracker.modelOrder[i] = modelInfo.InternalName
	}

	return tracker
}

// UpdateStatus updates the status of a specific model
func (t *ModelStatusTracker) UpdateStatus(modelName string, status ModelStatus, duration time.Duration, errorMsg string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if model, exists := t.models[modelName]; exists {
		model.Status = status
		model.Duration = duration
		model.ErrorMsg = errorMsg
		model.UpdateTime = time.Now()
	}
}

// UpdateRateLimited updates a model's status to rate limited with retry information
func (t *ModelStatusTracker) UpdateRateLimited(modelName string, retryAfter time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if model, exists := t.models[modelName]; exists {
		model.Status = StatusRateLimited
		model.RetryAfter = retryAfter
		model.UpdateTime = time.Now()
	}
}

// GetAllStates returns a snapshot of all model states in display order
func (t *ModelStatusTracker) GetAllStates() []*ModelState {
	t.mu.RLock()
	defer t.mu.RUnlock()

	states := make([]*ModelState, len(t.modelOrder))
	for i, modelName := range t.modelOrder {
		// Create a copy to avoid race conditions
		original := t.models[modelName]
		states[i] = &ModelState{
			Name:        original.Name,
			DisplayName: original.DisplayName,
			Index:       original.Index,
			Status:      original.Status,
			Duration:    original.Duration,
			RetryAfter:  original.RetryAfter,
			ErrorMsg:    original.ErrorMsg,
			StartTime:   original.StartTime,
			UpdateTime:  original.UpdateTime,
		}
	}
	return states
}

// GetSummary returns processing summary statistics
func (t *ModelStatusTracker) GetSummary() StatusSummary {
	t.mu.RLock()
	defer t.mu.RUnlock()

	summary := StatusSummary{
		TotalModels: t.totalModels,
		StartTime:   t.startTime,
	}

	for _, model := range t.models {
		switch model.Status {
		case StatusCompleted:
			summary.CompletedCount++
		case StatusFailed:
			summary.FailedCount++
		case StatusRateLimited:
			summary.RateLimitedCount++
		case StatusProcessing:
			summary.ProcessingCount++
		default:
			summary.QueuedCount++
		}
	}

	return summary
}

// IsAllComplete returns true if all models have finished processing (either completed or failed)
func (t *ModelStatusTracker) IsAllComplete() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, model := range t.models {
		if model.Status != StatusCompleted && model.Status != StatusFailed {
			return false
		}
	}
	return true
}

// StatusSummary provides aggregate statistics about model processing
type StatusSummary struct {
	TotalModels      int
	CompletedCount   int
	FailedCount      int
	RateLimitedCount int
	ProcessingCount  int
	QueuedCount      int
	StartTime        time.Time
}

// GetCompletionRate returns the percentage of models that have finished processing
func (s *StatusSummary) GetCompletionRate() float64 {
	finished := s.CompletedCount + s.FailedCount
	if s.TotalModels == 0 {
		return 0
	}
	return float64(finished) / float64(s.TotalModels) * 100
}

// GetSuccessRate returns the percentage of completed models among finished models
func (s *StatusSummary) GetSuccessRate() float64 {
	finished := s.CompletedCount + s.FailedCount
	if finished == 0 {
		return 0
	}
	return float64(s.CompletedCount) / float64(finished) * 100
}
