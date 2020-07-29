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

func (c *Client) PullRequest(num int) (*github.PullRequest, error) {
	pr, _, err := c.client.PullRequests.Get(c.ctx, c.owner, c.repo, num)
	if err != nil {
		return nil, fmt.Errorf("Get of PR %d failed: %v", num, err)
	}
	if glog.V(3) {
		glog.Infof("PR %d: %# v\n", num, pretty.Formatter(*pr))
	}
	return pr, nil
}

func (c *Client) MergePullRequest(num int, sha, method, msg string) (*github.PullRequest, error) {
	o := &github.PullRequestOptions{
		SHA:         sha,
		MergeMethod: method,
	}
	if glog.V(3) {
		glog.Infof("merge PR %d o: %# v\n", num, pretty.Formatter(*o))
	}
	res, resp, err := c.client.PullRequests.Merge(c.ctx, c.owner, c.repo, num, msg, o)
	if err != nil {
		return nil, fmt.Errorf("github merge of %d failed: %v", num, err)
	}
	if glog.V(3) {
		glog.Infof("merge PR %d res: %# v\n", num, pretty.Formatter(*res))
		glog.Infof("merge PR %d resp: %# v\n", num, pretty.Formatter(*resp))
	}
	pr, err := c.PullRequest(num)
	if err != nil {
		return nil, fmt.Errorf("Failed to get pr after updating base for %d: %v", num, err)
	}
	return pr, nil
}

func (c *Client) CreatePullRequest(npr *github.NewPullRequest) (*github.PullRequest, error) {
	pr, _, err := c.client.PullRequests.Create(c.ctx, c.owner, c.repo, npr)
	if err != nil {
		return nil, fmt.Errorf("pull request create failed: %v", err)
	}
	return pr, nil
}

func (c *Client) ChangePullRequestBase(num int, ref string) error {
	pr, err := c.PullRequest(num)
	if err != nil {
		return fmt.Errorf("Failed to get pr after updating base for %d: %v", num, err)
	}
	pr.Base.Ref = github.String(ref)
	if _, _, err := c.client.PullRequests.Edit(c.ctx, c.owner, c.repo, num, pr); err != nil {
		return fmt.Errorf("Failed to change base for pr %d: %v", num, err)
	}
	return nil
}
