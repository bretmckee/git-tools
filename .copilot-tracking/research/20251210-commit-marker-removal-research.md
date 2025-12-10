<!-- markdownlint-disable-file -->

# Task Research Notes: Commit Message Marker Removal Before PR Submission

## Research Executed

### Workflow Analysis

- `scripts/git-push-branches` (lines 14-115)
  - Uses directive marker `bretmckee-branch: <branch-name>` in commit messages
  - Extracts branch name with:
    `sed -n 's/^'"${DIRECTIVE}"': \([/A-Za-z0-9_.-]\+\)$/\1/p'`
  - Marker format: Line starting with `${DIRECTIVE}:` followed by branch name
  - Default directive: `bretmckee-branch`
  - Example from README: commit message ends with
    `__\nbretmckee-branch: update-readme`
- `cmd/submit-pr/main.go` (lines 19-44, 103-114)
  - `submitMsg()` builds squash commit message from all commits in PR
  - Uses `commit.GetMessage()` which includes the full commit message (with
    markers)
  - Message is passed to `MergePullRequest()` as the squash commit message
  - Current implementation includes directive markers in final commit
- `scripts/submit-prs` (lines 35-55)
  - Calls `submit-pr` command
  - After merge, rebases remaining commits onto base branch
  - Does not modify commit messages

### Code Search Results

- Marker usage pattern: Directive in commit message body
  - Format: `${DIRECTIVE}: ${BRANCH_NAME}` (e.g.,
    `bretmckee-branch: update-readme`)
  - Located at end of commit message after `__` separator
  - Used by `git-push-branches` to determine which commits get branches
- No existing commit message modification logic
  - No calls to git commit --amend or message rewriting
  - `submitMsg()` concatenates messages verbatim with
    `"* " + commit.GetMessage()`

### External Research

- #fetch:https://docs.github.com/en/rest/git/commits
  - GitHub API cannot modify existing commits (immutable)
  - Can create new commits with modified messages via `CreateCommit` API
  - Would require creating new commit objects with same tree but different
    message
  - This changes commit SHAs, breaking PR associations
- Git commit messages in GitHub
  - Once pushed, commits are immutable in GitHub's database
  - Modifying messages requires force-pushing new commits
  - PR branches can be force-updated, but this changes commit history

### Project Conventions

- Standards referenced: Go coding conventions, client interface patterns
- Instructions followed: Minimal code changes, preserve existing structure
- Message format: Title on first line, body after blank line, marker at end

### Testing Patterns

- Project uses table-driven tests (see `pkg/repo/client/status_test.go`)
- Tests use `httptest.NewServer` to mock GitHub API responses
- Tests verify both success and error cases
- No existing tests in `cmd/submit-pr/` directory (opportunity to add)
- Test naming convention: `TestFunctionName` for basic tests,
  `TestFunctionName_Scenario` for specific cases
- Mock pattern: Create test HTTP server, inject into client for isolated testing

## Key Discoveries

### Marker System Purpose and Usage

The directive marker (e.g., `bretmckee-branch: update-readme`) serves as a
development-time annotation:

- Used by `git-push-branches` to identify which commits should have branches
- Appears in local commit messages but not needed in final merged commits
- Workflow: Local commits → Branches created → PRs created → PRs merged (markers
  no longer needed)

### Current submitMsg() Behavior

```go
func submitMsg(c *client.Client, prBody string, first, last string) (string, error) {
    msg := ""
    l := 0
    for pos := first; pos != last; l++ {
        commit, err := c.Commit(pos)
        // ...
        msg = "* " + commit.GetMessage() + "\n\n" + msg  // Includes full message with markers
        pos = *commit.Parents[0].SHA
    }
    if l < 2 {
        return prBody, nil  // Single commit uses PR body
    }
    return msg, nil  // Multi-commit uses concatenated messages
}
```

### GitHub Merge Process

When `MergePullRequest()` is called:

- Merge method "squash" creates a single commit combining all PR commits
- The `msg` parameter becomes the commit message of the squashed commit
- This commit message appears in the main branch history
- Including markers in final commit pollutes the clean history

