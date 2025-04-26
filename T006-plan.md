# T006 Plan: Enforce logger.WithContext(ctx) at Entry Points

## Task Description
- **ID + Title**: T006 · Refactor · P2: enforce logger.WithContext(ctx) at entry points
- **Context**: cr-02 step 3
- **Action**:
  1. Ensure `main`, handlers, and CLI entrypoints assign `logger = logger.WithContext(ctx)` before any logging.
- **Done-when**:
  1. Every entrypoint uses `WithContext(ctx)` before logging.

## Analysis

This task aims to ensure that correlation IDs are properly propagated through context at all application entry points. The intent is to make sure that the logger has access to the context (which contains the correlation ID) before any logging occurs. This ensures consistent and traceable logs across the application.

While we've removed explicit correlation ID formatting in log calls (T005), we must now make sure every entrypoint properly attaches the context to the logger to preserve this information.

Entry points typically include:
1. `main.go` - The application's entry point
2. CLI command handlers
3. API handlers
4. Any other functions that serve as an entry to a workflow

## Implementation Plan

1. Identify all entrypoints by examining:
   - main.go
   - cmd/ directory for CLI entrypoints
   - API handlers if any

2. For each entrypoint:
   - Check if a context is created/available
   - Ensure the logger is initialized with this context using `logger = logger.WithContext(ctx)`
   - Verify this happens before any logging calls are made
   - Make necessary changes to ensure context is attached to logger

3. For main.go:
   - Ensure a context is created early (usually with `context.Background()`)
   - Ensure correlation ID is added to context
   - Make sure the logger is initialized with this context before first use

4. Run tests to verify everything still works properly

## Success Criteria
- All entry points have a context with correlation ID
- All entry points initialize their logger with this context
- No log calls occur before this initialization
- All tests pass