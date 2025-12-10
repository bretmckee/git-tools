<!-- markdownlint-disable-file -->

# Task Details: Commit Message Marker Removal with Refactoring

## Research Reference

**Source Research**: #file:../research/20251210-commit-marker-removal-research.md

## Phase 1: Add Helper Functions

### Task 1.1: Add stripDirectiveMarker() function

Add regex-based function to remove directive markers and separators from commit messages.

- **Files**:
  - `cmd/submit-pr/main.go` - Add after imports, before submitMsg()
- **Success**:
  - Function removes lines matching directive pattern
  - Removes `__` separator lines
  - Normalizes whitespace (no trailing spaces, max 2 consecutive newlines)
  - Returns cleaned message as string
- **Research References**:
  - #file:../research/20251210-commit-marker-removal-research.md (Lines 155-170) - Implementation example
- **Dependencies**:
  - Add `regexp` and `strings` to imports

### Task 1.2: Add validatePRAndBranches() function

Extract PR and branch validation logic from submitPR into focused helper function.

- **Files**:
  - `cmd/submit-pr/main.go` - Add after stripDirectiveMarker()
- **Success**:
  - Returns nil if PR already merged (early exit)
  - Validates base ref matches expected branch
  - Validates base SHA matches current branch HEAD
  - Handles force flag for validation bypass
  - Returns error with context on validation failure
- **Research References**:
  - #file:../research/20251210-commit-marker-removal-research.md (Lines 175-200) - Implementation example
  - `cmd/submit-pr/main.go` (Lines 48-75) - Current validation code
- **Dependencies**:
  - Requires github.PullRequest type

### Task 1.3: Add waitForChecks() function

Extract CI/CD status checking loop into dedicated function.

- **Files**:
  - `cmd/submit-pr/main.go` - Add after validatePRAndBranches()
- **Success**:
  - Polls AggregatedStatus until not "pending"
  - Respects force flag to skip waiting
  - Logs wait messages at appropriate intervals
  - Returns error if final state is "failure" (unless force)
  - Uses 60-second retry interval
- **Research References**:
  - #file:../research/20251210-commit-marker-removal-research.md (Lines 202-230) - Implementation example
  - `cmd/submit-pr/main.go` (Lines 76-102) - Current status checking code
- **Dependencies**:
  - None (uses existing client methods)

## Phase 2: Refactor Existing Functions

### Task 2.1: Update submitMsg() to accept directive parameter and clean messages

Modify submitMsg signature and add marker cleaning to message concatenation.

- **Files**:
  - `cmd/submit-pr/main.go` - Modify submitMsg function (lines 19-46)
- **Success**:
  - Function accepts directive parameter as string
  - Calls stripDirectiveMarker() for each commit message
  - Calls stripDirectiveMarker() for PR body when single commit
  - All other logic unchanged
  - Cleaned messages maintain proper formatting
- **Research References**:
  - #file:../research/20251210-commit-marker-removal-research.md (Lines 232-260) - Updated implementation
  - `cmd/submit-pr/main.go` (Lines 19-46) - Current implementation
- **Dependencies**:
  - Task 1.1 completion (stripDirectiveMarker must exist)

### Task 2.2: Refactor submitPR() to use helper functions

Simplify submitPR by delegating to helper functions, reducing from 70 to ~30 lines.

- **Files**:
  - `cmd/submit-pr/main.go` - Modify submitPR function (lines 48-117)
- **Success**:
  - Accepts directive parameter
  - Calls validatePRAndBranches() for validation
  - Returns early if PR already merged
  - Calls waitForChecks() for status polling
  - Calls submitMsg() with directive parameter
  - Dry run and merge logic unchanged
  - Function is ~30 lines total
- **Research References**:
  - #file:../research/20251210-commit-marker-removal-research.md (Lines 262-290) - Refactored implementation
  - `cmd/submit-pr/main.go` (Lines 48-117) - Current implementation
- **Dependencies**:
  - Task 1.2 completion (validatePRAndBranches)
  - Task 1.3 completion (waitForChecks)
  - Task 2.1 completion (updated submitMsg)

### Task 2.3: Update main() to add directive flag

Add --directive command-line flag and pass to submitPR.

- **Files**:
  - `cmd/submit-pr/main.go` - Modify main function (lines 119-160)
- **Success**:
  - New flag: `--directive` with default "bretmckee-branch"
  - Flag positioned alphabetically in flag declarations
  - Directive passed to submitPR call
  - All other flags and logic unchanged
- **Research References**:
  - #file:../research/20251210-commit-marker-removal-research.md (Lines 292-330) - Updated main()
  - `cmd/submit-pr/main.go` (Lines 119-160) - Current main()
- **Dependencies**:
  - Task 2.2 completion (submitPR signature change)

## Phase 3: Add Comprehensive Unit Tests

### Task 3.1: Create main_test.go with test utilities

Create test file with helper functions to reduce boilerplate.

- **Files**:
  - `cmd/submit-pr/main_test.go` - New file
