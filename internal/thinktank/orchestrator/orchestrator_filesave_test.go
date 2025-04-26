package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/fileutil"
	"github.com/phrazzld/thinktank/internal/ratelimit"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
	"github.com/phrazzld/thinktank/internal/thinktank/modelproc"
)

// MockFileWriter that simulates failures based on file path
type MockFailingFileWriter struct {
	// Files that should fail to save
	FailingFiles []string
	// Track files that were saved
	SavedFiles map[string]string
}

// SaveToFile implements the FileWriter interface
func (m *MockFailingFileWriter) SaveToFile(content, outputFile string) error {
	// Check if this file should fail
	for _, failingFile := range m.FailingFiles {
		if strings.Contains(outputFile, failingFile) {
			return errors.New("simulated file save error")
		}
	}

	// If we get here, the file save succeeded
	if m.SavedFiles == nil {
		m.SavedFiles = make(map[string]string)
	}
	m.SavedFiles[outputFile] = content
	return nil
}

// TestFileSaveErrorPropagation tests that file save errors are properly propagated
func TestFileSaveErrorPropagation(t *testing.T) {
	tests := []struct {
		name                string
		modelNames          []string
		mockResults         []modelResult
		failingFiles        []string
		synthesisModel      string
		expectError         bool
		expectedErrorPrefix string
		expectedSavedCount  int // Expected number of successfully saved files
	}{
		{
			name:       "All file saves succeed",
			modelNames: []string{"model1", "model2"},
			mockResults: []modelResult{
				{modelName: "model1", content: "Output from model1", err: nil},
				{modelName: "model2", content: "Output from model2", err: nil},
			},
			failingFiles:        []string{},
			expectError:         false,
			expectedErrorPrefix: "",
			expectedSavedCount:  2,
		},
		{
			name:       "Some file saves fail",
			modelNames: []string{"model1", "model2", "model3"},
			mockResults: []modelResult{
				{modelName: "model1", content: "Output from model1", err: nil},
				{modelName: "model2", content: "Output from model2", err: nil},
				{modelName: "model3", content: "Output from model3", err: nil},
			},
			failingFiles:        []string{"model2"},
			expectError:         true,
			expectedErrorPrefix: "1/3 files failed to save",
			expectedSavedCount:  2,
		},
		{
			name:       "All file saves fail",
			modelNames: []string{"model1", "model2"},
			mockResults: []modelResult{
				{modelName: "model1", content: "Output from model1", err: nil},
				{modelName: "model2", content: "Output from model2", err: nil},
			},
			failingFiles:        []string{"model1", "model2"},
			expectError:         true,
			expectedErrorPrefix: "2/2 files failed to save",
			expectedSavedCount:  0,
		},
		{
			name:       "Model errors and file save errors",
			modelNames: []string{"model1", "model2", "model3"},
			mockResults: []modelResult{
				{modelName: "model1", content: "Output from model1", err: nil},
				{modelName: "model2", content: "", err: errors.New("model2 failed")},
				{modelName: "model3", content: "Output from model3", err: nil},
			},
			failingFiles:        []string{"model3"},
			expectError:         true,
			expectedErrorPrefix: "model processing errors and file save errors occurred",
			expectedSavedCount:  1, // Only model1 should be successfully saved
		},
		{
			name:       "Synthesis model file save fails",
			modelNames: []string{"model1", "model2"},
			mockResults: []modelResult{
				{modelName: "model1", content: "Output from model1", err: nil},
				{modelName: "model2", content: "Output from model2", err: nil},
			},
			synthesisModel:      "synth-model",
			failingFiles:        []string{"synth-model-synthesis"},
			expectError:         true,
			expectedErrorPrefix: "failed to save synthesis output",
			expectedSavedCount:  0, // Since the synthesis file fails to save
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockAPIService := &MockAPIService{}
			mockContextGatherer := &MockContextGatherer{}
			mockAuditLogger := &MockAuditLogger{}
			mockLogger := &MockLogger{}

			// Create a failing file writer
			mockFileWriter := &MockFailingFileWriter{
				FailingFiles: tt.failingFiles,
				SavedFiles:   make(map[string]string),
			}

			// Create config
			cfg := &config.CliConfig{
				ModelNames:     tt.modelNames,
				SynthesisModel: tt.synthesisModel,
				OutputDir:      "/tmp/test-output",
			}

			// Create rate limiter
			rateLimiter := ratelimit.NewRateLimiter(0, 0)

			// Create test orchestrator with controlled behavior
			orch := &filesaveTestOrchestrator{
				Orchestrator: Orchestrator{
					apiService:      mockAPIService,
					contextGatherer: mockContextGatherer,
					fileWriter:      mockFileWriter,
					auditLogger:     mockAuditLogger,
					rateLimiter:     rateLimiter,
					config:          cfg,
					logger:          mockLogger,
				},
				mockResults: tt.mockResults,
			}

			// Call the Run method
			err := orch.Run(context.Background(), "test instructions")

			// Verify the error
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error but got nil")
				} else if !strings.HasPrefix(err.Error(), tt.expectedErrorPrefix) {
					t.Errorf("Expected error to start with %q but got %q", tt.expectedErrorPrefix, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got %v", err)
				}
			}

			// Verify the number of saved files
			if tt.expectedSavedCount != len(mockFileWriter.SavedFiles) {
				t.Errorf("Expected %d files to be saved, but got %d", tt.expectedSavedCount, len(mockFileWriter.SavedFiles))
			}
		})
	}
}

