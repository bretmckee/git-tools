<!-- markdownlint-disable-file -->
# Task Research Notes: GitHub Actions Check Status Integration

## Research Executed

### File Analysis
- `cmd/submit-pr/main.go` (lines 70-115)
  - Currently uses `CombinedStatus` API which doesn't work with GitHub Actions
  - Checks status in a loop until not "pending", then validates against "failure"
  - Uses ref from PR head to check status
- `pkg/repo/client/status.go`
  - Implements `CombinedStatus()` method calling `Repositories.GetCombinedStatus()`
  - Only method for status checking in the client
- `pkg/repo/repo.go`
  - Defines `Repo` interface with `CombinedStatus(ref string) (*github.CombinedStatus, error)`
  - Interface is implemented by client in `pkg/repo/client/client.go`

### Code Search Results
- `go-github` library version: `v28.1.1` (from `go.mod`)
  - Uses older v28 API which has Check Runs support via `Checks` service
  - `ChecksService.ListCheckRunsForRef()` method available

### External Research
- #fetch:https://docs.github.com/en/rest/checks/runs
  - Check Runs API is the modern way to report CI/CD status
  - Endpoint: `GET /repos/{owner}/{repo}/commits/{ref}/check-runs`
  - Returns: `ListCheckRunsResults` with array of `CheckRun` objects
  - Each CheckRun has:
    - `status`: "queued", "in_progress", "completed", "waiting", "requested", "pending"
    - `conclusion`: "success", "failure", "neutral", "cancelled", "skipped", "timed_out", "action_required", "stale"
  - Requires "Checks" repository permissions (read)

- #fetch:https://docs.github.com/en/rest/commits/statuses
  - Combined Status API is the legacy approach
  - Does NOT include GitHub Actions workflows
  - Only works with commit status API (older CI systems)
  - State values: "error", "failure", "pending", "success"
  - Note in docs: "If you are developing a GitHub App and want to provide more detailed information about an external service, you may want to use the REST API to manage checks."

- #githubRepo:"google/go-github" ListCheckRunsForRef status conclusion completed
  - Method signature: `ListCheckRunsForRef(ctx context.Context, owner, repo, ref string, opts *ListCheckRunsOptions) (*ListCheckRunsResults, *Response, error)`
  - Returns `ListCheckRunsResults` containing:
    - `Total *int` - total count of check runs
    - `CheckRuns []*CheckRun` - array of check run objects
  - CheckRun structure contains:
    - `Status *string` - current status of the check run
    - `Conclusion *string` - final conclusion (only set when status is "completed")
    - `Name *string` - name of the check
    - `App *App` - the GitHub App that created the check
  - Options struct `ListCheckRunsOptions`:
    - `CheckName *string` - filter by check name
    - `Status *string` - filter by status ("queued", "in_progress", "completed")
    - `Filter *string` - "latest" or "all"
    - `AppID *int64` - filter by GitHub App ID

## Key Discoveries

### Project Structure
- Uses interface-based abstraction in `pkg/repo/repo.go`
- Client implementation in `pkg/repo/client/` directory
- Main programs in `cmd/` directory (submit-pr, create-reviews, rebase-prs)
- Repo interface allows for testing and different implementations

### Implementation Patterns
The codebase follows these patterns:
- Interface definitions in `pkg/repo/repo.go`
- Concrete implementations in `pkg/repo/client/`
- Error handling with formatted errors including context
- glog for logging at various verbosity levels
- Uses `go-github` v28 library for GitHub API interactions

### Complete Examples

#### Current CombinedStatus Implementation (Legacy)
```go
func (c *Client) CombinedStatus(ref string) (*github.CombinedStatus, error) {
    o := &github.ListOptions{}
    s, _, err := c.client.Repositories.GetCombinedStatus(c.ctx, c.owner, c.repo, ref, o)
    if err != nil {
        return nil, fmt.Errorf("Failed to get statuses for %q: %v", ref, err)
    }
    if glog.V(3) {
        glog.Infof("combined status of %q: %# v\n", ref, pretty.Formatter(*s))
    }
    return s, nil
}
```