### Complete Example: Current Flow

```
Local Commit Message:
Update README.md

Add some more information to the read me.
__
bretmckee-branch: update-readme

↓ (git-push-branches finds marker and creates branch)

PR Branch: bretmckee/update-readme
PR Body: (same as commit message)

↓ (submit-pr merges with squash)

Main Branch Commit Message:
* Update README.md

Add some more information to the read me.
__
bretmckee-branch: update-readme
```

### Constraints

- Cannot modify commits via GitHub API (commits are immutable)
- Cannot force-push to modify commit history (would change SHAs, breaking PR
  association)
- Must clean message at merge time, not before
- Need to support configurable directive patterns (users may customize)

### Function Complexity Analysis

The `submitPR()` function (lines 48-117) has 70 lines and handles multiple
responsibilities:

1. **PR Retrieval and Validation** (lines 49-56): Get PR, check if already
   merged
2. **Base Branch Validation** (lines 57-63): Verify PR base ref matches expected
   base branch
3. **Base Branch SHA Validation** (lines 64-75): Verify PR base SHA matches
   current base branch SHA
4. **Status Checking Loop** (lines 76-95): Wait for CI/CD checks to complete
5. **Merge Message Building** (lines 103-106): Build squash commit message
6. **PR Merge Execution** (lines 107-117): Perform the actual merge

Each section is cohesive and can be extracted into a focused helper function.
This follows the Single Responsibility Principle and improves testability.

## Recommended Approach

### Refactored Architecture with Marker Removal

Apply both marker removal and function decomposition simultaneously. Extract
logical sections of `submitPR()` into focused helper functions while adding
marker cleaning capability.

#### New Helper Functions

**1. validatePRAndBranches** - Validates PR state and branch alignment

```go
func validatePRAndBranches(c *client.Client, pr *github.PullRequest, baseBranch string, force bool) error {
    if pr.GetMerged() {
        glog.Warningf("PR %d is already merged.", pr.GetNumber())
        return nil
    }

    if prRef := pr.GetBase().GetRef(); prRef != baseBranch {
        err := fmt.Errorf("pr base ref (%q) does not match base branch ref (%q)", prRef, baseBranch)
        if !force {
            return err
        }
        glog.Warningf("because force was specified, ignoring error %v", err)
    }

    bb, err := c.Branch(baseBranch)
    if err != nil {
        return fmt.Errorf("failed to get base branch %q: %v", baseBranch, err)
    }

    if prSHA, bbSHA := pr.GetBase().GetSHA(), bb.GetCommit().GetSHA(); prSHA != bbSHA {
        err := fmt.Errorf("pr base SHA (%q) does not match base branch SHA (%q)", prSHA, bbSHA)
        if !force {
            return err
        }
        glog.Warningf("because force was specified, ignoring error %v", err)
    }

    return nil
}
```

**2. waitForChecks** - Waits for CI/CD checks to complete

```go
func waitForChecks(c *client.Client, ref string, number int, force bool) error {
    const retrySeconds = 60
    var state string
    var err error

    for {
        state, err = c.AggregatedStatus(ref)
        if err != nil {
            return fmt.Errorf("failed to get status: %v", err)
        }
        if state != "pending" {
            break
        }
        if force {
            glog.Warningf("PR is pending, but not waiting because force was specified")
            break
        }
        glog.Warningf("pr %d status is pending: waiting %d seconds", number, retrySeconds)
        time.Sleep(time.Second * retrySeconds)
    }

    if state == "failure" {
        err := fmt.Errorf("pr %d cannot be submitted because it has status %s", number, state)
        if !force {
            return err
        }
        glog.Warningf("because force was specified, ignoring error %v", err)
    }

    return nil
}
```

**3. stripDirectiveMarker** - Removes directive markers from commit messages

