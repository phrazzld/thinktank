# Task Description

## Issue Details
Issue #61: 🔥 GORDIAN: Eliminate Registry System - Replace 10K Lines with Simple Map
URL: https://github.com/phrazzld/thinktank/issues/61

## Overview
The registry system (`internal/registry/`) consists of 10,803 lines of configuration-driven complexity to manage metadata for only 7 models. This radical simplification will delete the entire registry system and replace it with a simple Go map, eliminating YAML configuration, loading logic, validation, and complex initialization.

## Requirements
- Remove entire `internal/registry/` package (10,800+ lines)
- Replace with simple hardcoded map containing model metadata
- Maintain existing functionality for model info retrieval
- Remove YAML configuration files and loading logic
- Simplify model metadata access to direct map lookup
- Ensure all existing code that depends on registry continues to work

## Technical Context
### Current Registry System
- Registry manager with singleton pattern (646 lines)
- YAML model definitions (274 lines)
- Complex configuration loading and validation
- Provider definitions and mappings
- Extensive test infrastructure
- Configuration caching and initialization logic

### Models Currently Supported
- gpt-4.1
- o4-mini
- gemini-2.5-pro-preview-03-25
- (Plus ~4 more models)

### Key Registry Functions to Replace
- Model metadata retrieval (context window, max tokens, default params)
- Model existence validation
- Provider determination from model name

## Related Issues
- Part of GORDIAN Meta-Issue #65: Radical Codebase Simplification Initiative
- Related to Issue #60: Eliminate Provider Abstraction Layer
- Will simplify Issue #52: Centralize API Key Resolution Logic

## Priority & Complexity
- **Priority**: High (part of critical simplification initiative)
- **Complexity**: Extra Large (size:xl) - major architectural change
- **Domain**: Configuration, Architecture
- **Expected Impact**: Remove 10,800+ lines of code
