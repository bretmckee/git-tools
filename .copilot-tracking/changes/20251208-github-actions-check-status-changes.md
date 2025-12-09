<!-- markdownlint-disable-file -->

# Release Changes: GitHub Actions Check Status Integration

**Related Plan**: 20251208-github-actions-check-status-plan.instructions.md
**Implementation Date**: 2025-12-08

## Summary

Integration of GitHub Actions Check Runs API with existing PR submission tool to
support modern CI/CD workflows while maintaining backward compatibility with
legacy commit status API.

## Changes

### Added

- pkg/repo/client/status_test.go - Comprehensive unit tests for CheckRunsForRef
  method

### Modified

- pkg/repo/repo.go - Added CheckRunsForRef method to Repo interface for Check
  Runs API support
- pkg/repo/client/status.go - Implemented CheckRunsForRef method with GitHub
  Check Runs API integration
- pkg/repo/client/status.go - Implemented AggregatedStatus helper method to
  combine Check Runs and Combined Status APIs
- pkg/repo/client/status_test.go - Added comprehensive unit tests for
  AggregatedStatus method covering all status scenarios

### Removed