```go
func stripDirectiveMarker(message, directive string) string {
    pattern := fmt.Sprintf(`(?m)^%s:\s*[/A-Za-z0-9_.-]+\s*$`, regexp.QuoteMeta(directive))
    re := regexp.MustCompile(pattern)
    cleaned := re.ReplaceAllString(message, "")

    cleaned = regexp.MustCompile(`(?m)^__\s*$`).ReplaceAllString(cleaned, "")

    cleaned = strings.TrimSpace(cleaned)
    cleaned = regexp.MustCompile(`\n\n\n+`).ReplaceAllString(cleaned, "\n\n")

    return cleaned
}
```

#### Updated Core Functions

**submitMsg with marker removal:**

```go
func submitMsg(c *client.Client, prBody string, first, last string, directive string) (string, error) {
    msg := ""
    l := 0
    glog.V(2).Infof("submitMsg begins first=%s, last=%s", first, last)

    for pos := first; pos != last; l++ {
        commit, err := c.Commit(pos)
        if err != nil {
            return "", fmt.Errorf("submitMsg: failed to retrieve commit: %v", err)
        }
        glog.V(2).Infof("submitMsg processes commit: %s", pretty.Sprintf("%# v", commit))
        if parents := len(commit.Parents); parents != 1 {
            return "", fmt.Errorf("submitMsg: commit %s has %d parents", pos, parents)
        }

        cleanedMsg := stripDirectiveMarker(commit.GetMessage(), directive)
        msg = "* " + cleanedMsg + "\n\n" + msg
        glog.V(2).Infof("after commit %s, msg=[%v]", pos, msg)
        pos = *commit.Parents[0].SHA
        if l >= maxCommitChainLength {
            return "", fmt.Errorf("submitMsg: max chain length (%d) exceeded", maxCommitChainLength)
        }
    }

    if l < 2 {
        return stripDirectiveMarker(prBody, directive), nil
    }
    return msg, nil
}
```

**Refactored submitPR:**

```go
func submitPR(c *client.Client, dryRun, force bool, baseBranch string, number int, method, directive string) error {
    pr, err := c.PullRequest(number)
    if err != nil {
        return fmt.Errorf("submitPR: failed to get %d: %v", number, err)
    }

    if err := validatePRAndBranches(c, pr, baseBranch, force); err != nil {
        return err
    }

    if pr.GetMerged() {
        return nil
    }

    if err := waitForChecks(c, pr.GetHead().GetRef(), number, force); err != nil {
        return err
    }

    msg, err := submitMsg(c, *pr.Body, pr.GetHead().GetSHA(), pr.GetBase().GetSHA(), directive)
    if err != nil {
        return fmt.Errorf("submitPR failed to build summitMsg: %v", err)
    }

    if dryRun {
        glog.Warningf("skipping submission of %d because a dry run was requested", number)
        return nil
    }

    if _, err := c.MergePullRequest(number, pr.GetHead().GetSHA(), method, msg); err != nil {
        return fmt.Errorf("failed to submit PR %d: %v", number, err)
    }

    glog.Infof("Successfully submitted %d", number)
    return nil
}
```

**Updated main with directive flag:**

```go
func main() {
    var (
        baseBranch  = flag.String("base", "master", "Base branch")
        baseURL     = flag.String("url", "", "GitHub Base URL")
        directive   = flag.String("directive", "bretmckee-branch", "Directive marker to remove from commit messages")
        dryRun      = flag.Bool("dry-run", false, "Dry Run mode -- no pull requests will be created")
        force       = flag.Bool("force", false, "Submit even if not fully approved.")
        login       = flag.String("login", "", "Login of the user to submit for.")
        method      = flag.String("method", "squash", "github merge method -- [merge|rebase|squash]")
        pr          = flag.Int("pr", 0, "id of the pull request to submit")
        sourceOwner = flag.String("source-owner", "", "Name of the owner (user or org) of the repo to create the commit in.")
        sourceRepo  = flag.String("source-repo", "", "Name of repo to create the commit in.")
        token       = flag.String("token", "", "github auth token to use (also checks environment GITHUB_TOKEN")
        uploadURL   = flag.String("upload", "", "GitHub Upload URL")
    )
    flag.Parse()

    if *token == "" {
        *token = os.Getenv("GITHUB_TOKEN")
    }
    if *token == "" {
        glog.Exit("Unauthorized: No token present")
    }
    if *sourceOwner == "" || *sourceRepo == "" || *login == "" {
        glog.Exitf("A non-empty value must be specified for the flags `-source-owner (=%q)`, `-source-repo (=%q)` and `-login (=%q)`", *sourceOwner, *sourceRepo, *login)
    }
    if *pr <= 0 {
        glog.Exit("An positive integer value must be specified for `-pr`")
    }

    b, u, err := urls.Get(*baseURL, *uploadURL)
    if err != nil {
        glog.Exitf("failed to get URLs: %v", err)
    }

    c, err := client.Create(b, u, *sourceOwner, *sourceRepo, *login, *token)
    if err != nil {
        glog.Exitf("failed to create client: %v", err)
    }

    if err := submitPR(c, *dryRun, *force, *baseBranch, *pr, *method, *directive); err != nil {
        glog.Exitf("submitPR failed: %v", err)
    }
}
```

