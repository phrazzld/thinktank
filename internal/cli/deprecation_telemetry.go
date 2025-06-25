package cli

import (
	"sort"
	"sync"
	"time"
)

// DeprecationTelemetry tracks usage patterns for deprecated CLI flags
// Following John Carmack's performance principles with efficient data structures
type DeprecationTelemetry struct {
	// Thread-safe usage counters using RWMutex for high read throughput
	mu    sync.RWMutex
	usage map[string]int

	// Pattern tracking for migration guide generation
	patterns map[string]*UsagePattern
}

// UsagePattern represents a specific usage pattern with metadata
type UsagePattern struct {
	Pattern   string    `json:"pattern"`
	Count     int       `json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
	Args      []string  `json:"args"`
}

// FlagUsage represents usage statistics for a single flag
type FlagUsage struct {
	Flag  string `json:"flag"`
	Count int    `json:"count"`
}

// NewDeprecationTelemetry creates a new telemetry collector
// Following Rob Pike's simplicity: minimal initialization, clear purpose
func NewDeprecationTelemetry() *DeprecationTelemetry {
	return &DeprecationTelemetry{
		usage:    make(map[string]int),
		patterns: make(map[string]*UsagePattern),
	}
}

// RecordUsage increments the usage counter for a deprecated flag
// Thread-safe with minimal lock contention using RWMutex
func (dt *DeprecationTelemetry) RecordUsage(flag string) {
	dt.mu.Lock()
	dt.usage[flag]++
	dt.mu.Unlock()
}

// RecordUsagePattern records a specific usage pattern for migration analysis
// This helps generate context-aware migration suggestions
func (dt *DeprecationTelemetry) RecordUsagePattern(pattern string, args []string) {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	existing, exists := dt.patterns[pattern]
	if exists {
		existing.Count++
		existing.LastSeen = time.Now()
	} else {
		dt.patterns[pattern] = &UsagePattern{
			Pattern:   pattern,
			Count:     1,
			FirstSeen: time.Now(),
			LastSeen:  time.Now(),
			Args:      make([]string, len(args)), // Deep copy to avoid mutations
		}
		copy(dt.patterns[pattern].Args, args)
	}
}

// GetUsageStats returns a copy of current usage statistics
// Uses RWMutex for high-throughput concurrent reads
func (dt *DeprecationTelemetry) GetUsageStats() map[string]int {
	dt.mu.RLock()
	defer dt.mu.RUnlock()

	// Return defensive copy to prevent external mutations
	result := make(map[string]int, len(dt.usage))
	for flag, count := range dt.usage {
		result[flag] = count
	}
	return result
}

// GetUsagePatterns returns all recorded usage patterns
// Useful for generating migration guides and analytics
func (dt *DeprecationTelemetry) GetUsagePatterns() []UsagePattern {
	dt.mu.RLock()
	defer dt.mu.RUnlock()

	// Return defensive copy of pattern data
	result := make([]UsagePattern, 0, len(dt.patterns))
	for _, pattern := range dt.patterns {
		// Create copy with defensive arg slice copy
		patternCopy := UsagePattern{
			Pattern:   pattern.Pattern,
			Count:     pattern.Count,
			FirstSeen: pattern.FirstSeen,
			LastSeen:  pattern.LastSeen,
			Args:      make([]string, len(pattern.Args)),
		}
		copy(patternCopy.Args, pattern.Args)
		result = append(result, patternCopy)
	}

	// Sort by usage frequency for prioritized migration guidance
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	return result
}

// GetMostCommonFlags returns the most frequently used deprecated flags
// Optimized for analytics and migration planning
func (dt *DeprecationTelemetry) GetMostCommonFlags(limit int) []FlagUsage {
	dt.mu.RLock()
	defer dt.mu.RUnlock()

	// Convert to slice for sorting
	flags := make([]FlagUsage, 0, len(dt.usage))
	for flag, count := range dt.usage {
		flags = append(flags, FlagUsage{
			Flag:  flag,
			Count: count,
		})
	}

	// Sort by usage count (descending)
	sort.Slice(flags, func(i, j int) bool {
		return flags[i].Count > flags[j].Count
	})

	// Return top N results
	if limit > len(flags) {
		limit = len(flags)
	}

	return flags[:limit]
}

// Reset clears all telemetry data
// Useful for testing and periodic data rotation
func (dt *DeprecationTelemetry) Reset() {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	// Clear maps efficiently by recreating them
	dt.usage = make(map[string]int)
	dt.patterns = make(map[string]*UsagePattern)
}

// GetTotalUsageCount returns the total number of deprecated flag usages
// Useful for overall migration analytics
func (dt *DeprecationTelemetry) GetTotalUsageCount() int {
	dt.mu.RLock()
	defer dt.mu.RUnlock()

	total := 0
	for _, count := range dt.usage {
		total += count
	}
	return total
}

// GetUniqueFlags returns the number of unique deprecated flags used
// Helps understand the breadth of deprecated flag usage
func (dt *DeprecationTelemetry) GetUniqueFlags() int {
	dt.mu.RLock()
	defer dt.mu.RUnlock()

	return len(dt.usage)
}
