package logutil

import (
	"context"
	"testing"
)

func TestWithCorrelationID(t *testing.T) {
	// Test case 1: Adding correlation ID to empty context
	ctx := context.Background()
	ctxWithID := WithCorrelationID(ctx)
	id := GetCorrelationID(ctxWithID)

	if id == "" {
		t.Error("WithCorrelationID should generate a non-empty UUID when added to empty context")
	}

	// Verify generated ID follows UUID format (36 characters)
	if len(id) != 36 {
		t.Errorf("Generated correlation ID should be a UUID with 36 characters, got %d characters: %s", len(id), id)
	}

	// Test case 2: Preserving existing correlation ID
	ctx2 := WithCorrelationID(ctxWithID)
	id2 := GetCorrelationID(ctx2)

	if id2 != id {
		t.Errorf("WithCorrelationID should preserve existing ID. Expected %s, got %s", id, id2)
	}

	// Test case 3: Setting explicit correlation ID
	customID := "custom-correlation-id-123"
	ctxWithCustomID := WithCorrelationID(ctx, customID)
	resultID := GetCorrelationID(ctxWithCustomID)

	if resultID != customID {
		t.Errorf("WithCorrelationID with custom ID should use provided ID. Expected %s, got %s", customID, resultID)
	}

	// Test case 4: Overriding existing correlation ID
	overrideID := "override-correlation-id-456"
	ctxWithOverrideID := WithCorrelationID(ctxWithID, overrideID)
	resultOverrideID := GetCorrelationID(ctxWithOverrideID)

	if resultOverrideID != overrideID {
		t.Errorf("WithCorrelationID with custom ID should override existing ID. Expected %s, got %s", overrideID, resultOverrideID)
	}

	// Test case 5: Empty custom ID preserves existing
	ctxWithEmptyCustomID := WithCorrelationID(ctxWithID, "")
	resultEmptyID := GetCorrelationID(ctxWithEmptyCustomID)

	if resultEmptyID != id {
		t.Errorf("WithCorrelationID with empty custom ID should preserve existing ID. Expected %s, got %s", id, resultEmptyID)
	}
}

func TestWithCustomCorrelationID(t *testing.T) {
	// Test setting custom correlation ID
	ctx := context.Background()
	customID := "test-custom-id-789"
	ctxWithCustomID := WithCustomCorrelationID(ctx, customID)
	resultID := GetCorrelationID(ctxWithCustomID)

	if resultID != customID {
		t.Errorf("WithCustomCorrelationID should set the provided ID. Expected %s, got %s", customID, resultID)
	}

	// Test overriding existing ID
	existingCtx := WithCorrelationID(ctx)
	overrideID := "override-custom-id-012"
	overrideCtx := WithCustomCorrelationID(existingCtx, overrideID)
	resultOverrideID := GetCorrelationID(overrideCtx)

	if resultOverrideID != overrideID {
		t.Errorf("WithCustomCorrelationID should override existing ID. Expected %s, got %s", overrideID, resultOverrideID)
	}

	// Note: WithCustomCorrelationID with empty string will generate a new UUID
	// This behavior is defined in WithCorrelationID where empty string is treated
	// as a request to generate a new ID when no existing ID is present
}

func TestGetCorrelationID(t *testing.T) {
	// Test with nil context (using context.TODO() as recommended)
	if id := GetCorrelationID(context.TODO()); id != "" {
		t.Errorf("GetCorrelationID with nil/empty context should return empty string, got %s", id)
	}

	// Test with empty context
	if id := GetCorrelationID(context.Background()); id != "" {
		t.Errorf("GetCorrelationID with empty context should return empty string, got %s", id)
	}

	// Test with invalid type in context
	ctxWithInvalidType := context.WithValue(context.Background(), CorrelationIDKey, 123)
	if id := GetCorrelationID(ctxWithInvalidType); id != "" {
		t.Errorf("GetCorrelationID with invalid type should return empty string, got %s", id)
	}

	// Test with valid correlation ID
	expectedID := "test-correlation-id-345"
	ctxWithID := context.WithValue(context.Background(), CorrelationIDKey, expectedID)
	if id := GetCorrelationID(ctxWithID); id != expectedID {
		t.Errorf("GetCorrelationID should return the stored ID. Expected %s, got %s", expectedID, id)
	}
}

