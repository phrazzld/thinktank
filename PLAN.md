# Implementation Plan: Add Synthesis Step (T13)

## 1. Context Analysis

### Current Architecture
- The `thinktank` CLI currently runs multiple models in parallel via the `Orchestrator.Run` method, which processes each model through `processModelWithRateLimit`
- Each model's output is saved individually to a separate file (e.g., `<output_dir>/<model_name>.md`)
- There is no mechanism to collect and synthesize outputs from multiple models

### Key Components to Modify
- **CLI Parsing**: `cmd/thinktank/cli.go` - Need to add new `--synthesis-model` flag
- **Configuration**: `internal/config/config.go` - Need to add field for synthesis model name
- **Orchestration**: `internal/thinktank/orchestrator/orchestrator.go` - Major changes needed to collect outputs and add synthesis step
- **Model Processing**: `internal/thinktank/modelproc/processor.go` - Modify to return content in addition to writing to file
- **Prompt Generation**: `internal/thinktank/prompt/prompt.go` - Add new function for synthesis prompt
- **API Service**: Existing interfaces will be used for synthesis model interaction

### Interface Changes
- The `modelproc.Processor.Process` signature needs to change to return content
- The `Orchestrator.processModels` needs to collect and return outputs from all models

## 2. Implementation Strategy

### 2.1 Add CLI Flag and Configuration
1. **Add CLI Flag**:
   - Update `cmd/thinktank/cli.go`:
   ```go
   synthesisModelFlag := flagSet.String("synthesis-model", "", "Optional: Model to use for synthesizing results from multiple models.")
   ```
   - Update help/usage output

2. **Update Configuration**:
   - Add `SynthesisModel string` field to `CliConfig` in `internal/config/config.go`
   - Update `ParseFlagsWithEnv` to store flag value in config
   ```go
   cfg.SynthesisModel = *synthesisModelFlag
   ```
   - Add validation in `ValidateInputsWithEnv` to check if synthesis model exists in registry

### 2.2 Modify Model Processor to Return Content
1. **Update Process Method Signature**:
   - Change from `Process(ctx context.Context, modelName string, stitchedPrompt string) error` to `Process(ctx context.Context, modelName string, stitchedPrompt string) (string, error)`
   - Return the generated content instead of only writing to file
   ```go
   // After ProcessLLMResponse
   if err != nil {
       return "", fmt.Errorf(...) // Return empty string on error
   }
   return generatedOutput, nil
   ```

### 2.3 Update Orchestrator for Result Collection and Synthesis
1. **Modify `processModelWithRateLimit`**:
   - Create a result struct to capture model output and errors
   ```go
   type modelResult struct {
       modelName string
       content   string
       err       error
   }
   ```
   - Update channel to use result struct instead of just errors

2. **Update `processModels`**:
   - Collect and return both outputs and errors
   ```go
   modelOutputs := make(map[string]string)
   var modelErrors []error
   for result := range resultChan {
       if result.err != nil {
           modelErrors = append(modelErrors, result.err)
       } else {
           modelOutputs[result.modelName] = result.content
       }
   }
   return modelOutputs, modelErrors
   ```
   - Change return type to `(map[string]string, []error)`

3. **Modify `Run` Method**:
   - Call updated `processModels`: `modelOutputs, modelErrors := o.processModels(ctx, stitchedPrompt)`
   - Add conditional logic for synthesis:
   ```go
   if o.config.SynthesisModel == "" {
       // No synthesis: Write individual files
       for modelName, content := range modelOutputs {
           // Save individual files using o.fileWriter
       }
   } else {
       // Synthesis required
       if len(modelOutputs) > 0 {
           synthesisContent, err := o.synthesizeResults(ctx, instructions, modelOutputs)
           if err != nil {
               // Handle synthesis error
           } else {
               // Save synthesis file using o.fileWriter
           }
       }
   }
   ```

