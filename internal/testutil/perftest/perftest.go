// Package perftest provides utilities for CI-aware performance testing
package perftest

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

// Environment represents the testing environment characteristics
type Environment struct {
	IsCI           bool
	IsRaceEnabled  bool
	CPUCount       int
	OSArchitecture string
	RunnerType     string // e.g., "github-actions", "local", etc.
}

// Config holds performance test configuration adjusted for the environment
type Config struct {
	Environment Environment
	// Multipliers applied to base thresholds
	ThroughputMultiplier float64
	TimeoutMultiplier    float64
	MemoryMultiplier     float64
}

// DetectEnvironment determines the current testing environment
func DetectEnvironment() Environment {
	env := Environment{
		IsCI:           isCI(),
		IsRaceEnabled:  raceDetectionEnabled(),
		CPUCount:       runtime.NumCPU(),
		OSArchitecture: fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	// Detect specific CI runner type
	if env.IsCI {
		switch {
		case os.Getenv("GITHUB_ACTIONS") == "true":
			env.RunnerType = "github-actions"
		case os.Getenv("GITLAB_CI") != "":
			env.RunnerType = "gitlab-ci"
		case os.Getenv("CIRCLECI") == "true":
			env.RunnerType = "circleci"
		default:
			env.RunnerType = "generic-ci"
		}
	} else {
		env.RunnerType = "local"
	}

	return env
}

// NewConfig creates a performance test configuration for the current environment
func NewConfig() *Config {
	env := DetectEnvironment()
	cfg := &Config{
		Environment:          env,
		ThroughputMultiplier: 1.0,
		TimeoutMultiplier:    1.0,
		MemoryMultiplier:     1.0,
	}

	// Adjust multipliers based on environment
	if env.IsCI {
		// CI environments typically have:
		// - Shared resources (slower)
		// - More network latency
		// - Less predictable performance
		cfg.ThroughputMultiplier = 0.7 // Expect 30% lower throughput
		cfg.TimeoutMultiplier = 2.0    // Double timeouts
		cfg.MemoryMultiplier = 1.2     // Allow 20% more memory usage
	}

	if env.IsRaceEnabled {
		// Race detector adds significant overhead
		cfg.ThroughputMultiplier *= 0.5 // 50% reduction
		cfg.TimeoutMultiplier *= 2.0    // Double again
		cfg.MemoryMultiplier *= 3.0     // Race detector uses more memory
	}

	// Further adjustments based on specific CI runners
	switch env.RunnerType {
	case "github-actions":
		// GitHub Actions runners are known to have variable performance
		cfg.ThroughputMultiplier *= 0.9
	}

	return cfg
}

// AdjustThroughput adjusts a base throughput expectation for the environment
func (c *Config) AdjustThroughput(baseThroughput float64) float64 {
	return baseThroughput * c.ThroughputMultiplier
}

// AdjustTimeout adjusts a base timeout for the environment
func (c *Config) AdjustTimeout(baseTimeout time.Duration) time.Duration {
	adjusted := time.Duration(float64(baseTimeout) * c.TimeoutMultiplier)
	// Ensure minimum timeout
	if adjusted < 60*time.Second {
		adjusted = 60 * time.Second
	}
	return adjusted
}

// AdjustMemory adjusts a base memory expectation for the environment
func (c *Config) AdjustMemory(baseMemory int64) int64 {
	return int64(float64(baseMemory) * c.MemoryMultiplier)
}

// ShouldSkip determines if a performance test should be skipped in this environment
func (c *Config) ShouldSkip(testType string) (bool, string) {
	switch testType {
	case "heavy-cpu":
		if c.Environment.CPUCount < 4 {
			return true, fmt.Sprintf("Test requires at least 4 CPUs, found %d", c.Environment.CPUCount)
		}
	case "race-sensitive":
		if c.Environment.IsRaceEnabled {
			return true, "Test is not compatible with race detector"
		}
	case "local-only":
		if c.Environment.IsCI {
			return true, "Test is designed for local development only"
		}
	}
	return false, ""
}

// Helper functions

func isCI() bool {
	// Check common CI environment variables
	ciVars := []string{
		"CI",
		"CONTINUOUS_INTEGRATION",
		"BUILD_ID",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"CIRCLECI",
		"JENKINS_URL",
		"BUILDKITE",
	}

	for _, v := range ciVars {
		if val := os.Getenv(v); val == "true" || val == "1" || val != "" {
			return true
		}
	}
	return false
}

func raceDetectionEnabled() bool {
	// Primary method: Use build tags (most reliable)
	if raceDetectionEnabledByBuildTag() {
		return true
	}

	// Fallback: Check environment variable for manual override
	if os.Getenv("RACE_ENABLED") == "true" {
		return true
	}

	// Fallback: Check GORACE environment variable (set when race detector is active)
	if os.Getenv("GORACE") != "" {
		return true
	}

	// Fallback: Check command line arguments (may not work in all cases)
	for _, arg := range os.Args {
		if strings.Contains(arg, "-race") || strings.Contains(arg, "-test.race") {
			return true
		}
	}

	return false
}
