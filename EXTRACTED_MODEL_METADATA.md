# Extracted Model Metadata from config/models.yaml

This document captures the exact configuration values for all 7 models currently defined in `config/models.yaml` that will be converted to hardcoded Go definitions.

## Provider to Environment Variable Mapping

```yaml
api_key_sources:
  openai: "OPENAI_API_KEY"
  gemini: "GEMINI_API_KEY"
  openrouter: "OPENROUTER_API_KEY"
```

## Provider Base URLs

```yaml
providers:
  - name: openai        # No custom base_url (uses OpenAI default)
  - name: gemini        # No custom base_url (uses Gemini default)
  - name: openrouter    # Uses https://openrouter.ai/api/v1 (hardcoded in system)
```

## Model Definitions (7 Total)

### 1. gpt-4.1 (OpenAI)
```yaml
name: gpt-4.1
provider: openai
api_model_id: gpt-4.1
context_window: 1000000
max_output_tokens: 200000
parameters:
  temperature: {type: float, default: 0.7}
  top_p: {type: float, default: 1.0}
  frequency_penalty: {type: float, default: 0.0}
  presence_penalty: {type: float, default: 0.0}
```

### 2. o4-mini (OpenAI)
```yaml
name: o4-mini
provider: openai
api_model_id: o4-mini
context_window: 200000
max_output_tokens: 200000
parameters:
  temperature: {type: float, default: 1.0}
  top_p: {type: float, default: 1.0}
  frequency_penalty: {type: float, default: 0.0}
  presence_penalty: {type: float, default: 0.0}
  reasoning: {type: object, default: {effort: "high"}}
```

### 3. gemini-2.5-pro (Gemini)
```yaml
name: gemini-2.5-pro
provider: gemini
api_model_id: gemini-2.5-pro
context_window: 1000000
max_output_tokens: 65000
parameters:
  temperature: {type: float, default: 0.7}
  top_p: {type: float, default: 0.95}
  top_k: {type: int, default: 40}
```

### 4. gemini-2.5-flash (Gemini)
```yaml
name: gemini-2.5-flash
provider: gemini
api_model_id: gemini-2.5-flash
context_window: 1000000
max_output_tokens: 65000
parameters:
  temperature: {type: float, default: 0.7}
  top_p: {type: float, default: 0.95}
  top_k: {type: int, default: 40}
```

### 5. openrouter/deepseek/deepseek-chat-v3-0324 (OpenRouter)
```yaml
name: openrouter/deepseek/deepseek-chat-v3-0324
provider: openrouter
api_model_id: deepseek/deepseek-chat-v3-0324
context_window: 65536
max_output_tokens: 8192
parameters:
  temperature: {type: float, default: 0.7}
  top_p: {type: float, default: 0.95}
```

### 6. openrouter/deepseek/deepseek-r1 (OpenRouter)
```yaml
name: openrouter/deepseek/deepseek-r1
provider: openrouter
api_model_id: deepseek/deepseek-r1
context_window: 131072
max_output_tokens: 33792
parameters:
  temperature: {type: float, default: 0.7}
  top_p: {type: float, default: 0.95}
```

### 7. openrouter/x-ai/grok-3-beta (OpenRouter)
```yaml
name: openrouter/x-ai/grok-3-beta
provider: openrouter
api_model_id: x-ai/grok-3-beta
context_window: 131072
max_output_tokens: 131072
parameters:
  temperature: {type: float, default: 0.7}
  top_p: {type: float, default: 0.95}
```

## Summary for Go Implementation

**Total Models**: 7 active models (3 providers)
- **OpenAI**: 2 models (gpt-4.1, o4-mini)
- **Gemini**: 2 models (gemini-2.5-pro, gemini-2.5-flash)
- **OpenRouter**: 3 models (deepseek-chat-v3-0324, deepseek-r1, grok-3-beta)

**Default Parameters to Preserve**:
- All models have `temperature` (0.7 or 1.0)
- OpenAI models have `top_p`, `frequency_penalty`, `presence_penalty`
- o4-mini has special `reasoning` object parameter
- Gemini models have `top_p` and `top_k`
- OpenRouter models have `top_p`

**Environment Variables Required**:
- `OPENAI_API_KEY` for OpenAI models
- `GEMINI_API_KEY` for Gemini models
- `OPENROUTER_API_KEY` for OpenRouter models

**Base URLs**:
- OpenAI: Use SDK default
- Gemini: Use SDK default
- OpenRouter: "https://openrouter.ai/api/v1"