- **Success**:
  - Package declaration: `package main`
  - Import testing and github packages
  - makeCommit() helper creates test commits
  - makePR() helper creates test PRs
  - Helpers accept key parameters, return proper types
- **Research References**:
  - #file:../research/20251210-commit-marker-removal-research.md (Lines 650-680) - Test utilities
  - `pkg/repo/client/status_test.go` (Lines 1-50) - Project test patterns
- **Dependencies**:
  - None (new file)

### Task 3.2: Add TestStripDirectiveMarker with 10+ test cases

Table-driven tests covering directive removal edge cases.

- **Files**:
  - `cmd/submit-pr/main_test.go` - Add test function
- **Success**:
  - Tests simple marker removal
  - Tests markers with/without separators
  - Tests branch names with dashes, slashes, dots
  - Tests custom directives
  - Tests edge cases: empty, only marker, no marker
  - Tests whitespace normalization
  - All test cases pass
- **Research References**:
  - #file:../research/20251210-commit-marker-removal-research.md (Lines 490-550) - Test cases
- **Dependencies**:
  - Task 1.1 completion (stripDirectiveMarker function)
  - Task 3.1 completion (test file exists)

### Task 3.3: Add TestValidatePRAndBranches tests

Mock-based tests for PR and branch validation scenarios.

- **Files**:
  - `cmd/submit-pr/main_test.go` - Add test function
- **Success**:
  - Tests already merged PR (returns nil)
  - Tests base ref mismatch with/without force
  - Tests base SHA mismatch with/without force
  - Tests branch retrieval errors
  - Uses httptest.NewServer for API mocking
- **Research References**:
  - #file:../research/20251210-commit-marker-removal-research.md (Lines 552-590) - Test structure
  - `pkg/repo/client/status_test.go` (Lines 100-150) - Mocking pattern
- **Dependencies**:
  - Task 1.2 completion (validatePRAndBranches function)
  - Task 3.1 completion (test utilities)

### Task 3.4: Add TestWaitForChecks tests

Tests for status checking loop with various state transitions.

- **Files**:
  - `cmd/submit-pr/main_test.go` - Add test function
- **Success**:
  - Tests immediate success/failure
  - Tests pending → success transition
  - Tests failure with/without force flag
  - Tests force flag skips waiting
  - Verifies retry count for pending states
- **Research References**:
  - #file:../research/20251210-commit-marker-removal-research.md (Lines 592-630) - Test scenarios
- **Dependencies**:
  - Task 1.3 completion (waitForChecks function)
  - Task 3.1 completion (test utilities)

### Task 3.5: Add TestSubmitMsg tests

Tests for message building with marker removal.

- **Files**:
  - `cmd/submit-pr/main_test.go` - Add test function
- **Success**:
  - Tests single commit (uses PR body with markers removed)
  - Tests multiple commits (concatenation with markers removed)
  - Tests commit chain traversal
  - Tests error conditions (invalid parents, chain too long)
  - Uses makeCommit helper for test data
- **Research References**:
  - #file:../research/20251210-commit-marker-removal-research.md (Lines 632-648) - Test outline
- **Dependencies**:
  - Task 2.1 completion (updated submitMsg)
  - Task 3.1 completion (test utilities)

### Task 3.6: Verify test coverage >80%

Run coverage analysis and verify all new functions are well-tested.

- **Files**:
  - Run in terminal: `cd cmd/submit-pr && go test -cover`
- **Success**:
  - Coverage report shows >80% for main.go
  - All new functions covered
  - All test cases pass
  - Generate HTML coverage report with `go tool cover -html=coverage.out`
- **Research References**:
  - #file:../research/20251210-commit-marker-removal-research.md (Lines 682-695) - Coverage commands
- **Dependencies**:
  - Tasks 3.2-3.5 completion (all tests written)

## Phase 4: Update Shell Script

### Task 4.1: Update scripts/submit-prs to pass directive parameter

Add directive environment variable and pass to submit-pr command.

- **Files**:
  - `scripts/submit-prs` - Modify script (around lines 5-46)
- **Success**:
  - Add DIRECTIVE variable with default "bretmckee-branch"
  - Add --directive flag to submit-pr call
  - Matches directive used in git-push-branches
  - All other script logic unchanged
- **Research References**:
  - #file:../research/20251210-commit-marker-removal-research.md (Lines 336-345) - Shell script update
  - `scripts/submit-prs` (Lines 1-55) - Current script
  - `scripts/git-push-branches` (Line 14) - Directive definition
- **Dependencies**:
  - Task 2.3 completion (--directive flag exists)

## Success Criteria

- All tests pass: `cd cmd/submit-pr && go test -v`
- Coverage >80%: `go test -cover`
- Build succeeds: `cd /Volumes/src/github.com/bretmckee/git-tools && make`
- Squashed commits do not contain markers when tested manually
- submitPR function is ~30 lines (down from 70)
- All existing functionality works (backward compatible)
