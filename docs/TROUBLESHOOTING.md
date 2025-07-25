# Troubleshooting Guide

This guide helps you diagnose and resolve common issues when using thinktank. Issues are organized by symptoms you'll observe, with step-by-step solutions.

## Quick Diagnosis

If you're experiencing issues, start here for rapid diagnosis:

```bash
# Enable verbose logging to see detailed error information
thinktank --verbose --instructions task.txt ./your-project

# Or use dry-run to test configuration without API calls
thinktank --dry-run --instructions task.txt ./your-project
```

**Most Common Issues:**
1. [API key problems](#authentication-errors) (60% of failures)
2. [Rate limiting](#rate-limiting-errors) (20% of failures)
3. [Input too large](#input-size-errors) (10% of failures)
4. [Network connectivity](#network-and-connectivity-errors) (10% of failures)

---

## Authentication Errors

### Symptoms
- "Authentication failed" or "unauthorized" errors
- HTTP 401/403 status codes
- "Invalid API key" messages

### Quick Solutions

All models now use OpenRouter for unified API access.

#### OpenRouter Authentication
```bash
# Check if API key is set
echo $OPENROUTER_API_KEY

# Key should start with 'sk-or-' and be ~64 characters
# If missing or incorrect:
export OPENROUTER_API_KEY="sk-or-your-actual-openrouter-key-here"

# Get your key at: https://openrouter.ai/keys
```

**Migration Note**: If you have old `OPENAI_API_KEY` or `GEMINI_API_KEY` environment variables set, the tool will display helpful migration messages guiding you to use `OPENROUTER_API_KEY` instead.

### Advanced Diagnostics

#### Check API Key Validity
```bash
# Test OpenRouter key (all models now use this)
curl -H "Authorization: Bearer $OPENROUTER_API_KEY" \
  https://openrouter.ai/api/v1/models
```

#### Common Auth Issues

| Issue | Solution |
|-------|----------|
| Key expired/invalid | Visit [OpenRouter](https://openrouter.ai/keys) to regenerate |
| Wrong key format | OpenRouter keys start with 'sk-or-', not 'sk-' |
| Account not activated | Check [OpenRouter account](https://openrouter.ai/account) status |
| Insufficient credits | Add credits to your [OpenRouter account](https://openrouter.ai/account) |

---

## Rate Limiting Errors

### Symptoms
- "Rate limit exceeded" or "Too many requests" messages
- HTTP 429 status codes
- Long delays between model processing
- Models showing as "rate limited" in output

### Quick Solutions

#### Reduce Request Rate
```bash
# Conservative approach - lower concurrency and rate limits
thinktank --instructions task.txt \
  --max-concurrent 2 \
  --openai-rate-limit 100 \
  --gemini-rate-limit 15 \
  --openrouter-rate-limit 10 \
  ./your-project
```

#### Provider-Specific Quick Fixes

**OpenAI 429 Errors:**
```bash
# For free/low-tier accounts
thinktank --instructions task.txt \
  --openai-rate-limit 20 \
  --max-concurrent 2 \
  ./your-project

# For paid accounts experiencing issues
thinktank --instructions task.txt \
  --openai-rate-limit 500 \
  --max-concurrent 5 \
  ./your-project
```

**Gemini Rate Limiting:**
```bash
# For free tier
thinktank --instructions task.txt \
  --gemini-rate-limit 10 \
  --max-concurrent 2 \
  ./your-project

# For paid tier
thinktank --instructions task.txt \
  --gemini-rate-limit 1000 \
  --max-concurrent 8 \
  ./your-project
```

**OpenRouter Issues:**
```bash
# Conservative for mixed model usage
thinktank --instructions task.txt \
  --openrouter-rate-limit 5 \
  --max-concurrent 1 \
  ./your-project

# With sufficient account balance ($10+)
thinktank --instructions task.txt \
  --openrouter-rate-limit 50 \
  --max-concurrent 3 \
  ./your-project
```

### Advanced Rate Limiting Diagnostics

#### Check Your Account Tiers

**OpenAI Tier Check:**
- Visit [OpenAI Usage](https://platform.openai.com/usage)
- Look for "Rate limits" section
- Tier 1: 500 RPM, Tier 2: 5000 RPM, etc.

**Gemini Quota Check:**
- Visit [Google AI Studio](https://makersuite.google.com/)
- Check "Quota" in project settings
- Free: 15 RPM, Paid: 1000+ RPM

**OpenRouter Balance Check:**
- Visit [OpenRouter Account](https://openrouter.ai/account)
- $10+ balance required for higher limits
- Free models: Fixed 20 RPM limit

#### Model-Specific Rate Limits

Some models have special limits regardless of provider settings:

| Model | Rate Limit | Note |
|-------|------------|------|
| `openrouter/deepseek/deepseek-r1-0528` | 5 RPM | Reasoning model |
| `openrouter/deepseek/deepseek-r1-0528:free` | 3 RPM | Free tier |

**For these models, use:**
```bash
thinktank --instructions task.txt \
  --max-concurrent 1 \
  --model openrouter/deepseek/deepseek-r1-0528 \
  ./your-project
```

For comprehensive rate limiting guidance, see: [README.md - Rate Limiting & Performance Optimization](../README.md#rate-limiting--performance-optimization)

---

## Input Size Errors

### Symptoms
- "Token limit exceeded" or "Input too long" messages
- "Maximum context length" errors
- Models failing on large codebases

### Quick Solutions

#### Reduce Input Size
```bash
# Filter to specific file types
thinktank --instructions task.txt \
  --include .go,.md \
  ./your-project

# Exclude large or irrelevant files
thinktank --instructions task.txt \
  --exclude node_modules,vendor,dist,.git \
  ./your-project

# Exclude specific patterns
thinktank --instructions task.txt \
  --exclude-names "*.log,*.tmp,*_test.go" \
  ./your-project
```

#### Target Specific Directories
```bash
# Instead of analyzing entire project
thinktank --instructions task.txt ./entire-project

# Focus on specific components
thinktank --instructions task.txt ./src ./docs
```

### Advanced Input Management

#### Estimate Token Usage
```bash
# Use dry-run to see what files would be processed
thinktank --dry-run --instructions task.txt ./your-project
```

#### Model Context Windows

| Model | Context Window | Best For |
|-------|---------------|----------|
| `gpt-4.1` | 1M tokens | Large codebases |
| `gemini-2.5-pro` | 1M tokens | Large codebases |
| `gemini-2.5-flash` | 1M tokens | Large codebases |
| `o4-mini` | 200K tokens | Medium projects |
| Most OpenRouter models | 64K-200K tokens | Focused analysis |

#### Smart Filtering Strategies

**For Web Projects:**
```bash
thinktank --instructions task.txt \
  --include .js,.ts,.jsx,.tsx,.html,.css \
  --exclude node_modules,dist,build,.next \
  ./your-project
```

**For Go Projects:**
```bash
thinktank --instructions task.txt \
  --include .go,.md \
  --exclude vendor,_test.go \
  ./your-project
```

**For Documentation:**
```bash
thinktank --instructions task.txt \
  --include .md,.txt,.rst \
  ./your-project
```

---

## Network and Connectivity Errors

### Symptoms
- "Connection timeout" or "Network error" messages
- "Failed to connect" errors
- Intermittent failures that succeed on retry

### Quick Solutions

#### Test Connectivity
```bash
# Test basic internet connectivity
curl -I https://www.google.com

# Test specific provider endpoints
curl -I https://api.openai.com
curl -I https://generativelanguage.googleapis.com
curl -I https://openrouter.ai
```

#### Retry with Increased Timeouts
```bash
# Most network issues are temporary - try again
thinktank --instructions task.txt ./your-project

# If problems persist, retry individual models
thinktank --instructions task.txt \
  --model gpt-4.1 \
  ./your-project
```

### Advanced Network Diagnostics

#### Check for Proxy/Firewall Issues
```bash
# Check proxy settings
echo $HTTP_PROXY
echo $HTTPS_PROXY

# If behind corporate firewall, check DNS resolution
nslookup api.openai.com
nslookup generativelanguage.googleapis.com
nslookup openrouter.ai
```

#### Regional Connectivity Issues

Some providers may have regional restrictions or performance differences:

| Provider | Common Issues | Solutions |
|----------|---------------|-----------|
| OpenAI | Rate limits vary by region | Try different time of day |
| Gemini | Some regions restricted | Check [supported regions](https://ai.google.dev/available_regions) |
| OpenRouter | Varies by underlying model | Try different models |

---

## Model-Specific Issues

### DeepSeek R1 Models

**Symptoms:**
- High failure rate in batch processing
- Models work individually but fail in groups
- "Concurrency conflicts" in logs

**Solutions:**
```bash
# Force sequential processing for R1 models
thinktank --instructions task.txt \
  --max-concurrent 1 \
  --model openrouter/deepseek/deepseek-r1-0528 \
  ./your-project

# Use very low rate limits
thinktank --instructions task.txt \
  --openrouter-rate-limit 3 \
  --model openrouter/deepseek/deepseek-r1-0528:free \
  ./your-project
```

### Content Filtering Issues

**Symptoms:**
- "Content blocked by safety filters" messages
- "Content policy violation" errors
- Empty responses from models

**Provider-Specific Guidance:**

**OpenAI Content Policy:**
- Review [OpenAI Usage Policies](https://openai.com/policies/usage-policies)
- Avoid: violent content, illegal activities, personal info
- Try rephrasing prompts to be more neutral

**Gemini Safety Filters:**
- Google has strict content filtering
- Try different phrasing or break down complex requests
- Avoid: medical advice, legal advice, controversial topics

**OpenRouter Filtering:**
- Varies by underlying model provider
- Try different models if one model blocks content
- Some models (like Claude) have stricter filtering

**Workarounds:**
```bash
# Try multiple models to find one that works
thinktank --instructions task.txt \
  --model gpt-4.1 --model gemini-2.5-pro --model openrouter/meta-llama/llama-3.3-70b-instruct \
  --partial-success-ok \
  ./your-project

# Use synthesis to combine results
thinktank --instructions task.txt \
  --model gpt-4.1 --model gemini-2.5-flash \
  --synthesis-model gpt-4.1 \
  ./your-project
```

---

## Error Code Reference

### HTTP Status Codes

| Code | Meaning | Common Causes | Quick Fix |
|------|---------|---------------|-----------|
| 400 | Bad Request | Invalid parameters, malformed request | Check model name and parameters |
| 401 | Unauthorized | Invalid/missing API key | Check API key environment variable |
| 403 | Forbidden | Valid key but no access | Check account permissions/billing |
| 404 | Not Found | Invalid model name | Verify model name spelling |
| 429 | Too Many Requests | Rate limit exceeded | Reduce rate limits or wait |
| 500 | Server Error | Provider-side issue | Wait and retry |
| 502/503 | Service Unavailable | Provider maintenance/outage | Wait and retry |

### thinktank Error Categories

When you see categorized errors in output, here's what they mean:

| Category | Meaning | Typical Causes | Quick Action |
|----------|---------|----------------|--------------|
| `Auth` | Authentication failed | Invalid API key | Check environment variables |
| `RateLimit` | Too many requests | Exceeding provider limits | Reduce rate limits |
| `InvalidRequest` | Bad request format | Wrong parameters | Check model name/parameters |
| `NotFound` | Resource not found | Typo in model name | Verify model name |
| `Server` | Provider server issue | Temporary outage | Wait and retry |
| `Network` | Connectivity problem | Internet/proxy issues | Check network connection |
| `InputLimit` | Input too large | Codebase too big | Use filtering flags |
| `ContentFiltered` | Content blocked | Safety filters triggered | Rephrase prompt |
| `InsufficientCredits` | No credits/quota | Account billing issue | Check account balance |

---

## Advanced Debugging

### Enable Detailed Logging
```bash
# Get maximum diagnostic information
thinktank --verbose --instructions task.txt ./your-project

# Save logs for analysis
thinktank --verbose --instructions task.txt ./your-project 2> debug.log

# Check logs for specific error patterns
grep -i "error\|fail\|timeout" debug.log
```

### Incremental Testing

#### Test Individual Components
```bash
# Test single model
thinktank --instructions task.txt \
  --model gpt-4.1 \
  ./your-project

# Test single file
thinktank --instructions task.txt \
  ./your-project/single-file.go

# Test with minimal input
echo "Test content" > test.txt
thinktank --instructions "Analyze this file" test.txt
```

#### Isolate Variables
```bash
# Test each provider separately
thinktank --instructions task.txt --model gpt-4.1 ./project          # OpenAI
thinktank --instructions task.txt --model gemini-2.5-pro ./project   # Gemini
thinktank --instructions task.txt --model openrouter/meta-llama/llama-3.3-70b-instruct ./project  # OpenRouter

# Test different concurrency levels
thinktank --instructions task.txt --max-concurrent 1 ./project
thinktank --instructions task.txt --max-concurrent 3 ./project
thinktank --instructions task.txt --max-concurrent 10 ./project
```

### Performance Monitoring
```bash
# Monitor system resources during processing
top -p $(pgrep thinktank)

# Check network usage
netstat -i

# Monitor API response times with timestamps
thinktank --verbose --instructions task.txt ./project | ts '[%Y-%m-%d %H:%M:%S]'
```

---

## Tokenization Issues

### Symptoms
- "Tokenizer initialization failed" messages
- Inconsistent token count estimates between runs
- Models selected that can't handle input size
- Performance degradation during token counting

### Quick Solutions

#### Tokenizer Fallback to Estimation
```bash
# Check which tokenizer is being used
thinktank instructions.txt ./project --dry-run --verbose

# Look for logs like:
# "Using tiktoken for OpenAI models"
# "Falling back to estimation for model: model-name"
```

**Common Causes:**
- Missing tokenizer dependencies (rare - bundled with binary)
- Circuit breaker triggered due to repeated failures
- Network issues preventing tokenizer initialization

**Solutions:**
```bash
# Clear any cached tokenizer state
thinktank instructions.txt ./project --dry-run  # Fresh tokenizer initialization

# Force reinitialization by waiting for circuit breaker reset (30 seconds)
sleep 30 && thinktank instructions.txt ./project --dry-run
```

#### Inaccurate Model Selection
```bash
# Verify model compatibility with accurate counts
thinktank instructions.txt ./large-project --dry-run

# Check output for:
# "Compatible models: gpt-4.1 (180K tokens), gemini-2.5-pro (200K tokens)"
# "Skipped models: o4-mini (exceeds 128K context limit)"
```

**Token Count Accuracy by Provider:**
| Provider | Accuracy | Method |
|----------|----------|--------|
| OpenAI   | 99%+     | tiktoken |
| Gemini   | 99%+     | SentencePiece |
| Others   | ~75%     | Estimation |

### Advanced Tokenization Diagnostics

#### Circuit Breaker Errors

**Symptoms:**
- "Circuit breaker open" in logs
- Consistent fallback to estimation
- Recently working tokenizers suddenly unavailable

**Diagnosis:**
```bash
# Enable debug logging to see circuit breaker state
thinktank instructions.txt ./project --debug --dry-run 2>&1 | grep -i circuit

# Look for:
# "Circuit breaker OPEN for provider: openai"
# "Circuit breaker HALF_OPEN for provider: gemini"
```

**Recovery:**
```bash
# Wait for automatic recovery (30 seconds)
sleep 30

# Or restart the process to reset circuit breakers
# Circuit breakers are per-process, not persistent
```

**Prevention:**
- Ensure stable network connectivity
- Monitor system resources (memory, CPU)
- Check for conflicting tokenizer processes

#### Performance Issues

**Large Input Sets:**
```bash
# For projects with >100MB of files
thinktank instructions.txt ./huge-project --include .go,.md --exclude vendor,node_modules

# Use streaming tokenization automatically enabled for large inputs
# Monitor memory usage: tokenization should stay under 50MB additional
```

**Streaming Tokenizer Performance:**
```bash
# Expected performance characteristics:
# Production: 9-10 MB/s throughput
# Development (race detection): 0.4-0.6 MB/s throughput

# For very large inputs (>50MB), streaming tokenization is automatic
# Memory usage remains constant regardless of input size
# Adaptive chunking optimizes performance: 8KB → 32KB → 64KB chunks
```

**Repeated Token Counting:**
```bash
# Token counts are cached per-session
# Identical input + model = cached result

# Clear cache between runs if needed:
# (Cache is automatic and memory-only)
```

### Provider-Specific Issues

#### OpenAI Tokenization
- **Issue**: tiktoken encoding errors
- **Cause**: Unsupported model variant
- **Solution**: Check model name spelling, use supported models (gpt-4.1, o4-mini)

#### Gemini Tokenization
- **Issue**: SentencePiece initialization failures
- **Cause**: Rare bundling issue or memory constraints
- **Solution**: Falls back to estimation automatically, no action needed

#### OpenRouter Tokenization
- **Issue**: High variance in token count estimates
- **Cause**: Uses estimation by design (OpenRouter normalizes internally)
- **Solution**: Expected behavior, estimates are sufficient for filtering

---

## Getting Help

### Information to Include in Bug Reports

When reporting issues, include:

1. **Command used:**
   ```bash
   thinktank --verbose --instructions task.txt --model gpt-4.1 ./project
   ```

2. **Error output:**
   ```
   [Copy the exact error message and any relevant log output]
   ```

3. **Environment:**
   ```bash
   # System info
   thinktank --version
   go version
   uname -a

   # API key status (DO NOT include actual keys)
   echo "OpenRouter key set: $([ -n "$OPENROUTER_API_KEY" ] && echo "YES" || echo "NO")"
   ```

4. **Project details:**
   - Approximate number of files
   - File types being processed
   - Any special configuration

### Common Support Channels

- **GitHub Issues**: [thinktank issues](https://github.com/phrazzld/thinktank/issues)
- **Documentation**: Check existing docs in `/docs` directory
- **Rate Limiting Guide**: [README.md](../README.md#rate-limiting--performance-optimization)
- **Error Handling**: [ERROR_HANDLING_AND_LOGGING.md](ERROR_HANDLING_AND_LOGGING.md)

### Self-Help Checklist

Before asking for help, try:

- [ ] API keys are set correctly for your provider(s)
- [ ] You can access provider APIs directly (curl tests)
- [ ] You've tried with `--dry-run` to test configuration
- [ ] You've tested with a smaller input set
- [ ] You've checked recent [GitHub issues](https://github.com/phrazzld/thinktank/issues) for similar problems
- [ ] You've tried the conservative rate limiting settings shown above
- [ ] You've enabled `--verbose` logging to see detailed error information

Most issues are resolved by checking API keys and adjusting rate limiting settings based on your account tier with each provider.
