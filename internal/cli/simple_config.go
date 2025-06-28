// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/phrazzld/thinktank/internal/models"
)

// SimplifiedConfig represents the essential configuration in exactly 33 bytes.
// Following Go's principle of "less is more", this struct contains only the
// absolutely necessary fields with smart defaults for everything else.
//
// Memory layout (33 bytes total on 64-bit systems):
// - InstructionsFile: 16 bytes (string header: ptr+len)
// - TargetPath: 16 bytes (string header: ptr+len)
// - Flags: 1 byte (bitfield for DryRun|Verbose|Synthesis)
type SimplifiedConfig struct {
	InstructionsFile string // 16 bytes (pointer + length on 64-bit)
	TargetPath       string // 16 bytes (pointer + length on 64-bit)
	Flags            uint8  // 1 byte bitfield
	// Note: Actual struct size may include padding. The 33-byte target
	// refers to the logical data size, not the Go struct alignment.
}

// Flag constants for bitwise operations - O(1) validation
const (
	FlagDryRun     uint8 = 1 << iota // 0x01
	FlagVerbose                      // 0x02
	FlagSynthesis                    // 0x04
	FlagDebug                        // 0x08
	FlagQuiet                        // 0x10
	FlagJsonLogs                     // 0x20
	FlagNoProgress                   // 0x40
	// 1 bit remaining for future expansion
)

// Smart defaults - applied during parsing/conversion
const (
	DefaultModel     = "gemini-2.5-pro"
	DefaultOutputDir = "."
)

// HasFlag checks if a flag is set using bitwise AND - O(1) operation
func (s *SimplifiedConfig) HasFlag(flag uint8) bool {
	return s.Flags&flag != 0
}

// SetFlag sets a flag using bitwise OR - O(1) operation
func (s *SimplifiedConfig) SetFlag(flag uint8) {
	s.Flags |= flag
}

// ClearFlag clears a flag using bitwise AND NOT - O(1) operation
func (s *SimplifiedConfig) ClearFlag(flag uint8) {
	s.Flags &^= flag
}

// Validate performs essential validation with fail-fast behavior.
// Target: <1ms for typical inputs, ordered by likelihood of failure.
// Uses lazy evaluation to minimize syscalls and API key checks.
func (s *SimplifiedConfig) Validate() error {
	// 1. Path length validation first - O(1) check to prevent filesystem errors (~0.001ms)
	// Use 255 as a conservative limit that works across all filesystems
	const maxPathLength = 255
	if len(s.TargetPath) > maxPathLength {
		return fmt.Errorf("target path too long (max %d characters): %d characters", maxPathLength, len(s.TargetPath))
	}

	if s.InstructionsFile != "" && len(s.InstructionsFile) > maxPathLength {
		return fmt.Errorf("instructions file path too long (max %d characters): %d characters", maxPathLength, len(s.InstructionsFile))
	}

	// 2. Enhanced positional argument validation with file extension checks
	// For dry-run mode, only validate if instructions file is provided
	if s.HasFlag(FlagDryRun) {
		// In dry-run mode, validate positional arguments only if both are non-empty
		if s.InstructionsFile != "" && s.TargetPath != "" {
			if err := validatePositionalArgs(s.InstructionsFile, s.TargetPath); err != nil {
				return err
			}
		} else if s.TargetPath == "" {
			return fmt.Errorf("target path required")
		}
		// Instructions file can be empty in dry-run mode
	} else {
		// In normal mode, both arguments are required and must be validated
		if err := validatePositionalArgs(s.InstructionsFile, s.TargetPath); err != nil {
			return err
		}
	}

	// 5. API key validation - environment lookup only (~0.01ms)
	// Only validate if not in dry-run mode
	if !s.HasFlag(FlagDryRun) {
		if s.HasFlag(FlagSynthesis) {
			// For synthesis, pre-computed model list for efficiency
			return validateAPIKeysForModels([]string{"gemini-2.5-pro", "gpt-4.1"})
		} else {
			// Single model validation
			return validateAPIKeyForModel(DefaultModel)
		}
	}

	return nil
}

// validateAPIKeyForModel checks if the required API key is set for the given model.
// This is extracted to a helper to avoid code duplication and enable testing.
func validateAPIKeyForModel(modelName string) error {
	// Get provider for the model
	provider, err := models.GetProviderForModel(modelName)
	if err != nil {
		return fmt.Errorf("unknown model %s: %w", modelName, err)
	}

	// Get the environment variable name for this provider
	envVar := models.GetAPIKeyEnvVar(provider)
	if envVar == "" {
		return fmt.Errorf("unknown provider %s for model %s", provider, modelName)
	}

	// Check if the API key is set (not validating the key itself, just presence)
	if os.Getenv(envVar) == "" {
		return fmt.Errorf("API key not set: please set %s environment variable", envVar)
	}

	return nil
}

