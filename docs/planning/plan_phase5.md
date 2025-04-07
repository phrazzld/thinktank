# Thinktank Test Suite Refactoring - Phase 5

## Phase 5: Simplify CLI Testing

**Objective:** Extract command logic to enable direct testing without CLI framework overhead.

### Steps:

1. **Extract Command Handlers:**
   - Move action callback logic to dedicated functions:
     ```typescript
     // In cli/commands/run.ts
     export async function runCommandHandler(input: string, options: any): Promise<void> {
       // All the logic previously in .action()
       await runThinktank({ input, options });
     }

     // In CLI setup
     program.command('run <input>').action(runCommandHandler);
     ```

2. **Test Command Handlers Directly:**
   - Import and test like normal functions:
     ```typescript
     import { runCommandHandler } from '../commands/run';

     test('runCommand processes input', async () => {
       // Arrange
       const mockRunThinktank = jest.fn();
       jest.spyOn(workflowModule, 'runThinktank').mockImplementation(mockRunThinktank);
       
       // Act
       await runCommandHandler('prompt.txt', { model: 'gpt-4' });
       
       // Assert
       expect(mockRunThinktank).toHaveBeenCalledWith({
         input: 'prompt.txt', 
         options: { model: 'gpt-4' }
       });
     });
     ```
