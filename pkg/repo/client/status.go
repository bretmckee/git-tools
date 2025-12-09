package client

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/google/go-github/v28/github"
	"github.com/kr/pretty"
)

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

// AggregatedStatus returns an overall status by checking both Check Runs and Combined Status APIs
func (c *Client) AggregatedStatus(ref string) (string, error) {
	checkRuns, err := c.CheckRunsForRef(ref, nil)
	if err != nil {
		return "", fmt.Errorf("Failed to get check runs: %v", err)
	}

	combinedStatus, err := c.CombinedStatus(ref)
	if err != nil {
		return "", fmt.Errorf("Failed to get combined status: %v", err)
	}

	hasCheckRuns := checkRuns != nil && checkRuns.Total != nil && *checkRuns.Total > 0
	hasStatuses := combinedStatus != nil && combinedStatus.TotalCount != nil && *combinedStatus.TotalCount > 0

	if glog.V(2) {
		glog.Infof("AggregatedStatus for %q: check_runs=%d, statuses=%d", ref, 
			func() int { if hasCheckRuns { return *checkRuns.Total }; return 0 }(),
			func() int { if hasStatuses { return *combinedStatus.TotalCount }; return 0 }())
	}

	if !hasCheckRuns && !hasStatuses {
		if glog.V(2) {
			glog.Infof("No checks configured for %q, returning success", ref)
		}
		return "success", nil
	}

	if hasCheckRuns {
		for _, run := range checkRuns.CheckRuns {
			status := run.GetStatus()
			if status == "queued" || status == "in_progress" || status == "waiting" || status == "requested" || status == "pending" {
				if glog.V(2) {
					glog.Infof("Check run %q has status %q, overall status is pending", run.GetName(), status)
				}
				return "pending", nil
			}
		}
	}

	if hasStatuses {
		if combinedStatus.GetState() == "pending" {
			if glog.V(2) {
				glog.Infof("Combined status is pending")
			}
			return "pending", nil
		}
	}

	if hasCheckRuns {
		for _, run := range checkRuns.CheckRuns {
			if run.GetStatus() == "completed" {
				conclusion := run.GetConclusion()
				if conclusion == "failure" || conclusion == "timed_out" || conclusion == "action_required" {
					if glog.V(1) {
						glog.Warningf("Check run %q failed with conclusion %q", run.GetName(), conclusion)
					}
					return "failure", nil
				}
			}
		}
	}

	if hasStatuses {
		state := combinedStatus.GetState()
		if state == "failure" || state == "error" {
			if glog.V(1) {
				glog.Warningf("Combined status has state %q", state)
			}
			return "failure", nil
		}
	}

	if glog.V(2) {
		glog.Infof("All checks passed for %q", ref)
	}
	return "success", nil
}
