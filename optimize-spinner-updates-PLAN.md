# Optimize Spinner Updates: Reduce terminal flicker

## Goal
Implement debouncing or batching for spinner updates to reduce visual flicker and improve user experience.

## Analysis of Current Implementation
After examining the codebase, I found that the spinner text is updated very frequently in several places:
1. During model query execution, the spinner text updates for each model status change
2. During file output processing, the spinner updates for each file write operation
3. During various workflow steps with sequential status messages

These frequent updates likely cause visual flicker as the terminal rapidly redraws, which can be distracting and reduce the perceived quality of the user interface.

## Potential Approaches

### Approach 1: Debounce Spinner Updates
- Implement a debouncing mechanism that delays spinner text updates until a specified time has passed
- Only show the most recent update if multiple updates occur within the debounce period
- Pros: Reduces update frequency while still showing meaningful progress
- Cons: Adds complexity; may lose some granularity in status reporting

### Approach 2: Batch Updates with Rate Limiting
- Implement a rate-limiting mechanism that limits spinner updates to a maximum frequency
- Batch updates together if they occur more frequently than the limit
- Pros: Ensures consistent update frequency; maintains a balance between feedback and flicker
- Cons: May require more complex state management

### Approach 3: Smart Spinner with Update Throttling
- Create a wrapper around the ora spinner with built-in throttling capability
- Maintain an internal state of the "true" spinner text, but only update the visible text periodically
- Pros: Encapsulates optimization logic; easy to adjust throttling parameters
- Cons: Requires understanding ora's internal implementation

## Selected Approach
I'll implement **Approach 3: Smart Spinner with Update Throttling** for the following reasons:

1. It provides the best balance between responsive feedback and reduced flicker
2. By creating a wrapper class, the solution is encapsulated and doesn't require changes throughout the codebase
3. This approach aligns with the project's modular design principles
4. The throttling parameters can be easily adjusted if needed

The implementation will:
- Create a `ThrottledSpinner` class that wraps the ora spinner
- Limit visible updates to a reasonable frequency (e.g., maximum 5 updates per second)
- Maintain internal state of the "true" spinner text
- Ensure critical updates (like warnings, errors, or completion messages) bypass throttling
- Provide a drop-in replacement for ora's spinner with compatible API

This approach will significantly reduce visual flicker while maintaining a good user experience with appropriate status feedback.