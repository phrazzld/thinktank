package fileutil

import (
	"path/filepath"
	"slices"
	"strings"
	"unicode"
)

// FileFilterResult represents the result of file filtering operations
type FileFilterResult struct {
	ShouldProcess bool
	Reason        string
	FileType      string
}

// FilteringOptions contains pure filtering configuration without I/O dependencies
type FilteringOptions struct {
	IncludeExts  []string
	ExcludeExts  []string
	ExcludeNames []string
	// Git-related filtering options
	IgnoreHidden   bool
	IgnoreGitFiles bool
	GitIgnoreRules []string // For future expansion - explicit git ignore rules
}

// FilterFiles applies all filtering rules to a list of file paths and returns
// the paths that should be processed. This is a pure function with no side effects.
func FilterFiles(paths []string, opts FilteringOptions) []string {
	var filtered []string

	for _, path := range paths {
		if ShouldProcessFile(path, opts).ShouldProcess {
			filtered = append(filtered, path)
		}
	}

	return filtered
}

// ShouldProcessFile determines if a file should be processed based on filtering rules.
// This is a pure function that doesn't perform I/O or logging operations.
func ShouldProcessFile(path string, opts FilteringOptions) FileFilterResult {
	base := filepath.Base(path)
	ext := strings.ToLower(filepath.Ext(path))

	// Check if explicitly excluded by name
	if len(opts.ExcludeNames) > 0 && slices.Contains(opts.ExcludeNames, base) {
		return FileFilterResult{
			ShouldProcess: false,
			Reason:        "excluded by name",
			FileType:      ClassifyFileByExtension(ext),
		}
	}

	// Check if hidden file (when enabled)
	if opts.IgnoreHidden && IsHiddenPath(path) {
		return FileFilterResult{
			ShouldProcess: false,
			Reason:        "hidden file or directory",
			FileType:      ClassifyFileByExtension(ext),
		}
	}

	// Check if git-related file (when enabled)
	if opts.IgnoreGitFiles && IsGitRelatedPath(path) {
		return FileFilterResult{
			ShouldProcess: false,
			Reason:        "git-related file",
			FileType:      ClassifyFileByExtension(ext),
		}
	}

	// Check include extensions (if specified)
	if len(opts.IncludeExts) > 0 {
		included := false
		for _, includeExt := range opts.IncludeExts {
			if ext == strings.ToLower(includeExt) {
				included = true
				break
			}
		}
		if !included {
			return FileFilterResult{
				ShouldProcess: false,
				Reason:        "extension not in include list",
				FileType:      ClassifyFileByExtension(ext),
			}
		}
	}

	// Check exclude extensions
	if len(opts.ExcludeExts) > 0 {
		for _, excludeExt := range opts.ExcludeExts {
			if ext == strings.ToLower(excludeExt) {
				return FileFilterResult{
					ShouldProcess: false,
					Reason:        "extension in exclude list",
					FileType:      ClassifyFileByExtension(ext),
				}
			}
		}
	}

	return FileFilterResult{
		ShouldProcess: true,
		Reason:        "passed all filters",
		FileType:      ClassifyFileByExtension(ext),
	}
}

// IsHiddenPath determines if a path represents a hidden file or directory.
// A path is considered hidden if any component starts with a dot (except . and ..).
func IsHiddenPath(path string) bool {
	// Check each component of the path
	parts := strings.Split(filepath.Clean(path), string(filepath.Separator))
	for _, part := range parts {
		if strings.HasPrefix(part, ".") && part != "." && part != ".." {
			return true
		}
	}
	return false
}

// IsGitRelatedPath determines if a path is related to git version control.
func IsGitRelatedPath(path string) bool {
	base := filepath.Base(path)

	// Check if it's the .git directory itself
	if base == ".git" {
		return true
	}

	// Check if the path contains .git directory (handle both "/" and "\" separators)
	if strings.Contains(path, string(filepath.Separator)+".git"+string(filepath.Separator)) {
		return true
	}

	// Check if path starts with .git/ (for cases like ".git/HEAD")
	if strings.HasPrefix(path, ".git"+string(filepath.Separator)) {
		return true
	}

	// Check for git-related files
	gitFiles := []string{".gitignore", ".gitattributes", ".gitmodules", ".gitkeep"}
	for _, gitFile := range gitFiles {
		if base == gitFile {
			return true
		}
	}

	return false
}