#### Example from go-github for ListCheckRunsForRef
```go
func (s *ChecksService) ListCheckRunsForRef(ctx context.Context, owner, repo, ref string, opts *ListCheckRunsOptions) (*ListCheckRunsResults, *Response, error) {
    u := fmt.Sprintf("repos/%v/%v/commits/%v/check-runs", owner, repo, refURLEscape(ref))
    u, err := addOptions(u, opts)
    if err != nil {
        return nil, nil, err
    }

    req, err := s.client.NewRequest("GET", u, nil)
    if err != nil {
        return nil, nil, err
    }

    req.Header.Set("Accept", mediaTypeCheckRunsPreview)

    var checkRunResults *ListCheckRunsResults
    resp, err := s.client.Do(ctx, req, &checkRunResults)
    if err != nil {
        return nil, resp, err
    }

    return checkRunResults, resp, nil
}
```

### API and Schema Documentation

#### CheckRun Structure
```go
type CheckRun struct {
    ID           *int64          `json:"id,omitempty"`
    NodeID       *string         `json:"node_id,omitempty"`
    HeadSHA      *string         `json:"head_sha,omitempty"`
    ExternalID   *string         `json:"external_id,omitempty"`
    URL          *string         `json:"url,omitempty"`
    HTMLURL      *string         `json:"html_url,omitempty"`
    DetailsURL   *string         `json:"details_url,omitempty"`
    Status       *string         `json:"status,omitempty"`      // "queued", "in_progress", "completed", "waiting", "requested", "pending"
    Conclusion   *string         `json:"conclusion,omitempty"`  // "success", "failure", "neutral", "cancelled", "skipped", "timed_out", "action_required", "stale"
    StartedAt    *Timestamp      `json:"started_at,omitempty"`
    CompletedAt  *Timestamp      `json:"completed_at,omitempty"`
    Output       *CheckRunOutput `json:"output,omitempty"`
    Name         *string         `json:"name,omitempty"`
    CheckSuite   *CheckSuite     `json:"check_suite,omitempty"`
    App          *App            `json:"app,omitempty"`
    PullRequests []*PullRequest  `json:"pull_requests,omitempty"`
}

type ListCheckRunsResults struct {
    Total     *int        `json:"total_count,omitempty"`
    CheckRuns []*CheckRun `json:"check_runs,omitempty"`
}

type ListCheckRunsOptions struct {
    CheckName *string `url:"check_name,omitempty"` // filter by check name
    Status    *string `url:"status,omitempty"`     // "queued", "in_progress", or "completed"
    Filter    *string `url:"filter,omitempty"`     // "latest" or "all" (default: "latest")
    AppID     *int64  `url:"app_id,omitempty"`     // filter by GitHub App ID
    ListOptions
}
```

#### CombinedStatus Structure (Legacy)
```go
type CombinedStatus struct {
    State      *string        `json:"state,omitempty"`       // "success", "failure", "pending", "error"
    SHA        *string        `json:"sha,omitempty"`
    TotalCount *int           `json:"total_count,omitempty"`
    Statuses   []RepoStatus   `json:"statuses,omitempty"`
    // ... other fields
}
```

### Technical Requirements

#### Status Determination Logic
The code needs to determine PR readiness by checking ALL check runs for a ref:

1. **Pending State**: If ANY check run has status != "completed", PR is pending
2. **Success State**: All check runs must have:
   - `status == "completed"`
   - `conclusion == "success"`
3. **Failure State**: Any check run with `conclusion` in ["failure", "timed_out", "action_required"]
4. **Neutral/Skipped**: Check runs with "neutral", "cancelled", or "skipped" should not block PR

#### Backward Compatibility
- Must continue to work with repos using legacy commit status API
- Need to check BOTH Check Runs API and Combined Status API
- If Check Runs API returns no results, fall back to Combined Status
- Aggregate results from both APIs into single overall state

#### API Access Requirements
- "Checks" repository permissions (read) for Check Runs API
- "Commit statuses" repository permissions (read) for Combined Status API
- Personal access tokens already have `repo` scope which includes both

## Recommended Approach

### Hybrid Status Checking Implementation

Implement a new method that checks both Check Runs and Combined Status APIs, combining results into a unified status. This ensures compatibility with both modern GitHub Actions workflows and legacy CI systems.

#### Interface Updates
Add new method to `pkg/repo/repo.go`:
```go
type Repo interface {
    // ... existing methods ...
    
    // CheckRunsForRef returns check runs for a specific ref
    CheckRunsForRef(ref string, opts *github.ListCheckRunsOptions) (*github.ListCheckRunsResults, error)
}
```

