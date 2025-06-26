package cli

import (
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadEnvironmentDefaults_BasicConfiguration tests the core functionality
func TestLoadEnvironmentDefaults_BasicConfiguration(t *testing.T) {
	testCases := []struct {
		name           string
		envVars        map[string]string
		expectedConfig func(*config.CliConfig) bool
		description    string
	}{
		{
			name: "THINKTANK_MODEL sets default model",
			envVars: map[string]string{
				"THINKTANK_MODEL": "gpt-4o",
			},
			expectedConfig: func(cfg *config.CliConfig) bool {
				return len(cfg.ModelNames) == 1 && cfg.ModelNames[0] == "gpt-4o"
			},
			description: "Should set model name from environment variable",
		},
		{
			name: "THINKTANK_OUTPUT_DIR sets output directory",
			envVars: map[string]string{
				"THINKTANK_OUTPUT_DIR": "/tmp/thinktank",
			},
			expectedConfig: func(cfg *config.CliConfig) bool {
				return cfg.OutputDir == "/tmp/thinktank"
			},
			description: "Should set output directory from environment variable",
		},
		{
			name: "THINKTANK_DRY_RUN enables dry run mode",
			envVars: map[string]string{
				"THINKTANK_DRY_RUN": "true",
			},
			expectedConfig: func(cfg *config.CliConfig) bool {
				return cfg.DryRun
			},
			description: "Should enable dry run mode from environment variable",
		},
		{
			name: "THINKTANK_VERBOSE enables verbose mode",
			envVars: map[string]string{
				"THINKTANK_VERBOSE": "1",
			},
			expectedConfig: func(cfg *config.CliConfig) bool {
				return cfg.Verbose
			},
			description: "Should enable verbose mode from environment variable",
		},
		{
			name: "THINKTANK_QUIET enables quiet mode",
			envVars: map[string]string{
				"THINKTANK_QUIET": "yes",
			},
			expectedConfig: func(cfg *config.CliConfig) bool {
				return cfg.Quiet
			},
			description: "Should enable quiet mode from environment variable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create base config
			cfg := config.NewDefaultCliConfig()

			// Create mock environment getter
			getenv := func(key string) string {
				return tc.envVars[key]
			}

			// Apply environment defaults - this should fail initially (RED phase)
			err := LoadEnvironmentDefaults(cfg, getenv)

			require.NoError(t, err, "LoadEnvironmentDefaults should not error")
			assert.True(t, tc.expectedConfig(cfg), tc.description)
		})
	}
}

// TestLoadEnvironmentDefaults_RateLimiting tests rate limiting environment variables
func TestLoadEnvironmentDefaults_RateLimiting(t *testing.T) {
	testCases := []struct {
		name           string
		envVars        map[string]string
		expectedConfig func(*config.CliConfig) bool
		description    string
	}{
		{
			name: "THINKTANK_RATE_LIMIT sets global rate limit",
			envVars: map[string]string{
				"THINKTANK_RATE_LIMIT": "1000",
			},
			expectedConfig: func(cfg *config.CliConfig) bool {
				return cfg.RateLimitRequestsPerMinute == 1000
			},
			description: "Should set global rate limit from environment variable",
		},
		{
			name: "THINKTANK_MAX_CONCURRENT sets concurrent requests",
			envVars: map[string]string{
				"THINKTANK_MAX_CONCURRENT": "8",
			},
			expectedConfig: func(cfg *config.CliConfig) bool {
				return cfg.MaxConcurrentRequests == 8
			},
			description: "Should set max concurrent requests from environment variable",
		},
		{
			name: "THINKTANK_TIMEOUT sets operation timeout",
			envVars: map[string]string{
				"THINKTANK_TIMEOUT": "5m",
			},
			expectedConfig: func(cfg *config.CliConfig) bool {
				return cfg.Timeout == 5*time.Minute
			},
			description: "Should set timeout from environment variable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.NewDefaultCliConfig()

			getenv := func(key string) string {
				return tc.envVars[key]
			}

			err := LoadEnvironmentDefaults(cfg, getenv)

			require.NoError(t, err, "LoadEnvironmentDefaults should not error")
			assert.True(t, tc.expectedConfig(cfg), tc.description)
		})
	}
}

