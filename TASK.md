# Task: Replace Gemini-based Name Generation with Deterministic Random Name Generation

## Overview
Currently, ThinkTank generates a friendly name for output directories using Google's Gemini API. This approach has the following issues:
- Requires an API call for each run
- Often produces repetitive names (observed "jazzy" as the first word consistently)
- Adds an external dependency
- Increases latency during startup

Instead, we'll implement a simple, deterministic approach using hardcoded arrays of adjectives and nouns, selecting one randomly from each to generate a name.

## Requirements
1. Create arrays of adjectives and nouns in the nameGenerator module
2. Replace the current API-based implementation with a randomized selection from these arrays
3. Keep the same hyphenated format (e.g., "clever-otter")
4. Maintain the fallback name generation for error cases

## Files to Modify
- `/Users/phaedrus/Development/thinktank/src/utils/nameGenerator.ts` - Main implementation
- `/Users/phaedrus/Development/thinktank/src/utils/__tests__/nameGenerator.test.ts` - Test updates
- `/Users/phaedrus/Development/thinktank/src/workflow/runThinktank.ts` - Integration point (lines 324-331)

## Implementation Details
1. Create large arrays of adjectives and nouns (at least 50 of each) at the top of nameGenerator.ts
2. Modify `generateFunName()` to:
   - No longer require an API call
   - Select a random adjective and noun
   - Return in the format "adjective-noun"
   - Remove Google API dependency
3. Update tests to verify the new deterministic approach
4. Remove Google GenerativeAI import and associated code

## Expected Outcome
- Faster, more reliable name generation
- No external API dependency
- More diverse output directory names
- Consistent hyphenated format

## Testing
- Verify that names are in the correct format
- Ensure proper randomization across runs
- Confirm fallback still works when needed

## Bonus
Consider adding theme-based sets of words for more coherent naming patterns.