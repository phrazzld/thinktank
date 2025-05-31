package orchestrator

import (
	"context"
	"sync"
)

// BaseMockOutputWriter is the base implementation for mock output writers
type BaseMockOutputWriter struct {
	savedCount           int
	saveError            error
	capturedModelOutputs map[string]string
	capturedOutputDir    string

	// mutex for thread-safe access to fields
	mutex sync.RWMutex
}

// SaveIndividualOutputs is a mock implementation that returns configured results
func (m *BaseMockOutputWriter) SaveIndividualOutputs(ctx context.Context, modelOutputs map[string]string, outputDir string) (int, map[string]string, error) {
	// Lock for writing to captured fields
	m.mutex.Lock()
	m.capturedModelOutputs = modelOutputs
	m.capturedOutputDir = outputDir
	m.mutex.Unlock()

	// Create mock file paths map - one path per model
	filePaths := make(map[string]string)
	for modelName := range modelOutputs {
		filePaths[modelName] = outputDir + "/" + modelName + ".md"
	}

	// Read lock for accessing savedCount and saveError
	m.mutex.RLock()
	count := m.savedCount
	err := m.saveError
	m.mutex.RUnlock()

	return count, filePaths, err
}

// SetSavedCount sets the saved count in a thread-safe manner
func (m *BaseMockOutputWriter) SetSavedCount(count int) {
	m.mutex.Lock()
	m.savedCount = count
	m.mutex.Unlock()
}

// SetSaveError sets the save error in a thread-safe manner
func (m *BaseMockOutputWriter) SetSaveError(err error) {
	m.mutex.Lock()
	m.saveError = err
	m.mutex.Unlock()
}

// GetCapturedModelOutputs returns the captured model outputs in a thread-safe manner
func (m *BaseMockOutputWriter) GetCapturedModelOutputs() map[string]string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Create a copy to avoid data races
	result := make(map[string]string)
	for k, v := range m.capturedModelOutputs {
		result[k] = v
	}
	return result
}

// GetCapturedOutputDir returns the captured output directory in a thread-safe manner
func (m *BaseMockOutputWriter) GetCapturedOutputDir() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.capturedOutputDir
}
