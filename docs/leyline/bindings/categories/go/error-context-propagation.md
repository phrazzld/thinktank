---
id: error-context-propagation
last_modified: '2025-06-02'
derived_from: fix-broken-windows
enforced_by: 'Go error handling patterns, code review, static analysis'
---

# Binding: Propagate Error Context Through Go's Error Wrapping Patterns

Use Go's error wrapping mechanisms to preserve context and create clear error chains that enable effective debugging and monitoring. Properly wrapped errors maintain the full context of failure paths while providing actionable information for both developers and operations teams.

## Rationale

This binding implements our fix-broken-windows tenet by preventing the decay of error information quality over time. When errors lose context as they bubble up through the call stack, debugging becomes exponentially more difficult, leading to longer incident resolution times and recurring problems that could have been prevented with better error visibility.

Think of error context propagation like a crime scene investigation. Each piece of evidence (stack frame, operation context, user information) tells part of the story. When investigators arrive at a crime scene, they don't just note "something bad happened here"—they carefully document every detail, preserve the chain of custody, and maintain detailed records that future investigators can follow. Similarly, well-wrapped errors preserve the "chain of custody" for failures, documenting exactly what was happening at each level when things went wrong.

Without proper error context, debugging production issues becomes like trying to solve a mystery with only the final outcome and no clues about how you got there. Go's error wrapping capabilities, introduced in Go 1.13, provide the tools to maintain this investigative trail. When we fail to use these tools consistently, we create systems where errors surface as generic messages that require extensive detective work to understand. This investigative overhead compounds over time, making the system increasingly difficult to operate and maintain.

## Rule Definition

Error context propagation must establish these Go-specific practices:

- **Error Wrapping**: Use `fmt.Errorf` with the `%w` verb to wrap errors when adding context. This preserves the original error while adding contextual information about what operation was being performed.

- **Context Preservation**: Include relevant operational context (user IDs, request IDs, resource identifiers, operation parameters) when wrapping errors, but avoid exposing sensitive information.

- **Error Chain Inspection**: Use `errors.Is()` and `errors.As()` to inspect error chains rather than string comparison, allowing for proper error type checking even through multiple wrap layers.

- **Sentinel Error Patterns**: Define package-level sentinel errors using `errors.New()` for common error conditions that callers need to handle specifically, making them checkable with `errors.Is()`.

- **Structured Error Types**: Create custom error types that implement the `error` interface when errors need to carry structured data (codes, metadata, retry information) while remaining inspectable through the error chain.

- **Error Boundary Management**: Establish clear error boundaries where technical errors are translated into user-facing messages, while preserving the full technical context for logging and monitoring.

**Error Wrapping Patterns:**
- Operation context wrapping (`"processing user order: %w"`)
- Resource identification (`"failed to update user %s: %w"`, userID)
- Temporal context (`"timeout after %v while connecting: %w"`, duration)
- Retry context (`"failed after %d attempts: %w"`, attempts)

**Context Information Types:**
- Request identifiers and correlation IDs
- User/tenant identifiers (non-sensitive)
- Resource identifiers (file paths, database keys)
- Operation parameters (timeouts, retry counts)
- Environmental context (hostname, service version)

## Practical Implementation

1. **Implement Consistent Error Wrapping**: Always add context when propagating errors up the call stack:

   ```go
   func (s *OrderService) ProcessOrder(ctx context.Context, orderID string) error {
       order, err := s.orderRepo.GetOrder(ctx, orderID)
       if err != nil {
           // Wrap with operation context
           return fmt.Errorf("failed to retrieve order %s: %w", orderID, err)
       }

       if err := s.validateOrder(order); err != nil {
           // Wrap with validation context
           return fmt.Errorf("order %s validation failed: %w", orderID, err)
       }

       if err := s.chargePayment(ctx, order); err != nil {
           // Wrap with payment context
           return fmt.Errorf("payment processing failed for order %s: %w", orderID, err)
       }

       return nil
   }
   ```

