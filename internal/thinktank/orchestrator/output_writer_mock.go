package orchestrator

import (
	"context"
)

// BaseMockOutputWriter is the base implementation for mock output writers
type BaseMockOutputWriter struct {
	savedCount           int
	saveError            error
	capturedModelOutputs map[string]string
	capturedOutputDir    string
}

// SaveIndividualOutputs is a mock implementation that returns configured results
func (m *BaseMockOutputWriter) SaveIndividualOutputs(ctx context.Context, modelOutputs map[string]string, outputDir string) (int, map[string]string, error) {
	m.capturedModelOutputs = modelOutputs
	m.capturedOutputDir = outputDir

	// Create mock file paths map - one path per model
	filePaths := make(map[string]string)
	for modelName := range modelOutputs {
		filePaths[modelName] = outputDir + "/" + modelName + ".md"
	}

	return m.savedCount, filePaths, m.saveError
}