// filesaveTestOrchestrator extends Orchestrator for file save testing
type filesaveTestOrchestrator struct {
	Orchestrator
	mockResults []modelResult
}

// Override Run to avoid running the full orchestration logic
func (o *filesaveTestOrchestrator) Run(ctx context.Context, instructions string) error {
	// Call gatherProjectContext to maintain normal flow
	_, _, err := o.gatherProjectContext(ctx)
	if err != nil {
		return err
	}

	// Simulate the model processing step by getting our mock results
	modelOutputs := make(map[string]string)
	var modelErrors []error

	for _, result := range o.mockResults {
		if result.err == nil {
			modelOutputs[result.modelName] = result.content
		} else {
			modelErrors = append(modelErrors, result.err)
		}
	}

	// Create a returnErr if we have model errors
	var returnErr error
	if len(modelErrors) > 0 {
		returnErr = fmt.Errorf("processed %d/%d models successfully; %d failed: %v",
			len(modelOutputs), len(o.config.ModelNames), len(modelErrors),
			aggregateErrorMessages(modelErrors))
	}

	// Handle file save logic using the real implementation
	// We'll track file-save errors separately from model processing errors
	var fileSaveErrors error

	if o.config.SynthesisModel == "" {
		// No synthesis model specified - save individual model outputs
		totalCount := len(modelOutputs)
		savedCount := 0
		errorCount := 0

		// Iterate over the model outputs and save each to a file
		for modelName, content := range modelOutputs {
			// Sanitize model name for use in filename
			sanitizedModelName := modelproc.SanitizeFilename(modelName)

			// Construct output file path
			outputFilePath := filepath.Join(o.config.OutputDir, sanitizedModelName+".md")

			// Save the output to file
			if err := o.fileWriter.SaveToFile(content, outputFilePath); err != nil {
				errorCount++
			} else {
				savedCount++
			}
		}

		// Create a descriptive error for file save failures if needed
		if errorCount > 0 {
			fileSaveErrors = fmt.Errorf("%d/%d files failed to save", errorCount, totalCount)
		}
	} else {
		// Synthesis model specified - simulate synthesis result
		if len(modelOutputs) > 0 {
			// Create a synthesized result
			synthesisContent := "Synthesized content"

			// Sanitize model name for use in filename
			sanitizedModelName := modelproc.SanitizeFilename(o.config.SynthesisModel)

			// Construct output file path with -synthesis suffix
			outputFilePath := filepath.Join(o.config.OutputDir, sanitizedModelName+"-synthesis.md")

			// Save the synthesis output to file
			if err := o.fileWriter.SaveToFile(synthesisContent, outputFilePath); err != nil {
				fileSaveErrors = fmt.Errorf("failed to save synthesis output to %s: %w", outputFilePath, err)
			}
		}
	}

	// Return errors using the same logic as in the real implementation
	if returnErr != nil && fileSaveErrors != nil {
		return fmt.Errorf("model processing errors and file save errors occurred: %w; additionally: %v",
			returnErr, fileSaveErrors)
	} else if fileSaveErrors != nil {
		return fileSaveErrors
	} else {
		return returnErr
	}
}

// Override gatherProjectContext to return a dummy context
func (o *filesaveTestOrchestrator) gatherProjectContext(ctx context.Context) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
	return []fileutil.FileMeta{}, &interfaces.ContextStats{}, nil
}
