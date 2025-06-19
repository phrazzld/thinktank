package logutil

import "fmt"

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
