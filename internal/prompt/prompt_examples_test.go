package prompt

import (
	"strings"
	"testing"
)

func TestListExampleTemplates(t *testing.T) {
	// Create a manager
	logger := newMockLogger()
	manager := NewManager(logger)

	// Get example templates
	examples, err := manager.ListExampleTemplates()
	if err != nil {
		t.Fatalf("ListExampleTemplates returned error: %v", err)
	}

	// Verify that we have some examples
	if len(examples) == 0 {
		t.Error("No example templates found, expected at least one")
	}

	// Verify that we have the expected example templates
	expectedTemplates := []string{"basic.tmpl", "detailed.tmpl", "bugfix.tmpl", "feature.tmpl"}
	for _, expected := range expectedTemplates {
		found := false
		for _, actual := range examples {
			if expected == actual {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find example template %s, but it was not in the list", expected)
		}
	}
}

func TestGetExampleTemplate(t *testing.T) {
	// Create a manager
	logger := newMockLogger()
	manager := NewManager(logger)

	// Test cases
	testCases := []struct {
		name          string
		templateName  string
		expectedError bool
		expectedText  string
	}{
		{
			name:          "Valid template with .tmpl extension",
			templateName:  "basic.tmpl",
			expectedError: false,
			expectedText:  "You are a skilled software engineer",
		},
		{
			name:          "Valid template without .tmpl extension",
			templateName:  "basic",
			expectedError: false,
			expectedText:  "You are a skilled software engineer",
		},
		{
			name:          "Non-existent template",
			templateName:  "nonexistent.tmpl",
			expectedError: true,
			expectedText:  "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content, err := manager.GetExampleTemplate(tc.templateName)

			// Check error
			if tc.expectedError && err == nil {
				t.Fatalf("Expected an error, but got none")
			}
			if !tc.expectedError && err != nil {
				t.Fatalf("Expected no error, but got: %v", err)
			}

			// If we expect content, check it
			if tc.expectedText != "" && !strings.Contains(content, tc.expectedText) {
				t.Errorf("Expected template content to contain '%s', but got: %s", tc.expectedText, content)
			}
		})
	}
}
