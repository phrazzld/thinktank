# Strategic Task Executor - Systematic TODO Completion

Methodically execute tasks from TODO.md with expert-level strategic planning and implementation.

**Usage**: `/project:execute`

## GOAL

Select and complete the next available task from TODO.md using comprehensive analysis, strategic planning, and flawless execution for the thinktank Go CLI project.

## ACQUISITION

Select the next available ticket from TODO.md following this priority:
1. In-progress tasks marked with `[~]` - Continue where work was paused
2. Unblocked tasks marked with `[ ]` - Start fresh work
3. Consider task dependencies and critical path
4. Skip blocked tasks until dependencies are resolved

If all tasks in TODO.md are completed:
- Celebrate completion appropriately
- Suggest next strategic moves
- Halt

## CONTEXT GATHERING

Conduct comprehensive review before execution:

### 1. **Codebase Analysis**
- Read all files mentioned in or relevant to the task
- Understand existing Go patterns and conventions from CLAUDE.md
- Identify potential impact areas and dependencies
- Review relevant packages in `internal/` directory structure

### 2. **Documentation Review**
- Study CLAUDE.md for project-specific guidelines (TDD, conventional commits, coverage requirements)
- Check dependency injection patterns already implemented in `internal/cli/run_interfaces.go`
- Review testing strategies from completed phases in TODO.md
- Examine any architectural decision records in `docs/`

### 3. **External Research**
- Use WebSearch for Go best practices and framework documentation
- Research similar implementations in respected Go codebases
- Check latest Go tooling recommendations (go vet, golangci-lint patterns)

### 4. **Advanced Analysis** (when task complexity warrants)
- Invoke `thinktank` CLI for multi-perspective analysis on complex problems
- Consider security implications (vulnerability scanning with govulncheck)
- Evaluate performance impact (race detection, benchmarking)
- Assess long-term maintainability within existing architecture

## STRATEGIC PLANNING

### Multi-Expert Planning Session

For complex tasks, use the Task tool to consult expert perspectives:

**Task 1: John Carmack - Systems Engineering Excellence**
- Prompt: "As John Carmack, analyze this Go implementation task. What's the most elegant, performant solution considering Go's strengths? Focus on algorithmic efficiency, memory management, and system design. What would you optimize for in a CLI tool?"

**Task 2: Rob Pike - Go Philosophy & Simplicity**
- Prompt: "As Rob Pike, review this Go task. How does this align with Go's design philosophy? What's the most idiomatic Go approach? Focus on simplicity, readability, and leveraging Go's built-in features effectively."

**Task 3: Kent Beck - Test-Driven Excellence**
- Prompt: "As Kent Beck, plan this implementation using TDD principles. How would you approach this test-first? What's the smallest change that could possibly work? How do we ensure 90% test coverage as required by CLAUDE.md?"

### Plan Synthesis
- Combine expert insights into a cohesive Go-idiomatic strategy
- Create step-by-step implementation plan following TDD principles
- Identify checkpoints for validation (test coverage, lint checks, race detection)
- Plan for rollback if issues arise

## IMPLEMENTATION

Execute the approved plan with precision:

### 1. **Pre-Implementation Setup**
- Create feature branch if required (following git workflow in CLAUDE.md)
- Set up test infrastructure using existing patterns from `internal/cli/run_*_test.go`
- Prepare any necessary mock interfaces using patterns from `run_mocks.go`

### 2. **Incremental TDD Execution**
- Write failing tests first (as mandated by CLAUDE.md)
- Implement in small, testable increments
- Run tests after each significant change: `go test ./...`
- Follow dependency injection patterns from existing `RunConfig/RunResult` architecture
- Commit working states frequently with conventional commit messages

### 3. **Continuous Validation**
- Run project-specific quality gates:
  ```bash
  go build ./...                    # Verify builds
  go test ./...                     # Run all tests
  go test -race ./...              # Race detection (required before committing)
  golangci-lint run ./...          # Lint checks (catches errcheck, staticcheck)
  go fmt ./...                     # Format code
  go vet ./...                     # Vet checks
  ./scripts/check-coverage.sh      # Verify 90% coverage threshold
  ```
