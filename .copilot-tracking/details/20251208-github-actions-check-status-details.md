<!-- markdownlint-disable-file -->

# Task Details: GitHub Actions Check Status Integration

## Research Reference

**Source Research**: #file:../research/20251208-github-actions-check-status-research.md

## Phase 1: Add Check Runs API Support

### Task 1.1: Add CheckRunsForRef method to Repo interface

Add new method to support Check Runs API in the Repo interface.

- **Files**:
  - pkg/repo/repo.go - Add CheckRunsForRef method to Repo interface
- **Success**:
  - Method signature added: `CheckRunsForRef(ref string, opts *github.ListCheckRunsOptions) (*github.ListCheckRunsResults, error)`
  - Interface compiles without errors
- **Research References**:
  - #file:../research/20251208-github-actions-check-status-research.md (Lines 165-173) - Recommended interface updates
  - #file:../research/20251208-github-actions-check-status-research.md (Lines 119-151) - CheckRun structure and API schema
- **Dependencies**:
  - None (foundational change)

### Task 1.2: Implement CheckRunsForRef in client

Implement the CheckRunsForRef method in the Client struct to call GitHub's Check Runs API.

- **Files**:
  - pkg/repo/client/status.go - Add CheckRunsForRef implementation
- **Success**:
  - Method calls `c.client.Checks.ListCheckRunsForRef()` with appropriate parameters
  - Uses "latest" filter by default if opts is nil
  - Includes glog.V(3) logging for debugging
  - Returns formatted error on failure with context
  - Follows existing code patterns from CombinedStatus method
- **Research References**:
  - #file:../research/20251208-github-actions-check-status-research.md (Lines 175-195) - Implementation example with error handling
  - #file:../research/20251208-github-actions-check-status-research.md (Lines 85-103) - go-github library example
  - #file:../research/20251208-github-actions-check-status-research.md (Lines 69-83) - Current CombinedStatus pattern
- **Dependencies**:
  - Task 1.1 completion (interface definition)

## Phase 2: Implement Aggregated Status Logic

### Task 2.1: Add AggregatedStatus helper method to client

Create aggregation logic that combines Check Runs API and Combined Status API results.

- **Files**:
  - pkg/repo/client/status.go - Add AggregatedStatus method (internal helper, not part of interface)
- **Success**:
  - Method calls both CheckRunsForRef and CombinedStatus
  - Returns "pending" if any check run has status != "completed"
  - Returns "pending" if combined status is "pending"
  - Returns "failure" if any check run has conclusion in ["failure", "timed_out", "action_required"]
  - Returns "failure" if combined status is "failure" or "error"
  - Returns "success" only if all checks pass
  - Handles case where no checks are configured (returns "success")
  - Properly aggregates results from both APIs
  - Includes appropriate error handling
- **Research References**:
  - #file:../research/20251208-github-actions-check-status-research.md (Lines 197-265) - Complete AggregatedStatus implementation
  - #file:../research/20251208-github-actions-check-status-research.md (Lines 153-163) - Status determination logic requirements
  - #file:../research/20251208-github-actions-check-status-research.md (Lines 107-117) - CheckRun status and conclusion values
- **Dependencies**:
  - Task 1.2 completion (CheckRunsForRef must exist)

## Phase 3: Update submit-pr Command

### Task 3.1: Replace CombinedStatus with AggregatedStatus in submit-pr

Update the PR submission logic to use the new aggregated status checking.

- **Files**:
  - cmd/submit-pr/main.go - Modify status checking loop (approximately lines 78-95)
- **Success**:
  - Replace call to `c.CombinedStatus(ref)` with `c.AggregatedStatus(ref)`
  - Update variable handling to work with string return type instead of *github.CombinedStatus
  - Preserve existing logic for "pending" wait loop
  - Preserve existing logic for "failure" handling with force flag
  - All existing behavior maintained
  - Code compiles and runs without errors
- **Research References**:
  - #file:../research/20251208-github-actions-check-status-research.md (Lines 267-290) - Example usage in submit-pr
  - #file:../research/20251208-github-actions-check-status-research.md (Lines 18-22) - Current implementation location
- **Dependencies**:
  - Task 2.1 completion (AggregatedStatus must exist)

### Task 3.2: Add enhanced logging for status checks

Improve logging to show details about check runs during status polling.

- **Files**:
  - pkg/repo/client/status.go - Enhance logging in AggregatedStatus method
- **Success**:
  - Add glog.V(2) logging showing count of check runs and statuses
  - Add glog.V(2) logging showing combined status results
  - Log individual failing checks at glog.V(1) level
  - Logging provides useful debugging information without being too verbose
- **Research References**:
  - #file:../research/20251208-github-actions-check-status-research.md (Lines 69-83) - Existing logging patterns
- **Dependencies**:
  - Task 2.1 completion (AggregatedStatus must exist)

## Dependencies

- go-github v28.1.1 (already present)
- golang/glog (already present)

## Success Criteria

- All code compiles without errors
- submit-pr waits for GitHub Actions workflows to complete
- Backward compatibility with legacy commit status API maintained
- Enhanced logging provides visibility into status checks
- No breaking changes to existing API consumers