### 2.4 Implement Synthesis Logic
1. **Add Prompt Function**:
   - Create `StitchSynthesisPrompt` in `internal/thinktank/prompt/prompt.go`:
   ```go
   func StitchSynthesisPrompt(originalInstructions string, modelOutputs map[string]string) string {
       var builder strings.Builder
       // Format original instructions and model outputs with clear delimiters
       builder.WriteString("<instructions>\n")
       builder.WriteString(originalInstructions)
       builder.WriteString("\n</instructions>\n\n")

       builder.WriteString("<model_outputs>\n")
       for modelName, output := range modelOutputs {
           builder.WriteString(fmt.Sprintf("<output model=\"%s\">\n", modelName))
           builder.WriteString(output)
           builder.WriteString("\n</output>\n\n")
       }
       builder.WriteString("</model_outputs>\n\n")

       builder.WriteString("Please synthesize these outputs into a single, consolidated summary that addresses the original instructions...")

       return builder.String()
   }
   ```

2. **Implement Synthesis Method**:
   - Add `synthesizeResults` method to `Orchestrator`:
   ```go
   func (o *Orchestrator) synthesizeResults(ctx context.Context, originalInstructions string, modelOutputs map[string]string) (string, error) {
       // Build synthesis prompt
       synthesisPrompt := prompt.StitchSynthesisPrompt(originalInstructions, modelOutputs)

       // Get client for synthesis model
       client, err := o.apiService.InitLLMClient(ctx, o.config.SynthesisModel)
       if err != nil {
           return "", fmt.Errorf("failed to initialize synthesis model client: %w", err)
       }
       defer client.Close()

       // Call model API
       result, err := client.GenerateContent(ctx, synthesisPrompt)
       if err != nil {
           return "", fmt.Errorf("synthesis model API call failed: %w", err)
       }

       // Process response
       synthesisOutput, err := o.apiService.ProcessLLMResponse(result)
       if err != nil {
           return "", fmt.Errorf("failed to process synthesis response: %w", err)
       }

       return synthesisOutput, nil
   }
   ```

3. **Add Audit Logging**:
   - Include log entries for synthesis start/end, model used, success/failure

## 3. Testing Strategy

### 3.1 Unit Tests
1. **CLI and Config Tests**:
   - Test parsing and validation of `--synthesis-model` flag
   - Test `config.CliConfig` handles `SynthesisModel` field correctly

2. **Prompt Tests**:
   - Test `StitchSynthesisPrompt` with various inputs:
     - Empty map of outputs
     - Single model output
     - Multiple model outputs
     - Long content

3. **Orchestrator Tests**:
   - Mock `APIService`, `ContextGatherer`, `FileWriter`
   - Test `Run` method logic:
     - Without synthesis flag -> individual files saved
     - With synthesis flag -> synthesis occurs and single file saved
   - Test `synthesizeResults` method:
     - Mock API calls and verify correct prompt is passed
     - Verify error handling for API failures

### 3.2 Integration Tests
1. **End-to-End Flow Tests**:
   - Test without synthesis flag -> multiple output files
   - Test with synthesis flag -> single synthesis output file
   - Test with failed model + synthesis flag -> synthesis proceeds with available outputs

2. **Edge Case Tests**:
   - Test with invalid synthesis model name
   - Test with missing API key for synthesis model
   - Test with synthesis model failing

### 3.3 E2E Tests
1. **Create New E2E Test File**:
   - `internal/e2e/cli_synthesis_test.go`
   - Configure mock API to handle synthesis model
   - Test full workflow with multiple models and synthesis

2. **Test Cases**:
   - Basic synthesis workflow
   - Handling of primary model failures
   - Error propagation from synthesis step

## 4. Documentation

### 4.1 User Documentation
1. **Update README.md**:
   - Add `--synthesis-model <model_name>` to options table
   - Explain purpose: "Optional flag to specify a model for summarizing outputs from primary models into a single result"
   - Add example command showing synthesis usage

2. **CLI Help**:
   - Update flag descriptions in CLI help output

### 4.2 Code Documentation
1. **Add/Update GoDoc Comments**:
   - Document new `SynthesisModel` field in `CliConfig`
   - Document updated method signatures
   - Document new methods and functions
   - Explain synthesis workflow in orchestrator

## 5. Implementation Sequence

1. Add CLI flag and config field
2. Modify `ModelProcessor.Process` to return content
3. Update orchestrator to collect model outputs
4. Implement synthesis prompt builder
5. Implement synthesis method in orchestrator
6. Update result handling in orchestrator
7. Add unit tests for all components
8. Add integration and E2E tests
9. Update documentation