func TestContextPropagation(t *testing.T) {
	// Test correlation ID propagation through context chains
	ctx := context.Background()

	// Create a root context with correlation ID
	rootID := "root-correlation-id"
	rootCtx := WithCorrelationID(ctx, rootID)

	// Define type for context key to avoid using string directly
	type childKeyType string
	childKey := childKeyType("child_key")

	// Create a child context with additional values
	childCtx := context.WithValue(rootCtx, childKey, "child_value")

	// Verify correlation ID is preserved in child context
	childID := GetCorrelationID(childCtx)
	if childID != rootID {
		t.Errorf("Correlation ID should be preserved in child context. Expected %s, got %s", rootID, childID)
	}

	// Add more context values and verify preservation
	type grandchildKeyType string
	grandchildKey := grandchildKeyType("grandchild_key")
	grandchildCtx := context.WithValue(childCtx, grandchildKey, "grandchild_value")
	grandchildID := GetCorrelationID(grandchildCtx)
	if grandchildID != rootID {
		t.Errorf("Correlation ID should be preserved through multiple context layers. Expected %s, got %s", rootID, grandchildID)
	}
}

func TestCorrelationIDUniqueness(t *testing.T) {
	// Test that auto-generated correlation IDs are unique
	ctx := context.Background()

	// Generate multiple IDs
	ctx1 := WithCorrelationID(ctx)
	ctx2 := WithCorrelationID(ctx)
	ctx3 := WithCorrelationID(ctx)

	id1 := GetCorrelationID(ctx1)
	id2 := GetCorrelationID(ctx2)
	id3 := GetCorrelationID(ctx3)

	// Check all IDs are unique
	if id1 == id2 || id1 == id3 || id2 == id3 {
		t.Errorf("Auto-generated correlation IDs should be unique. Got: %s, %s, %s", id1, id2, id3)
	}

	// Check all IDs have proper UUID format
	for i, id := range []string{id1, id2, id3} {
		if len(id) != 36 {
			t.Errorf("Generated correlation ID %d should be a UUID with 36 characters, got %d characters: %s", i+1, len(id), id)
		}
	}
}

func TestContextWithMultipleCorrelationIDs(t *testing.T) {
	// Test behavior with multiple correlation ID assignments
	ctx := context.Background()

	// Chain of correlation ID assignments
	id1 := "first-correlation-id"
	id2 := "second-correlation-id"
	id3 := "third-correlation-id"

	ctx = WithCorrelationID(ctx, id1)
	if GetCorrelationID(ctx) != id1 {
		t.Errorf("First correlation ID not set correctly. Expected %s, got %s", id1, GetCorrelationID(ctx))
	}

	ctx = WithCorrelationID(ctx, id2)
	if GetCorrelationID(ctx) != id2 {
		t.Errorf("Second correlation ID should override first. Expected %s, got %s", id2, GetCorrelationID(ctx))
	}

	ctx = WithCorrelationID(ctx, id3)
	if GetCorrelationID(ctx) != id3 {
		t.Errorf("Third correlation ID should override second. Expected %s, got %s", id3, GetCorrelationID(ctx))
	}

	// Test empty string preserves previous ID
	ctx = WithCorrelationID(ctx, "")
	if GetCorrelationID(ctx) != id3 {
		t.Errorf("Empty correlation ID should preserve previous ID. Expected %s, got %s", id3, GetCorrelationID(ctx))
	}

	// Test auto-generated ID preserves previous ID when it exists
	ctx = WithCorrelationID(ctx)
	if GetCorrelationID(ctx) != id3 {
		t.Errorf("Auto-generated correlation ID should preserve previous ID. Expected %s, got %s", id3, GetCorrelationID(ctx))
	}
}

func TestCorrelationIDWithCancelledContext(t *testing.T) {
	// Test behavior with cancelled contexts
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set correlation ID
	id := "correlation-id-with-cancel"
	ctxWithID := WithCorrelationID(ctx, id)

	// Verify ID is set
	if GetCorrelationID(ctxWithID) != id {
		t.Errorf("Correlation ID not set correctly. Expected %s, got %s", id, GetCorrelationID(ctxWithID))
	}

	// Cancel the context
	cancel()

	// Correlation ID should still be accessible after cancellation
	if GetCorrelationID(ctxWithID) != id {
		t.Errorf("Correlation ID should be accessible after context cancellation. Expected %s, got %s", id, GetCorrelationID(ctxWithID))
	}
}

func TestCorrelationIDWithDeadlineContext(t *testing.T) {
	// Test behavior with deadline contexts
	ctx, cancel := context.WithTimeout(context.Background(), 0) // Immediate timeout
	defer cancel()

	// Set correlation ID
	id := "correlation-id-with-deadline"
	ctxWithID := WithCorrelationID(ctx, id)

	// Verify ID is set
	if GetCorrelationID(ctxWithID) != id {
		t.Errorf("Correlation ID not set correctly. Expected %s, got %s", id, GetCorrelationID(ctxWithID))
	}

	// Even with an expired deadline, correlation ID should be accessible
	if GetCorrelationID(ctxWithID) != id {
		t.Errorf("Correlation ID should be accessible after context deadline. Expected %s, got %s", id, GetCorrelationID(ctxWithID))
	}
}
