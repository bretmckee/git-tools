package client

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/google/go-github/v28/github"
	"github.com/kr/pretty"
)

func (c *Client) Commit(sha string) (*github.Commit, error) {
	commit, _, err := c.client.Git.GetCommit(c.ctx, c.owner, c.repo, sha)
	if err != nil {
		return nil, fmt.Errorf("Get of commit %q failed: %v", sha, err)
	}
	if glog.V(3) {
		glog.Infof("Commit %q: %# v\n", sha, pretty.Formatter(*commit))
	}
	return commit, nil
}