// validateAPIKeysForModels efficiently validates API keys for multiple models.
// Uses a set to avoid duplicate provider checks, optimizing for synthesis mode.
func validateAPIKeysForModels(modelNames []string) error {
	// Use map to deduplicate providers and cache env vars
	checkedProviders := make(map[string]string, 3) // Expected max: openai, gemini, openrouter

	for _, modelName := range modelNames {
		// Get provider for the model
		provider, err := models.GetProviderForModel(modelName)
		if err != nil {
			return fmt.Errorf("unknown model %s: %w", modelName, err)
		}

		// Skip if we already checked this provider
		if envVar, alreadyChecked := checkedProviders[provider]; alreadyChecked {
			if envVar == "" {
				return fmt.Errorf("API key not set: please set %s environment variable",
					models.GetAPIKeyEnvVar(provider))
			}
			continue
		}

		// Get and cache the environment variable for this provider
		envVar := models.GetAPIKeyEnvVar(provider)
		if envVar == "" {
			return fmt.Errorf("unknown provider %s for model %s", provider, modelName)
		}

		// Check if the API key is set
		keyValue := os.Getenv(envVar)
		checkedProviders[provider] = keyValue

		if keyValue == "" {
			return fmt.Errorf("API key not set: please set %s environment variable", envVar)
		}
	}

	return nil
}

// parseOptionalFlags processes optional flags with abbreviation support and validation.
// Supports both --flag=value and --flag value formats, plus short flags: -m, -v
// Returns positional arguments and any parsing errors.
func parseOptionalFlags(args []string, config *SimplifiedConfig) ([]string, error) {
	positionalArgs := make([]string, 0, 2) // Pre-allocate for performance

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Handle positional arguments (not flags)
		if len(arg) < 2 || arg[0] != '-' {
			positionalArgs = append(positionalArgs, arg)
			continue
		}

		// Handle short flags (single dash)
		if len(arg) == 2 && arg[0] == '-' && arg[1] != '-' {
			switch arg {
			case "-m":
				// -m requires a value for model
				if i+1 >= len(args) {
					return nil, fmt.Errorf("flag needs an argument: %s", arg)
				}
				i++ // Skip the model value - ignored in SimplifiedConfig
			case "-v":
				config.SetFlag(FlagVerbose)
			default:
				return nil, fmt.Errorf("unknown flag: %s", arg)
			}
			continue
		}

		// Handle long flags (double dash)
		if !strings.HasPrefix(arg, "--") {
			// Single dash with more than one character is an error
			return nil, fmt.Errorf("unknown flag: %s", arg)
		}

		// Handle long flags
		switch {
		case arg == "--dry-run":
			config.SetFlag(FlagDryRun)
		case arg == "--verbose":
			config.SetFlag(FlagVerbose)
		case arg == "--synthesis":
			config.SetFlag(FlagSynthesis)
		case arg == "--debug":
			config.SetFlag(FlagDebug)
		case arg == "--quiet":
			config.SetFlag(FlagQuiet)
		case arg == "--json-logs":
			config.SetFlag(FlagJsonLogs)
		case arg == "--no-progress":
			config.SetFlag(FlagNoProgress)
		case arg == "--model":
			// --model requires a value, but we ignore it for SimplifiedConfig
			if i+1 >= len(args) {
				return nil, fmt.Errorf("flag needs an argument: %s", arg)
			}
			i++ // Skip the value
		case arg == "--output-dir":
			// --output-dir requires a value, but we ignore it for SimplifiedConfig
			if i+1 >= len(args) {
				return nil, fmt.Errorf("flag needs an argument: %s", arg)
			}
			i++ // Skip the value
		case strings.HasPrefix(arg, "--model="):
			// Handle --model=value format
			value := strings.TrimPrefix(arg, "--model=")
			if value == "" {
				return nil, fmt.Errorf("flag has empty value: %s", arg)
			}
			// Value ignored for SimplifiedConfig
		case strings.HasPrefix(arg, "--output-dir="):
			// Handle --output-dir=value format
			value := strings.TrimPrefix(arg, "--output-dir=")
			if value == "" {
				return nil, fmt.Errorf("flag has empty value: %s", arg)
			}
			// Value ignored for SimplifiedConfig
		default:
			return nil, fmt.Errorf("unknown flag: %s", arg)
		}
	}

	return positionalArgs, nil
}