2. **Create Actionable Custom Error Types**: Define structured errors that carry metadata while remaining inspectable:

   ```go
   // ValidationError carries details about what validation failed
   type ValidationError struct {
       Field   string
       Value   interface{}
       Rule    string
       Message string
       Cause   error
   }

   func (e *ValidationError) Error() string {
       if e.Cause != nil {
           return fmt.Sprintf("validation failed for field '%s': %s (caused by: %v)",
                              e.Field, e.Message, e.Cause)
       }
       return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
   }

   func (e *ValidationError) Unwrap() error {
       return e.Cause
   }

   // RetryableError indicates whether an operation can be retried
   type RetryableError struct {
       Operation string
       Retryable bool
       Cause     error
   }

   func (e *RetryableError) Error() string {
       retryStatus := "non-retryable"
       if e.Retryable {
           retryStatus = "retryable"
       }
       return fmt.Sprintf("%s operation failed (%s): %v", e.Operation, retryStatus, e.Cause)
   }

   func (e *RetryableError) Unwrap() error {
       return e.Cause
   }
   ```

3. **Establish Error Boundary Patterns**: Create clear boundaries between technical and user-facing errors:

   ```go
   // ErrorBoundary translates internal errors to user-appropriate messages
   type ErrorBoundary struct {
       logger Logger
   }

   func (eb *ErrorBoundary) HandleError(ctx context.Context, err error) error {
       // Extract correlation ID for logging
       correlationID := getCorrelationID(ctx)

       // Log the full technical error with context
       eb.logger.Error("operation failed",
           "correlation_id", correlationID,
           "error", err.Error(),
           "error_type", fmt.Sprintf("%T", err))

       // Check for specific error types and return appropriate user messages
       var validationErr *ValidationError
       if errors.As(err, &validationErr) {
           return fmt.Errorf("invalid %s: %s", validationErr.Field, validationErr.Message)
       }

       var retryableErr *RetryableError
       if errors.As(err, &retryableErr) {
           if retryableErr.Retryable {
               return errors.New("service temporarily unavailable, please try again")
           }
       }

       // For unexpected errors, provide a generic message but log details
       eb.logger.Error("unexpected error occurred",
           "correlation_id", correlationID,
           "stack_trace", fmt.Sprintf("%+v", err))

       return errors.New("an unexpected error occurred")
   }
   ```

4. **Use Proper Error Chain Inspection**: Always use `errors.Is()` and `errors.As()` for error checking:

   ```go
   func (s *Service) handleDatabaseError(err error) error {
       // Check for specific sentinel errors
       if errors.Is(err, sql.ErrNoRows) {
           return &NotFoundError{Resource: "user", Cause: err}
       }

       // Check for custom error types
       var timeoutErr *TimeoutError
       if errors.As(err, &timeoutErr) {
           return &RetryableError{
               Operation: "database_query",
               Retryable: true,
               Cause:     err,
           }
       }

       // Check for connection-related errors (these are usually retryable)
       if isConnectionError(err) {
           return &RetryableError{
               Operation: "database_connection",
               Retryable: true,
               Cause:     err,
           }
       }

       // Unknown database error - not retryable
       return &RetryableError{
           Operation: "database_operation",
           Retryable: false,
           Cause:     err,
       }
   }

   func isConnectionError(err error) bool {
       // Check for common connection error patterns
       errStr := err.Error()
       return strings.Contains(errStr, "connection refused") ||
              strings.Contains(errStr, "timeout") ||
              strings.Contains(errStr, "network")
   }
   ```

5. **Implement Request Context Propagation**: Include correlation IDs and request context in error messages:

   ```go
   type ContextKey string

   const (
       CorrelationIDKey ContextKey = "correlation_id"
       UserIDKey        ContextKey = "user_id"
       RequestIDKey     ContextKey = "request_id"
   )

   func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
       correlationID := getCorrelationID(ctx)

       user, err := s.userRepo.FindByID(ctx, userID)
       if err != nil {
           // Include correlation ID in error context
           return nil, fmt.Errorf("failed to get user %s (correlation_id: %s): %w",
                                  userID, correlationID, err)
       }

       return user, nil
   }

   func getCorrelationID(ctx context.Context) string {
       if id, ok := ctx.Value(CorrelationIDKey).(string); ok {
           return id
       }
       return "unknown"
   }

   // Middleware to inject correlation ID
   func CorrelationMiddleware(next http.Handler) http.Handler {
       return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
           correlationID := r.Header.Get("X-Correlation-ID")
           if correlationID == "" {
               correlationID = generateCorrelationID()
           }

           ctx := context.WithValue(r.Context(), CorrelationIDKey, correlationID)
           w.Header().Set("X-Correlation-ID", correlationID)

           next.ServeHTTP(w, r.WithContext(ctx))
       })
   }
   ```

