# Extract plan generation and saving to output.go

## Task Goal
Move the plan generation and saving functionality from main.go to output.go, maintaining consistent behavior while creating a proper separation of concerns.

## Implementation Approach

### Selected Approach: Progressive Extraction with Interface Enhancement
1. Enhance the `OutputWriter` interface with method signatures that match the patterns established in main.go
2. Implement the plan generation and saving functionality in output.go by moving the code from main.go
3. Replace the direct function calls in main.go with calls to the OutputWriter interface
4. Maintain error handling patterns consistent with the existing SaveToFile implementation
5. Update main.go functions with transitional implementation comments

### Alternative Approaches Considered:

#### Alternative 1: Complete Rewrite with New Design
- Redesign the plan generation API completely with a more object-oriented approach
- Create new interfaces and types to represent plans and generators
- Implement entirely new code in output.go
- Replace main.go functionality with calls to the new API

**Rejected because:** A complete redesign would violate the "pure refactoring" principle established in the project guidelines. The goal is to reorganize code without changing behavior, not introduce new designs.

#### Alternative 2: Function-by-Function Direct Copy
- Copy each function directly with minimal adaptation
- Maintain all internal implementation details unchanged
- Update main.go to call the new functions

**Rejected because:** While this would be simpler, it wouldn't fully leverage the interface-based design pattern established with OutputWriter. It would also miss the opportunity to improve error handling consistency.

## Implementation Reasoning
The selected approach maintains the project's guidelines for "pure refactoring" while also improving maintainability through:

1. **Interface-Based Design**: Enhancing the OutputWriter interface allows for better testing and potential replacement of implementations in the future
2. **Consistent Error Handling**: Converting log.Fatal calls to returned errors allows the caller to decide how to handle failures
3. **Progressive Transition**: Maintaining compatibility with existing code while working toward a cleaner architecture
4. **Minimal Behavioral Change**: Preserving the exact same user-facing behavior while improving code organization

This approach also aligns with the project's established pattern of moving functionality to properly modularized components.