#### Client Implementation
Add to `pkg/repo/client/status.go`:
```go
func (c *Client) CheckRunsForRef(ref string, opts *github.ListCheckRunsOptions) (*github.ListCheckRunsResults, error) {
    if opts == nil {
        opts = &github.ListCheckRunsOptions{
            Filter: github.String("latest"),
        }
    }
    
    results, _, err := c.client.Checks.ListCheckRunsForRef(c.ctx, c.owner, c.repo, ref, opts)
    if err != nil {
        return nil, fmt.Errorf("Failed to get check runs for %q: %v", ref, err)
    }
    
    if glog.V(3) {
        glog.Infof("check runs for %q: %# v\n", ref, pretty.Formatter(*results))
    }
    
    return results, nil
}

// AggregatedStatus returns an overall status by checking both Check Runs and Combined Status
func (c *Client) AggregatedStatus(ref string) (state string, err error) {
    // Check GitHub Actions (Check Runs API)
    checkRuns, err := c.CheckRunsForRef(ref, nil)
    if err != nil {
        return "", fmt.Errorf("Failed to get check runs: %v", err)
    }
    
    // Check legacy statuses (Combined Status API)
    combinedStatus, err := c.CombinedStatus(ref)
    if err != nil {
        return "", fmt.Errorf("Failed to get combined status: %v", err)
    }
    
    // Aggregate: if either has pending/in_progress, overall is pending
    // If either has failure, overall is failure
    // Only if all are success, overall is success
    
    hasCheckRuns := checkRuns != nil && checkRuns.Total != nil && *checkRuns.Total > 0
    hasStatuses := combinedStatus != nil && combinedStatus.TotalCount != nil && *combinedStatus.TotalCount > 0
    
    if !hasCheckRuns && !hasStatuses {
        return "success", nil // No checks configured
    }
    
    // Check if any check runs are pending or in progress
    if hasCheckRuns {
        for _, run := range checkRuns.CheckRuns {
            status := run.GetStatus()
            if status == "queued" || status == "in_progress" || status == "waiting" || status == "requested" || status == "pending" {
                return "pending", nil
            }
        }
    }
    
    // Check combined status
    if hasStatuses {
        if combinedStatus.GetState() == "pending" {
            return "pending", nil
        }
    }
    
    // Check for failures in check runs
    if hasCheckRuns {
        for _, run := range checkRuns.CheckRuns {
            if run.GetStatus() == "completed" {
                conclusion := run.GetConclusion()
                if conclusion == "failure" || conclusion == "timed_out" || conclusion == "action_required" {
                    return "failure", nil
                }
            }
        }
    }
    
    // Check for failures in combined status
    if hasStatuses {
        state := combinedStatus.GetState()
        if state == "failure" || state == "error" {
            return "failure", nil
        }
    }
    
    // All checks completed successfully
    return "success", nil
}
```

#### Update submit-pr to use new method
Modify `cmd/submit-pr/main.go` to use `AggregatedStatus`:
```go
// Replace current status checking loop (lines 78-95)
ref := pr.GetHead().GetRef()
var state string
for {
    // TODO(bretmckee): Consider an argument to terminate this loop after a timeout.
    state, err = c.AggregatedStatus(ref)
    if err != nil {
        return fmt.Errorf("submitPR: failed to get status: %v", err)
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
```

## Implementation Guidance

### Objectives
1. Support GitHub Actions workflows by using Check Runs API
2. Maintain backward compatibility with legacy commit status API
3. Provide unified status determination for PR submission
4. Minimize breaking changes to existing interface

### Key Tasks
1. Add `CheckRunsForRef` method to Repo interface
2. Implement `CheckRunsForRef` in client
3. Implement `AggregatedStatus` helper method in client (not in interface - internal use)
4. Update `submit-pr` to use new aggregated status checking
5. Add appropriate logging for debugging status checks

### Dependencies
- `go-github` v28.1.1 already in use supports Check Runs API
- No new external dependencies required
- Uses existing `github.Checks` service from go-github

### Success Criteria
1. `submit-pr` correctly waits for GitHub Actions workflows to complete
2. Backward compatibility: repos without GitHub Actions continue to work
3. Repos with both GitHub Actions and legacy statuses work correctly
4. Appropriate logging shows status of all checks during wait
5. All existing tests continue to pass