#### Update Shell Script

Modify `scripts/submit-prs` to pass directive from environment:

```bash
DIRECTIVE=${DIRECTIVE:-bretmckee-branch}

# In the submit-pr call (around line 46), add --directive flag:
submit-pr ${FORCE} ${DRY_RUN} --base="${BASE_BRANCH}" --directive="${DIRECTIVE}" \
  --source-owner="${SOURCE_OWNER}" --source-repo="${SOURCE_REPO}" \
  --login="${LOGIN}" --url="${URL}" --stderrthreshold="${THRESHOLD}" \
  -v="${VERBOSITY}" --pr="${PR}"
```

### Complete Example: After Implementation

```
Local Commit Message:
Update README.md

Add some more information to the read me.
__
bretmckee-branch: update-readme

↓ (git-push-branches finds marker and creates branch)

PR Branch: bretmckee/update-readme
PR Body: (same as commit message)

↓ (submit-pr merges with squash, stripDirectiveMarker removes marker)

Main Branch Commit Message:
* Update README.md

Add some more information to the read me.
```

## Implementation Guidance

### Objectives

1. Remove directive markers from final squash commit messages
2. Maintain clean git history on main branch
3. Preserve all other commit message content
4. Provide configurability for different directive patterns
5. **Improve code maintainability by decomposing submitPR into focused
   functions**
6. **Enhance testability by extracting validation and status checking logic**

### Key Tasks

1. Add `stripDirectiveMarker()` helper function to clean commit messages
2. Add `validatePRAndBranches()` to extract PR and branch validation logic
3. Add `waitForChecks()` to extract status checking and retry logic
4. Update `submitMsg()` signature to accept directive parameter and clean
   messages
5. Refactor `submitPR()` to use new helper functions and accept directive
   parameter
6. Add `--directive` command-line flag with default value matching
   git-push-branches
7. Update `scripts/submit-prs` to pass directive from environment variable
8. Add necessary imports: `regexp`, `strings`

### Refactoring Benefits

- **Single Responsibility**: Each function has one clear purpose
- **Reduced Complexity**: submitPR goes from 70 lines to ~30 lines
- **Testability**: Helper functions can be unit tested independently
- **Readability**: Function names document what each section does
- **Reusability**: Validation and status checking logic can be reused

### Dependencies

- Uses Go standard library `regexp` and `strings` packages
- Requires `github.com/google/go-github/v28/github` for PR type
- No external dependencies required
- Compatible with existing `go-github` integration
- No interface changes (changes are internal to cmd/submit-pr)

### Success Criteria

1. Squashed commit messages on main branch do not contain directive markers
2. Commit message titles and bodies are preserved (except markers and
   separators)
3. Directive pattern is configurable via command-line flag
4. Single-commit PRs also have markers removed from PR body
5. Shell script passes directive consistently with `git-push-branches`
   configuration
6. Cleaned messages have proper whitespace (no trailing blanks or multiple blank
   lines)
