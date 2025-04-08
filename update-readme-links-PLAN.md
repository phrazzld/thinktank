# Update README.md Links - Implementation Plan

## Task Title
Update README.md links

## Goal
Fix broken links in README.md, ensuring they point to valid documentation resources.

## Chosen Approach
**Pragmatic Fix with Internal Linking** - A hybrid approach combining elements from several suggested solutions, focusing on fixing broken links while maintaining usability and documentation integrity.

### Implementation Steps

1. **Fix TESTING.md link (line 5)**
   - Change `[testing guide](TESTING.md)` to `[testing philosophy](TESTING_PHILOSOPHY.md)`
   - This redirects to the existing philosophy document while accurately describing its content

2. **Fix Context Formatting Documentation links (lines 5 & 130)**
   - Determine if the "Including Context Files and Directories" section in README.md already explains this feature
   - If yes: Convert both links to internal anchors pointing to that section
   - Change line 5 link to `[context formatting](#including-context-files-and-directories)`
   - Change line 130 link to internal reference `See the [section above](#including-context-files-and-directories) for details.`

3. **Review Project Planning link (line 5)**
   - Verify the link to `docs/planning/master-plan.md` points to the existing file
   - Add a note in the master-plan.md file indicating its relationship to other planning documents if needed

4. **Add comment to TODO.md for future documentation improvement**
   - Create a reminder in TODO.md to potentially restore full documentation files later
   - This acknowledges the temporary nature of the fix while ensuring it's tracked

## Reasoning for This Approach

1. **Testability Considerations**
   - This approach requires minimal changes, reducing the risk of introducing new errors
   - The changes can be easily tested by verifying all links now point to existing resources
   - Internal links (anchors) work reliably within Markdown and can be tested directly in GitHub's preview

2. **Maintainability Benefits**
   - Uses existing resources rather than creating stub files that could become outdated
   - Clearly separates immediate fixes from longer-term documentation needs
   - Internal links keep related content together rather than fragmenting it
   - Avoids documentation duplication by leveraging already-written sections

3. **Practical Implementation**
   - Balances the need for immediate fixes with the reality of the project's current state
   - Acknowledges documentation evolution by using accurate link text that matches content
   - Keeps users within the main README for key feature explanations, improving user experience

This approach prioritizes fixing the immediate issues with minimal changes while ensuring all links point to useful, relevant content. By leveraging internal linking, we maintain documentation coherence without requiring extensive content creation or restructuring.