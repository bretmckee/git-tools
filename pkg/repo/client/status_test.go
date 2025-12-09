package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v28/github"
)

func TestCheckRunsForRef(t *testing.T) {
	tests := []struct {
		name           string
		ref            string
		opts           *github.ListCheckRunsOptions
		mockResponse   string
		mockStatusCode int
		wantErr        bool
		wantCount      int
		wantFilter     string
	}{
		{
			name:           "Successful request with nil options",
			ref:            "main",
			opts:           nil,
			mockStatusCode: http.StatusOK,
			mockResponse: `{
				"total_count": 2,
				"check_runs": [
					{
						"id": 1,
						"status": "completed",
						"conclusion": "success",
						"name": "test"
					},
					{
						"id": 2,
						"status": "completed",
						"conclusion": "success",
						"name": "build"
					}
				]
			}`,
			wantErr:    false,
			wantCount:  2,
			wantFilter: "latest",
		},
		{
			name: "Successful request with custom options",
			ref:  "feature-branch",
			opts: &github.ListCheckRunsOptions{
				Filter: github.String("all"),
			},
			mockStatusCode: http.StatusOK,
			mockResponse: `{
				"total_count": 1,
				"check_runs": [
					{
						"id": 3,
						"status": "in_progress",
						"name": "lint"
					}
				]
			}`,
			wantErr:    false,
			wantCount:  1,
			wantFilter: "all",
		},
		{
			name:           "Empty check runs",
			ref:            "empty-branch",
			opts:           nil,
			mockStatusCode: http.StatusOK,
			mockResponse: `{
				"total_count": 0,
				"check_runs": []
			}`,
			wantErr:    false,
			wantCount:  0,
			wantFilter: "latest",
		},
		{
			name:           "API error response",
			ref:            "invalid-ref",
			opts:           nil,
			mockStatusCode: http.StatusNotFound,
			mockResponse:   `{"message": "Not Found"}`,
			wantErr:        true,
			wantCount:      0,
			wantFilter:     "latest",
		},
		{
			name:           "Server error",
			ref:            "main",
			opts:           nil,
			mockStatusCode: http.StatusInternalServerError,
			mockResponse:   `{"message": "Internal Server Error"}`,
			wantErr:        true,
			wantCount:      0,
			wantFilter:     "latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.opts != nil && tt.opts.Filter != nil {
					filter := r.URL.Query().Get("filter")
					if filter != *tt.opts.Filter {
						t.Errorf("expected filter %q, got %q", *tt.opts.Filter, filter)
					}
				} else {
					filter := r.URL.Query().Get("filter")
					if filter != tt.wantFilter {
						t.Errorf("expected default filter %q, got %q", tt.wantFilter, filter)
					}
				}

				w.WriteHeader(tt.mockStatusCode)
				fmt.Fprint(w, tt.mockResponse)
			}))
			defer server.Close()

			client, err := github.NewEnterpriseClient(server.URL, server.URL, nil)
			if err != nil {
				t.Fatalf("failed to create test client: %v", err)
			}

			c := &Client{
				owner:  "test-owner",
				repo:   "test-repo",
				login:  "test-user",
				ctx:    context.Background(),
				client: client,
			}

			results, err := c.CheckRunsForRef(tt.ref, tt.opts)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if results == nil {
				t.Fatal("expected non-nil results")
			}

			if len(results.CheckRuns) != tt.wantCount {
				t.Errorf("expected %d check runs, got %d", tt.wantCount, len(results.CheckRuns))
			}

			if *results.Total != tt.wantCount {
				t.Errorf("expected total count %d, got %d", tt.wantCount, *results.Total)
			}
		})
	}
}

func TestCheckRunsForRef_DefaultOptions(t *testing.T) {
	filterCalled := false
	expectedFilter := "latest"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter")
		if filter == expectedFilter {
			filterCalled = true
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total_count": 0, "check_runs": []}`)
	}))
	defer server.Close()

	client, err := github.NewEnterpriseClient(server.URL, server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	c := &Client{
		owner:  "test-owner",
		repo:   "test-repo",
		login:  "test-user",
		ctx:    context.Background(),
		client: client,
	}

	_, err = c.CheckRunsForRef("main", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !filterCalled {
		t.Errorf("expected default filter %q to be applied", expectedFilter)
	}
}
