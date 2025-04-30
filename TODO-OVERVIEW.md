# Code Size Optimization - Work Breakdown

This directory contains task lists for the code size optimization project, broken down into logical sets for separate branches/PRs.

## Sets Overview

1. **[TODO-DEAD-CODE.md](TODO-DEAD-CODE.md)** - Removing obsolete token code, disabled files, legacy API services (P0, first priority)
2. **[TODO-MOCKS.md](TODO-MOCKS.md)** - Consolidating mock implementations in test files (P1)
3. **[TODO-PROVIDERS.md](TODO-PROVIDERS.md)** - Centralizing provider logic and error handling (P2)
4. **[TODO-API-KEYS.md](TODO-API-KEYS.md)** - Centralizing API key resolution (P2)
5. **[TODO-ADAPTERS.md](TODO-ADAPTERS.md)** - Flattening adapter layers (P2)
6. **[TODO-STDLIB.md](TODO-STDLIB.md)** - Using standard library helpers (P3, low impact)
7. **[TODO-LOGGING.md](TODO-LOGGING.md)** - Pruning verbose logging and unnecessary comments (P3)

## Implementation Order

The suggested implementation order follows the priority levels in the tasks:

1. **Dead Code Elimination** - Has the highest impact for the lowest risk
2. **Mock Consolidation** - Major code reduction with moderate risk
3. **Provider Logic/Error Handling** - Medium impact with moderate complexity
4. **API Key Resolution & Adapter Flattening** - Smaller improvements with low risk
5. **Standard Library & Logging** - Minor improvements with very low risk

## Dependencies

Note that each set maintains internal dependencies between its tasks, but there are minimal cross-set dependencies, allowing parallel work if desired.

## Task IDs

Task IDs (T001, T002, etc.) are maintained across all files to ensure unique identification and proper dependency tracking.
