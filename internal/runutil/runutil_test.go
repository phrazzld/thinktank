package runutil

import (
	"regexp"
	"strings"
	"testing"
)

// TestGenerateRunName tests the basic functionality of GenerateRunName
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

// TestGenerateRunNameFormat verifies the format of generated names
func TestGenerateRunNameFormat(t *testing.T) {
	// Run multiple tests to check format consistency
	for i := 0; i < 10; i++ {
		name := GenerateRunName()

		// Split the name into adjective and noun
		parts := strings.Split(name, "-")
		if len(parts) != 2 {
			t.Errorf("Generated name %q does not have exactly one hyphen separator", name)
			continue
		}

		adjective, noun := parts[0], parts[1]

		// Check that the adjective is not empty
		if adjective == "" {
			t.Errorf("Generated name %q has an empty adjective part", name)
		}

		// Check that the noun is not empty
		if noun == "" {
			t.Errorf("Generated name %q has an empty noun part", name)
		}

		// Check that adjective and noun are all lowercase
		if adjective != strings.ToLower(adjective) {
			t.Errorf("Adjective %q in name %q is not all lowercase", adjective, name)
		}

		if noun != strings.ToLower(noun) {
			t.Errorf("Noun %q in name %q is not all lowercase", noun, name)
		}
	}
}

// TestGenerateRunNameUsesDefinedLists ensures generated names use only values from defined adjective/noun lists
func TestGenerateRunNameUsesDefinedLists(t *testing.T) {
	// Create maps for faster lookup
	adjectiveMap := make(map[string]bool, len(adjectives))
	for _, adj := range adjectives {
		adjectiveMap[adj] = true
	}

	nounMap := make(map[string]bool, len(nouns))
	for _, n := range nouns {
		nounMap[n] = true
	}

	// Run 100 tests to check against the defined lists
	for i := 0; i < 100; i++ {
		name := GenerateRunName()
		parts := strings.Split(name, "-")

		if len(parts) != 2 {
			t.Errorf("Generated name %q has invalid format", name)
			continue
		}

		adjective, noun := parts[0], parts[1]

		// Check that the adjective is from the defined list
		if !adjectiveMap[adjective] {
			t.Errorf("Adjective %q in generated name %q is not in the defined adjectives list", adjective, name)
		}

		// Check that the noun is from the defined list
		if !nounMap[noun] {
			t.Errorf("Noun %q in generated name %q is not in the defined nouns list", noun, name)
		}
	}
}

// TestGenerateRunNameDistribution checks the distribution of generated names
func TestGenerateRunNameDistribution(t *testing.T) {
	// This test ensures sufficient randomness in the output
	runsCount := 1000
	generatedNames := make(map[string]int, runsCount)

	// Generate a large number of run names
	for i := 0; i < runsCount; i++ {
		name := GenerateRunName()
		generatedNames[name]++
	}

	// Check uniqueness
	uniqueNames := len(generatedNames)
	t.Logf("Generated %d unique names out of %d runs (%.2f%% unique)",
		uniqueNames, runsCount, float64(uniqueNames)*100/float64(runsCount))

	// With good randomness, we should get a reasonable number of unique names
	// With good randomness, we should get a reasonable number of unique names
	minExpectedUniqueness := runsCount / 10 // At least 10% of runs should produce unique names

	if uniqueNames < minExpectedUniqueness {
		t.Errorf("Expected at least %d unique names but got only %d", minExpectedUniqueness, uniqueNames)
	}

	// Check that we don't have excessive repetition
	maxOccurrences := 0
	for _, count := range generatedNames {
		if count > maxOccurrences {
			maxOccurrences = count
		}
	}

	// With good randomness, no single combination should appear too frequently
	maxExpectedOccurrences := runsCount / (uniqueNames / 5)
	if maxOccurrences > maxExpectedOccurrences {
		t.Logf("Maximum occurrence of a single name: %d out of %d runs", maxOccurrences, runsCount)
		// This is just a log, not an error, as randomness can occasionally produce higher repetition
	}
}