// ClassifyFileByExtension classifies a file based on its extension.
func ClassifyFileByExtension(ext string) string {
	ext = strings.ToLower(ext)

	// Default case for no extension
	if ext == "" {
		return "no_extension"
	}

	// Programming languages
	programmingExts := map[string]string{
		".go":    "go",
		".py":    "python",
		".js":    "javascript",
		".ts":    "typescript",
		".java":  "java",
		".cpp":   "cpp",
		".c":     "c",
		".rs":    "rust",
		".rb":    "ruby",
		".php":   "php",
		".cs":    "csharp",
		".swift": "swift",
		".kt":    "kotlin",
		".scala": "scala",
	}

	if lang, exists := programmingExts[ext]; exists {
		return lang
	}

	// Configuration and data files
	configExts := map[string]string{
		".json": "json",
		".yaml": "yaml",
		".yml":  "yaml",
		".toml": "toml",
		".xml":  "xml",
		".ini":  "config",
		".conf": "config",
		".cfg":  "config",
		".env":  "config",
	}

	if configType, exists := configExts[ext]; exists {
		return configType
	}

	// Documentation
	docExts := map[string]string{
		".md":   "markdown",
		".txt":  "text",
		".rst":  "restructuredtext",
		".adoc": "asciidoc",
	}

	if docType, exists := docExts[ext]; exists {
		return docType
	}

	// Web files
	webExts := map[string]string{
		".html": "html",
		".css":  "css",
		".scss": "scss",
		".sass": "sass",
		".less": "less",
	}

	if webType, exists := webExts[ext]; exists {
		return webType
	}

	// Shell scripts
	if ext == ".sh" || ext == ".bash" || ext == ".zsh" || ext == ".fish" {
		return "shell"
	}

	return "other"
}

// ValidateFilePath validates and normalizes a file path for processing.
func ValidateFilePath(path string) (string, bool) {
	if path == "" {
		return "", false
	}

	// Clean the path to remove redundant elements
	cleaned := filepath.Clean(path)

	// Check for invalid characters or patterns
	if strings.Contains(cleaned, "..") {
		return "", false // Path traversal attempt
	}

	return cleaned, true
}

// FileStatistics contains comprehensive file statistics
type FileStatistics struct {
	CharCount         int
	LineCount         int
	WordCount         int
	TokenCount        int
	BlankLineCount    int
	NonBlankLines     int
	AverageLineLength float64
}

// CalculateFileStatistics calculates comprehensive statistics for file content.
// This is an enhanced version of the original CalculateStatistics function.
func CalculateFileStatistics(content string) FileStatistics {
	if content == "" {
		return FileStatistics{}
	}

	lines := strings.Split(content, "\n")
	lineCount := len(lines)
	charCount := len(content)

	// Count words and tokens
	wordCount := 0
	tokenCount := 0
	blankLineCount := 0
	nonBlankLines := 0
	totalLineLength := 0

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			blankLineCount++
		} else {
			nonBlankLines++
			totalLineLength += len(line)
		}

		// Count words in this line
		words := strings.Fields(line)
		wordCount += len(words)
	}

	// Estimate tokens using a more sophisticated method
	tokenCount = estimateAdvancedTokenCount(content)

	// Calculate average line length
	averageLineLength := 0.0
	if nonBlankLines > 0 {
		averageLineLength = float64(totalLineLength) / float64(nonBlankLines)
	}

	return FileStatistics{
		CharCount:         charCount,
		LineCount:         lineCount,
		WordCount:         wordCount,
		TokenCount:        tokenCount,
		BlankLineCount:    blankLineCount,
		NonBlankLines:     nonBlankLines,
		AverageLineLength: averageLineLength,
	}
}

// estimateAdvancedTokenCount provides a more sophisticated token count estimation
// that considers programming language constructs and punctuation.
func estimateAdvancedTokenCount(text string) int {
	if text == "" {
		return 0
	}

	count := 0
	inToken := false
	inString := false
	stringChar := byte(0)

	for i, r := range text {
		b := byte(r)

		// Handle string literals
		if !inString && (b == '"' || b == '\'' || b == '`') {
			inString = true
			stringChar = b
			if inToken {
				count++
				inToken = false
			}
			count++ // Count the string as a token
			continue
		}

		if inString {
			if b == stringChar && (i == 0 || text[i-1] != '\\') {
				inString = false
				stringChar = 0
			}
			continue
		}

		// Handle token boundaries
		if unicode.IsSpace(r) || isPunctuation(r) {
			if inToken {
				count++
				inToken = false
			}
			// Count significant punctuation as tokens
			if isPunctuation(r) && !unicode.IsSpace(r) {
				count++
			}
		} else {
			inToken = true
		}
	}

	// Count final token if we ended in one
	if inToken {
		count++
	}

	return count
}

// isPunctuation checks if a rune is significant punctuation for token counting
func isPunctuation(r rune) bool {
	switch r {
	case '(', ')', '[', ']', '{', '}', ';', ',', '.', ':', '=', '+', '-', '*', '/', '%', '&', '|', '^', '!', '?', '<', '>', '~':
		return true
	default:
		return false
	}
}

// CreateFilteringOptions creates FilteringOptions from the legacy Config struct.
// This helps bridge the gap between the old I/O-mixed code and the new pure functions.
func CreateFilteringOptions(config *Config) FilteringOptions {
	return FilteringOptions{
		IncludeExts:    config.IncludeExts,
		ExcludeExts:    config.ExcludeExts,
		ExcludeNames:   config.ExcludeNames,
		IgnoreHidden:   true, // Default behavior
		IgnoreGitFiles: true, // Default behavior
	}
}
