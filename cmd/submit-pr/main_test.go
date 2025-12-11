package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bretmckee/git-tools/pkg/repo/client"
	"github.com/google/go-github/v28/github"
)

func TestStripDirectiveMarker(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		directive string
		want      string
	}{
		{
			name:      "Simple marker removal",
			message:   "Fix bug in handler\n__\nbretmckee-branch: fix-bug",
			directive: "bretmckee-branch",
			want:      "Fix bug in handler",
		},
		{
			name:      "Marker without separator",
			message:   "Add new feature\nbretmckee-branch: new-feature",
			directive: "bretmckee-branch",
			want:      "Add new feature",
		},
		{
			name:      "Branch name with dashes",
			message:   "Update docs\n__\nbretmckee-branch: update-user-docs",
			directive: "bretmckee-branch",
			want:      "Update docs",
		},
		{
			name:      "Branch name with slashes",
			message:   "Feature work\n__\nbretmckee-branch: feature/user-auth",
			directive: "bretmckee-branch",
			want:      "Feature work",
		},
		{
			name:      "Branch name with dots",
			message:   "Version bump\n__\nbretmckee-branch: release-1.2.3",
			directive: "bretmckee-branch",
			want:      "Version bump",
		},
		{
			name:      "Custom directive",
			message:   "Custom work\n__\ncustom-directive: branch-name",
			directive: "custom-directive",
			want:      "Custom work",
		},
		{
			name:      "No marker present",
			message:   "Simple commit message",
			directive: "bretmckee-branch",
			want:      "Simple commit message",
		},
		{
			name:      "Empty message",
			message:   "",
			directive: "bretmckee-branch",
			want:      "",
		},
		{
			name:      "Only marker",
			message:   "__\nbretmckee-branch: test",
			directive: "bretmckee-branch",
			want:      "",
		},
		{
			name:      "Multiple newlines normalized",
			message:   "Line 1\n\n\n\nLine 2\n__\nbretmckee-branch: test",
			directive: "bretmckee-branch",
			want:      "Line 1\n\nLine 2",
		},
		{
			name:      "Whitespace around marker",
			message:   "Message\n__  \nbretmckee-branch: test  ",
			directive: "bretmckee-branch",
			want:      "Message",
		},
		{
			name:      "Marker in middle preserved if not at end",
			message:   "Before\nbretmckee-branch: test\nAfter",
			directive: "bretmckee-branch",
			want:      "Before\nbretmckee-branch: test\nAfter",
		},
		{
			name:      "Three underscores separator",
			message:   "Message\n___\nbretmckee-branch: test",
			directive: "bretmckee-branch",
			want:      "Message",
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

func TestValidatePRAndBranches(t *testing.T) {
	tests := []struct {
		name       string
		pr         *github.PullRequest
		baseBranch string
		force      bool
		branchSHA  string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "Already merged PR",
			pr: &github.PullRequest{
				Number: github.Int(1),
				Merged: github.Bool(true),
			},
			baseBranch: "main",
			force:      false,
			wantErr:    true,
			wantErrMsg: "PR is already merged",
		},
		{
			name: "Base ref mismatch without force",
			pr: &github.PullRequest{
				Number: github.Int(2),
				Merged: github.Bool(false),
				Base: &github.PullRequestBranch{
					Ref: github.String("develop"),
					SHA: github.String("abc123"),
				},
			},
			baseBranch: "main",
			branchSHA:  "abc123",
			force:      false,
			wantErr:    true,
			wantErrMsg: "pr base ref",
		},
		{
			name: "Base ref mismatch with force",
			pr: &github.PullRequest{
				Number: github.Int(3),
				Merged: github.Bool(false),
				Base: &github.PullRequestBranch{
					Ref: github.String("develop"),
					SHA: github.String("abc123"),
				},
			},
			baseBranch: "main",
			branchSHA:  "abc123",
			force:      true,
			wantErr:    false,
		},
		{
			name: "Base SHA mismatch without force",
			pr: &github.PullRequest{
				Number: github.Int(4),
				Merged: github.Bool(false),
				Base: &github.PullRequestBranch{
					Ref: github.String("main"),
					SHA: github.String("abc123"),
				},
			},
			baseBranch: "main",
			branchSHA:  "def456",
			force:      false,
			wantErr:    true,
			wantErrMsg: "pr base SHA",
		},
		{
			name: "Base SHA mismatch with force",
			pr: &github.PullRequest{
				Number: github.Int(5),
				Merged: github.Bool(false),
				Base: &github.PullRequestBranch{
					Ref: github.String("main"),
					SHA: github.String("abc123"),
				},
			},
			baseBranch: "main",
			branchSHA:  "def456",
			force:      true,
			wantErr:    false,
		},
		{
			name: "Valid PR",
			pr: &github.PullRequest{
				Number: github.Int(6),
				Merged: github.Bool(false),
				Base: &github.PullRequestBranch{
					Ref: github.String("main"),
					SHA: github.String("abc123"),
				},
			},
			baseBranch: "main",
			branchSHA:  "abc123",
			force:      false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/repos/test/test/branches/"+tt.baseBranch {
					branch := &github.Branch{
						Commit: &github.RepositoryCommit{
							SHA: github.String(tt.branchSHA),
						},
					}
					json.NewEncoder(w).Encode(branch)
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			c, err := client.Create(server.URL, server.URL, "test", "test", "user", "token")
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			err = validatePRAndBranches(c, tt.pr, tt.baseBranch, tt.force)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePRAndBranches() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErrMsg != "" {
				if err.Error() != tt.wantErrMsg && err != errAlreadyMerged {
					errContains := false
					if tt.wantErrMsg != "" {
						errContains = contains(err.Error(), tt.wantErrMsg)
					}
					if !errContains {
						t.Errorf("validatePRAndBranches() error = %q, want to contain %q", err.Error(), tt.wantErrMsg)
					}
				}
			}
		})
	}
}

