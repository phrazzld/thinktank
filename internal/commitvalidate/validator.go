// Package commitvalidate provides Go-native conventional commit validation
// This package eliminates external dependencies and provides consistent
// validation across all environments (local, CI, pre-commit hooks)
package commitvalidate

import (
	"fmt"
	"regexp"
	"strings"
)

// CommitMessage represents a parsed conventional commit
type CommitMessage struct {
	Type        string
	Scope       string
	Breaking    bool
	Description string
	Body        string
	FooterLines []string
	Raw         string
}

// ValidationResult contains the result of commit validation
type ValidationResult struct {
	Valid   bool
	Errors  []string
	Message *CommitMessage
}

// Validator validates conventional commit messages
type Validator struct {
	// AllowedTypes defines valid commit types
	AllowedTypes []string
	// MaxBodyLineLength defines maximum line length for body
	MaxBodyLineLength int
	// RequireFooterBlankLine requires blank line before footer
	RequireFooterBlankLine bool
}

// NewValidator creates a new validator with standard conventional commit rules
func NewValidator() *Validator {
	return &Validator{
		AllowedTypes: []string{
			"feat", "fix", "docs", "style", "refactor",
			"perf", "test", "chore", "ci", "build", "revert",
		},
		MaxBodyLineLength:      0, // 0 means no limit
		RequireFooterBlankLine: true,
	}
}

// conventionalCommitRegex matches the standard conventional commit format
// Format: <type>[optional scope]: <description>
var conventionalCommitRegex = regexp.MustCompile(`^([a-z]+)(\([a-z0-9/-]+\))?(!)?:\s(.+)`)

// Parse parses a commit message into structured components
func (v *Validator) Parse(message string) (*CommitMessage, error) {
	if message == "" {
		return nil, fmt.Errorf("commit message cannot be empty")
	}

	lines := strings.Split(strings.TrimSpace(message), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("commit message cannot be empty")
	}

	subject := lines[0]

	// Parse the subject line using regex
	matches := conventionalCommitRegex.FindStringSubmatch(subject)
	if len(matches) < 5 {
		return nil, fmt.Errorf("commit message does not follow conventional commit format")
	}

	commit := &CommitMessage{
		Type:        matches[1],
		Description: matches[4],
		Breaking:    matches[3] == "!",
		Raw:         message,
	}

	// Extract scope if present (matches[2] will be like "(scope)" or empty)
	if matches[2] != "" {
		scopeText := matches[2]
		// Remove parentheses
		commit.Scope = scopeText[1 : len(scopeText)-1]
	}

	// Parse body and footer if present
	if len(lines) > 1 {
		// Skip empty line after subject if present
		bodyStart := 1
		if bodyStart < len(lines) && lines[bodyStart] == "" {
			bodyStart = 2
		}

		if bodyStart < len(lines) {
			// Find footer (lines that contain ":")
			footerStart := -1
			for i := bodyStart; i < len(lines); i++ {
				if strings.Contains(lines[i], ":") && !strings.HasPrefix(lines[i], " ") {
					footerStart = i
					break
				}
			}

			if footerStart > bodyStart {
				commit.Body = strings.Join(lines[bodyStart:footerStart], "\n")
				commit.FooterLines = lines[footerStart:]
			} else if footerStart == -1 {
				commit.Body = strings.Join(lines[bodyStart:], "\n")
			} else {
				commit.FooterLines = lines[footerStart:]
			}
		}
	}

	return commit, nil
}

// Validate validates a commit message against conventional commit rules
func (v *Validator) Validate(message string) *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	// Parse the commit message
	commit, err := v.Parse(message)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
		return result
	}

	result.Message = commit

	// Validate type
	if !v.isValidType(commit.Type) {
		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("invalid commit type '%s'. Valid types: %s",
				commit.Type, strings.Join(v.AllowedTypes, ", ")))
	}

	// Validate scope format (if present)
	if commit.Scope != "" && !v.isValidScope(commit.Scope) {
		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("scope '%s' must contain only lowercase letters, numbers, hyphens, and slashes", commit.Scope))
	}

	// Validate description
	if err := v.validateDescription(commit.Description); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
	}

	// Validate body line length
	if commit.Body != "" {
		if err := v.validateBodyLineLength(commit.Body); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, err.Error())
		}
	}

	// Note: Footer validation is lenient for now
	// Most conventional commit tools don't strictly validate footer formats
	// and there are many valid variations (git trailers, GitHub formats, etc.)

	return result
}

// isValidType checks if the commit type is allowed
func (v *Validator) isValidType(commitType string) bool {
	for _, allowedType := range v.AllowedTypes {
		if commitType == allowedType {
			return true
		}
	}
	return false
}

// isValidScope checks if the scope format is valid (lowercase, alphanumeric, hyphens, slashes)
func (v *Validator) isValidScope(scope string) bool {
	validScopeRegex := regexp.MustCompile(`^[a-z0-9/-]+$`)
	return validScopeRegex.MatchString(scope)
}

// validateDescription validates the commit description
func (v *Validator) validateDescription(description string) error {
	if len(description) == 0 {
		return fmt.Errorf("description cannot be empty")
	}

	if strings.HasSuffix(description, ".") {
		return fmt.Errorf("description should not end with a period")
	}

	// Check that description starts with lowercase (unless it's an acronym)
	firstChar := description[0]
	if firstChar >= 'A' && firstChar <= 'Z' {
		// Allow acronyms/proper nouns at start, but warn about general convention
		words := strings.Fields(description)
		if len(words) > 0 {
			firstWord := words[0]
			// If it's all caps and reasonable length, it's likely an acronym (API, HTTP, etc.)
			if len(firstWord) <= 5 && strings.ToUpper(firstWord) == firstWord {
				// Allow common acronyms
				return nil
			}
			// Otherwise, it should start with lowercase
			return fmt.Errorf("description should start with lowercase letter")
		}
	}

	return nil
}

// validateBodyLineLength validates body line lengths
func (v *Validator) validateBodyLineLength(body string) error {
	// Skip validation if MaxBodyLineLength is 0 (no limit)
	if v.MaxBodyLineLength == 0 {
		return nil
	}

	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if len(line) > v.MaxBodyLineLength {
			return fmt.Errorf("body line %d exceeds maximum length of %d characters", i+1, v.MaxBodyLineLength)
		}
	}
	return nil
}
