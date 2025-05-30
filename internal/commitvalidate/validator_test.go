package commitvalidate

import (
	"testing"
)

func TestValidator_Parse_ValidCommits(t *testing.T) {
	validator := NewValidator()
	tests := []struct {
		name             string
		message          string
		expectedType     string
		expectedScope    string
		expectedBreaking bool
		expectedDesc     string
	}{
		{
			name:         "simple feat commit",
			message:      "feat: add new feature",
			expectedType: "feat",
			expectedDesc: "add new feature",
		},
		{
			name:          "feat with scope",
			message:       "feat(api): add new endpoint",
			expectedType:  "feat",
			expectedScope: "api",
			expectedDesc:  "add new endpoint",
		},
		{
			name:             "breaking change",
			message:          "feat!: breaking change",
			expectedType:     "feat",
			expectedBreaking: true,
			expectedDesc:     "breaking change",
		},
		{
			name:             "breaking change with scope",
			message:          "feat(api)!: breaking change",
			expectedType:     "feat",
			expectedScope:    "api",
			expectedBreaking: true,
			expectedDesc:     "breaking change",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commit, err := validator.Parse(tt.message)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if commit.Type != tt.expectedType {
				t.Errorf("expected type %s, got %s", tt.expectedType, commit.Type)
			}
			if commit.Scope != tt.expectedScope {
				t.Errorf("expected scope %s, got %s", tt.expectedScope, commit.Scope)
			}
			if commit.Breaking != tt.expectedBreaking {
				t.Errorf("expected breaking %v, got %v", tt.expectedBreaking, commit.Breaking)
			}
			if commit.Description != tt.expectedDesc {
				t.Errorf("expected description %s, got %s", tt.expectedDesc, commit.Description)
			}
		})
	}
}

func TestValidator_Parse_InvalidCommits(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		message string
	}{
		{"empty message", ""},
		{"invalid format - no colon", "feat add new feature"},
		{"invalid format - no description", "feat:"},
		{"invalid format - uppercase type", "FEAT: add new feature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.Parse(tt.message)
			if err == nil {
				t.Errorf("expected error but got none")
			}
		})
	}
}

func TestValidator_Validate_ValidCommits(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		message string
	}{
		{"valid feat commit", "feat: add new feature"},
		{"valid fix commit with scope", "fix(auth): resolve login issue"},
		{"valid breaking change", "feat!: breaking api change"},
		{"valid commit with body", "feat: add new feature\n\nThis is a longer description of the feature."},
		{"valid acronym at start", "feat: API endpoint added"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.message)
			if !result.Valid {
				t.Errorf("expected valid commit, got errors: %v", result.Errors)
			}
		})
	}
}

func TestValidator_Validate_InvalidCommits(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name       string
		message    string
		errorCount int
	}{
		{"invalid type", "invalid: bad commit type", 1},
		{"invalid scope format", "feat(Invalid-Scope): bad scope", 1},
		{"description with period", "feat: add new feature.", 1},
		{"description starting with uppercase", "feat: Add new feature", 1},
		{"empty description", "feat: ", 1},
		{"multiple errors", "invalid(BAD-scope): Bad description.", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.message)
			if result.Valid {
				t.Errorf("expected invalid commit")
			}
			if len(result.Errors) != tt.errorCount {
				t.Errorf("expected %d errors, got %d: %v", tt.errorCount, len(result.Errors), result.Errors)
			}
		})
	}
}

func TestValidator_ValidateBodyLineLength(t *testing.T) {
	validator := NewValidator()
	validator.MaxBodyLineLength = 50 // Set a shorter limit for testing

	tests := []struct {
		name        string
		message     string
		expectValid bool
	}{
		{
			name:        "body within limit",
			message:     "feat: add feature\n\nShort body line",
			expectValid: true,
		},
		{
			name:        "body exceeds limit",
			message:     "feat: add feature\n\nThis is a very long body line that exceeds the maximum allowed length",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.message)

			if result.Valid != tt.expectValid {
				t.Errorf("expected valid %v, got %v. Errors: %v", tt.expectValid, result.Valid, result.Errors)
			}
		})
	}
}

func TestValidator_AllowedTypes(t *testing.T) {
	validator := NewValidator()

	// Test all standard allowed types
	allowedTypes := []string{
		"feat", "fix", "docs", "style", "refactor",
		"perf", "test", "chore", "ci", "build", "revert",
	}

	for _, commitType := range allowedTypes {
		t.Run("valid_type_"+commitType, func(t *testing.T) {
			message := commitType + ": test message"
			result := validator.Validate(message)

			if !result.Valid {
				t.Errorf("type %s should be valid, got errors: %v", commitType, result.Errors)
			}
		})
	}

	// Test invalid types
	invalidTypes := []string{"invalid", "bad", "wrong", "feature", "bugfix"}

	for _, commitType := range invalidTypes {
		t.Run("invalid_type_"+commitType, func(t *testing.T) {
			message := commitType + ": test message"
			result := validator.Validate(message)

			if result.Valid {
				t.Errorf("type %s should be invalid", commitType)
			}
		})
	}
}

func TestValidator_ScopeValidation(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		scope       string
		expectValid bool
	}{
		{"valid lowercase", "api", true},
		{"valid with numbers", "api2", true},
		{"valid with hyphens", "api-v2", true},
		{"valid with slashes", "api/v2", true},
		{"valid complex", "api-v2/auth", true},
		{"invalid uppercase", "API", false},
		{"invalid mixed case", "Api", false},
		{"invalid spaces", "api auth", false},
		{"invalid special chars", "api@v2", false},
		{"invalid dots", "api.v2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := "feat(" + tt.scope + "): test message"
			result := validator.Validate(message)

			if tt.expectValid && !result.Valid {
				t.Errorf("scope %s should be valid, got errors: %v", tt.scope, result.Errors)
			} else if !tt.expectValid && result.Valid {
				t.Errorf("scope %s should be invalid", tt.scope)
			}
		})
	}
}

func TestValidator_ComplexCommitMessages(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		message     string
		expectValid bool
	}{
		{
			name: "complex valid commit",
			message: `feat(api): add user authentication

This commit adds JWT-based authentication to the API.
It includes login, logout, and token refresh endpoints.

Closes #123
Reviewed-by: John Doe`,
			expectValid: true,
		},
		{
			name: "breaking change with body",
			message: `feat(api)!: remove deprecated endpoints

BREAKING CHANGE: The /v1/old-endpoint has been removed.
Use /v2/new-endpoint instead.`,
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.message)

			if result.Valid != tt.expectValid {
				t.Errorf("expected valid %v, got %v. Errors: %v", tt.expectValid, result.Valid, result.Errors)
			}
		})
	}
}