func TestWaitForChecks(t *testing.T) {
	tests := []struct {
		name        string
		states      []string
		force       bool
		wantErr     bool
		wantErrMsg  string
		wantRetries int
	}{
		{
			name:        "Immediate success",
			states:      []string{"success"},
			force:       false,
			wantErr:     false,
			wantRetries: 0,
		},
		{
			name:        "Immediate failure without force",
			states:      []string{"failure"},
			force:       false,
			wantErr:     true,
			wantErrMsg:  "cannot be submitted",
			wantRetries: 0,
		},
		{
			name:        "Immediate failure with force",
			states:      []string{"failure"},
			force:       true,
			wantErr:     false,
			wantRetries: 0,
		},
		{
			name:        "Pending then success",
			states:      []string{"pending", "success"},
			force:       false,
			wantErr:     false,
			wantRetries: 1,
		},
		{
			name:        "Pending with force skips wait",
			states:      []string{"pending"},
			force:       true,
			wantErr:     false,
			wantRetries: 0,
		},
		{
			name:        "Multiple pending then success",
			states:      []string{"pending", "pending", "pending", "success"},
			force:       false,
			wantErr:     false,
			wantRetries: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/repos/test/test/commits/testref/check-runs" {
					state := tt.states[callCount]
					if callCount < len(tt.states)-1 {
						callCount++
					}
					results := &github.ListCheckRunsResults{
						Total: github.Int(1),
						CheckRuns: []*github.CheckRun{
							{
								Name:       github.String("test"),
								Status:     github.String(mapStateToCheckStatus(state)),
								Conclusion: github.String(mapStateToConclusion(state)),
							},
						},
					}
					json.NewEncoder(w).Encode(results)
					return
				}
				if r.URL.Path == "/repos/test/test/commits/testref/status" {
					status := &github.CombinedStatus{
						State:      github.String("success"),
						TotalCount: github.Int(0),
					}
					json.NewEncoder(w).Encode(status)
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			c, err := client.Create(server.URL, server.URL, "test", "test", "user", "token")
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			noopSleep := func(d time.Duration) {}

			err = waitForChecks(c, "testref", 123, tt.force, noopSleep)
			if (err != nil) != tt.wantErr {
				t.Errorf("waitForChecks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErrMsg != "" {
				if !contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("waitForChecks() error = %q, want to contain %q", err.Error(), tt.wantErrMsg)
				}
			}
		})
	}
}

func mapStateToCheckStatus(state string) string {
	if state == "pending" {
		return "in_progress"
	}
	return "completed"
}