7. **submitPR function is reduced to ~30 lines with clear, focused logic**
8. **Each helper function has a single, well-defined responsibility**
9. **All existing functionality is preserved (backward compatible)**
10. **Unit tests cover all new helper functions**
11. **Unit tests verify marker removal with various input patterns**

## Unit Testing Strategy

### Test File Structure

Create `cmd/submit-pr/main_test.go` following project conventions:

```go
package main

import (
    "testing"
    "github.com/google/go-github/v28/github"
)
```

### Test Cases for stripDirectiveMarker()

Table-driven tests covering various marker patterns and edge cases:

```go
func TestStripDirectiveMarker(t *testing.T) {
    tests := []struct {
        name      string
        message   string
        directive string
        want      string
    }{
        {
            name:      "Simple marker removal",
            message:   "Update README\n\nAdd docs\n__\nbretmckee-branch: update-readme",
            directive: "bretmckee-branch",
            want:      "Update README\n\nAdd docs",
        },
        {
            name:      "Marker without separator",
            message:   "Fix bug\n\nbretmckee-branch: fix-bug",
            directive: "bretmckee-branch",
            want:      "Fix bug",
        },
        {
            name:      "Multiple separators",
            message:   "Title\n\nBody\n__\n__\nbretmckee-branch: test",
            directive: "bretmckee-branch",
            want:      "Title\n\nBody",
        },
        {
            name:      "No marker present",
            message:   "Regular commit\n\nNo marker here",
            directive: "bretmckee-branch",
            want:      "Regular commit\n\nNo marker here",
        },
        {
            name:      "Marker with dashes in branch name",
            message:   "Title\n__\nbretmckee-branch: feature-name-123",
            directive: "bretmckee-branch",
            want:      "Title",
        },
        {
            name:      "Marker with slashes in branch name",
            message:   "Title\n__\nbretmckee-branch: feature/sub-feature",
            directive: "bretmckee-branch",
            want:      "Title",
        },
        {
            name:      "Custom directive",
            message:   "Title\n__\ncustom-dir: branch-name",
            directive: "custom-dir",
            want:      "Title",
        },
        {
            name:      "Trailing whitespace cleanup",
            message:   "Title\n\n\n\n__\nbretmckee-branch: test",
            directive: "bretmckee-branch",
            want:      "Title",
        },
        {
            name:      "Empty message",
            message:   "",
            directive: "bretmckee-branch",
            want:      "",
        },
        {
            name:      "Only marker",
            message:   "bretmckee-branch: test",
            directive: "bretmckee-branch",
            want:      "",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := stripDirectiveMarker(tt.message, tt.directive)
            if got != tt.want {
                t.Errorf("stripDirectiveMarker() = %q, want %q", got, tt.want)
            }
        })
    }
}
```

### Test Cases for validatePRAndBranches()

Mock-based tests using httptest for API interactions:

```go
func TestValidatePRAndBranches(t *testing.T) {
    tests := []struct {
        name        string
        pr          *github.PullRequest
        baseBranch  string
        force       bool
        setupServer func(*testing.T) *httptest.Server
        wantErr     bool
        errContains string
    }{
        {
            name:       "Already merged PR returns nil",
            pr:         &github.PullRequest{Number: github.Int(1), Merged: github.Bool(true)},
            baseBranch: "main",
            force:      false,
            wantErr:    false,
        },
        {
            name: "Base ref mismatch without force",
            pr: &github.PullRequest{
                Number: github.Int(1),
                Merged: github.Bool(false),
                Base:   &github.PullRequestBranch{Ref: github.String("develop")},
            },
            baseBranch:  "main",
            force:       false,
            wantErr:     true,
            errContains: "does not match base branch ref",
        },
        {
            name: "Base ref mismatch with force",
            pr: &github.PullRequest{
                Number: github.Int(1),
                Merged: github.Bool(false),
                Base:   &github.PullRequestBranch{Ref: github.String("develop")},
            },
            baseBranch: "main",
            force:      true,
            wantErr:    false,
        },
        // Additional test cases for SHA validation, branch retrieval errors, etc.
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var server *httptest.Server
            if tt.setupServer != nil {
                server = tt.setupServer(t)
                defer server.Close()
            }

            // Create test client and call validatePRAndBranches
            // Verify error matches expectations
        })
    }
}
```

