# Create test migration guide

## Goal
Create a comprehensive guide that helps developers migrate from using the old `mockFsUtils` approach to the new `virtualFsUtils` approach for filesystem testing. This will be done by updating the existing `src/__tests__/utils/README.md` file with clear instructions and examples.

## Implementation Approach
I'll create a detailed migration guide that focuses on:

1. Creating a new section in the existing README.md that covers migrating from mockFsUtils to virtualFsUtils
2. Providing clear step-by-step instructions with code examples showing before/after transformations
3. Addressing common patterns and edge cases with concrete solutions
4. Including best practices for working with the new virtualFsUtils approach

The guide will be structured to assist developers at different levels of experience, with particular attention to common file operations that need to be migrated, including:
- Reading files
- Writing files
- Checking file existence
- Getting file stats
- Handling directories
- Simulating errors

## Reasoning

I considered three potential approaches:

1. **Update existing README.md with migration section (CHOSEN)**
   - Pros: Keeps all test utility documentation in one place
   - Pros: Allows for direct comparison between old and new approaches
   - Pros: Leverages existing documentation structure
   - Cons: Makes the README.md file longer and potentially more complex

2. **Create a separate TEST_MIGRATION.md file**
   - Pros: Keeps migration guide separate and focused
   - Pros: Could be easier to find for developers specifically looking for migration instructions
   - Cons: Fragments documentation across multiple files
   - Cons: Requires duplication of some content that exists in README.md

3. **Replace README.md entirely with new content**
   - Pros: Clean slate for documentation
   - Pros: Avoids confusion between old and new approaches
   - Cons: Loses valuable information about existing patterns
   - Cons: More disruptive for developers familiar with current documentation

I chose the first approach because it provides the most comprehensive solution while maintaining continuity with existing documentation. By adding a migration section to the README.md, developers can see both the old and new patterns in the same file, which is helpful when updating tests. This approach also allows for a gradual transition where both approaches can be used side-by-side during the migration period.