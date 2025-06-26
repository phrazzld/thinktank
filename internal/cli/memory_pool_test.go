package cli

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStringIntern_BasicFunctionality tests basic string interning behavior
func TestStringIntern_BasicFunctionality(t *testing.T) {
	si := &StringIntern{
		pool: make(map[string]string),
	}

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"simple string", "test", "test"},
		{"model name", "gpt-4.1", "gpt-4.1"},
		{"long model name", "openrouter/deepseek/deepseek-chat-v3-0324", "openrouter/deepseek/deepseek-chat-v3-0324"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := si.Intern(tc.input)
			assert.Equal(t, tc.expected, result, "Interned string should match input")

			// Verify same string is returned for subsequent calls
			if tc.input != "" {
				result2 := si.Intern(tc.input)
				assert.Equal(t, result, result2, "Subsequent intern calls should return same string")
			}
		})
	}
}

// TestStringIntern_CommonModelNames tests interning of common model names
func TestStringIntern_CommonModelNames(t *testing.T) {
	commonModels := []string{
		"gpt-4.1",
		"o4-mini",
		"o3",
		"gemini-2.5-pro",
		"gemini-2.5-flash",
		"openrouter/deepseek/deepseek-chat-v3-0324",
	}

	for _, model := range commonModels {
		t.Run("model_"+model, func(t *testing.T) {
			result1 := InternModelName(model)
			result2 := InternModelName(model)

			assert.Equal(t, model, result1, "Interned model name should match input")
			assert.Equal(t, result1, result2, "Multiple calls should return same string")
		})
	}
}

// TestStringIntern_ThreadSafety tests concurrent access to string interning
func TestStringIntern_ThreadSafety(t *testing.T) {
	si := &StringIntern{
		pool: make(map[string]string),
	}

	const numGoroutines = 100
	const numOperations = 100

	var wg sync.WaitGroup
	testStrings := []string{"model1", "model2", "model3", "model4", "model5"}

	// Launch multiple goroutines that intern strings concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Use different strings to test concurrent map access
				testStr := testStrings[j%len(testStrings)]
				result := si.Intern(testStr)
				assert.Equal(t, testStr, result, "Interned string should match input")
			}
		}(i)
	}

	wg.Wait()

	// Verify all test strings are in the pool
	for _, str := range testStrings {
		result := si.Intern(str)
		assert.Equal(t, str, result, "String should be in pool after concurrent access")
	}
}

// TestStringSlicePool tests the string slice pool functionality
func TestStringSlicePool(t *testing.T) {
	t.Run("basic get/put", func(t *testing.T) {
		slice1 := GetStringSlice()
		require.NotNil(t, slice1, "Should get non-nil slice from pool")
		assert.Equal(t, 0, len(slice1), "Slice should be empty")
		assert.GreaterOrEqual(t, cap(slice1), 16, "Slice should have minimum capacity")

		// Use the slice
		slice1 = append(slice1, "test1", "test2")
		assert.Equal(t, 2, len(slice1), "Slice should contain added elements")

		// Return to pool
		PutStringSlice(slice1)

		// Get another slice (might be the same one, cleared)
		slice2 := GetStringSlice()
		require.NotNil(t, slice2, "Should get non-nil slice from pool")
		assert.Equal(t, 0, len(slice2), "Returned slice should be cleared")
	})

	t.Run("concurrent access", func(t *testing.T) {
		const numGoroutines = 50
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					slice := GetStringSlice()
					slice = append(slice, "test")
					PutStringSlice(slice)
				}
			}()
		}

		wg.Wait()
	})
}

// TestArgumentsCopyPool tests the arguments copy pool functionality
func TestArgumentsCopyPool(t *testing.T) {
	t.Run("basic get/put", func(t *testing.T) {
		slice1 := GetArgumentsCopy()
		require.NotNil(t, slice1, "Should get non-nil slice from pool")
		assert.Equal(t, 0, len(slice1), "Slice should be empty")
		assert.GreaterOrEqual(t, cap(slice1), 32, "Slice should have minimum capacity for arguments")

		// Use the slice
		slice1 = append(slice1, "thinktank", "--model", "gpt-4.1", "file.txt")
		assert.Equal(t, 4, len(slice1), "Slice should contain added elements")

		// Return to pool
		PutArgumentsCopy(slice1)

		// Get another slice
		slice2 := GetArgumentsCopy()
		require.NotNil(t, slice2, "Should get non-nil slice from pool")
		assert.Equal(t, 0, len(slice2), "Returned slice should be cleared")
	})
}

// TestPatternPool tests the pattern pool functionality
func TestPatternPool(t *testing.T) {
	t.Run("basic get/put", func(t *testing.T) {
		slice1 := GetPatternSlice()
		require.NotNil(t, slice1, "Should get non-nil slice from pool")
		assert.Equal(t, 0, len(slice1), "Slice should be empty")
		assert.GreaterOrEqual(t, cap(slice1), 8, "Slice should have minimum capacity for patterns")

		// Use the slice
		slice1 = append(slice1, "--model", "gpt-4.1", "--verbose")
		assert.Equal(t, 3, len(slice1), "Slice should contain added elements")

		// Return to pool
		PutPatternSlice(slice1)

		// Get another slice
		slice2 := GetPatternSlice()
		require.NotNil(t, slice2, "Should get non-nil slice from pool")
		assert.Equal(t, 0, len(slice2), "Returned slice should be cleared")
	})
}

