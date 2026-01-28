package logutil

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// FormatFileSize formats a file size in bytes to a human-readable string
// using binary units (1024 bytes = 1K). Returns formats like "512B", "4.2K", "1.3M".
//
// The function follows binary unit conventions:
// - B (bytes): 0-1023 bytes
// - K (kilobytes): 1024 bytes and above
// - M (megabytes): 1024^2 bytes and above
// - G (gigabytes): 1024^3 bytes and above
// - T (terabytes): 1024^4 bytes and above
// - P (petabytes): 1024^5 bytes and above
// - E (exabytes): 1024^6 bytes and above
//
// Values 1K and above are displayed with one decimal place precision.
// Negative values are handled by preserving the sign and formatting the absolute value.
func FormatFileSize(bytes int64) string {
	const unit = 1024

	// Handle negative values by preserving sign and working with absolute value
	negative := bytes < 0
	if negative {
		bytes = -bytes
	}

	if bytes < unit {
		if negative {
			return fmt.Sprintf("-%dB", bytes)
		}
		return fmt.Sprintf("%dB", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	result := fmt.Sprintf("%.1f%c", float64(bytes)/float64(div), "KMGTPE"[exp])
	if negative {
		return "-" + result
	}
	return result
}

// FormatDuration formats a time.Duration into a human-readable string
// like "1.2s", "850ms", etc.
func FormatDuration(d time.Duration) string {
	if d >= time.Second {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}

// FormatToWidth formats a message to fit within the specified width.
// In interactive mode, it truncates long messages with ellipsis.
// In non-interactive mode, it returns the message unchanged.
func FormatToWidth(message string, width int, isInteractive bool) string {
	// If message fits within terminal width, return as-is
	if len(message) <= width {
		return message
	}

	// In non-interactive mode, don't truncate messages
	if !isInteractive {
		return message
	}

	// Truncate message and add ellipsis
	if width <= 3 {
		return "..."
	}
	return message[:width-3] + "..."
}

// ColorizeStatus applies appropriate colors to status text based on content.
// It examines the status text for Unicode symbols and applies semantic coloring.
func ColorizeStatus(status string, colors *ColorScheme) string {
	// Determine status type based on Unicode symbols and keywords
	if strings.Contains(status, "✓") {
		// Success status - apply success color to the entire status
		return colors.ColorSuccess(status)
	} else if strings.Contains(status, "✗") {
		// Error status - apply error color to the entire status
		return colors.ColorError(status)
	} else if strings.Contains(status, "⚠") {
		// Warning status (rate limited, etc.) - apply warning color
		return colors.ColorWarning(status)
	}
	lowerStatus := strings.ToLower(status)
	if strings.Contains(lowerStatus, "processing") ||
		strings.Contains(lowerStatus, "queued") ||
		strings.Contains(lowerStatus, "starting") {
		return colors.ColorInfo(status)
	}

	// Default status - no special coloring
	return status
}

// DetectInteractiveEnvironment determines if we're running in an interactive
// environment based on TTY detection and CI environment variables.
func DetectInteractiveEnvironment(isTerminalFunc func() bool, getEnvFunc func(string) string) bool {
	// Check common CI environment variables
	ciVars := []string{
		"CI",
		"GITHUB_ACTIONS",
		"CONTINUOUS_INTEGRATION",
		"GITLAB_CI",
		"TRAVIS",
		"CIRCLECI",
		"JENKINS_URL",
		"BUILDKITE",
	}

	for _, envVar := range ciVars {
		value := getEnvFunc(envVar)
		if value != "" && (value == "true" || envVar == "JENKINS_URL") {
			return false
		}
	}

	// If not in CI and stdout is a terminal, we're interactive
	return isTerminalFunc()
}

// FormatConsoleMessage formats a message for console display with optional
// symbol and color formatting based on message type and environment.
type ConsoleMessageType int

const (
	MessageTypeInfo ConsoleMessageType = iota
	MessageTypeSuccess
	MessageTypeWarning
	MessageTypeError
	MessageTypeStatus
)

// ConsoleMessageOptions contains options for formatting console messages
type ConsoleMessageOptions struct {
	Type          ConsoleMessageType
	IsInteractive bool
	Width         int
	Colors        *ColorScheme
	Symbols       *SymbolProvider
}

// FormatConsoleMessage formats a message for console display with appropriate
// coloring, symbols, and width constraints based on the message type and environment.
func FormatConsoleMessage(message string, opts ConsoleMessageOptions) string {
	// Format message to fit terminal width
	formattedMessage := FormatToWidth(message, opts.Width, opts.IsInteractive)

	if opts.IsInteractive && opts.Colors != nil && opts.Symbols != nil {
		symbols := opts.Symbols.GetSymbols()

		switch opts.Type {
		case MessageTypeSuccess:
			coloredMessage := opts.Colors.ColorSuccess(formattedMessage)
			successSymbol := opts.Colors.ColorSuccess(symbols.Success)
			return fmt.Sprintf("%s %s", successSymbol, coloredMessage)
		case MessageTypeWarning:
			coloredMessage := opts.Colors.ColorWarning(formattedMessage)
			warningSymbol := opts.Colors.ColorWarning(symbols.Warning)
			return fmt.Sprintf("%s %s", warningSymbol, coloredMessage)
		case MessageTypeError:
			coloredMessage := opts.Colors.ColorError(formattedMessage)
			errorSymbol := opts.Colors.ColorError(symbols.Error)
			return fmt.Sprintf("%s %s", errorSymbol, coloredMessage)
		case MessageTypeStatus:
			bulletSymbol := opts.Colors.ColorSymbol(symbols.Bullet)
			return fmt.Sprintf("%s %s", bulletSymbol, formattedMessage)
		default:
			return formattedMessage
		}
	} else {
		// Non-interactive mode with prefixes
		switch opts.Type {
		case MessageTypeSuccess:
			return fmt.Sprintf("SUCCESS: %s", formattedMessage)
		case MessageTypeWarning:
			return fmt.Sprintf("WARNING: %s", formattedMessage)
		case MessageTypeError:
			return fmt.Sprintf("ERROR: %s", formattedMessage)
		default:
			return formattedMessage
		}
	}
}

// FormatJSON formats structured data as JSON with optional pretty printing.
// This provides a structured output format for data that needs to be consumed
// by other tools or when human-readable formatting is not appropriate.
func FormatJSON(data interface{}, pretty bool) (string, error) {
	var jsonBytes []byte
	var err error

	if pretty {
		jsonBytes, err = json.MarshalIndent(data, "", "  ")
	} else {
		jsonBytes, err = json.Marshal(data)
	}

	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// I/O Operations - Pure functions for writing to the console
// These functions handle only the actual writing operations, keeping I/O
// separate from formatting logic following Carmack's philosophy.

// WriteToConsole writes a formatted message to stdout.
// This is a pure I/O operation that should be used for all console output.
func WriteToConsole(message string) {
	fmt.Println(message)
}

// WriteToConsoleF writes a formatted message to stdout using Printf.
// This is a pure I/O operation for formatted output.
func WriteToConsoleF(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// WriteLineToConsole writes a message to stdout with a newline.
// Identical to WriteToConsole but more explicit about the newline behavior.
func WriteLineToConsole(message string) {
	fmt.Println(message)
}

// WriteEmptyLineToConsole writes an empty line to stdout.
// Used for visual separation between sections.
func WriteEmptyLineToConsole() {
	fmt.Println()
}

// WriteToStderr writes a message to stderr for errors and warnings.
// This ensures error messages don't interfere with structured stdout output.
func WriteToStderr(message string) {
	fmt.Fprintln(os.Stderr, message)
}

// WriteToStderrF writes a formatted message to stderr.
// Used for error reporting that should not be captured with stdout.
func WriteToStderrF(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}
