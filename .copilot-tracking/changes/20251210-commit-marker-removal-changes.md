<!-- markdownlint-disable-file -->

# Release Changes: Commit Message Marker Removal with Refactoring

**Related Plan**: 20251210-commit-marker-removal-plan.instructions.md
**Implementation Date**: 2025-12-10

## Summary

Refactoring submit-pr command to remove directive markers from squashed commit
messages while decomposing the submitPR function into focused, testable helper
functions with comprehensive unit tests.

## Changes

### Added

- cmd/submit-pr/main_test.go - Created comprehensive unit test file with
  table-driven tests
- cmd/submit-pr/main_test.go - Added TestStripDirectiveMarker with 12 test cases
  covering edge cases
- cmd/submit-pr/main_test.go - Added TestValidatePRAndBranches with 6 test cases
  using httptest mocking
- cmd/submit-pr/main_test.go - Added TestWaitForChecks with 6 test cases for
  status checking scenarios
- cmd/submit-pr/main_test.go - Added helper functions (contains,
  mapStateToCheckStatus, mapStateToConclusion)

### Modified

- cmd/submit-pr/main.go - Added regexp and strings imports for marker removal
  functionality
- cmd/submit-pr/main.go - Added stripDirectiveMarker() helper function to remove
  directive markers from commit messages
- cmd/submit-pr/main.go - Added validatePRAndBranches() helper function to
  extract PR and branch validation logic
- cmd/submit-pr/main.go - Added waitForChecks() helper function to extract
  status checking and retry logic
- cmd/submit-pr/main.go - Added errAlreadyMerged sentinel error to preserve
  short-circuit behavior for already-merged PRs
- cmd/submit-pr/main.go - Refactored waitForChecks() to use direct returns
  instead of break statements for more idiomatic Go
- cmd/submit-pr/main.go - Fixed validatePRAndBranches() to return
  errAlreadyMerged to preserve original short-circuit semantics

### Removed
