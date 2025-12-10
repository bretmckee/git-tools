---
mode: agent
model: Claude Sonnet 4.5
---

<!-- markdownlint-disable-file -->

# Implementation Prompt: Commit Message Marker Removal with Refactoring

## Implementation Instructions

### Step 1: Create Changes Tracking File

You WILL create `20251210-commit-marker-removal-changes.md` in #file:../changes/ if it does not exist.

### Step 2: Execute Implementation

You WILL follow #file:../../.github/instructions/task-implementation.instructions.md
You WILL systematically implement #file:../plans/20251210-commit-marker-removal-plan.instructions.md task-by-task
You WILL follow ALL project standards and conventions:

- Go coding conventions for all Go code
- #file:../../.github/instructions/self-explanatory-code-commenting.instructions.md for commenting style
- #file:../../.github/instructions/coding-best-practices.instructions.md for implementation patterns
- #file:../../.github/instructions/taming-copilot.instructions.md for interaction patterns

**CRITICAL**: If ${input:phaseStop:true} is true, you WILL stop after each Phase for user review.
**CRITICAL**: If ${input:taskStop:true} is true, you WILL stop after each Task for user review.

### Step 3: Implementation Order

You WILL implement in this exact order:

1. **Phase 1**: Add all helper functions first (stripDirectiveMarker, validatePRAndBranches, waitForChecks)
2. **Phase 2**: Refactor existing functions to use helpers (submitMsg, submitPR, main)
3. **Phase 3**: Add comprehensive unit tests with >80% coverage
4. **Phase 4**: Update shell script to pass directive parameter

### Step 4: Verification

After EACH task you WILL:
- Verify code compiles: `cd /Volumes/src/github.com/bretmckee/git-tools && make`
- Run tests after Phase 3: `cd cmd/submit-pr && go test -v`
- Check coverage after Phase 3: `go test -cover`

### Step 5: Cleanup

When ALL Phases are checked off (`[x]`) and completed you WILL do the following:

1. You WILL provide a markdown style link and a summary of all changes from #file:../changes/20251210-commit-marker-removal-changes.md to the user:

   - You WILL keep the overall summary brief
   - You WILL add spacing around any lists
   - You MUST wrap any reference to a file in a markdown style link

2. You WILL provide markdown style links to:
   - .copilot-tracking/plans/20251210-commit-marker-removal-plan.instructions.md
   - .copilot-tracking/details/20251210-commit-marker-removal-details.md
   - .copilot-tracking/research/20251210-commit-marker-removal-research.md

   You WILL recommend cleaning these files up as well.

3. **MANDATORY**: You WILL attempt to delete .copilot-tracking/prompts/implement-commit-marker-removal.prompt.md

## Success Criteria

- [ ] Changes tracking file created
- [ ] All plan items implemented with working code
- [ ] All detailed specifications satisfied
- [ ] Project conventions followed
- [ ] All relevant coding conventions followed
- [ ] All new code has unit tests
- [ ] Test coverage >80% for cmd/submit-pr/main.go
- [ ] All tests pass
- [ ] Build succeeds with make
- [ ] Changes file updated continuously
- [ ] submitPR reduced from 70 to ~30 lines
- [ ] Directive markers removed from squashed commits
