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

func TestAggregatedStatus(t *testing.T) {
	tests := []struct {
		name              string
		checkRunsResponse string
		statusResponse    string
		checkRunsCode     int
		statusCode        int
		wantState         string
		wantErr           bool
	}{
		{
			name: "No checks configured returns success",
			checkRunsResponse: `{
				"total_count": 0,
				"check_runs": []
			}`,
			statusResponse: `{
				"state": "success",
				"total_count": 0,
				"statuses": []
			}`,
			checkRunsCode: http.StatusOK,
			statusCode:    http.StatusOK,
			wantState:     "success",
			wantErr:       false,
		},
		{
			name: "Check run queued returns pending",
			checkRunsResponse: `{
				"total_count": 1,
				"check_runs": [
					{
						"id": 1,
						"status": "queued",
						"name": "test"
					}
				]
			}`,
			statusResponse: `{
				"state": "success",
				"total_count": 0,
				"statuses": []
			}`,
			checkRunsCode: http.StatusOK,
			statusCode:    http.StatusOK,
			wantState:     "pending",
			wantErr:       false,
		},
		{
			name: "Check run in_progress returns pending",
			checkRunsResponse: `{
				"total_count": 1,
				"check_runs": [
					{
						"id": 1,
						"status": "in_progress",
						"name": "test"
					}
				]
			}`,
			statusResponse: `{
				"state": "success",
				"total_count": 0,
				"statuses": []
			}`,
			checkRunsCode: http.StatusOK,
			statusCode:    http.StatusOK,
			wantState:     "pending",
			wantErr:       false,
		},
		{
			name: "Combined status pending returns pending",
			checkRunsResponse: `{
				"total_count": 0,
				"check_runs": []
			}`,
			statusResponse: `{
				"state": "pending",
				"total_count": 1,
				"statuses": [
					{
						"state": "pending",
						"context": "ci/test"
					}
				]
			}`,
			checkRunsCode: http.StatusOK,
			statusCode:    http.StatusOK,
			wantState:     "pending",
			wantErr:       false,
		},
		{
			name: "Check run failure returns failure",
			checkRunsResponse: `{
				"total_count": 1,
				"check_runs": [
					{
						"id": 1,
						"status": "completed",
						"conclusion": "failure",
						"name": "test"
					}
				]
			}`,
			statusResponse: `{
				"state": "success",
				"total_count": 0,
				"statuses": []
			}`,
			checkRunsCode: http.StatusOK,
			statusCode:    http.StatusOK,
			wantState:     "failure",
			wantErr:       false,
		},
		{
			name: "Check run timed_out returns failure",
			checkRunsResponse: `{
				"total_count": 1,
				"check_runs": [
					{
						"id": 1,
						"status": "completed",
						"conclusion": "timed_out",
						"name": "test"
					}
				]
			}`,
			statusResponse: `{
				"state": "success",
				"total_count": 0,
				"statuses": []
			}`,
			checkRunsCode: http.StatusOK,
			statusCode:    http.StatusOK,
			wantState:     "failure",
			wantErr:       false,
		},
		{
			name: "Check run action_required returns failure",
			checkRunsResponse: `{
				"total_count": 1,
				"check_runs": [
					{
						"id": 1,
						"status": "completed",
						"conclusion": "action_required",
						"name": "test"
					}
				]
			}`,
			statusResponse: `{
				"state": "success",
				"total_count": 0,
				"statuses": []
			}`,
			checkRunsCode: http.StatusOK,
			statusCode:    http.StatusOK,
			wantState:     "failure",
			wantErr:       false,
		},
		{
			name: "Combined status failure returns failure",
			checkRunsResponse: `{
				"total_count": 0,
				"check_runs": []
			}`,
			statusResponse: `{
				"state": "failure",
				"total_count": 1,
				"statuses": [
					{
						"state": "failure",
						"context": "ci/test"
					}
				]
			}`,
			checkRunsCode: http.StatusOK,
			statusCode:    http.StatusOK,
			wantState:     "failure",
			wantErr:       false,
		},
		{
			name: "Combined status error returns failure",
			checkRunsResponse: `{
				"total_count": 0,
				"check_runs": []
			}`,
			statusResponse: `{
				"state": "error",
				"total_count": 1,
				"statuses": [
					{
						"state": "error",
						"context": "ci/test"
					}
				]
			}`,
			checkRunsCode: http.StatusOK,
			statusCode:    http.StatusOK,
			wantState:     "failure",
			wantErr:       false,
		},
		{
			name: "All checks successful returns success",
			checkRunsResponse: `{
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
			statusResponse: `{
				"state": "success",
				"total_count": 1,
				"statuses": [
					{
						"state": "success",
						"context": "ci/legacy"
					}
				]
			}`,
			checkRunsCode: http.StatusOK,
			statusCode:    http.StatusOK,
			wantState:     "success",
			wantErr:       false,
		},
		{
			name: "Check run neutral does not block",
			checkRunsResponse: `{
				"total_count": 1,
				"check_runs": [
					{
						"id": 1,
						"status": "completed",
						"conclusion": "neutral",
						"name": "optional-check"
					}
				]
			}`,
			statusResponse: `{
				"state": "success",
				"total_count": 0,
				"statuses": []
			}`,
			checkRunsCode: http.StatusOK,
			statusCode:    http.StatusOK,
			wantState:     "success",
			wantErr:       false,
		},
		{
			name: "Check run skipped does not block",
			checkRunsResponse: `{
				"total_count": 1,
				"check_runs": [
					{
						"id": 1,
						"status": "completed",
						"conclusion": "skipped",
						"name": "conditional-check"
					}
				]
			}`,
			statusResponse: `{
				"state": "success",
				"total_count": 0,
				"statuses": []
			}`,
			checkRunsCode: http.StatusOK,
			statusCode:    http.StatusOK,
			wantState:     "success",
			wantErr:       false,
		},
		{
			name: "Mixed success with pending returns pending",
			checkRunsResponse: `{
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
						"status": "in_progress",
						"name": "build"
					}
				]
			}`,
			statusResponse: `{
				"state": "success",
				"total_count": 0,
				"statuses": []
			}`,
			checkRunsCode: http.StatusOK,
			statusCode:    http.StatusOK,
			wantState:     "pending",
			wantErr:       false,
		},
		{
			name: "Mixed success with failure returns failure",
			checkRunsResponse: `{
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
						"conclusion": "failure",
						"name": "build"
					}
				]
			}`,
			statusResponse: `{
				"state": "success",
				"total_count": 0,
				"statuses": []
			}`,
			checkRunsCode: http.StatusOK,
			statusCode:    http.StatusOK,
			wantState:     "failure",
			wantErr:       false,
		},
		{
			name:              "Check runs API error",
			checkRunsResponse: `{"message": "Not Found"}`,
			statusResponse: `{
				"state": "success",
				"total_count": 0,
				"statuses": []
			}`,
			checkRunsCode: http.StatusNotFound,
			statusCode:    http.StatusOK,
			wantState:     "",
			wantErr:       true,
		},
		{
			name: "Combined status API error",
			checkRunsResponse: `{
				"total_count": 0,
				"check_runs": []
			}`,
			statusResponse:    `{"message": "Internal Server Error"}`,
			checkRunsCode:     http.StatusOK,
			statusCode:        http.StatusInternalServerError,
			wantState:         "",
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/repos/test-owner/test-repo/commits/test-ref/check-runs" {
					w.WriteHeader(tt.checkRunsCode)
					fmt.Fprint(w, tt.checkRunsResponse)
				} else if r.URL.Path == "/repos/test-owner/test-repo/commits/test-ref/status" {
					w.WriteHeader(tt.statusCode)
					fmt.Fprint(w, tt.statusResponse)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
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

			state, _, err := c.AggregatedStatus("test-ref")

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

			if state != tt.wantState {
				t.Errorf("expected state %q, got %q", tt.wantState, state)
			}
		})
	}
}

func TestAggregatedStatusHasChecks(t *testing.T) {
	tests := []struct {
		name              string
		checkRunsResponse string
		statusResponse    string
		wantHasChecks     bool
	}{
		{
			name:              "No check runs and no statuses",
			checkRunsResponse: `{"total_count": 0, "check_runs": []}`,
			statusResponse:    `{"state": "success", "total_count": 0, "statuses": []}`,
			wantHasChecks:     false,
		},
		{
			name: "Has check runs only",
			checkRunsResponse: `{
				"total_count": 1,
				"check_runs": [{"id": 1, "status": "completed", "conclusion": "success", "name": "test"}]
			}`,
			statusResponse: `{"state": "success", "total_count": 0, "statuses": []}`,
			wantHasChecks:  true,
		},
		{
			name:              "Has combined status only",
			checkRunsResponse: `{"total_count": 0, "check_runs": []}`,
			statusResponse: `{
				"state": "success",
				"total_count": 1,
				"statuses": [{"state": "success", "context": "ci/legacy"}]
			}`,
			wantHasChecks: true,
		},
		{
			name: "Has both check runs and combined status",
			checkRunsResponse: `{
				"total_count": 1,
				"check_runs": [{"id": 1, "status": "completed", "conclusion": "success", "name": "test"}]
			}`,
			statusResponse: `{
				"state": "success",
				"total_count": 1,
				"statuses": [{"state": "success", "context": "ci/legacy"}]
			}`,
			wantHasChecks: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/repos/test-owner/test-repo/commits/test-ref/check-runs" {
					fmt.Fprint(w, tt.checkRunsResponse)
				} else if r.URL.Path == "/repos/test-owner/test-repo/commits/test-ref/status" {
					fmt.Fprint(w, tt.statusResponse)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			gh, err := github.NewEnterpriseClient(server.URL, server.URL, nil)
			if err != nil {
				t.Fatalf("failed to create test client: %v", err)
			}

			c := &Client{
				owner:  "test-owner",
				repo:   "test-repo",
				login:  "test-user",
				ctx:    context.Background(),
				client: gh,
			}

			_, hasChecks, err := c.AggregatedStatus("test-ref")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if hasChecks != tt.wantHasChecks {
				t.Errorf("hasChecks = %v, want %v", hasChecks, tt.wantHasChecks)
			}
		})
	}
}
