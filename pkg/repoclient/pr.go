package client

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/google/go-github/v28/github"
	"github.com/kr/pretty"
)

func (c *RESTClient) PullRequests() ([]*github.PullRequest, error) {
	o := &github.PullRequestListOptions{}
	prs, _, err := c.client.PullRequests.List(c.ctx, c.owner, c.repo, o)
	if err != nil {
		return nil, fmt.Errorf("Failed to list pull requests: %v", err)
	}
	return prs, nil
}

func (c *RESTClient) PullRequest(id int) (*github.PullRequest, error) {
	pr, _, err := c.client.PullRequests.Get(c.ctx, c.owner, c.repo, id)
	if err != nil {
		return nil, fmt.Errorf("Get of PR %d failed: %v", id, err)
	}
	glog.V(2).Infof("PR %d: %# v\n", id, pretty.Formatter(*pr))
	return pr, nil
}

func (c *RESTClient) MergePullRequest(id int, sha, msg string) (*github.PullRequest, error) {
	o := &github.PullRequestOptions{
		SHA:         sha,
		MergeMethod: "rebase",
	}
	res, resp, err := c.client.PullRequests.Merge(c.ctx, c.owner, c.repo, id, msg, o)
	if err != nil {
		return nil, fmt.Errorf("github merge of %d failed: %v", id, err)
	}
	glog.V(2).Infof("merge PR %d res: %# v\n", id, pretty.Formatter(*res))
	glog.V(2).Infof("merge PR %d resp: %# v\n", id, pretty.Formatter(*resp))
	pr, err := c.PullRequest(id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get pr after updating base for %d: %v", id, err)
	}
	return pr, nil
}

func (c *Connection) ChangePullRequestBase(id int, sha string) error {
	glog.Warningf("ChaneePullRequestBase returning unchanged pr %d", id)
	// TODO(bretmckee): Actually change the base.
	pr, err := c.GetPr(id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get pr after updating base for %d: %v", id, err)
	}
	return pr, nil
}
