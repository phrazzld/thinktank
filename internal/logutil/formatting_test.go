package logutil

import (
	"testing"
)

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		// Bytes (less than 1024)
		{"zero bytes", 0, "0B"},
		{"single byte", 1, "1B"},
		{"small bytes", 512, "512B"},
		{"max bytes", 1023, "1023B"},

		// Kilobytes
		{"exact 1K", 1024, "1.0K"},
		{"1.5K", 1536, "1.5K"},
		{"large K", 1048575, "1024.0K"}, // Just under 1M

		// Megabytes
		{"exact 1M", 1048576, "1.0M"},      // 1024^2
		{"4.2M", 4404019, "4.2M"},          // ~4.2MB
		{"large M", 1073741823, "1024.0M"}, // Just under 1G

		// Gigabytes
		{"exact 1G", 1073741824, "1.0G"},      // 1024^3
		{"2.5G", 2684354560, "2.5G"},          // ~2.5GB
		{"large G", 1099511627775, "1024.0G"}, // Just under 1T

		// Terabytes
		{"exact 1T", 1099511627776, "1.0T"},      // 1024^4
		{"1.2T", 1319413953331, "1.2T"},          // ~1.2TB
		{"large T", 1125899906842623, "1024.0T"}, // Just under 1P

		// Petabytes
		{"exact 1P", 1125899906842624, "1.0P"},      // 1024^5
		{"2.3P", 2589569785253478, "2.3P"},          // ~2.3PB
		{"large P", 1152921504606846975, "1024.0P"}, // Just under 1E

		// Exabytes
		{"exact 1E", 1152921504606846976, "1.0E"}, // 1024^6
		{"3.7E", 4265267724775055360, "3.7E"},     // ~3.7EB
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatFileSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatFileSize(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestFormatFileSize_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		// Negative values (edge case - should handle gracefully)
		{"negative bytes", -1, "-1B"},
		{"negative large", -1024, "-1.0K"},

		// Very large values
		{"max int64", 9223372036854775807, "8.0E"}, // Close to max int64
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatFileSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatFileSize(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

// Test that covers specific decimal precision requirements
func TestFormatFileSize_DecimalPrecision(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		// Test that we get exactly 1 decimal place
		{"1.1K", 1126, "1.1K"},           // Should be 1.1, not 1.10
		{"2.0M", 2097152, "2.0M"},        // Should show .0 for exact values
		{"3.14159G", 3373259499, "3.1G"}, // Should round to 1 decimal
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatFileSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatFileSize(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}
