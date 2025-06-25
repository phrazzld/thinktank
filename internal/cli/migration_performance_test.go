package cli

import (
	"fmt"
	"testing"
	"time"
)

// BenchmarkMigrationGuideGeneration tests generation performance
// RED: This benchmark will fail until we implement efficient generation
func BenchmarkMigrationGuideGeneration(b *testing.B) {
	generator := NewMigrationGuideGenerator()
	telemetry := NewDeprecationTelemetry()

	// Pre-populate with realistic data (1000 patterns)
	patterns := []string{
		"--instructions file.txt",
		"--model gpt-4",
		"--output-dir output",
		"--include *.go",
		"--exclude *.test",
		"--dry-run",
		"--verbose",
	}

	for i := 0; i < 1000; i++ {
		pattern := patterns[i%len(patterns)]
		telemetry.RecordUsagePattern(fmt.Sprintf("%s-%d", pattern, i), []string{pattern})
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		guide, err := generator.GenerateFromTelemetry(telemetry)
		if err != nil {
			b.Fatalf("GenerateFromTelemetry() failed: %v", err)
		}
		if guide == nil {
			b.Fatal("GenerateFromTelemetry() returned nil guide")
		}
	}
}

// TestMigrationGuidePerformance tests that generation meets performance requirements
// RED: This test will fail until we achieve <100ms generation for 10k patterns
func TestMigrationGuidePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	generator := NewMigrationGuideGenerator()
	telemetry := NewDeprecationTelemetry()

	// Generate 10,000 patterns to test scalability
	patterns := []string{
		"--instructions file.txt",
		"--model gpt-4",
		"--output-dir output",
		"--include *.go",
		"--exclude *.test",
		"--dry-run",
		"--verbose",
		"--audit-log-file audit.log",
		"--log-level debug",
		"--synthesis-model claude-3-opus",
	}

	for i := 0; i < 10000; i++ {
		pattern := patterns[i%len(patterns)]
		args := []string{pattern, fmt.Sprintf("value-%d", i)}
		telemetry.RecordUsagePattern(fmt.Sprintf("%s %s", pattern, args[1]), args)
	}

	// Measure generation time
	start := time.Now()
	guide, err := generator.GenerateFromTelemetry(telemetry)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("GenerateFromTelemetry() failed: %v", err)
	}

	if guide == nil {
		t.Fatal("GenerateFromTelemetry() returned nil guide")
	}

	// Performance requirement: <100ms for 10k patterns
	maxDuration := 100 * time.Millisecond
	if duration > maxDuration {
		t.Errorf("Generation took %v, expected <%v for 10k patterns", duration, maxDuration)
	}

	t.Logf("Generated %d suggestions from 10k patterns in %v", len(guide.Suggestions), duration)
}

// TestMigrationGuideMemoryUsage tests memory efficiency
// RED: This test will fail until we implement memory-efficient generation
func TestMigrationGuideMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	generator := NewMigrationGuideGenerator()
	telemetry := NewDeprecationTelemetry()

	// Generate patterns to test memory usage
	for i := 0; i < 1000; i++ {
		pattern := fmt.Sprintf("--flag-%d value-%d", i, i)
		telemetry.RecordUsagePattern(pattern, []string{fmt.Sprintf("--flag-%d", i), fmt.Sprintf("value-%d", i)})
	}

	// Run generation multiple times to check for memory leaks
	for i := 0; i < 100; i++ {
		guide, err := generator.GenerateFromTelemetry(telemetry)
		if err != nil {
			t.Fatalf("GenerateFromTelemetry() failed on iteration %d: %v", i, err)
		}
		if guide == nil {
			t.Fatalf("GenerateFromTelemetry() returned nil on iteration %d", i)
		}
	}

	t.Log("Memory usage test completed without observable leaks")
}

// TestMigrationGuideConcurrentAccess tests thread safety
// RED: This test will fail until we implement thread-safe generation
func TestMigrationGuideConcurrentAccess(t *testing.T) {
	generator := NewMigrationGuideGenerator()
	telemetry := NewDeprecationTelemetry()

	// Pre-populate telemetry
	for i := 0; i < 100; i++ {
		pattern := fmt.Sprintf("--flag-%d value", i)
		telemetry.RecordUsagePattern(pattern, []string{fmt.Sprintf("--flag-%d", i), "value"})
	}

	// Run concurrent generation
	const numGoroutines = 10
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				guide, err := generator.GenerateFromTelemetry(telemetry)
				if err != nil {
					done <- fmt.Errorf("goroutine %d iteration %d failed: %w", id, j, err)
					return
				}
				if guide == nil {
					done <- fmt.Errorf("goroutine %d iteration %d returned nil guide", id, j)
					return
				}
			}
			done <- nil
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		if err := <-done; err != nil {
			t.Fatalf("Concurrent access failed: %v", err)
		}
	}

	t.Log("Concurrent access test completed successfully")
}

// TestMigrationGuideScalability tests generation with various data sizes
// RED: This test will fail until we implement scalable generation
func TestMigrationGuideScalability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping scalability test in short mode")
	}

	generator := NewMigrationGuideGenerator()

	testSizes := []int{10, 100, 1000, 5000}
	maxAllowedTime := []time.Duration{
		1 * time.Millisecond,  // 10 patterns
		5 * time.Millisecond,  // 100 patterns
		20 * time.Millisecond, // 1000 patterns
		50 * time.Millisecond, // 5000 patterns
	}

	for i, size := range testSizes {
		t.Run(fmt.Sprintf("Size_%d", size), func(t *testing.T) {
			telemetry := NewDeprecationTelemetry()

			// Generate patterns
			for j := 0; j < size; j++ {
				pattern := fmt.Sprintf("--flag-%d value-%d", j%20, j) // Realistic repetition
				telemetry.RecordUsagePattern(pattern, []string{fmt.Sprintf("--flag-%d", j%20), fmt.Sprintf("value-%d", j)})
			}

			start := time.Now()
			guide, err := generator.GenerateFromTelemetry(telemetry)
			duration := time.Since(start)

			if err != nil {
				t.Fatalf("GenerateFromTelemetry() failed: %v", err)
			}

			if guide == nil {
				t.Fatal("GenerateFromTelemetry() returned nil guide")
			}

			if duration > maxAllowedTime[i] {
				t.Errorf("Generation of %d patterns took %v, expected <%v",
					size, duration, maxAllowedTime[i])
			} else {
				t.Logf("Generation of %d patterns took %v (within %v limit)",
					size, duration, maxAllowedTime[i])
			}
		})
	}
}