## Examples

```go
// ❌ BAD: Losing error context and making debugging difficult
func (s *OrderService) ProcessOrder(orderID string) error {
    order, err := s.repo.GetOrder(orderID)
    if err != nil {
        return err // Lost: which order failed to load?
    }

    err = s.validateOrder(order)
    if err != nil {
        return errors.New("validation failed") // Lost: what validation failed? Original error?
    }

    err = s.processPayment(order)
    if err != nil {
        return err // Lost: context about payment processing step
    }

    return nil
}

// When this error surfaces, you only know something failed, but not:
// - Which order was being processed
// - Which step in the process failed
// - What the underlying error was
// - Whether the error is retryable
```

```go
// ✅ GOOD: Comprehensive error context propagation
func (s *OrderService) ProcessOrder(ctx context.Context, orderID string) error {
    correlationID := getCorrelationID(ctx)

    order, err := s.repo.GetOrder(ctx, orderID)
    if err != nil {
        return fmt.Errorf("failed to retrieve order %s (correlation_id: %s): %w",
                          orderID, correlationID, err)
    }

    if err := s.validateOrder(order); err != nil {
        return fmt.Errorf("validation failed for order %s (correlation_id: %s): %w",
                          orderID, correlationID, err)
    }

    if err := s.processPayment(ctx, order); err != nil {
        return fmt.Errorf("payment processing failed for order %s amount $%.2f (correlation_id: %s): %w",
                          orderID, order.Amount, correlationID, err)
    }

    return nil
}

// This error chain provides:
// - Specific order being processed
// - Step that failed
// - Correlation ID for request tracing
// - Original error details
// - Relevant context (amount for payment failures)
```

```go
// ❌ BAD: Generic error handling that loses actionable information
func (s *PaymentService) ChargeCard(amount float64, cardToken string) error {
    resp, err := s.paymentAPI.Charge(amount, cardToken)
    if err != nil {
        return errors.New("payment failed")
    }

    if resp.Status != "success" {
        return errors.New("payment was declined")
    }

    return nil
}

// Problems:
// - No indication whether failure is retryable
// - No error codes for specific handling
// - No context about the payment amount
// - Original API errors are lost
```

```go
// ✅ GOOD: Structured error handling with actionable information
type PaymentError struct {
    Code      string
    Message   string
    Amount    float64
    Retryable bool
    Cause     error
}

func (e *PaymentError) Error() string {
    retryStatus := "non-retryable"
    if e.Retryable {
        retryStatus = "retryable"
    }
    return fmt.Sprintf("payment of $%.2f failed (code: %s, %s): %s",
                       e.Amount, e.Code, retryStatus, e.Message)
}

func (e *PaymentError) Unwrap() error {
    return e.Cause
}

func (s *PaymentService) ChargeCard(ctx context.Context, amount float64, cardToken string) error {
    correlationID := getCorrelationID(ctx)

    resp, err := s.paymentAPI.Charge(amount, cardToken)
    if err != nil {
        // Network or API errors are usually retryable
        return &PaymentError{
            Code:      "API_ERROR",
            Message:   fmt.Sprintf("payment API call failed (correlation_id: %s)", correlationID),
            Amount:    amount,
            Retryable: true,
            Cause:     err,
        }
    }

    switch resp.Status {
    case "success":
        return nil
    case "insufficient_funds":
        return &PaymentError{
            Code:      "INSUFFICIENT_FUNDS",
            Message:   "card has insufficient funds",
            Amount:    amount,
            Retryable: false,
            Cause:     nil,
        }
    case "card_declined":
        return &PaymentError{
            Code:      "CARD_DECLINED",
            Message:   "card was declined by issuer",
            Amount:    amount,
            Retryable: false,
            Cause:     nil,
        }
    case "temporary_failure":
        return &PaymentError{
            Code:      "TEMPORARY_FAILURE",
            Message:   "temporary payment processing issue",
            Amount:    amount,
            Retryable: true,
            Cause:     nil,
        }
    default:
        return &PaymentError{
            Code:      "UNKNOWN_STATUS",
            Message:   fmt.Sprintf("unexpected payment status: %s", resp.Status),
            Amount:    amount,
            Retryable: false,
            Cause:     nil,
        }
    }
}

// Usage in calling code:
func (s *OrderService) handlePaymentError(err error) error {
    var paymentErr *PaymentError
    if errors.As(err, &paymentErr) {
        // Log structured payment error details
        s.logger.Error("payment processing failed",
            "error_code", paymentErr.Code,
            "amount", paymentErr.Amount,
            "retryable", paymentErr.Retryable,
            "underlying_error", paymentErr.Cause)

        // Retry logic based on error metadata
        if paymentErr.Retryable {
            return s.schedulePaymentRetry(paymentErr)
        }

        // Return user-appropriate message
        return fmt.Errorf("payment failed: %s", paymentErr.Message)
    }

    return err
}
```

