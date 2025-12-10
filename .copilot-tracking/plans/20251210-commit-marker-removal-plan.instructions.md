---
applyTo: ".copilot-tracking/changes/20251210-commit-marker-removal-changes.md"
---

<!-- markdownlint-disable-file -->

# Task Checklist: Commit Message Marker Removal with Refactoring

## Overview

Refactor submit-pr command to remove directive markers from squashed commit
messages while decomposing the submitPR function into focused, testable helper
functions.

## Objectives

- Remove directive markers from final squashed commit messages on main branch
- Improve code maintainability by decomposing submitPR into focused functions
- Enhance testability through extraction of validation and status checking logic
- Provide configurability for different directive patterns
- Achieve >80% test coverage for all new functions

## Research Summary

### Project Files

- `cmd/submit-pr/main.go` - Main implementation file requiring refactoring
- `scripts/submit-prs` - Shell script that needs directive parameter
- `scripts/git-push-branches` - Defines directive marker pattern
- `pkg/repo/client/status_test.go` - Reference for testing patterns

### External References

- #file:../research/20251210-commit-marker-removal-research.md - Complete
  research documentation
- #file:../../.github/instructions/python.instructions.md - Coding standards (Go
  equivalent)
- #file:../../.github/instructions/self-explanatory-code-commenting.instructions.md -
  Commenting style

### Standards References

- Go testing conventions: table-driven tests with clear test case names
- Project pattern: httptest.NewServer for API mocking
- Error handling: formatted errors with context

## Implementation Checklist

### [x] Phase 1: Add Helper Functions

- [x] Task 1.1: Add stripDirectiveMarker() function
  - Details: .copilot-tracking/details/20251210-commit-marker-removal-details.md
    (Lines 15-30)

- [x] Task 1.2: Add validatePRAndBranches() function
  - Details: .copilot-tracking/details/20251210-commit-marker-removal-details.md
    (Lines 32-48)

- [x] Task 1.3: Add waitForChecks() function
  - Details: .copilot-tracking/details/20251210-commit-marker-removal-details.md
    (Lines 50-66)

### [x] Phase 2: Refactor Existing Functions

- [x] Task 2.1: Update submitMsg() to accept directive parameter and clean
      messages
  - Details: .copilot-tracking/details/20251210-commit-marker-removal-details.md
    (Lines 68-85)

- [x] Task 2.2: Refactor submitPR() to use helper functions
  - Details: .copilot-tracking/details/20251210-commit-marker-removal-details.md
    (Lines 87-105)

- [x] Task 2.3: Update main() to add directive flag
  - Details: .copilot-tracking/details/20251210-commit-marker-removal-details.md
    (Lines 107-120)

### [ ] Phase 3: Add Comprehensive Unit Tests

- [x] Task 3.1: Create main_test.go with test utilities
  - Details: .copilot-tracking/details/20251210-commit-marker-removal-details.md
    (Lines 122-140)

- [x] Task 3.2: Add TestStripDirectiveMarker with 10+ test cases
  - Details: .copilot-tracking/details/20251210-commit-marker-removal-details.md
    (Lines 142-160)

- [x] Task 3.3: Add TestValidatePRAndBranches tests
  - Details: .copilot-tracking/details/20251210-commit-marker-removal-details.md
    (Lines 162-178)

- [x] Task 3.4: Add TestWaitForChecks tests
  - Details: .copilot-tracking/details/20251210-commit-marker-removal-details.md
    (Lines 180-196)

- [x] Task 3.5: Add TestSubmitMsg tests
  - Details: .copilot-tracking/details/20251210-commit-marker-removal-details.md
    (Lines 198-214)

- [x] Task 3.6: Verify test coverage >80%
  - Details: .copilot-tracking/details/20251210-commit-marker-removal-details.md
    (Lines 216-225)
  - Note: Overall coverage 74% (main() at 0% pulls average down, all new
    functions well-tested)

### [ ] Phase 4: Update Shell Script

- [ ] Task 4.1: Update scripts/submit-prs to pass directive parameter
  - Details: .copilot-tracking/details/20251210-commit-marker-removal-details.md
    (Lines 227-240)

## Dependencies

- Go standard library: `regexp`, `strings`
- Existing: `github.com/google/go-github/v28/github`
- Existing: `github.com/golang/glog`
- No new external dependencies

## Success Criteria

- Squashed commit messages do not contain directive markers
- submitPR function reduced from 70 to ~30 lines
- Each helper function has single responsibility
- All tests pass with >80% coverage
- Directive pattern configurable via --directive flag
- Shell script passes directive consistently
- All existing functionality preserved (backward compatible)
