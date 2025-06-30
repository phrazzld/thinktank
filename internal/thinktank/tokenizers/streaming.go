package tokenizers

import (
	"context"
	"io"
)

// StreamingTokenizerManager extends TokenizerManager with streaming capabilities
type StreamingTokenizerManager interface {
	TokenizerManager
	GetStreamingTokenizer(provider string) (StreamingTokenCounter, error)
}

// streamingTokenizerManagerImpl implements StreamingTokenizerManager
type streamingTokenizerManagerImpl struct {
	*tokenizerManagerImpl
	chunkSize int
}

// NewStreamingTokenizerManager creates a new streaming tokenizer manager
func NewStreamingTokenizerManager() StreamingTokenizerManager {
	return &streamingTokenizerManagerImpl{
		tokenizerManagerImpl: &tokenizerManagerImpl{},
		chunkSize:            8 * 1024, // 8KB chunks for better cancellation responsiveness
	}
}

// GetStreamingTokenizer returns a streaming tokenizer for the given provider
func (m *streamingTokenizerManagerImpl) GetStreamingTokenizer(provider string) (StreamingTokenCounter, error) {
	baseTokenizer, err := m.GetTokenizer(provider)
	if err != nil {
		return nil, err
	}

	return &streamingTokenizerImpl{
		underlying: baseTokenizer,
		chunkSize:  m.chunkSize,
	}, nil
}

// streamingTokenizerImpl implements StreamingTokenCounter by wrapping an AccurateTokenCounter
type streamingTokenizerImpl struct {
	underlying AccurateTokenCounter
	chunkSize  int
}

// CountTokens delegates to the underlying tokenizer (implements AccurateTokenCounter)
func (s *streamingTokenizerImpl) CountTokens(ctx context.Context, text string, modelName string) (int, error) {
	return s.underlying.CountTokens(ctx, text, modelName)
}

// SupportsModel delegates to the underlying tokenizer
func (s *streamingTokenizerImpl) SupportsModel(modelName string) bool {
	return s.underlying.SupportsModel(modelName)
}

// GetEncoding delegates to the underlying tokenizer
func (s *streamingTokenizerImpl) GetEncoding(modelName string) (string, error) {
	return s.underlying.GetEncoding(modelName)
}

// GetChunkSizeForInput returns the optimal chunk size based on input size
// Uses adaptive chunking: 8KB for small inputs, 32KB for medium, 64KB for large
func (s *streamingTokenizerImpl) GetChunkSizeForInput(inputSizeBytes int) int {
	const (
		smallChunkSize  = 8 * 1024  // 8KB for small inputs (< 5MB)
		mediumChunkSize = 32 * 1024 // 32KB for medium inputs (5MB - 20MB)
		largeChunkSize  = 64 * 1024 // 64KB for large inputs (> 20MB)

		mediumThreshold = 5 * 1024 * 1024  // 5MB
		largeThreshold  = 20 * 1024 * 1024 // 20MB (reduced from 25MB for CI reliability)
	)

	switch {
	case inputSizeBytes >= largeThreshold:
		return largeChunkSize
	case inputSizeBytes >= mediumThreshold:
		return mediumChunkSize
	default:
		return smallChunkSize
	}
}

// CountTokensStreaming implements streaming tokenization by processing input in chunks
func (s *streamingTokenizerImpl) CountTokensStreaming(ctx context.Context, reader io.Reader, modelName string) (int, error) {
	totalTokens := 0
	buffer := make([]byte, s.chunkSize)

	for {
		// Check for cancellation before each chunk read
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
		}

		n, err := reader.Read(buffer)
		if n > 0 {
			// Check for cancellation before tokenization (expensive operation)
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			default:
			}

			// Convert chunk to string and tokenize with timeout protection
			chunk := string(buffer[:n])

			// Create a channel to receive tokenization result
			resultChan := make(chan struct {
				tokens int
				err    error
			}, 1)

			// Run tokenization in a goroutine to enable cancellation
			go func() {
				tokens, tokenErr := s.underlying.CountTokens(ctx, chunk, modelName)
				resultChan <- struct {
					tokens int
					err    error
				}{tokens, tokenErr}
			}()

			// Wait for either completion or cancellation
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			case result := <-resultChan:
				if result.err != nil {
					return 0, NewTokenizerErrorWithDetails("streaming", modelName, "chunk tokenization failed", result.err, "streaming")
				}
				totalTokens += result.tokens
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, NewTokenizerErrorWithDetails("streaming", modelName, "read error", err, "streaming")
		}
	}

	return totalTokens, nil
}

// CountTokensStreamingWithAdaptiveChunking implements streaming tokenization with adaptive chunk sizing
// This method uses input size to determine optimal chunk size for better performance
func (s *streamingTokenizerImpl) CountTokensStreamingWithAdaptiveChunking(ctx context.Context, reader io.Reader, modelName string, inputSizeBytes int) (int, error) {
	totalTokens := 0
	// Use adaptive chunk size based on input size
	adaptiveChunkSize := s.GetChunkSizeForInput(inputSizeBytes)
	buffer := make([]byte, adaptiveChunkSize)

	for {
		// Check for cancellation before each chunk read
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
		}

		n, err := reader.Read(buffer)
		if n > 0 {
			// Check for cancellation before tokenization (expensive operation)
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			default:
			}

			// Convert chunk to string and tokenize with timeout protection
			chunk := string(buffer[:n])

			// Create a channel to receive tokenization result
			resultChan := make(chan struct {
				tokens int
				err    error
			}, 1)

			// Run tokenization in a goroutine to enable cancellation
			go func() {
				tokens, tokenErr := s.underlying.CountTokens(ctx, chunk, modelName)
				resultChan <- struct {
					tokens int
					err    error
				}{tokens, tokenErr}
			}()

			// Wait for either completion or cancellation
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			case result := <-resultChan:
				if result.err != nil {
					return 0, NewTokenizerErrorWithDetails("streaming", modelName, "chunk tokenization failed", result.err, "streaming")
				}
				totalTokens += result.tokens
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, NewTokenizerErrorWithDetails("streaming", modelName, "read error", err, "streaming")
		}
	}

	return totalTokens, nil
}
