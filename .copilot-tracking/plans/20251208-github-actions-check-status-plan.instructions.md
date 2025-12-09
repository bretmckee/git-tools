---
applyTo: ".copilot-tracking/changes/20251208-github-actions-check-status-changes.md"
---

<!-- markdownlint-disable-file -->

# Task Checklist: GitHub Actions Check Status Integration

## Overview

Integrate GitHub Actions workflow status checking with existing PR submission
tool by implementing Check Runs API alongside legacy Combined Status API.

## Objectives

- Support GitHub Actions workflows using the Check Runs API
- Maintain backward compatibility with legacy commit status API
- Provide unified status determination for PR submission decisions
- Minimize breaking changes to existing Repo interface

## Research Summary

### Project Files

- pkg/repo/repo.go - Defines Repo interface for repository operations
- pkg/repo/client/status.go - Contains CombinedStatus implementation (legacy)
- cmd/submit-pr/main.go - Uses status checking to determine PR submission
  readiness

### External References

- #file:../research/20251208-github-actions-check-status-research.md - Complete
  API research and implementation patterns
- #githubRepo:"google/go-github" ListCheckRunsForRef - Check Runs API
  implementation examples
- #fetch:https://docs.github.com/en/rest/checks/runs - GitHub Check Runs API
  documentation

### Standards References

- #file:../../.github/copilot-instructions.md - Go coding conventions
- #file:../../.github/instructions/coding-best-practices.instructions.md -
  General coding practices

## Implementation Checklist

### [x] Phase 1: Add Check Runs API Support

- [x] Task 1.1: Add CheckRunsForRef method to Repo interface
  - Details:
    .copilot-tracking/details/20251208-github-actions-check-status-details.md
    (Lines 8-20)

- [x] Task 1.2: Implement CheckRunsForRef in client
  - Details:
    .copilot-tracking/details/20251208-github-actions-check-status-details.md
    (Lines 22-43)

### [x] Phase 2: Implement Aggregated Status Logic

- [x] Task 2.1: Add AggregatedStatus helper method to client
  - Details:
    .copilot-tracking/details/20251208-github-actions-check-status-details.md
    (Lines 45-90)

### [ ] Phase 3: Update submit-pr Command

- [ ] Task 3.1: Replace CombinedStatus with AggregatedStatus in submit-pr
  - Details:
    .copilot-tracking/details/20251208-github-actions-check-status-details.md
    (Lines 92-115)

- [ ] Task 3.2: Add enhanced logging for status checks
  - Details:
    .copilot-tracking/details/20251208-github-actions-check-status-details.md
    (Lines 117-128)

## Dependencies

- go-github v28.1.1 (already in use, supports Check Runs API)
- golang/glog (for logging)
- No new external dependencies required

## Success Criteria

- submit-pr correctly waits for GitHub Actions workflows to complete
- Backward compatibility maintained for repos using legacy commit status API
- Repos with both GitHub Actions and legacy statuses work correctly
- Appropriate logging shows status of all checks during wait loops
- All existing functionality continues to work without breaking changes
