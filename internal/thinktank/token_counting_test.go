package thinktank

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED: Write the smallest failing test first
func TestTokenCountingService_CountsEmptyContext(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()

	result, err := service.CountTokens(context.Background(), TokenCountingRequest{
		Instructions: "",
		Files:        []FileContent{},
	})

	require.NoError(t, err)
	assert.Equal(t, 0, result.TotalTokens)
}

// Second RED: Add test for single file counting
func TestTokenCountingService_CountsSingleFile(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()

	result, err := service.CountTokens(context.Background(), TokenCountingRequest{
		Instructions: "analyze this",
		Files: []FileContent{
			{Path: "test.go", Content: "package main\nfunc main() {}"},
		},
	})

	require.NoError(t, err)
	assert.Greater(t, result.TotalTokens, 0)
	assert.Greater(t, result.InstructionTokens, 0)
	assert.Greater(t, result.FileTokens, 0)

	// Verify components sum to total
	expected := result.InstructionTokens + result.FileTokens + result.Overhead
	assert.Equal(t, expected, result.TotalTokens)
}

// Third RED: Add table-driven test for comprehensive scenarios
func TestTokenCountingService_TableDriven(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()

	tests := []struct {
		name         string
		instructions string
		files        []FileContent
		expectTokens int
	}{
		{
			name:         "empty context",
			instructions: "",
			files:        []FileContent{},
			expectTokens: 0,
		},
		{
			name:         "instructions only",
			instructions: "test", // 4 chars * 0.75 + 1000 overhead = 1003
			files:        []FileContent{},
			expectTokens: 1503, // instruction tokens (1003) + overhead (500)
		},
		{
			name:         "files only",
			instructions: "",
			files: []FileContent{
				{Path: "test.go", Content: "test"}, // 4 chars * 0.75 = 3
			},
			expectTokens: 503, // 0 instruction tokens + 3 file tokens + 500 overhead
		},
		{
			name:         "instructions and files",
			instructions: "analyze", // 7 chars * 0.75 + 1000 = 1005.25 -> 1005
			files: []FileContent{
				{Path: "main.go", Content: "package main"}, // 12 chars * 0.75 = 9
			},
			expectTokens: 1519, // 1005 + 9 + 500 + 5 (rounding)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := service.CountTokens(context.Background(), TokenCountingRequest{
				Instructions: tt.instructions,
				Files:        tt.files,
			})

			require.NoError(t, err)
			// Allow small variance due to rounding in token estimation
			assert.InDelta(t, tt.expectTokens, result.TotalTokens, 10,
				"Expected ~%d tokens, got %d", tt.expectTokens, result.TotalTokens)
		})
	}
}
