# BACKLOG
* **[Feature]** Always write all output to a logfile.
* **[Feature]** Allow users to add a final *synthesize* step, sending all model outputs to a chosen model for summarization.
* **[Feature]** Allow users to pass system prompts.
* **[Feature]** Allow users to pass full model configs.
* **[Feature]** Allow users to define an arbitrary number of steps in their task (e.g., Plan -> Critique -> Revise Plan).
* **[Feature]** Estimate cost per LLM request and log it. Integrate with provider cost APIs if possible.
* **[Feature]** Implement better token count handling (e.g., set max tokens per model, warn if exceeded).
* **[Feature]** Add optional context summarization step before sending prompt. Perhaps auto-summarize context for models with smaller context windows.
* **[Feature]** Automatically assemble models to query based on task and context.
* **[DX]** Review and improve clarity and actionability of error messages across the application.
* **[Feature]** Get, and keep in sync, all relevant info about every model we could be using -- max tokens, cost, general usage tips, etc
