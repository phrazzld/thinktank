package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/logutil"
)

func TestListExampleTemplatesCommand(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a logger
	logger := logutil.NewLogger(logutil.DebugLevel, os.Stderr, "[test] ", false)

	// Create a config manager
	configManager := config.NewManager(logger)

	// Call the function
	listExampleTemplates(logger, configManager)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify the output contains expected text
	if !strings.Contains(output, "Available Example Templates:") {
		t.Errorf("Expected output to contain 'Available Example Templates:', got: %s", output)
	}

	// Check for the expected templates
	expectedTemplates := []string{"basic.tmpl", "detailed.tmpl", "bugfix.tmpl", "feature.tmpl"}
	for _, tmpl := range expectedTemplates {
		if !strings.Contains(output, tmpl) {
			t.Errorf("Expected output to contain template '%s', got: %s", tmpl, output)
		}
	}
}

func TestShowExampleTemplateCommand(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a logger
	logger := logutil.NewLogger(logutil.DebugLevel, os.Stderr, "[test] ", false)

	// Create a config manager
	configManager := config.NewManager(logger)

	// Call the function with a known template name
	showExampleTemplate("basic.tmpl", logger, configManager)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify the output contains expected text from the template
	if !strings.Contains(output, "You are a skilled software engineer") {
		t.Errorf("Expected output to contain template content, got: %s", output)
	}
}

func TestParsingExampleFlags(t *testing.T) {
	// Save old command line
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Save old flagset and restore after test
	oldFlagCommandLine := flag.CommandLine
	defer func() { flag.CommandLine = oldFlagCommandLine }()

	testCases := []struct {
		name                string
		args                []string
		expectedListFlag    bool
		expectedShowExample string
	}{
		{
			name:                "No example flags",
			args:                []string{"architect", "--task-file", "task.txt", "./"},
			expectedListFlag:    false,
			expectedShowExample: "",
		},
		{
			name:                "With list-examples flag",
			args:                []string{"architect", "--list-examples"},
			expectedListFlag:    true,
			expectedShowExample: "",
		},
		{
			name:                "With show-example flag",
			args:                []string{"architect", "--show-example", "basic.tmpl"},
			expectedListFlag:    false,
			expectedShowExample: "basic.tmpl",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset flag set for each test
			flag.CommandLine = flag.NewFlagSet(tc.args[0], flag.ExitOnError)

			// Set command line args
			os.Args = tc.args

			// Parse flags
			config := parseFlags()

			// Check flag values
			if config.ListExamples != tc.expectedListFlag {
				t.Errorf("Expected ListExamples to be %v, got %v", tc.expectedListFlag, config.ListExamples)
			}

			if config.ShowExample != tc.expectedShowExample {
				t.Errorf("Expected ShowExample to be %q, got %q", tc.expectedShowExample, config.ShowExample)
			}
		})
	}
}