```go
// ❌ BAD: Error boundary that loses important debugging information
func (h *HTTPHandler) handleError(w http.ResponseWriter, err error) {
    log.Println("Error occurred:", err)
    http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}
```

```go
// ✅ GOOD: Error boundary that preserves context while providing appropriate responses
func (h *HTTPHandler) handleError(ctx context.Context, w http.ResponseWriter, err error) {
    correlationID := getCorrelationID(ctx)

    // Log the full error with context
    h.logger.Error("request failed",
        "correlation_id", correlationID,
        "error", err.Error(),
        "error_chain", fmt.Sprintf("%+v", err),
        "stack_trace", string(debug.Stack()))

    // Set correlation ID header for client debugging
    w.Header().Set("X-Correlation-ID", correlationID)

    // Handle specific error types with appropriate HTTP responses
    var validationErr *ValidationError
    if errors.As(err, &validationErr) {
        http.Error(w, fmt.Sprintf("Invalid %s: %s", validationErr.Field, validationErr.Message),
                   http.StatusBadRequest)
        return
    }

    var notFoundErr *NotFoundError
    if errors.As(err, &notFoundErr) {
        http.Error(w, fmt.Sprintf("%s not found", notFoundErr.Resource),
                   http.StatusNotFound)
        return
    }

    var paymentErr *PaymentError
    if errors.As(err, &paymentErr) {
        if paymentErr.Code == "INSUFFICIENT_FUNDS" || paymentErr.Code == "CARD_DECLINED" {
            http.Error(w, paymentErr.Message, http.StatusPaymentRequired)
            return
        }

        // Other payment errors are server errors
        http.Error(w, "Payment processing temporarily unavailable",
                   http.StatusServiceUnavailable)
        return
    }

    // Generic server error for unexpected errors
    http.Error(w, "An unexpected error occurred", http.StatusInternalServerError)
}
```

## Related Bindings

- [error-wrapping.md](../../docs/bindings/categories/go/error-wrapping.md): Error context propagation builds on Go's basic error wrapping patterns to create comprehensive error chains. While error wrapping covers the mechanics of `fmt.Errorf` and `%w`, this binding focuses on the systematic preservation of operational context throughout error propagation.

- [technical-debt-tracking.md](../../core/technical-debt-tracking.md): Poor error context is a form of technical debt that compounds over time. This binding prevents the accumulation of debugging debt by ensuring error information quality doesn't degrade. Both bindings support the fix-broken-windows principle of addressing quality issues immediately.

- [use-structured-logging.md](../../core/use-structured-logging.md): Error context propagation and structured logging work together to create comprehensive observability. Error contexts provide the data that structured logging captures, while correlation IDs link errors to specific request traces across distributed systems.

- [fix-broken-windows.md](../../tenets/fix-broken-windows.md): This binding directly implements the fix-broken-windows tenet by preventing the degradation of error information quality. Each properly wrapped error preserves debugging capability, while poor error handling creates "broken windows" that make the entire system harder to maintain.
