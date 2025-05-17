# SVU Version Calculation Test Results

This document contains test results for verifying `svu next` version calculation behavior across different commit scenarios.

Generated on: Sat May 17 10:06:18 CDT 2025

## Test Results

| Starting Version | Commit(s) | Expected Version | Actual Version | Status |
|-----------------|-----------|------------------|----------------|--------|
| none | fix: resolve bug | v0.0.1 | v0.0.1 | ✅ PASS |
| none | feat: add new feature | v0.1.0 | v0.1.0 | ✅ PASS |
| none | feat!: breaking change | v1.0.0 | v0.1.0 | ❌ FAIL |
| v1.2.3 | fix: bug fix | v1.2.4 | v1.2.4 | ✅ PASS |
| v1.2.3 | feat: new feature | v1.3.0 | v1.3.0 | ✅ PASS |
| v1.2.3 | fix!: breaking fix | v2.0.0 | v2.0.0 | ✅ PASS |
| v1.0.0 | fix: bug fix|feat: new feature | v1.1.0 | v1.1.0 | ✅ PASS |
| v1.0.0 | fix: bug fix|feat: new feature|refactor!: breaking refactor | v2.0.0 | v2.0.0 | ✅ PASS |
| v1.0.0 | chore: update deps | v1.0.1 | v1.0.0 | ❌ FAIL |
| v1.0.0 | feat(api): add endpoint | v1.1.0 | v1.1.0 | ✅ PASS |
| v1.0.0-alpha.1 | feat: new feature | v1.0.0-alpha.2 | v1.1.0 | ❌ FAIL |
| v1.0.0 | docs: update readme|style: format code|chore: cleanup | v1.0.1 | v1.0.0 | ❌ FAIL |
