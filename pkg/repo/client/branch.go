package client

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/google/go-github/v28/github"
	"github.com/kr/pretty"
)

func (c *Client) Branches() ([]*github.Branch, error) {
	var branches []*github.Branch
	for thisPage, lastPage := 1, 1; thisPage <= lastPage; thisPage++ {
		glog.V(2).Infof("loading branches page %d", thisPage)
		o := &github.ListOptions{Page: thisPage}
		page, resp, err := c.client.Repositories.ListBranches(c.ctx, c.owner, c.repo, o)
		if err != nil {
			return nil, fmt.Errorf("Failed to list branches: %v", err)
		}
		for i, b := range page {
			glog.V(3).Infof("branch %d: %# v\n", i, pretty.Formatter(*b))
			branches = append(branches, b)
		}
		glog.V(3).Infof("resp=%# v\n", resp)
		lastPage = resp.LastPage
	}
	return branches, nil
}

func (c *Client) Branch(name string) (*github.Branch, error) {
	b, _, err := c.client.Repositories.GetBranch(c.ctx, c.owner, c.repo, name)
	if err != nil {
		return nil, fmt.Errorf("get of branch %q failed: %v", name, err)
	}
	if glog.V(3) {
		glog.Infof("Get of Branch %q: %# v\n", name, pretty.Formatter(*b))
	}
	return b, nil
}
