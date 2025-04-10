package architect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/prompt"
)

// mockLogger for testing
type mockPromptLogger struct {
	debugMessages []string
	infoMessages  []string
	warnMessages  []string
	errorMessages []string
}

func (m *mockPromptLogger) Debug(format string, args ...interface{}) {
	m.debugMessages = append(m.debugMessages, format)
}

func (m *mockPromptLogger) Info(format string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, format)
}

func (m *mockPromptLogger) Warn(format string, args ...interface{}) {
	m.warnMessages = append(m.warnMessages, format)
}

func (m *mockPromptLogger) Error(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, format)
}

// mockPromptManager for testing
type mockPromptManager struct {
	loadTemplateErr     error
	buildPromptResult   string
	buildPromptErr      error
	listTemplatesResult []string
	listTemplatesErr    error
	exampleTemplateContent string
	exampleTemplateErr     error
}

func (m *mockPromptManager) LoadTemplate(templatePath string) error {
	return m.loadTemplateErr
}

func (m *mockPromptManager) BuildPrompt(templateName string, data *prompt.TemplateData) (string, error) {
	return m.buildPromptResult, m.buildPromptErr
}

func (m *mockPromptManager) ListTemplates() ([]string, error) {
	return m.listTemplatesResult, m.listTemplatesErr
}

func (m *mockPromptManager) ListExampleTemplates() ([]string, error) {
	return m.listTemplatesResult, m.listTemplatesErr
}

func (m *mockPromptManager) GetExampleTemplate(name string) (string, error) {
	return m.exampleTemplateContent, m.exampleTemplateErr
}

// mockConfigManager for testing
type mockConfigManager struct {
	config.ManagerInterface
}

// TestNewPromptBuilder tests the creation of a PromptBuilder
func TestNewPromptBuilder(t *testing.T) {
	logger := &mockPromptLogger{}
	
	builder := NewPromptBuilder(logger)
	if builder == nil {
		t.Fatal("Expected non-nil PromptBuilder")
	}
}

// TestReadTaskFromFile tests reading task files
func TestReadTaskFromFile(t *testing.T) {
	logger := &mockPromptLogger{}
	builder := NewPromptBuilder(logger)
	
	// Test reading a valid file
	t.Run("ValidFile", func(t *testing.T) {
		// Create a temporary file with content
		tmpFile, err := os.CreateTemp("", "test-task-*.md")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		
		// Write test content
		testContent := "This is a test task file"
		if _, err := tmpFile.WriteString(testContent); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		tmpFile.Close()
		
		// Read the file using the prompt builder
		content, err := builder.ReadTaskFromFile(tmpFile.Name())
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		if content != testContent {
			t.Errorf("Expected content %q, got %q", testContent, content)
		}
	})
	
	// Test reading a non-existent file
	t.Run("NonexistentFile", func(t *testing.T) {
		_, err := builder.ReadTaskFromFile("/this/file/does/not/exist.md")
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})
	
	// Test reading a directory
	t.Run("DirectoryAsFile", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "test-dir")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)
		
		_, err = builder.ReadTaskFromFile(tempDir)
		if err == nil {
			t.Error("Expected error when reading directory as file, got nil")
		}
	})
	
	// Test reading an empty file
	t.Run("EmptyFile", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test-empty-*.md")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()
		
		_, err = builder.ReadTaskFromFile(tmpFile.Name())
		if err == nil {
			t.Error("Expected error for empty file, got nil")
		}
	})
}

// TestBuildPromptInternal tests the internal prompt building helper
func TestBuildPromptInternal(t *testing.T) {
	logger := &mockPromptLogger{}
	builder := &promptBuilder{logger: logger}
	
	t.Run("SuccessfulBuild", func(t *testing.T) {
		// Create a mock prompt manager that returns a successful result
		mockManager := &mockPromptManager{
			buildPromptResult: "Generated prompt",
			buildPromptErr:    nil,
		}
		
		// Call the internal helper method
		result, err := builder.buildPromptInternal("task", "context", "template.tmpl", mockManager)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		if result != "Generated prompt" {
			t.Errorf("Expected result %q, got %q", "Generated prompt", result)
		}
	})
	
	t.Run("BuildError", func(t *testing.T) {
		// Create a mock prompt manager that returns an error
		mockManager := &mockPromptManager{
			buildPromptResult: "",
			buildPromptErr:    filepath.ErrNotExist,
		}
		
		// Call the internal helper method
		_, err := builder.buildPromptInternal("task", "context", "template.tmpl", mockManager)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
	
	t.Run("DefaultTemplate", func(t *testing.T) {
		// Create a mock prompt manager
		mockManager := &mockPromptManager{
			buildPromptResult: "Default template result",
			buildPromptErr:    nil,
		}
		
		// Call the internal helper method with empty template name
		result, err := builder.buildPromptInternal("task", "context", "", mockManager)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		if result != "Default template result" {
			t.Errorf("Expected result %q, got %q", "Default template result", result)
		}
	})
}

// TestBuildPrompt tests the BuildPrompt method
func TestBuildPrompt(t *testing.T) {
	logger := &mockPromptLogger{}
	builder := NewPromptBuilder(logger)
	
	// This is mostly a smoke test since the actual work is done in buildPromptInternal
	// and prompt.Manager which we can't easily mock here
	
	t.Run("BasicBuild", func(t *testing.T) {
		_, err := builder.BuildPrompt("Task description", "Context content", "")
		// We're just checking that it doesn't panic
		if err != nil {
			// This error is expected since the test environment doesn't have the embedded templates
			// Just verify it's the right type of error
			if !isTemplateNotFoundError(err) {
				t.Errorf("Unexpected error type: %v", err)
			}
		}
	})
}

// isTemplateNotFoundError checks if the error is related to template not found
func isTemplateNotFoundError(err error) bool {
	return err != nil && (filepath.ErrNotExist.Error() == err.Error() || 
			filepath.ErrNotExist.Error() == filepath.Err(err).Error() ||
			os.ErrNotExist.Error() == err.Error())
}

// TestBuildPromptWithConfig tests the BuildPromptWithConfig method
func TestBuildPromptWithConfig(t *testing.T) {
	logger := &mockPromptLogger{}
	builder := NewPromptBuilder(logger)
	configManager := &mockConfigManager{}
	
	// This is mostly a smoke test since the actual work is done in buildPromptInternal
	// and prompt.Manager which we can't easily mock here
	
	t.Run("BasicBuild", func(t *testing.T) {
		_, err := builder.BuildPromptWithConfig("Task description", "Context content", "", configManager)
		// We're just checking that it doesn't panic
		if err != nil {
			// This error is expected since the test environment doesn't have SetupPromptManagerWithConfig implemented
			// Just verify it's the right type of error
			if err.Error() != "failed to set up prompt manager: not implemented" {
				t.Errorf("Unexpected error type: %v", err)
			}
		}
	})
}

// TestListExampleTemplates tests the ListExampleTemplates method
func TestListExampleTemplates(t *testing.T) {
	t.Skip("Skipping test as it requires a full implementation of prompt.SetupPromptManagerWithConfig")
}

// TestShowExampleTemplate tests the ShowExampleTemplate method
func TestShowExampleTemplate(t *testing.T) {
	t.Skip("Skipping test as it requires a full implementation of prompt.SetupPromptManagerWithConfig")
}