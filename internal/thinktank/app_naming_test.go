// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestGenerateTimestampedRunNameFormat verifies that the generateTimestampedRunName function
// produces names in the expected format: thinktank_YYYYMMDD_HHMMSS_NNNNNNN
func TestGenerateTimestampedRunNameFormat(t *testing.T) {
	// Generate a name
	name := generateTimestampedRunName()

	// Define the expected regex pattern
	pattern := `^thinktank_\d{8}_\d{6}_\d{7}$`
	re := regexp.MustCompile(pattern)

	// Verify the name matches the expected pattern
	if !re.MatchString(name) {
		t.Errorf("Generated name %q does not match the expected format %q", name, pattern)
	}

	// Verify the prefix
	if !strings.HasPrefix(name, "thinktank_") {
		t.Errorf("Generated name %q does not have the expected prefix 'thinktank_'", name)
	}

	// Extract and verify the components
	parts := strings.Split(name, "_")
	if len(parts) != 4 {
		t.Fatalf("Generated name %q does not have 4 parts separated by underscores", name)
	}

	dateStr := parts[1]   // YYYYMMDD
	timeStr := parts[2]   // HHMMSS
	uniqueStr := parts[3] // NNNNNNN

	// Verify the date part
	year, err := strconv.Atoi(dateStr[:4])
	if err != nil || year < 2000 || year > 2100 {
		t.Errorf("Year component %q of date %q is invalid", dateStr[:4], dateStr)
	}

	month, err := strconv.Atoi(dateStr[4:6])
	if err != nil || month < 1 || month > 12 {
		t.Errorf("Month component %q of date %q is invalid", dateStr[4:6], dateStr)
	}

	day, err := strconv.Atoi(dateStr[6:8])
	if err != nil || day < 1 || day > 31 {
		t.Errorf("Day component %q of date %q is invalid", dateStr[6:8], dateStr)
	}

	// Verify the time part
	hour, err := strconv.Atoi(timeStr[:2])
	if err != nil || hour < 0 || hour > 23 {
		t.Errorf("Hour component %q of time %q is invalid", timeStr[:2], timeStr)
	}

	minute, err := strconv.Atoi(timeStr[2:4])
	if err != nil || minute < 0 || minute > 59 {
		t.Errorf("Minute component %q of time %q is invalid", timeStr[2:4], timeStr)
	}

	second, err := strconv.Atoi(timeStr[4:6])
	if err != nil || second < 0 || second > 59 {
		t.Errorf("Second component %q of time %q is invalid", timeStr[4:6], timeStr)
	}

	// Verify the uniqueness component
	uniqueNum, err := strconv.Atoi(uniqueStr)
	if err != nil || uniqueNum < 0 || uniqueNum > 9999999 {
		t.Errorf("Uniqueness component %q is not a valid 7-digit number", uniqueStr)
	}

	// Verify the uniqueness component is formatted with exactly 7 digits (with leading zeros if needed)
	if len(uniqueStr) != 7 {
		t.Errorf("Uniqueness component %q does not have exactly 7 digits", uniqueStr)
	}
}

// TestGenerateTimestampedRunNameUniqueness verifies that consecutive calls
// to generateTimestampedRunName produce different results
func TestGenerateTimestampedRunNameUniqueness(t *testing.T) {
	// Generate multiple run names
	runs := 100 // Increase from 10 to 100 for more thorough testing
	generatedNames := make(map[string]bool, runs)

	for i := 0; i < runs; i++ {
		name := generateTimestampedRunName()

		// Check that we're not getting duplicate names
		if generatedNames[name] {
			t.Errorf("Name %q was generated more than once (iteration %d)", name, i)
			// Log all generated names to help diagnose the issue
			t.Logf("All generated names so far: %v", generatedNames)
		}

		generatedNames[name] = true

		// No need for a sleep delay anymore since our implementation now combines
		// nanoseconds, random numbers, and an incremental counter to ensure uniqueness
		// even within the same millisecond
	}

	// Verify we got unique names
	if len(generatedNames) < runs {
		t.Errorf("Expected %d unique names, but only got %d", runs, len(generatedNames))
	}
}
