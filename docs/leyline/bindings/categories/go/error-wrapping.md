---
derived_from: explicit-over-implicit
enforced_by: golangci-lint("wrapcheck") & code review
id: error-wrapping
last_modified: '2025-05-14'
---
# Binding: Add Context to Errors as They Travel Upward

When errors cross package boundaries in Go, wrap them with contextual information using
`fmt.Errorf("context: %w", err)` or custom error types. Never return raw errors from
exported functions.

## Rationale

This binding implements our explicit-over-implicit tenet by making error context and
propagation paths visible rather than hidden.

Think of error wrapping like a travel journal for an error's journey through your
codebase. When a raw error travels across your application without being wrapped, it's
like a mysterious visitor with no record of where they've been or what they were trying
to do. By wrapping the error at each significant boundary—adding an entry to its travel
journal—you create a clear path of breadcrumbs showing exactly where it originated and
what operations failed along the way.

## Rule Definition

Error wrapping means adding contextual information to an error as it travels up the call
stack, while preserving the original error for type checking and root cause analysis. At
minimum, this context should include:

1. The operation that was attempted (e.g., "fetching user profile")
1. Any relevant identifiers (e.g., user IDs, record numbers)
1. Additional information that would help with debugging

This binding specifically requires:

- **Always wrap errors at package boundaries**: Any error returned from an exported
  function must be wrapped with context
- **Use the `%w` verb with fmt.Errorf**: This preserves the original error for later
  unwrapping and type checking
- **Custom error types must implement Unwrap()**: If using custom error types, they must
  properly implement the Unwrap() method
- **Don't wrap errors within package internals**: Internal functions generally don't
  need to wrap errors unless additional context is truly valuable

## Practical Implementation

### When to Wrap Errors

Always wrap errors when:

- Returning an error from an exported function
- Crossing major component boundaries
- Adding significant context would help with debugging

Generally avoid wrapping when:

- The error is already wrapped with the same context
- The function is internal to a package and doesn't add meaningful context
- Creating sentinel errors meant to be checked by type/value (these should be returned
  directly)

### Implementation Patterns

1. **Simple wrapping with fmt.Errorf**:

   ```go
   if err != nil {
       return fmt.Errorf("operation description: %w", err)
   }
   ```

1. **Custom error types** (when you need to include structured data):

   ```go
   type MyError struct {
       Operation string
       ResourceID string
       Err error
   }

   func (e *MyError) Error() string {
       return fmt.Sprintf("%s %s: %v", e.Operation, e.ResourceID, e.Err)
   }

   func (e *MyError) Unwrap() error {
       return e.Err
   }
   ```

1. **Error handling with wrapping**:

   ```go
   // Working with wrapped errors
   if errors.Is(err, sql.ErrNoRows) {
       // Handle specific error type
   }

   var myErr *MyError
   if errors.As(err, &myErr) {
       // Access fields in custom error
   }
   ```

### Common Mistakes to Avoid

1. **Losing the original error**:

   ```go
   // ❌ BAD: Lost original error
   return fmt.Errorf("operation failed: %v", err) // Using %v loses error type
   ```

1. **Wrapping with insufficient context**:

   ```go
   // ❌ BAD: Too generic
   return fmt.Errorf("failed: %w", err) // Not enough context
   ```

1. **Duplicate wrapping**:

   ```go
   // ❌ BAD: Duplicate context
   if err != nil {
       wrappedErr := fmt.Errorf("getting user: %w", err)
       return fmt.Errorf("getting user: %w", wrappedErr) // Redundant
   }
   ```

## Examples

```go
// ❌ BAD: Returning raw error with no context
func ProcessOrder(id string) error {
    order, err := db.GetOrder(id)
    if err != nil {
        return err // No context about what failed
    }
    // ...
    return nil
}

// ✅ GOOD: Wrapping with context
func ProcessOrder(id string) error {
    order, err := db.GetOrder(id)
    if err != nil {
        return fmt.Errorf("fetching order %s: %w", id, err)
    }

    if err := order.Validate(); err != nil {
        return fmt.Errorf("validating order %s: %w", id, err)
    }

    if err := payment.Process(order); err != nil {
        return fmt.Errorf("processing payment for order %s: %w", id, err)
    }

    return nil
}

// ✅ GOOD: Custom error type with structured context
type OrderError struct {
    OrderID string
    Operation string
    Err error
}

func (e *OrderError) Error() string {
    return fmt.Sprintf("order %s - %s: %v", e.OrderID, e.Operation, e.Err)
}

func (e *OrderError) Unwrap() error {
    return e.Err
}

func ProcessOrder(id string) error {
    order, err := db.GetOrder(id)
    if err != nil {
        return &OrderError{
            OrderID: id,
            Operation: "database fetch",
            Err: err,
        }
    }
    // ...
    return nil
}
```

### Error Trace Example

Here's how a wrapped error trace might look with proper context added at each level:

```
failed to process customer request: fetching order details for order 12345: connecting to order database: dial tcp: connection refused
```

This error trace tells us:

1. What high-level operation failed (processing customer request)
1. What specific step failed (fetching order details)
1. Which order was affected (12345)
1. What lower-level operation failed (connecting to database)
1. The root cause (connection refused)

Without proper wrapping, we might only see "connection refused" with no context.

## Related Bindings

- [use-structured-logging](../../docs/bindings/core/use-structured-logging.md) - Error context should be
  included in logs using structured formats
- [external-configuration](../../docs/bindings/core/external-configuration.md) - Error messages shouldn't
  contain hardcoded configuration values
- [hex-domain-purity](../../docs/bindings/core/hex-domain-purity.md) - Domain logic shouldn't depend on
  specific error implementation details
