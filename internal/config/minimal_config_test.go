package config

import (
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/logutil"
)

func TestNewDefaultMinimalConfig(t *testing.T) {
	t.Parallel()
	cfg := NewDefaultMinimalConfig()

	// Verify defaults
	if cfg == nil {
		t.Fatal("NewDefaultMinimalConfig() returned nil")
	}

	// Check default values
	if len(cfg.ModelNames) != 1 || cfg.ModelNames[0] != DefaultModel {
		t.Errorf("Expected ModelNames to be [%q], got %v", DefaultModel, cfg.ModelNames)
	}

	if cfg.LogLevel != logutil.InfoLevel {
		t.Errorf("Expected LogLevel to be %v, got %v", logutil.InfoLevel, cfg.LogLevel)
	}

	if cfg.Timeout != DefaultTimeout {
		t.Errorf("Expected Timeout to be %v, got %v", DefaultTimeout, cfg.Timeout)
	}

	if cfg.Format != DefaultFormat {
		t.Errorf("Expected Format to be %q, got %q", DefaultFormat, cfg.Format)
	}

	if cfg.Exclude != DefaultExcludes {
		t.Errorf("Expected Exclude to be %q, got %q", DefaultExcludes, cfg.Exclude)
	}

	if cfg.ExcludeNames != DefaultExcludeNames {
		t.Errorf("Expected ExcludeNames to be %q, got %q", DefaultExcludeNames, cfg.ExcludeNames)
	}

	// Check zero values
	if cfg.InstructionsFile != "" {
		t.Errorf("Expected InstructionsFile to be empty, got %q", cfg.InstructionsFile)
	}

	if cfg.OutputDir != "" {
		t.Errorf("Expected OutputDir to be empty, got %q", cfg.OutputDir)
	}

	if cfg.DryRun != false {
		t.Errorf("Expected DryRun to be false, got %v", cfg.DryRun)
	}

	if cfg.Verbose != false {
		t.Errorf("Expected Verbose to be false, got %v", cfg.Verbose)
	}
}

func TestMinimalConfigInterface(t *testing.T) {
	t.Parallel()
	cfg := &MinimalConfig{
		InstructionsFile: "test.md",
		TargetPaths:      []string{"src/", "docs/"},
		ModelNames:       []string{"model1", "model2"},
		OutputDir:        "output",
		DryRun:           true,
		Verbose:          true,
		SynthesisModel:   "synthesis-model",
		LogLevel:         logutil.DebugLevel,
		Timeout:          5 * time.Minute,
		Quiet:            true,
		NoProgress:       true,
		Format:           "custom-format",
		Exclude:          ".txt,.log",
		ExcludeNames:     "temp,cache",
	}

	// Test interface implementation
	var iface ConfigInterface = cfg

	// Test all getter methods
	if iface.GetInstructionsFile() != "test.md" {
		t.Errorf("GetInstructionsFile() = %q, want %q", iface.GetInstructionsFile(), "test.md")
	}

	if paths := iface.GetTargetPaths(); len(paths) != 2 || paths[0] != "src/" || paths[1] != "docs/" {
		t.Errorf("GetTargetPaths() = %v, want [src/ docs/]", paths)
	}

	if models := iface.GetModelNames(); len(models) != 2 || models[0] != "model1" || models[1] != "model2" {
		t.Errorf("GetModelNames() = %v, want [model1 model2]", models)
	}

	if iface.GetOutputDir() != "output" {
		t.Errorf("GetOutputDir() = %q, want %q", iface.GetOutputDir(), "output")
	}

	if !iface.IsDryRun() {
		t.Error("IsDryRun() = false, want true")
	}

	if !iface.IsVerbose() {
		t.Error("IsVerbose() = false, want true")
	}

	if iface.GetSynthesisModel() != "synthesis-model" {
		t.Errorf("GetSynthesisModel() = %q, want %q", iface.GetSynthesisModel(), "synthesis-model")
	}

	if iface.GetLogLevel() != logutil.DebugLevel {
		t.Errorf("GetLogLevel() = %v, want %v", iface.GetLogLevel(), logutil.DebugLevel)
	}

	if iface.GetTimeout() != 5*time.Minute {
		t.Errorf("GetTimeout() = %v, want %v", iface.GetTimeout(), 5*time.Minute)
	}

	if !iface.IsQuiet() {
		t.Error("IsQuiet() = false, want true")
	}

	if iface.ShouldShowProgress() {
		t.Error("ShouldShowProgress() = true, want false")
	}

	if iface.GetFormat() != "custom-format" {
		t.Errorf("GetFormat() = %q, want %q", iface.GetFormat(), "custom-format")
	}

	if iface.GetExclude() != ".txt,.log" {
		t.Errorf("GetExclude() = %q, want %q", iface.GetExclude(), ".txt,.log")
	}

	if iface.GetExcludeNames() != "temp,cache" {
		t.Errorf("GetExcludeNames() = %q, want %q", iface.GetExcludeNames(), "temp,cache")
	}
}

func TestCliConfigInterface(t *testing.T) {
	t.Parallel()
	// Test that CliConfig also implements ConfigInterface (for transition period)
	cfg := &CliConfig{
		InstructionsFile: "test.md",
		Paths:            []string{"src/"},
		ModelNames:       []string{"model1"},
		OutputDir:        "output",
		DryRun:           true,
		Verbose:          true,
		SynthesisModel:   "synthesis",
		LogLevel:         logutil.WarnLevel,
		Timeout:          10 * time.Minute,
		Quiet:            false,
		NoProgress:       false,
		Format:           "format",
		Exclude:          ".exe",
		ExcludeNames:     "node_modules",
	}

	var iface ConfigInterface = cfg

	// Basic smoke test to ensure interface is implemented
	if iface.GetInstructionsFile() != "test.md" {
		t.Errorf("GetInstructionsFile() = %q, want %q", iface.GetInstructionsFile(), "test.md")
	}

	if paths := iface.GetTargetPaths(); len(paths) != 1 || paths[0] != "src/" {
		t.Errorf("GetTargetPaths() = %v, want [src/]", paths)
	}

	if !iface.IsDryRun() {
		t.Error("IsDryRun() = false, want true")
	}

	if iface.ShouldShowProgress() != true {
		t.Error("ShouldShowProgress() = false, want true")
	}
}