### Test Cases for waitForChecks()

Tests for status checking loop with timeout scenarios:

```go
func TestWaitForChecks(t *testing.T) {
    tests := []struct {
        name           string
        statusSequence []string // Sequence of status responses
        force          bool
        wantErr        bool
        errContains    string
        maxIterations  int
    }{
        {
            name:           "Success on first check",
            statusSequence: []string{"success"},
            force:          false,
            wantErr:        false,
        },
        {
            name:           "Pending then success",
            statusSequence: []string{"pending", "pending", "success"},
            force:          false,
            wantErr:        false,
            maxIterations:  3,
        },
        {
            name:           "Failure without force",
            statusSequence: []string{"failure"},
            force:          false,
            wantErr:        true,
            errContains:    "cannot be submitted",
        },
        {
            name:           "Failure with force",
            statusSequence: []string{"failure"},
            force:          true,
            wantErr:        false,
        },
        {
            name:           "Pending with force skips wait",
            statusSequence: []string{"pending"},
            force:          true,
            wantErr:        false,
            maxIterations:  1, // Should not retry
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Mock AggregatedStatus to return sequence of statuses
            // Verify correct number of iterations
            // Verify error handling
        })
    }
}
```

### Test Cases for submitMsg()

Tests for message building with marker removal:

```go
func TestSubmitMsg(t *testing.T) {
    tests := []struct {
        name      string
        prBody    string
        commits   []*github.Commit // Mock commit chain
        directive string
        want      string
        wantErr   bool
    }{
        {
            name:      "Single commit uses PR body with marker removed",
            prBody:    "Title\n\nBody\n__\nbretmckee-branch: test",
            commits:   []*github.Commit{makeCommit("sha1", "parent", "Commit msg")},
            directive: "bretmckee-branch",
            want:      "Title\n\nBody",
            wantErr:   false,
        },
        {
            name:   "Multiple commits with markers removed",
            prBody: "PR Body",
            commits: []*github.Commit{
                makeCommit("sha2", "sha1", "Second\n__\nbretmckee-branch: test2"),
                makeCommit("sha1", "base", "First\n__\nbretmckee-branch: test1"),
            },
            directive: "bretmckee-branch",
            want:      "* First\n\n* Second\n\n",
            wantErr:   false,
        },
        // Additional test cases for error conditions, chain length limits, etc.
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Create mock client that returns test commits
            // Call submitMsg
            // Verify output matches expected cleaned message
        })
    }
}
```

### Integration Test Considerations

While unit tests cover individual functions, consider adding integration test:

```go
func TestSubmitPR_Integration(t *testing.T) {
    // End-to-end test with mock GitHub server
    // Verifies complete flow: validation → status check → message build → merge
    // Ensures all components work together correctly
}
```

### Test Utilities

Helper functions to reduce test boilerplate:

```go
func makeCommit(sha, parentSHA, message string) *github.Commit {
    return &github.Commit{
        SHA:     github.String(sha),
        Message: github.String(message),
        Parents: []github.Commit{{SHA: github.String(parentSHA)}},
    }
}

func makePR(number int, baseRef, baseSHA, headRef, headSHA string, merged bool, body string) *github.PullRequest {
    return &github.PullRequest{
        Number: github.Int(number),
        Merged: github.Bool(merged),
        Body:   github.String(body),
        Base: &github.PullRequestBranch{
            Ref: github.String(baseRef),
            SHA: github.String(baseSHA),
        },
        Head: &github.PullRequestBranch{
            Ref: github.String(headRef),
            SHA: github.String(headSHA),
        },
    }
}
```

### Running Tests

```bash
# Run all tests
cd cmd/submit-pr && go test -v

# Run specific test
go test -v -run TestStripDirectiveMarker

# Run with coverage
go test -cover -coverprofile=coverage.out
go tool cover -html=coverage.out

# Target: >80% coverage for new functions
```