- Verify no regressions introduced
- Check performance implications if relevant

### 4. **Adaptive Response**
If encountering unexpected situations:
- **HALT** implementation immediately
- Document the specific issue encountered
- Analyze implications for the current approach
- Consider using `thinktank` CLI for additional perspective
- Present findings to user with recommendations
- Wait for guidance before proceeding

## QUALITY ASSURANCE

Before marking task complete:

### 1. **Go-Specific Code Quality Checks**
- All tests pass: `go test ./...`
- Race detection clean: `go test -race ./...`
- Linting compliance: `golangci-lint run ./...`
- Code formatted: `go fmt ./...`
- Vet checks pass: `go vet ./...`
- Coverage maintained: `./scripts/check-coverage.sh` (≥90%)
- No errcheck violations (never ignore errors with `_`)

### 2. **Functional Validation**
- Task requirements fully met according to TODO.md specifications
- Edge cases handled appropriately with Go's error handling patterns
- Performance acceptable (no unnecessary allocations, efficient algorithms)
- Security considerations addressed (no secrets in code, input validation)

### 3. **Integration Verification**
- Changes work with existing `internal/cli` architecture
- Dependency injection patterns maintained (`RunConfig/RunResult`)
- No breaking changes unless intended and documented
- API contracts maintained (interface compatibility)
- Integration tests pass if applicable

## CLEANUP

Upon successful completion:

### 1. **Task Management**
- Update task status to `[x]` in TODO.md
- Add completion notes with commit references if helpful
- Check for any follow-up tasks that are now unblocked

### 2. **Code Finalization**
- Ensure all changes committed with conventional commit messages
- Update relevant documentation in `docs/` if needed
- Clean up any temporary files or test artifacts
- Remove any debugging code or temporary logging

### 3. **Progress Assessment**
- Review remaining tasks in TODO.md
- Consider if any new tasks emerged from implementation
- Assess overall project health and test coverage
- Prepare concise summary of accomplishments

## SUCCESS CRITERIA

- Task completed according to TODO.md specifications
- Code quality meets project standards (90% coverage, lint-free, race-free)
- All tests pass: `go test ./...` and `go test -race ./...`
- Implementation follows Go idioms and project conventions from CLAUDE.md
- No technical debt introduced
- Dependency injection architecture maintained
- Ready for code review and integration

## FAILURE PROTOCOLS

If unable to complete task:
- Document specific blockers encountered
- Update task with `[!]` blocked status in TODO.md
- Create new tasks for unblocking work
- Run diagnostic commands to gather information:
  ```bash
  go test -v ./...                 # Verbose test output
  golangci-lint run ./... --verbose # Detailed lint issues
  go mod tidy && go mod verify     # Dependency verification
  ```
- Communicate clearly about obstacles
- Suggest alternative approaches or request guidance

## PROJECT-SPECIFIC CONSIDERATIONS

### Architecture Patterns
- Follow established dependency injection patterns (`RunConfig/RunResult`)
- Use existing mock infrastructure in `internal/cli/run_mocks.go`
- Maintain separation between CLI layer and business logic
- Leverage adapter patterns from `internal/thinktank/adapters.go`

### Testing Strategy
- Prefer direct function testing over subprocess testing (established pattern)
- Use table-driven tests for comprehensive coverage
- Mock external dependencies using established interfaces
- Target 90% code coverage as enforced by CI

### Development Commands
- Build: `go build ./...`
- Test: `go test ./...`
- Race Detection: `go test -race ./...`
- Coverage: `./scripts/check-coverage.sh`
- Lint: `golangci-lint run ./...`
- E2E Tests: `./internal/e2e/run_e2e_tests.sh`

Execute the next task with strategic excellence, Go idioms, and systematic precision.
