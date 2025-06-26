package cli

import (
	"strings"
	"sync"
)

// StringIntern provides string interning for commonly used strings to reduce memory allocations
type StringIntern struct {
	mu   sync.RWMutex
	pool map[string]string
}

// commonModelNames contains the most frequently used model names for interning
var commonModelNames = map[string]string{
	"gpt-4.1":          "gpt-4.1",
	"o4-mini":          "o4-mini",
	"o3":               "o3",
	"gemini-2.5-pro":   "gemini-2.5-pro",
	"gemini-2.5-flash": "gemini-2.5-flash",
	"openrouter/deepseek/deepseek-chat-v3-0324":      "openrouter/deepseek/deepseek-chat-v3-0324",
	"openrouter/deepseek/deepseek-r1-0528":           "openrouter/deepseek/deepseek-r1-0528",
	"openrouter/deepseek/deepseek-chat-v3-0324:free": "openrouter/deepseek/deepseek-chat-v3-0324:free",
	"openrouter/deepseek/deepseek-r1-0528:free":      "openrouter/deepseek/deepseek-r1-0528:free",
	"openrouter/meta-llama/llama-3.3-70b-instruct":   "openrouter/meta-llama/llama-3.3-70b-instruct",
	"openrouter/meta-llama/llama-4-maverick":         "openrouter/meta-llama/llama-4-maverick",
	"openrouter/meta-llama/llama-4-scout":            "openrouter/meta-llama/llama-4-scout",
	"openrouter/x-ai/grok-3-mini-beta":               "openrouter/x-ai/grok-3-mini-beta",
	"openrouter/x-ai/grok-3-beta":                    "openrouter/x-ai/grok-3-beta",
	"openrouter/google/gemma-3-27b-it":               "openrouter/google/gemma-3-27b-it",
}

// Global string interning instance
var globalStringIntern = &StringIntern{
	pool: make(map[string]string, len(commonModelNames)),
}

func init() {
	// Pre-populate the intern pool with common model names
	for name := range commonModelNames {
		globalStringIntern.pool[name] = name
	}
}

// Intern returns an interned version of the string to reduce memory usage
func (si *StringIntern) Intern(s string) string {
	if s == "" {
		return ""
	}

	// Fast path: check if we have this string already (read lock)
	si.mu.RLock()
	if interned, exists := si.pool[s]; exists {
		si.mu.RUnlock()
		return interned
	}
	si.mu.RUnlock()

	// Slow path: add new string (write lock)
	si.mu.Lock()
	defer si.mu.Unlock()

	// Double-check in case another goroutine added it
	if interned, exists := si.pool[s]; exists {
		return interned
	}

	// Add new string to pool
	si.pool[s] = s
	return s
}

// InternModelName returns an interned version of a model name
func InternModelName(name string) string {
	return globalStringIntern.Intern(name)
}

// Memory pools for temporary structures to reduce allocations

// StringSlicePool provides a pool of string slices for temporary use
var StringSlicePool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate with capacity for typical argument count
		return make([]string, 0, 16)
	},
}

// GetStringSlice returns a string slice from the pool
func GetStringSlice() []string {
	return StringSlicePool.Get().([]string)
}

// PutStringSlice returns a string slice to the pool after clearing it
func PutStringSlice(slice []string) {
	// Clear the slice but keep the capacity
	slice = slice[:0]
	//nolint:staticcheck // SA6002: slice pooling requires value type, not pointer
	StringSlicePool.Put(slice)
}

// ArgumentsCopyPool provides a pool for copying argument slices
var ArgumentsCopyPool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate with capacity for typical argument count
		return make([]string, 0, 32)
	},
}

// GetArgumentsCopy returns a slice from the pool for copying arguments
func GetArgumentsCopy() []string {
	return ArgumentsCopyPool.Get().([]string)
}

// PutArgumentsCopy returns an arguments slice to the pool after clearing it
func PutArgumentsCopy(slice []string) {
	// Clear the slice but keep the capacity
	slice = slice[:0]
	//nolint:staticcheck // SA6002: slice pooling requires value type, not pointer
	ArgumentsCopyPool.Put(slice)
}

// PatternPool provides a pool for deprecation pattern strings
var PatternPool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate with capacity for typical pattern length
		return make([]string, 0, 8)
	},
}

// GetPatternSlice returns a string slice from the pattern pool
func GetPatternSlice() []string {
	return PatternPool.Get().([]string)
}

// PutPatternSlice returns a pattern slice to the pool after clearing it
func PutPatternSlice(slice []string) {
	// Clear the slice but keep the capacity
	slice = slice[:0]
	//nolint:staticcheck // SA6002: slice pooling requires value type, not pointer
	PatternPool.Put(slice)
}

// StringBuilderPool provides a pool for string builders to reduce allocations
var StringBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// GetStringBuilder returns a string builder from the pool
func GetStringBuilder() *strings.Builder {
	return StringBuilderPool.Get().(*strings.Builder)
}

// PutStringBuilder returns a string builder to the pool after resetting it
func PutStringBuilder(sb *strings.Builder) {
	sb.Reset()
	StringBuilderPool.Put(sb)
}
