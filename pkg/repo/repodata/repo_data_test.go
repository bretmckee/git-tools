package repodata

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bretmckee/git-tools/pkg/repo/client"
)

func TestSyncAnchor_SameRepoIsNoOp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected API call in same-repo mode: %s %s", r.Method, r.URL.Path)
		http.NotFound(w, r)
	}))
	defer server.Close()

	fork, upstream, err := client.CreatePair(server.URL, server.URL, "owner", "repo", "owner", "repo", "user", "token")
	if err != nil {
		t.Fatalf("CreatePair failed: %v", err)
	}
	if fork != upstream {
		t.Fatalf("expected same-pointer clients in same-repo mode")
	}

	r := &RepoData{Fork: fork, Upstream: upstream}

	if err := r.SyncAnchor("some-branch", "abc123"); err != nil {
		t.Errorf("SyncAnchor same-repo mode returned error: %v", err)
	}
}

func TestSyncAnchor_ForkModeCallsUpstreamEnsureRef(t *testing.T) {
	const (
		forkOwner     = "myuser"
		upstreamOwner = "some-org"
		repoName      = "proj"
		branchName    = "users/myuser/feature-1"
		sha           = "deadbeefcafe"
	)

	var patchCalled bool
	var patchOwner, patchRepo, patchRef string

	expectedPatchPath := fmt.Sprintf("/repos/%s/%s/git/refs/heads/%s", upstreamOwner, repoName, branchName)
	forkPathPrefix := fmt.Sprintf("/repos/%s/%s/", forkOwner, repoName)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= len(forkPathPrefix) && r.URL.Path[:len(forkPathPrefix)] == forkPathPrefix {
			t.Errorf("SyncAnchor must not touch the fork; got %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.Method == "PATCH" && r.URL.Path == expectedPatchPath {
			patchCalled = true
			patchOwner = upstreamOwner
			patchRepo = repoName
			patchRef = branchName
			fmt.Fprintf(w, `{"ref": "refs/heads/%s", "object": {"sha": %q}}`, branchName, sha)
			return
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		http.NotFound(w, r)
	}))
	defer server.Close()

	fork, upstream, err := client.CreatePair(server.URL, server.URL, forkOwner, repoName, upstreamOwner, repoName, "user", "token")
	if err != nil {
		t.Fatalf("CreatePair failed: %v", err)
	}
	if fork == upstream {
		t.Fatalf("expected distinct clients in fork mode")
	}

	r := &RepoData{Fork: fork, Upstream: upstream}

	if err := r.SyncAnchor(branchName, sha); err != nil {
		t.Fatalf("SyncAnchor fork mode returned error: %v", err)
	}

	if !patchCalled {
		t.Fatalf("expected PATCH to upstream refs endpoint, none seen")
	}
	if patchOwner != upstreamOwner || patchRepo != repoName || patchRef != branchName {
		t.Errorf("PATCH targeted %s/%s ref %s, want %s/%s ref %s", patchOwner, patchRepo, patchRef, upstreamOwner, repoName, branchName)
	}
}