// BenchmarkStringIntern benchmarks string interning performance
func BenchmarkStringIntern(b *testing.B) {
	si := &StringIntern{
		pool: make(map[string]string),
	}

	testStrings := []string{
		"gpt-4.1",
		"gemini-2.5-pro",
		"o4-mini",
		"openrouter/deepseek/deepseek-chat-v3-0324",
		"new-model-name",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		str := testStrings[i%len(testStrings)]
		_ = si.Intern(str)
	}
}

// BenchmarkStringIntern_GlobalPool benchmarks the global interning function
func BenchmarkStringIntern_GlobalPool(b *testing.B) {
	testStrings := []string{
		"gpt-4.1",
		"gemini-2.5-pro",
		"o4-mini",
		"openrouter/deepseek/deepseek-chat-v3-0324",
		"new-model-name",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		str := testStrings[i%len(testStrings)]
		_ = InternModelName(str)
	}
}

// BenchmarkStringSlicePool benchmarks string slice pool performance
func BenchmarkStringSlicePool(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		slice := GetStringSlice()
		slice = append(slice, "test1", "test2", "test3")
		PutStringSlice(slice)
	}
}

// BenchmarkStringSlicePool_vs_DirectAllocation compares pool vs direct allocation
func BenchmarkStringSlicePool_vs_DirectAllocation(b *testing.B) {
	b.Run("pool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			slice := GetStringSlice()
			slice = append(slice, "test1", "test2", "test3")
			PutStringSlice(slice)
		}
	})

	b.Run("direct", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			slice := make([]string, 0, 16)
			slice = append(slice, "test1", "test2", "test3")
			_ = slice // Simulate usage
		}
	})
}

// BenchmarkStringBuilderPool benchmarks string builder pool performance
func BenchmarkStringBuilderPool(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sb := GetStringBuilder()
		sb.WriteString("thinktank instructions.txt target_path --model ")
		sb.WriteString("gpt-4.1")
		_ = sb.String()
		PutStringBuilder(sb)
	}
}

// BenchmarkStringBuilderPool_vs_DirectAllocation compares pool vs direct allocation
func BenchmarkStringBuilderPool_vs_DirectAllocation(b *testing.B) {
	b.Run("pool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sb := GetStringBuilder()
			sb.WriteString("thinktank instructions.txt target_path --model ")
			sb.WriteString("gpt-4.1")
			_ = sb.String()
			PutStringBuilder(sb)
		}
	})

	b.Run("direct", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var sb strings.Builder
			sb.WriteString("thinktank instructions.txt target_path --model ")
			sb.WriteString("gpt-4.1")
			_ = sb.String()
		}
	})
}

// TestInternedModelNamesPreloaded tests that common model names are pre-interned
func TestInternedModelNamesPreloaded(t *testing.T) {
	// These should all be pre-loaded in the global intern pool
	preloadedModels := []string{
		"gpt-4.1",
		"o4-mini",
		"o3",
		"gemini-2.5-pro",
		"gemini-2.5-flash",
	}

	for _, model := range preloadedModels {
		t.Run("preloaded_"+model, func(t *testing.T) {
			// First call should return the pre-interned string
			result1 := InternModelName(model)
			assert.Equal(t, model, result1)

			// Second call should return the same string
			result2 := InternModelName(model)
			assert.Equal(t, result1, result2, "Pre-loaded model should return same string")
		})
	}
}

// TestStringBuilderPool tests the string builder pool functionality
func TestStringBuilderPool(t *testing.T) {
	t.Run("basic get/put", func(t *testing.T) {
		sb1 := GetStringBuilder()
		require.NotNil(t, sb1, "Should get non-nil string builder from pool")
		assert.Equal(t, 0, sb1.Len(), "String builder should be empty")

		// Use the string builder
		sb1.WriteString("test")
		sb1.WriteString(" string")
		assert.Equal(t, "test string", sb1.String())

		// Return to pool
		PutStringBuilder(sb1)

		// Get another string builder (might be the same one, reset)
		sb2 := GetStringBuilder()
		require.NotNil(t, sb2, "Should get non-nil string builder from pool")
		assert.Equal(t, 0, sb2.Len(), "Returned string builder should be reset")
	})

	t.Run("concurrent access", func(t *testing.T) {
		const numGoroutines = 50
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					sb := GetStringBuilder()
					sb.WriteString("test")
					PutStringBuilder(sb)
				}
			}()
		}

		wg.Wait()
	})
}

// TestMemoryPoolIntegration tests integration between different pool types
func TestMemoryPoolIntegration(t *testing.T) {
	// Simulate a parsing workflow that uses multiple pools
	args := GetArgumentsCopy()
	args = append(args, "thinktank", "--model", "gpt-4.1", "instructions.txt", "target/")

	// Extract model name and intern it
	var modelName string
	for i, arg := range args {
		if arg == "--model" && i+1 < len(args) {
			modelName = InternModelName(args[i+1])
			break
		}
	}

	// Use pattern pool for deprecation detection
	pattern := GetPatternSlice()
	pattern = append(pattern, "--model", modelName)

	// Use string builder pool for suggestion generation
	sb := GetStringBuilder()
	sb.WriteString("thinktank instructions.txt target_path --model ")
	sb.WriteString(modelName)
	suggestion := sb.String()

	// Verify everything works
	assert.Equal(t, "gpt-4.1", modelName)
	assert.Equal(t, []string{"--model", "gpt-4.1"}, pattern)
	assert.Equal(t, "thinktank instructions.txt target_path --model gpt-4.1", suggestion)

	// Clean up pools
	PutArgumentsCopy(args)
	PutPatternSlice(pattern)
	PutStringBuilder(sb)
}
