# Create virtualFsUtils.ts utility

## Goal
Implement a new utility file that provides functions for creating, manipulating, and accessing an in-memory filesystem using memfs, replacing the custom mockFsUtils.ts.

## Implementation Approach
I'll create a simple yet powerful abstraction over memfs that maintains compatibility with the existing test patterns while simplifying the API. This approach will use memfs's Volume class to provide a virtual filesystem that fully implements the Node.js fs API, making it easier to test filesystem interactions without complex mocking.

The implementation will:
1. Export a Volume instance that test files can import and use directly
2. Provide high-level helper functions to simplify common operations (reset filesystem, create files/directories, simulate errors)
3. Support transparent mocking of fs and fs/promises through Jest without test-specific code in production

### Core Functions:
- `createVirtualFs()`: Initialize a virtual filesystem with a specified structure
- `resetVirtualFs()`: Clear the virtual filesystem between tests
- `getVirtualFs()`: Get the current memfs volume for direct manipulation
- `mockFsModules()`: Helper to set up proper mocking in Jest

## Reasoning

I considered three potential approaches:

1. **Direct memfs Usage with Minimal Abstraction (CHOSEN)**
   - Pros: Simple, transparent API that closely matches the Node.js fs module
   - Pros: Minimal learning curve for developers familiar with fs module
   - Pros: Direct access to the Volume instance allows for maximum flexibility
   - Pros: Works with existing Jest mocking patterns
   - Cons: Requires slightly more setup in test files compared to highly abstracted APIs

2. **Comprehensive API that Mirrors mockFsUtils.ts**
   - Pros: Minimal changes required to existing tests
   - Pros: Familiar patterns for the team
   - Cons: Complex implementation that would largely replicate the issues with the current approach
   - Cons: More difficult to maintain long-term
   - Cons: Less idiomatic usage of memfs

3. **Class-Based Implementation with Chainable API**
   - Pros: Modern, fluent API with method chaining
   - Pros: Could potentially make tests more readable
   - Cons: Significant departure from current patterns
   - Cons: More code to maintain
   - Cons: Higher learning curve

I chose the first approach because it strikes an optimal balance between simplicity, compatibility, and maintainability. By providing a thin wrapper around memfs with a few helpful utility functions, we can make tests easier to write and understand while minimizing the maintenance burden. This approach also better aligns with the project's goals of simplifying the testing strategy and improving reliability.