func mapStateToConclusion(state string) string {
	if state == "failure" {
		return "failure"
	}
	if state == "pending" {
		return ""
	}
	return "success"
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestSubmitMsg(t *testing.T) {
	tests := []struct {
		name      string
		prBody    string
		commits   map[string]*github.Commit
		first     string
		last      string
		directive string
		want      string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "Single commit uses PR body with marker removed",
			prBody:    "Title\n\nBody\n__\nbretmckee-branch: test",
			commits:   map[string]*github.Commit{},
			first:     "sha1",
			last:      "sha1",
			directive: "bretmckee-branch",
			want:      "Title\n\nBody",
			wantErr:   false,
		},
		{
			name:   "Two commits with markers removed",
			prBody: "PR Body",
			commits: map[string]*github.Commit{
				"sha2": {
					SHA:     github.String("sha2"),
					Message: github.String("Second commit\n__\nbretmckee-branch: test2"),
					Parents: []github.Commit{{SHA: github.String("sha1")}},
				},
				"sha1": {
					SHA:     github.String("sha1"),
					Message: github.String("First commit\n__\nbretmckee-branch: test1"),
					Parents: []github.Commit{{SHA: github.String("base")}},
				},
			},
			first:     "sha2",
			last:      "base",
			directive: "bretmckee-branch",
			want:      "* First commit\n\n* Second commit\n\n",
			wantErr:   false,
		},
		{
			name:   "Multiple commits concatenated",
			prBody: "PR Body",
			commits: map[string]*github.Commit{
				"sha3": {
					SHA:     github.String("sha3"),
					Message: github.String("Third\n__\nbretmckee-branch: c3"),
					Parents: []github.Commit{{SHA: github.String("sha2")}},
				},
				"sha2": {
					SHA:     github.String("sha2"),
					Message: github.String("Second\n__\nbretmckee-branch: c2"),
					Parents: []github.Commit{{SHA: github.String("sha1")}},
				},
			},
			first:     "sha3",
			last:      "sha1",
			directive: "bretmckee-branch",
			want:      "* Second\n\n* Third\n\n",
			wantErr:   false,
		},
		{
			name:   "Commit with multiple parents fails",
			prBody: "PR Body",
			commits: map[string]*github.Commit{
				"merge": {
					SHA:     github.String("merge"),
					Message: github.String("Merge commit"),
					Parents: []github.Commit{
						{SHA: github.String("sha1")},
						{SHA: github.String("sha2")},
					},
				},
			},
			first:     "merge",
			last:      "sha1",
			directive: "bretmckee-branch",
			wantErr:   true,
			errMsg:    "has 2 parents",
		},
		{
			name:   "No marker in commits",
			prBody: "Clean PR body",
			commits: map[string]*github.Commit{
				"sha2": {
					SHA:     github.String("sha2"),
					Message: github.String("Second clean message"),
					Parents: []github.Commit{{SHA: github.String("sha1")}},
				},
				"sha1": {
					SHA:     github.String("sha1"),
					Message: github.String("First clean message"),
					Parents: []github.Commit{{SHA: github.String("base")}},
				},
			},
			first:     "sha2",
			last:      "base",
			directive: "bretmckee-branch",
			want:      "* First clean message\n\n* Second clean message\n\n",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for sha, commit := range tt.commits {
					if r.URL.Path == "/repos/test/test/git/commits/"+sha {
						json.NewEncoder(w).Encode(commit)
						return
					}
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			c, err := client.Create(server.URL, server.URL, "test", "test", "user", "token")
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			got, err := submitMsg(c, tt.prBody, tt.first, tt.last, tt.directive)
			if (err != nil) != tt.wantErr {
				t.Errorf("submitMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
				t.Errorf("submitMsg() error = %q, want to contain %q", err.Error(), tt.errMsg)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("submitMsg() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSubmitPR(t *testing.T) {
	tests := []struct {
		name        string
		pr          *github.PullRequest
		branch      *github.Branch
		checkStates []string
		commits     map[string]*github.Commit
		dryRun      bool
		force       bool
		baseBranch  string
		directive   string
		wantErr     bool
		errMsg      string
		wantMerged  bool
	}{
		{
			name: "Successful submission",
			pr: &github.PullRequest{
				Number: github.Int(1),
				Merged: github.Bool(false),
				Body:   github.String("PR body\n__\nbretmckee-branch: test"),
				Base: &github.PullRequestBranch{
					Ref: github.String("main"),
					SHA: github.String("base-sha"),
				},
				Head: &github.PullRequestBranch{
					Ref: github.String("feature"),
					SHA: github.String("head-sha"),
				},
			},
			branch: &github.Branch{
				Commit: &github.RepositoryCommit{
					SHA: github.String("base-sha"),
				},
			},
			checkStates: []string{"success"},
			commits: map[string]*github.Commit{
				"head-sha": {
					SHA:     github.String("head-sha"),
					Message: github.String("Test commit"),
					Parents: []github.Commit{{SHA: github.String("base-sha")}},
				},
			},
			dryRun:     false,
			force:      false,
			baseBranch: "main",
			directive:  "bretmckee-branch",
			wantErr:    false,
			wantMerged: true,
		},
		{
			name: "Already merged PR returns early",
			pr: &github.PullRequest{
				Number: github.Int(2),
				Merged: github.Bool(true),
				Base: &github.PullRequestBranch{
					Ref: github.String("main"),
					SHA: github.String("base-sha"),
				},
			},
			baseBranch: "main",
			directive:  "bretmckee-branch",
			wantErr:    false,
			wantMerged: false,
		},
		{
			name: "Dry run skips merge",
			pr: &github.PullRequest{
				Number: github.Int(3),
				Merged: github.Bool(false),
				Body:   github.String("PR body"),
				Base: &github.PullRequestBranch{
					Ref: github.String("main"),
					SHA: github.String("base-sha"),
				},
				Head: &github.PullRequestBranch{
					Ref: github.String("feature"),
					SHA: github.String("head-sha"),
				},
			},
			branch: &github.Branch{
				Commit: &github.RepositoryCommit{
					SHA: github.String("base-sha"),
				},
			},
			checkStates: []string{"success"},
			commits: map[string]*github.Commit{
				"head-sha": {
					SHA:     github.String("head-sha"),
					Message: github.String("Test commit"),
					Parents: []github.Commit{{SHA: github.String("base-sha")}},
				},
			},
			dryRun:     true,
			force:      false,
			baseBranch: "main",
			directive:  "bretmckee-branch",
			wantErr:    false,
			wantMerged: false,
		},
		{
			name: "Failed checks without force",
			pr: &github.PullRequest{
				Number: github.Int(4),
				Merged: github.Bool(false),
				Base: &github.PullRequestBranch{
					Ref: github.String("main"),
					SHA: github.String("base-sha"),
				},
				Head: &github.PullRequestBranch{
					Ref: github.String("feature"),
					SHA: github.String("head-sha"),
				},
			},
			branch: &github.Branch{
				Commit: &github.RepositoryCommit{
					SHA: github.String("base-sha"),
				},
			},
			checkStates: []string{"failure"},
			force:       false,
			baseBranch:  "main",
			directive:   "bretmckee-branch",
			wantErr:     true,
			errMsg:      "cannot be submitted",
		},
		{
			name: "Failed checks with force proceeds",
			pr: &github.PullRequest{
				Number: github.Int(5),
				Merged: github.Bool(false),
				Body:   github.String("PR body"),
				Base: &github.PullRequestBranch{
					Ref: github.String("main"),
					SHA: github.String("base-sha"),
				},
				Head: &github.PullRequestBranch{
					Ref: github.String("feature"),
					SHA: github.String("head-sha"),
				},
			},
			branch: &github.Branch{
				Commit: &github.RepositoryCommit{
					SHA: github.String("base-sha"),
				},
			},
			checkStates: []string{"failure"},
			commits: map[string]*github.Commit{
				"head-sha": {
					SHA:     github.String("head-sha"),
					Message: github.String("Test commit"),
					Parents: []github.Commit{{SHA: github.String("base-sha")}},
				},
			},
			force:      true,
			baseBranch: "main",
			directive:  "bretmckee-branch",
			wantErr:    false,
			wantMerged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			merged := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" && contains(r.URL.Path, "/pulls/") {
					json.NewEncoder(w).Encode(tt.pr)
					return
				}
				if r.Method == "GET" && contains(r.URL.Path, "/branches/") {
					if tt.branch != nil {
						json.NewEncoder(w).Encode(tt.branch)
					} else {
						http.NotFound(w, r)
					}
					return
				}
				if contains(r.URL.Path, "/check-runs") {
					state := "success"
					if tt.checkStates != nil && len(tt.checkStates) > 0 {
						state = tt.checkStates[callCount]
						if callCount < len(tt.checkStates)-1 {
							callCount++
						}
					}
					results := &github.ListCheckRunsResults{
						Total: github.Int(1),
						CheckRuns: []*github.CheckRun{
							{
								Name:       github.String("test"),
								Status:     github.String(mapStateToCheckStatus(state)),
								Conclusion: github.String(mapStateToConclusion(state)),
							},
						},
					}
					json.NewEncoder(w).Encode(results)
					return
				}
				if r.Method == "GET" && contains(r.URL.Path, "/status") {
					json.NewEncoder(w).Encode(&github.CombinedStatus{
						State:      github.String("success"),
						TotalCount: github.Int(0),
					})
					return
				}
				if r.Method == "GET" && contains(r.URL.Path, "/git/commits/") {
					for sha, commit := range tt.commits {
						if contains(r.URL.Path, sha) {
							json.NewEncoder(w).Encode(commit)
							return
						}
					}
					http.NotFound(w, r)
					return
				}
				if r.Method == "PUT" && contains(r.URL.Path, "/merge") {
					merged = true
					result := &github.PullRequestMergeResult{
						Merged: github.Bool(true),
					}
					json.NewEncoder(w).Encode(result)
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			c, err := client.Create(server.URL, server.URL, "test", "test", "user", "token")
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			err = submitPR(c, tt.dryRun, tt.force, tt.baseBranch, tt.pr.GetNumber(), "squash", tt.directive)
			if (err != nil) != tt.wantErr {
				t.Errorf("submitPR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
				t.Errorf("submitPR() error = %q, want to contain %q", err.Error(), tt.errMsg)
				return
			}
			if merged != tt.wantMerged {
				t.Errorf("submitPR() merged = %v, want %v", merged, tt.wantMerged)
			}
		})
	}
}
