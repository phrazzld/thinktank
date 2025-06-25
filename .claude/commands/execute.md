# Strategic Task Executor - Systematic TODO Completion

Methodically execute tasks from TODO.md with expert-level strategic planning and implementation using TDD and Go best practices.

**Usage**: `/project:execute`

## GOAL

Select and complete the next available task from TODO.md using comprehensive analysis, strategic planning, and flawless execution following the project's leyline principles of simplicity, testability, and maintainability.

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
- Understand existing patterns and conventions from CLAUDE.md
- Identify potential impact areas and dependencies
- Check test coverage and existing test patterns

### 2. **Documentation Review**
- Study relevant leyline documents for foundational principles:
  - `docs/leyline/tenets/simplicity.md` - Fight complexity demon, prefer boring solutions
  - `docs/leyline/tenets/testability.md` - Design for testability, avoid internal mocking
  - `docs/leyline/tenets/maintainability.md` - Code that lasts and evolves gracefully
  - `docs/leyline/bindings/go/` - Go-specific patterns and practices
- Check project README and CLAUDE.md guidelines
- Review any architectural decision records (ADRs)

### 3. **External Research**
- Use web searches for Go best practices and common pitfalls
- Research similar implementations in respected codebases
- Verify current API documentation for dependencies

## STRATEGIC PLANNING

### Multi-Expert Planning Session

For complex tasks, use the Task tool to consult expert perspectives. Each subagent must run independently and in parallel for maximum efficiency. CRITICAL: All subagents operate in research/investigation mode only - they should NOT modify code, use plan mode, or create files. They must output all research and brainstorming directly to chat.

**Task 1: Rob Pike - Go Philosophy & Simplicity**
- Prompt: "As Rob Pike, analyze this Go implementation task. What's the most idiomatic Go approach? How can we keep it simple, clear, and focused? Consider Go's philosophy of explicit error handling, interfaces, and composition over inheritance. IMPORTANT: You are in research mode only - do not modify any code, do not use plan mode, and output all your analysis directly to chat."

**Task 2: Dave Cheney - Go Testing & Performance**
- Prompt: "As Dave Cheney, review this task from a testing and performance perspective. How would you structure this for excellent testability? What are the performance implications? Focus on table-driven tests, benchmarks, and avoiding premature optimization. IMPORTANT: You are in research mode only - do not modify any code, do not use plan mode, and output all your analysis directly to chat."

**Task 3: Kent Beck - TDD & Incremental Design**
- Prompt: "As Kent Beck, plan this implementation using Test-Driven Development. What's the smallest failing test we can write first? How do we ensure each step adds value? Focus on the RED-GREEN-REFACTOR cycle and incremental design. IMPORTANT: You are in research mode only - do not modify any code, do not use plan mode, and output all your analysis directly to chat."

### Plan Synthesis
- Combine expert insights into a cohesive Go-idiomatic strategy
- Create step-by-step TDD implementation plan
- Identify checkpoints for validation
- Plan for rollback if issues arise

## IMPLEMENTATION

Execute the approved plan with precision following Go and project conventions:

### 1. **Pre-Implementation Setup**
- Create feature branch if required
- Set up test infrastructure following existing patterns
- Prepare any necessary tooling

### 2. **TDD Incremental Execution**
- **RED**: Write failing test first that captures desired behavior
- **GREEN**: Implement minimal code to make test pass
- **REFACTOR**: Clean up code while keeping tests green
- Run tests after each cycle: `go test ./...`
- Check race conditions: `go test -race ./...` for concurrent code
- Commit working states frequently with conventional commit messages
- Follow project's code style and conventions from CLAUDE.md

### 3. **Continuous Validation**
- Run linters: `go vet ./...` and `golangci-lint run ./...`
- Execute full test suite: `go test ./...`
- Check coverage: `go test -cover ./...` (maintain 90%+ target)
- Verify no regressions introduced
- Check performance implications with benchmarks

### 4. **Adaptive Response**
If encountering unexpected situations:
- **HALT** implementation immediately
- Document the specific issue encountered
- Analyze implications for the current approach
- Present findings to user with recommendations
- Wait for guidance before proceeding

## QUALITY ASSURANCE

Before marking task complete:

### 1. **Code Quality Checks**
- All tests pass: `go test ./...`
- Race detection clean: `go test -race ./...`
- Linting compliance: `go vet ./...` and `golangci-lint run ./...`
- Code formatting: `go fmt ./...`
- Coverage threshold met: 90%+ via `./scripts/check-coverage.sh`
- No commented-out code or unresolved TODOs left

### 2. **Go-Specific Validation**
- Follows Go idioms: clear error handling, interfaces, composition
- Uses appropriate Go standard library patterns
- Proper package structure and naming conventions
- Efficient memory usage and garbage collection friendly
- Thread-safe concurrent access where applicable

### 3. **Integration Verification**
- Changes work with existing codebase
- No breaking changes unless intended
- API contracts maintained
- Backward compatibility preserved
- Integration tests pass if applicable

## CLEANUP

Upon successful completion:

### 1. **Task Management**
- Update task status to `[x]` in TODO.md
- Add completion notes if helpful for future reference
- Check for any follow-up tasks that are now unblocked

### 2. **Code Finalization**
- Ensure all changes committed with conventional commit messages
- Update any relevant documentation
- Clean up any temporary files or branches
- Run final vulnerability scan: `govulncheck -scan=module`

### 3. **Progress Assessment**
- Review remaining tasks in TODO.md
- Consider if any emergent tasks surfaced during implementation
- Prepare summary of what was accomplished

## SUCCESS CRITERIA

- Task completed according to specifications
- Code quality meets or exceeds project standards (90%+ coverage)
- All tests pass including race detection
- Implementation follows Go idioms and project conventions
- TDD methodology demonstrated with test-first development
- No technical debt introduced
- Clear documentation of any decisions made
- Ready for code review and integration

## FAILURE PROTOCOLS

If unable to complete task:
- Document specific blockers encountered
- Update task with `[!]` blocked status
- Create new tasks for unblocking work
- Communicate clearly about obstacles
- Suggest alternative approaches

## PROJECT-SPECIFIC COMMANDS

- **Build**: `go build ./...`
- **Test**: `go test ./...`
- **Race Detection**: `go test -race ./...`
- **Coverage**: `go test -cover ./...` or `./scripts/check-coverage.sh`
- **Lint**: `go vet ./...` and `golangci-lint run ./...`
- **Format**: `go fmt ./...`
- **Vulnerability Scan**: `govulncheck -scan=module`

Execute the next task with strategic excellence, Go idioms, and systematic TDD precision.
