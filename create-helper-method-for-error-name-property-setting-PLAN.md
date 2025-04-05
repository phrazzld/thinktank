# Create Helper Method for Error Name Property Setting

## Task Goal
Add a helper method in `ThinktankError` base class to standardize setting the error name property across all error subclasses, ensuring proper inheritance and correct `instanceof` checks.

## Current Implementation Analysis
Currently, each error subclass in the error hierarchy manually sets its `name` property in its constructor:

```typescript
// ThinktankError base class
constructor(message: string, options?: {/*...*/}) {
  super(message);
  this.name = 'ThinktankError';
  // ...
}

// ConfigError subclass
constructor(message: string, options?: {/*...*/}) {
  super(message, {/*...*/});
  this.name = 'ConfigError';
}

// ApiError subclass
constructor(message: string, options?: {/*...*/}) {
  super(formattedMessage, {/*...*/});
  this.name = 'ApiError';
  this.providerId = options?.providerId;
}

// And so on for other subclasses
```

This approach:
1. Violates the DRY principle by duplicating the name-setting logic
2. Requires maintenance when adding new error subclasses
3. Could lead to inconsistencies if a developer forgets to set the name
4. Doesn't guarantee that the name reflects the actual class hierarchy

## Implementation Approach

I will add a protected method in the `ThinktankError` base class that:
1. Automatically sets the name property based on the constructor's name
2. Is called by the base class constructor, so subclasses inherit this behavior
3. Can be overridden by subclasses if needed

The implementation will:
1. Extract the constructor name using `this.constructor.name`
2. Set the name property to this value
3. Make sure the method is called during object initialization

This approach has several advantages:
- Centralizes the name-setting logic
- Ensures consistent naming across all error classes
- Maintains the correct class name for `instanceof` checks
- Requires minimal changes to existing code
- Makes adding new error subclasses easier and less error-prone

## Key Design Decisions

1. **Protected Method vs. Private Property**: Using a protected method allows subclasses to override the name-setting behavior if needed, while maintaining encapsulation.

2. **Using constructor.name**: This approach automatically reflects the actual class name, ensuring consistency between the error name and its class.

3. **Setting Name in Base Constructor**: By making the base constructor call the method, we ensure that all subclasses inherit this behavior without needing to modify their constructors.

4. **Backward Compatibility**: This implementation maintains backward compatibility with existing code that relies on the error name property.

## Potential Impacts

This change is low-risk and should have minimal impact on the rest of the codebase:
- Error subclasses will still have the same name property values
- No changes required to error handling code
- Instance checks (`instanceof`) will continue to work as expected

The only notable change is that error classes created by extending existing error classes will automatically get the correct name without requiring explicit assignment.