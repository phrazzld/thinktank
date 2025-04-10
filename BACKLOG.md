# backlog

* enforce convention over configuration
* implement persistent, structured logging (e.g., json lines) to a configurable file path, detailing operations, inputs, outputs, token counts, errors, and final results for better auditability and programmatic use by tools like claude code.
* don't hardcode the plan prompt. force the user to pass a file with the full prompt.
* support an arbitrary number of context files and directories that we can append to the core prompt file. we want to maintain a sense, in the prompt file, of "this is the main prompt / request / ask" and "this other stuff is extra context for you to use to execute"
* set up github actions
* refactor output handling to use standard streams (stdout for primary results, stderr for logs/errors) and add a json output mode flag (`--output-format json`) for machine-readable results.
* use distinct exit codes for different outcomes (success, user error, api error, file system error) to allow programmatic result checking.
* make writing plan output to a file optional (`--output <path>`) and default to printing the plan to stdout if the flag is omitted
* allow users to define and pass custom system prompts to the models.
* enable users to provide full model configuration parameters (temperature, top-p, max output tokens, etc.) via cli flags or a configuration file
* fetch, cache, and keep updated model metadata (e.g., max tokens, cost per token, input/output limits, usage tips) from provider apis or documentation for all supported models
* implement robust token counting using provider apis, display usage relative to model limits, warn if limits are approached, and potentially halt execution if limits are exceeded
* estimate request cost based on token count and model pricing information, log the estimated cost, and potentially integrate with provider billing apis for accuracy.
* develop a flexible multi-step workflow engine allowing users to define arbitrary sequences of operations (e.g., plan -> critique -> revise -> synthesize) via configuration or cli flags.
* add a built-in synthesis step where outputs from multiple preceding steps (e.g., multiple model responses, critiques) are sent to a final model for summarization or consolidation.
* support querying multiple models concurrently for the same task (`--model model_a,model_b`) to get a "council of experts" perspective, returning labeled outputs.
* implement a built-in plan -> critique -> refine workflow activated by a specific flag (e.g., `--critique-refine`).
* add an optional context preprocessing step to summarize large context inputs before sending them to the llm, potentially triggered automatically for models with smaller context windows.
* explore automatically selecting the most appropriate model or models based on the task description and context size/type.
* integrate abstract syntax tree (ast) parsing for supported languages to provide richer structural context to the llm beyond raw text.
* add git integration to use `git diff` output as context (`--context-git-diff <ref>`) or optionally include `git blame` / commit history for context files.
* investigate integrating with language servers or symbol indexing tools (ctags, sourcegraph api) to provide symbol definition/usage information as context.
* support generating output plans in structured formats like json or yaml in addition to markdown.
* add a code generation mode to attempt generating code snippets or files based on the generated plan.
* provide an option to output suggested code changes (especially for refactoring) as a `.patch` file.
* make `architect` aware of claude code's memory files (`CLAUDE.md`, `CLAUDE.local.md`) to read configuration settings, respecting the same hierarchy
* investigate presenting `architect` itself as a tool to claude code, potentially via mcp, defining its capabilities for planning, critique, and refinement
* review and significantly improve the clarity, detail, and actionability of all error messages throughout the application
* remove user-facing niceties like spinners and excessive color formatting in favor of clean, programmable output
* extend the interactive task clarification feature (`--clarify`) to allow interactive refinement of the *generated plan* itself
* enhance token count handling to allow setting maximum tokens per model and provide clearer warnings or errors if limits are exceeded
* add metadata (file paths, git status) to the context provided to the llm.
