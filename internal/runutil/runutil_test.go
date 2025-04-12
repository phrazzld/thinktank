package runutil

import (
	"regexp"
	"testing"
)

func TestGenerateRunName(t *testing.T) {
	// Run the test multiple times to ensure we get different values
	runs := 50
	generatedNames := make(map[string]bool, runs)

	// Compile the regex pattern once outside the loop
	pattern := regexp.MustCompile(`^[a-z]+-[a-z]+$`)

	for i := 0; i < runs; i++ {
		name := GenerateRunName()

		// Check that the name matches the expected pattern (adjective-noun)
		if !pattern.MatchString(name) {
			t.Errorf("Generated name %q does not match the expected format 'adjective-noun' with hyphen", name)
		}

		// Check that we're not always returning the same name
		if i > 0 && generatedNames[name] {
			// This is not a fatal error because there's a small chance of random collision
			// But we check many times to make sure we get different values
			t.Logf("Name %q was generated more than once", name)
		}
		generatedNames[name] = true
	}

	// Ensure we got at least a few different names (allowing for some collisions)
	if len(generatedNames) < runs/2 {
		t.Errorf("Expected to generate at least %d unique names, but only got %d", runs/2, len(generatedNames))
	}

	t.Logf("Generated %d unique names out of %d runs", len(generatedNames), runs)
}