// TestLoadEnvironmentDefaults_FilePatterns tests file pattern environment variables
func TestLoadEnvironmentDefaults_FilePatterns(t *testing.T) {
	testCases := []struct {
		name           string
		envVars        map[string]string
		expectedConfig func(*config.CliConfig) bool
		description    string
	}{
		{
			name: "THINKTANK_INCLUDE sets include patterns",
			envVars: map[string]string{
				"THINKTANK_INCLUDE": "*.go,*.ts",
			},
			expectedConfig: func(cfg *config.CliConfig) bool {
				return cfg.Include == "*.go,*.ts"
			},
			description: "Should set include patterns from environment variable",
		},
		{
			name: "THINKTANK_EXCLUDE sets exclude patterns",
			envVars: map[string]string{
				"THINKTANK_EXCLUDE": "*.test.go,node_modules",
			},
			expectedConfig: func(cfg *config.CliConfig) bool {
				return cfg.Exclude == "*.test.go,node_modules"
			},
			description: "Should set exclude patterns from environment variable",
		},
		{
			name: "THINKTANK_EXCLUDE_NAMES sets exclude name patterns",
			envVars: map[string]string{
				"THINKTANK_EXCLUDE_NAMES": ".git,dist",
			},
			expectedConfig: func(cfg *config.CliConfig) bool {
				return cfg.ExcludeNames == ".git,dist"
			},
			description: "Should set exclude name patterns from environment variable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.NewDefaultCliConfig()

			getenv := func(key string) string {
				return tc.envVars[key]
			}

			err := LoadEnvironmentDefaults(cfg, getenv)

			require.NoError(t, err, "LoadEnvironmentDefaults should not error")
			assert.True(t, tc.expectedConfig(cfg), tc.description)
		})
	}
}

// TestLoadEnvironmentDefaults_InvalidValues tests error handling for invalid environment values
func TestLoadEnvironmentDefaults_InvalidValues(t *testing.T) {
	testCases := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		errorSubstr string
		description string
	}{
		{
			name: "Invalid timeout format",
			envVars: map[string]string{
				"THINKTANK_TIMEOUT": "not-a-duration",
			},
			expectError: true,
			errorSubstr: "invalid timeout format",
			description: "Should return error for invalid timeout format",
		},
		{
			name: "Invalid rate limit value",
			envVars: map[string]string{
				"THINKTANK_RATE_LIMIT": "not-a-number",
			},
			expectError: true,
			errorSubstr: "invalid rate limit",
			description: "Should return error for invalid rate limit",
		},
		{
			name: "Invalid max concurrent value",
			envVars: map[string]string{
				"THINKTANK_MAX_CONCURRENT": "-5",
			},
			expectError: true,
			errorSubstr: "max concurrent requests must be positive",
			description: "Should return error for negative max concurrent value",
		},
		{
			name: "Valid boolean values accepted",
			envVars: map[string]string{
				"THINKTANK_DRY_RUN": "false",
				"THINKTANK_VERBOSE": "0",
				"THINKTANK_QUIET":   "",
			},
			expectError: false,
			description: "Should accept valid boolean values without error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.NewDefaultCliConfig()

			getenv := func(key string) string {
				return tc.envVars[key]
			}

			err := LoadEnvironmentDefaults(cfg, getenv)

			if tc.expectError {
				require.Error(t, err, tc.description)
				if tc.errorSubstr != "" {
					assert.Contains(t, err.Error(), tc.errorSubstr, "Error should contain expected substring")
				}
			} else {
				require.NoError(t, err, tc.description)
			}
		})
	}
}

