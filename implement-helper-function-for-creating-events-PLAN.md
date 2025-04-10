# Implement helper function for creating events

## Goal
Create a helper function `NewAuditEvent` in the auditlog package that simplifies the creation of AuditEvent instances by providing sensible defaults and a clean API.

## Implementation Approach
I'll implement the `NewAuditEvent` function in the `event.go` file with the following characteristics:

1. It will accept the minimum required parameters (level, operation, message) to create a valid event
2. It will automatically set the timestamp to the current UTC time
3. It will return a properly initialized AuditEvent instance ready for further customization
4. I'll also add some convenience methods to make it easier to add inputs, outputs, metadata, and error details to the event

This approach simplifies event creation while still allowing for complete flexibility when needed. The builder-pattern style methods will make code more readable when constructing complex events.

## Reasoning
I chose this approach because:

1. It balances simplicity and flexibility - basic events are easy to create, but complex events are still possible
2. It reduces repetitive code in client code (like setting timestamps)
3. The builder pattern provides a clean API and improves readability at call sites
4. It's consistent with the existing package design and Go idioms
5. It helps ensure consistency in logging by setting reasonable defaults