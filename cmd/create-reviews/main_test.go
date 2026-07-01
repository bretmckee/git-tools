package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bretmckee/git-tools/pkg/repo/client"
	"github.com/bretmckee/git-tools/pkg/repo/repodata"
	"github.com/google/go-github/v28/github"
)

// buildRepoData constructs a RepoData against a test server. When
// upstreamOwner/upstreamRepo match source, both fields point at the same
// underlying client so IsFork() returns false (matches production behaviour
// established in client.CreatePair).
func buildRepoData(t *testing.T, serverURL, sourceOwner, sourceRepo, upstreamOwner, upstreamRepo, login string) *repodata.RepoData {
	t.Helper()
	fork, upstream, err := client.CreatePair(serverURL, serverURL, sourceOwner, sourceRepo, upstreamOwner, upstreamRepo, login, "token")
	if err != nil {
		t.Fatalf("CreatePair failed: %v", err)
	}
	return &repodata.RepoData{
		Fork:     fork,
		Upstream: upstream,
	}
}

func TestCreatePR_HeadFormat(t *testing.T) {
	const (
		sourceOwner   = "myuser"
		sourceRepo    = "proj"
		upstreamOwner = "some-org"
		login         = "myuser"
		branch        = "users/myuser/feature-1"
		base          = "main"
		oldestSHA     = "abc123"
		commitMessage = "Add feature 1\n\nDetailed description here."
	)

	tests := []struct {
		name          string
		upstreamOwner string
		upstreamRepo  string
		wantHead      string
		wantPullPath  string
	}{
		{
			name:          "Same-repo mode uses bare branch name for Head",
			upstreamOwner: sourceOwner,
			upstreamRepo:  sourceRepo,
			wantHead:      branch,
			wantPullPath:  "/repos/" + sourceOwner + "/" + sourceRepo + "/pulls",
		},
		{
			name:          "Fork mode prefixes Head with login and POSTs to upstream",
			upstreamOwner: upstreamOwner,
			upstreamRepo:  sourceRepo,
			wantHead:      login + ":" + branch,
			wantPullPath:  "/repos/" + upstreamOwner + "/" + sourceRepo + "/pulls",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedHead, capturedBase, capturedPath string
			var capturedMethod string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/git/commits/"+oldestSHA) {
					msg := commitMessage
					json.NewEncoder(w).Encode(&github.Commit{
						SHA:     github.String(oldestSHA),
						Message: &msg,
					})
					return
				}
				if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/pulls") {
					body, _ := ioutil.ReadAll(r.Body)
					var npr github.NewPullRequest
					if err := json.Unmarshal(body, &npr); err != nil {
						t.Fatalf("failed to decode PR body: %v", err)
					}
					capturedHead = npr.GetHead()
					capturedBase = npr.GetBase()
					capturedPath = r.URL.Path
					capturedMethod = r.Method
					json.NewEncoder(w).Encode(&github.PullRequest{
						Number: github.Int(42),
					})
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			r := buildRepoData(t, server.URL, sourceOwner, sourceRepo, tt.upstreamOwner, tt.upstreamRepo, login)

			err := createPR(r, branch, base, oldestSHA, oldestSHA, false, false)
			if err != nil {
				t.Fatalf("createPR failed: %v", err)
			}

			if capturedMethod != "POST" {
				t.Errorf("PR create was not called; captured method = %q", capturedMethod)
			}
			if capturedPath != tt.wantPullPath {
				t.Errorf("PR POST path = %q, want %q", capturedPath, tt.wantPullPath)
			}
			if capturedHead != tt.wantHead {
				t.Errorf("Head = %q, want %q", capturedHead, tt.wantHead)
			}
			if capturedBase != base {
				t.Errorf("Base = %q, want %q (base branch name is unqualified in both modes)", capturedBase, base)
			}
		})
	}
}

func TestCreatePR_DryRunSkipsCreate(t *testing.T) {
	const (
		owner         = "myuser"
		repo          = "proj"
		login         = "myuser"
		oldestSHA     = "abc123"
		commitMessage = "Title\n\nBody"
	)

	var createCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/git/commits/"+oldestSHA) {
			msg := commitMessage
			json.NewEncoder(w).Encode(&github.Commit{
				SHA:     github.String(oldestSHA),
				Message: &msg,
			})
			return
		}
		if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/pulls") {
			createCalled = true
			json.NewEncoder(w).Encode(&github.PullRequest{Number: github.Int(1)})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	r := buildRepoData(t, server.URL, owner, repo, owner, repo, login)

	if err := createPR(r, "feature", "main", oldestSHA, oldestSHA, false, true); err != nil {
		t.Fatalf("createPR failed: %v", err)
	}

	if createCalled {
		t.Errorf("PR create should not be called in dry-run mode")
	}
}