// ParseSimplifiedArgs parses command line arguments into a SimplifiedConfig
// Expected format: [flags...] <instructions-file> <target-path>
// Supported flags: --model, --output-dir, --dry-run, --verbose, --synthesis
// Supported abbreviations: -m (--model), -v (--verbose)
func ParseSimplifiedArgs(args []string) (*SimplifiedConfig, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("insufficient arguments: instructions file and target path required")
	}

	config := &SimplifiedConfig{
		Flags: 0,
	}

	// Parse optional flags and extract positional arguments
	positionalArgs, err := parseOptionalFlags(args, config)
	if err != nil {
		return nil, err
	}

	// Validate we have enough positional arguments
	if len(positionalArgs) < 2 {
		if len(positionalArgs) == 0 {
			return nil, fmt.Errorf("insufficient arguments: instructions file and target path required")
		}
		return nil, fmt.Errorf("target path required")
	}

	// Assign positional arguments
	config.InstructionsFile = positionalArgs[0]
	config.TargetPath = positionalArgs[1]

	return config, nil
}

// validatePositionalArgs validates the two required positional arguments.
// Following Go's principle of explicit validation with clear error messages.
// This function extracts and enhances validation logic for reusability.
func validatePositionalArgs(instructionsFile, targetPath string) error {
	// 1. Essential field validation - O(1) string checks
	if targetPath == "" {
		return fmt.Errorf("target path required: specify a file or directory to analyze")
	}

	if instructionsFile == "" {
		return fmt.Errorf("instructions file required: specify a .txt or .md file with analysis instructions")
	}

	// 2. Target path validation - filesystem checks first
	if err := validateTargetPathAccess(targetPath); err != nil {
		return err
	}

	// 3. Instructions file validation - filesystem checks first (directory check before extension)
	if err := validateInstructionsFileAccess(instructionsFile); err != nil {
		return err
	}

	// 4. File extension validation with helpful suggestions (only for non-directories)
	if err := validateInstructionsFileExtension(instructionsFile); err != nil {
		return err
	}

	return nil
}

// validateInstructionsFileExtension checks file extension with user-friendly errors
func validateInstructionsFileExtension(filePath string) error {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".txt", ".md":
		return nil
	case "":
		return fmt.Errorf("instructions file missing extension: use .txt or .md format (got: %s)", filePath)
	default:
		return fmt.Errorf("unsupported instructions file extension: %s (supported: .txt, .md)", ext)
	}
}

// validateTargetPathAccess checks path existence and readability
func validateTargetPathAccess(targetPath string) error {
	info, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("target path does not exist: %s (check spelling and permissions)", targetPath)
		}
		return fmt.Errorf("cannot access target path %s: %w", targetPath, err)
	}

	// Check readability based on file type
	if info.IsDir() {
		return validateDirectoryAccess(targetPath, info)
	}
	return validateFileAccess(targetPath, info)
}

// validateInstructionsFileAccess checks instructions file existence and readability
func validateInstructionsFileAccess(instructionsFile string) error {
	info, err := os.Stat(instructionsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("instructions file does not exist: %s (check spelling and permissions)", instructionsFile)
		}
		return fmt.Errorf("cannot access instructions file %s: %w", instructionsFile, err)
	}

	// Check if it's a directory first - this takes precedence over extension validation
	if info.IsDir() {
		return fmt.Errorf("instructions file is a directory: %s", instructionsFile)
	}

	// Permission check only - no need to open file
	if info.Mode().Perm()&0444 == 0 {
		return fmt.Errorf("instructions file has no read permissions: %s", instructionsFile)
	}

	return nil
}

// validateDirectoryAccess validates directory permissions
func validateDirectoryAccess(targetPath string, info os.FileInfo) error {
	// For directories, use permission bits first (faster than os.Open)
	if info.Mode().Perm()&0444 == 0 {
		return fmt.Errorf("target directory has no read permissions: %s", targetPath)
	}
	// Only if permissions look OK, verify with actual read attempt
	// This catches cases where parent dirs are unreadable
	if dir, err := os.Open(targetPath); err != nil {
		return fmt.Errorf("target directory is not readable: %w", err)
	} else {
		_ = dir.Close() // Ignore close error for read-only operation
	}
	return nil
}

// validateFileAccess validates file permissions
func validateFileAccess(targetPath string, info os.FileInfo) error {
	// For files, permission check is sufficient
	if info.Mode().Perm()&0444 == 0 {
		return fmt.Errorf("target file has no read permissions: %s", targetPath)
	}
	return nil
}
