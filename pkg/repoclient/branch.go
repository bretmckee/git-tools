package client

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/google/go-github/v28/github"
	"github.com/kr/pretty"
)

func (c *RESTClient) Branches() ([]*github.Branch, error) {
	o := &github.ListOptions{}
	branches, resp, err := c.client.Repositories.ListBranches(c.ctx, c.owner, c.repo, o)
	if err != nil {
		return nil, fmt.Errorf("Failed to list branches: %v", err)
	}
	if glog.V(2) {
		for i, b := range branches {
			glog.V(2).Infof("branch %d: %# v\n", i, pretty.Formatter(*b))
		}
		glog.V(2).Infof("resp=%# v\n", resp)
	}
	return branches, nil
}

func (c *RESTClient) Branch(name string) (*github.Branch, error) {
	b, _, err := c.client.Repositories.GetBranch(c.ctx, c.owner, c.repo, name)
	if err != nil {
		return nil, fmt.Errorf("get of branch %q failed: %v", name, err)
	}
	glog.V(2).Infof("Get of Branch %q: %# v\n", name, pretty.Formatter(*b))
	return b, nil
}
