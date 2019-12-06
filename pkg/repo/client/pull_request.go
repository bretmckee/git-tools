package client

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/google/go-github/v28/github"
	"github.com/kr/pretty"
)

func (c *Client) PullRequests() ([]*github.PullRequest, error) {
	o := &github.PullRequestListOptions{}
	prs, _, err := c.client.PullRequests.List(c.ctx, c.owner, c.repo, o)
	if err != nil {
		return nil, fmt.Errorf("Failed to list pull requests: %v", err)
	}
	return prs, nil
}

func (c *Client) PullRequest(id int) (*github.PullRequest, error) {
	pr, _, err := c.client.PullRequests.Get(c.ctx, c.owner, c.repo, id)
	if err != nil {
		return nil, fmt.Errorf("Get of PR %d failed: %v", id, err)
	}
	if glog.V(3) {
		glog.Infof("PR %d: %# v\n", id, pretty.Formatter(*pr))
	}
	return pr, nil
}

func (c *Client) MergePullRequest(id int, sha, msg string) (*github.PullRequest, error) {
	o := &github.PullRequestOptions{
		SHA:         sha,
		MergeMethod: "rebase",
	}
	res, resp, err := c.client.PullRequests.Merge(c.ctx, c.owner, c.repo, id, msg, o)
	if err != nil {
		return nil, fmt.Errorf("github merge of %d failed: %v", id, err)
	}
	if glog.V(3) {
		glog.Infof("merge PR %d res: %# v\n", id, pretty.Formatter(*res))
		glog.Infof("merge PR %d resp: %# v\n", id, pretty.Formatter(*resp))
	}
	pr, err := c.PullRequest(id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get pr after updating base for %d: %v", id, err)
	}
	return pr, nil
}

func (c *Client) ChangePullRequestBase(id int, sha string) error {
	glog.Warningf("ChangePullRequestBase returning unchanged pr %d", id)
	_, err := c.PullRequest(id)
	if err != nil {
		return fmt.Errorf("Failed to get pr after updating base for %d: %v", id, err)
	}
	// TODO(bretmckee): Actually change the base.
	return nil
}
