# Execution Flow

This document describes thinktank's main execution flow with state diagrams.

## High-Level Flow

```mermaid
flowchart TD
    A[CLI: Parse Args] --> B{Valid?}
    B -->|No| E[Exit 1: Error]
    B -->|Yes| C[App: Execute]
    C --> D[Orchestrator: Run]
    D --> F{Dry Run?}
    F -->|Yes| G[Display Stats]
    G --> H[Exit 0]
    F -->|No| I[Gather Context]
    I --> J[Build Prompt]
    J --> K[Process Models]
    K --> L{Synthesis?}
    L -->|Yes| M[Synthesize]
    L -->|No| N[Write Outputs]
    M --> N
    N --> O{All OK?}
    O -->|Yes| H
    O -->|Partial| P[Exit 0/1 based on config]
    O -->|All Failed| E
```

## Orchestrator State Machine

The orchestrator transitions through these states during execution:

```mermaid
stateDiagram-v2
    [*] --> Setup: Run() called
    Setup --> GatheringContext: Setup complete
    Setup --> Failed: Setup error

    GatheringContext --> DryRun: Dry run mode
    GatheringContext --> BuildingPrompt: Files gathered
    GatheringContext --> Failed: Gather error

    DryRun --> [*]: Display stats

    BuildingPrompt --> ProcessingModels: Prompt built

    ProcessingModels --> SavingOutputs: All models done
    ProcessingModels --> PartialSuccess: Some models failed

    PartialSuccess --> SavingOutputs: Has usable outputs
    PartialSuccess --> Failed: All models failed

    SavingOutputs --> Synthesizing: Synthesis enabled
    SavingOutputs --> DisplayingSummary: No synthesis

    Synthesizing --> DisplayingSummary: Done

    DisplayingSummary --> [*]: Success
    Failed --> [*]: Return error
```

## Model Processing (Concurrent)

Each model is processed concurrently with rate limiting:

```mermaid
sequenceDiagram
    participant O as Orchestrator
    participant RL as RateLimiter
    participant API as OpenRouter
    participant W as OutputWriter

    O->>O: Start goroutines for N models

    par Model 1
        O->>RL: Acquire token
        RL-->>O: OK
        O->>API: GenerateContent
        API-->>O: Response
        O->>W: Write output
    and Model 2
        O->>RL: Acquire token
        RL-->>O: Wait (rate limited)
        RL-->>O: OK
        O->>API: GenerateContent
        API-->>O: Response
        O->>W: Write output
    and Model N
        O->>RL: Acquire token
        RL-->>O: OK
        O->>API: GenerateContent
        API-->>O: Error (rate limit)
        O->>O: Record failure
    end

    O->>O: Collect results
    O->>O: Check success threshold
```

## Error Handling States

```mermaid
stateDiagram-v2
    [*] --> Processing
    Processing --> Success: All models OK
    Processing --> PartialSuccess: Some models OK
    Processing --> AllFailed: No models OK

    Success --> WriteOutputs
    PartialSuccess --> CheckConfig: --partial-success-ok?
    AllFailed --> ReturnError

    CheckConfig --> WriteOutputs: Flag set
    CheckConfig --> ReturnError: Flag not set

    WriteOutputs --> DisplaySummary
    DisplaySummary --> [*]
    ReturnError --> [*]
```

## Key Decision Points

1. **Dry Run Check**: Short-circuits before any API calls
2. **Partial Success**: Configurable behavior via `--partial-success-ok`
3. **Synthesis**: Only runs if `--synthesis` flag or auto-selected for large inputs
4. **Rate Limiting**: Per-model limits for models with concurrency restrictions