// TestLoadEnvironmentDefaults_PrecedenceBehavior tests that environment variables are only applied as defaults
func TestLoadEnvironmentDefaults_PrecedenceBehavior(t *testing.T) {
	t.Run("Environment variables set defaults but don't override explicit values", func(t *testing.T) {
		cfg := config.NewDefaultCliConfig()
		// Pre-set some explicit values
		cfg.ModelNames = []string{"explicit-model"}
		cfg.OutputDir = "explicit-output"
		cfg.DryRun = true

		getenv := func(key string) string {
			switch key {
			case "THINKTANK_MODEL":
				return "env-model"
			case "THINKTANK_OUTPUT_DIR":
				return "env-output"
			case "THINKTANK_DRY_RUN":
				return "false"
			default:
				return ""
			}
		}

		err := LoadEnvironmentDefaults(cfg, getenv)

		require.NoError(t, err)
		// Explicit values should remain unchanged
		assert.Equal(t, []string{"explicit-model"}, cfg.ModelNames, "Should not override explicit model")
		assert.Equal(t, "explicit-output", cfg.OutputDir, "Should not override explicit output dir")
		assert.True(t, cfg.DryRun, "Should not override explicit dry run setting")
	})

	t.Run("Environment variables set defaults for empty values", func(t *testing.T) {
		cfg := config.NewDefaultCliConfig()
		// Ensure fields are empty/default
		cfg.ModelNames = nil
		cfg.OutputDir = ""
		cfg.DryRun = false

		getenv := func(key string) string {
			switch key {
			case "THINKTANK_MODEL":
				return "env-model"
			case "THINKTANK_OUTPUT_DIR":
				return "env-output"
			case "THINKTANK_DRY_RUN":
				return "true"
			default:
				return ""
			}
		}

		err := LoadEnvironmentDefaults(cfg, getenv)

		require.NoError(t, err)
		// Environment values should be applied as defaults
		assert.Equal(t, []string{"env-model"}, cfg.ModelNames, "Should set default model from env")
		assert.Equal(t, "env-output", cfg.OutputDir, "Should set default output dir from env")
		assert.True(t, cfg.DryRun, "Should set default dry run from env")
	})
}

// TestLoadEnvironmentDefaults_EmptyEnvironment tests behavior with no environment variables set
func TestLoadEnvironmentDefaults_EmptyEnvironment(t *testing.T) {
	cfg := config.NewDefaultCliConfig()
	originalModel := cfg.ModelNames
	originalOutput := cfg.OutputDir
	originalDryRun := cfg.DryRun

	getenv := func(key string) string {
		return "" // No environment variables set
	}

	err := LoadEnvironmentDefaults(cfg, getenv)

	require.NoError(t, err, "Should handle empty environment without error")
	assert.Equal(t, originalModel, cfg.ModelNames, "Should not change model names")
	assert.Equal(t, originalOutput, cfg.OutputDir, "Should not change output dir")
	assert.Equal(t, originalDryRun, cfg.DryRun, "Should not change dry run setting")
}

// TestBooleanConversion tests the boolean value conversion utility
func TestBooleanConversion(t *testing.T) {
	testCases := []struct {
		value    string
		expected bool
		name     string
	}{
		{"true", true, "true"},
		{"True", true, "True"},
		{"TRUE", true, "TRUE"},
		{"1", true, "1"},
		{"yes", true, "yes"},
		{"YES", true, "YES"},
		{"on", true, "on"},
		{"ON", true, "ON"},
		{"false", false, "false"},
		{"False", false, "False"},
		{"FALSE", false, "FALSE"},
		{"0", false, "0"},
		{"no", false, "no"},
		{"NO", false, "NO"},
		{"off", false, "off"},
		{"OFF", false, "OFF"},
		{"", false, "empty string"},
		{"invalid", false, "invalid value"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseBooleanEnvVar(tc.value)
			assert.Equal(t, tc.expected, result, "parseBooleanEnvVar(%q) should return %v", tc.value, tc.expected)
		})
	}
}

// TestStringSliceConversion tests the string slice conversion utility
func TestStringSliceConversion(t *testing.T) {
	testCases := []struct {
		value    string
		expected []string
		name     string
	}{
		{"", nil, "empty string"},
		{"single", []string{"single"}, "single value"},
		{"one,two,three", []string{"one", "two", "three"}, "comma separated"},
		{"one, two, three", []string{"one", "two", "three"}, "comma separated with spaces"},
		{" first , second , third ", []string{"first", "second", "third"}, "with extra spaces"},
		{",,", nil, "only commas"},
		{",value,", []string{"value"}, "leading and trailing commas"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseStringSliceEnvVar(tc.value)
			assert.Equal(t, tc.expected, result, "parseStringSliceEnvVar(%q) should return %v", tc.value, tc.expected)
		})
	}
}
