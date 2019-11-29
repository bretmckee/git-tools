package client

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/google/go-github/v28/github"
	"github.com/kr/pretty"
)

func (c *RESTClient) Commit(sha string) (*github.Commit, error) {
	commit, _, err := c.client.Git.GetCommit(c.ctx, c.owner, c.repo, sha)
	if err != nil {
		return nil, fmt.Errorf("Get of commit %q failed: %v", sha, err)
	}
	glog.V(2).Infof("Commit %q: %# v\n", sha, pretty.Formatter(*commit))
	return commit, nil
}
