# Define AuditEvent and ErrorDetails structs

## Goal
Implement the `AuditEvent` and `ErrorDetails` structs with appropriate fields and JSON tags in the `event.go` file to establish the core data structures for structured logging.

## Implementation Approach
I'll implement the following data structures:

1. **ErrorDetails struct**:
   - A focused structure for representing error information
   - Fields will include `Message` (string), `Type` (string, optional), and `Details` (string, optional)
   - Each field will have appropriate JSON tags including omitempty for optional fields

2. **AuditEvent struct**:
   - The main structure representing a structured log entry
   - Core fields:
     - `Timestamp` (time.Time)
     - `Level` (string) for log level (INFO, ERROR, etc.)
     - `Operation` (string) to identify the operation being performed
     - `Message` (string) for human-readable summary
   - Extension fields (all with omitempty):
     - `Inputs` (map[string]interface{}) for operation inputs
     - `Outputs` (map[string]interface{}) for operation results
     - `Metadata` (map[string]interface{}) for additional contextual information
     - `Error` (*ErrorDetails) for error information when applicable

All fields will have appropriate JSON tags to ensure proper serialization to JSON format.

## Reasoning
I selected this approach for several reasons:

1. **Flexibility with Typed Structure**: Using a structured type system provides compile-time validation while keeping the nested maps flexible enough to accommodate various data types for inputs, outputs, and metadata.

2. **Separate Error Type**: Having a dedicated `ErrorDetails` struct allows for standardized error reporting with optional fields for different types of errors.

3. **JSON Serialization**: The JSON tags ensure consistent serialization format and allow for omitting empty fields to keep the logs concise.

4. **Alignment with Requirements**: This structure directly addresses AC 1.1 (JSON lines format) and AC 2.1 (logging operations, inputs, outputs, errors), providing a balance between structure and flexibility.

5. **Compatibility with Go Ecosystem**: The approach uses standard Go types and patterns, making it compatible with existing Go libraries and tools.

Alternative approaches considered:

1. **Using a single map for all data**: This would be more flexible but would lose type safety and make the code less self-documenting.

2. **More specialized event types for different operations**: This would provide more type safety but introduce excessive complexity for the current requirements.

The selected approach provides a good balance of structure, flexibility, and compatibility with the requirements.