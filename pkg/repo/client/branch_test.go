package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-github/v28/github"
)

func TestEnsureRef(t *testing.T) {
	const (
		owner  = "test-owner"
		repo   = "test-repo"
		branch = "users/me/feature-1"
		sha    = "abc123def456"
	)

	tests := []struct {
		name           string
		updateStatus   int
		updateBody     string
		createStatus   int
		createBody     string
		wantErr        bool
		wantErrSubstr  string
		wantUpdateCall bool
		wantCreateCall bool
	}{
		{
			name:           "Update succeeds on first try",
			updateStatus:   http.StatusOK,
			updateBody:     `{"ref": "refs/heads/users/me/feature-1", "object": {"sha": "abc123def456"}}`,
			wantErr:        false,
			wantUpdateCall: true,
			wantCreateCall: false,
		},
		{
			name:           "Update fails (ref does not exist), create succeeds",
			updateStatus:   http.StatusUnprocessableEntity,
			updateBody:     `{"message": "Reference does not exist"}`,
			createStatus:   http.StatusCreated,
			createBody:     `{"ref": "refs/heads/users/me/feature-1", "object": {"sha": "abc123def456"}}`,
			wantErr:        false,
			wantUpdateCall: true,
			wantCreateCall: true,
		},
		{
			name:           "Both fail - error mentions both attempts",
			updateStatus:   http.StatusUnprocessableEntity,
			updateBody:     `{"message": "Reference does not exist"}`,
			createStatus:   http.StatusUnprocessableEntity,
			createBody:     `{"message": "Object does not exist"}`,
			wantErr:        true,
			wantErrSubstr:  "update err",
			wantUpdateCall: true,
			wantCreateCall: true,
		},
		{
			name:           "Update returns 404 (branch doesn't exist), create succeeds",
			updateStatus:   http.StatusNotFound,
			updateBody:     `{"message": "Not Found"}`,
			createStatus:   http.StatusCreated,
			createBody:     `{"ref": "refs/heads/users/me/feature-1", "object": {"sha": "abc123def456"}}`,
			wantErr:        false,
			wantUpdateCall: true,
			wantCreateCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var updateCalled, createCalled bool

			expectedRefPath := fmt.Sprintf("/repos/%s/%s/git/refs/heads/%s", owner, repo, branch)
			createPath := fmt.Sprintf("/repos/%s/%s/git/refs", owner, repo)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == "PATCH" && r.URL.Path == expectedRefPath:
					updateCalled = true
					w.WriteHeader(tt.updateStatus)
					fmt.Fprint(w, tt.updateBody)
				case r.Method == "POST" && r.URL.Path == createPath:
					createCalled = true
					w.WriteHeader(tt.createStatus)
					fmt.Fprint(w, tt.createBody)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()

			gh, err := github.NewEnterpriseClient(server.URL, server.URL, nil)
			if err != nil {
				t.Fatalf("failed to create test client: %v", err)
			}

			c := &Client{
				owner:  owner,
				repo:   repo,
				login:  "test-user",
				ctx:    context.Background(),
				client: gh,
			}

			err = c.EnsureRef(branch, sha)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tt.wantErrSubstr != "" && !strings.Contains(err.Error(), tt.wantErrSubstr) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.wantErrSubstr)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if updateCalled != tt.wantUpdateCall {
				t.Errorf("updateCalled = %v, want %v", updateCalled, tt.wantUpdateCall)
			}
			if createCalled != tt.wantCreateCall {
				t.Errorf("createCalled = %v, want %v", createCalled, tt.wantCreateCall)
			}
		})
	}
}
