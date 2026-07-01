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

// EnsureRef tries UpdateRef first (the common path once the anchor exists),
// falling back to CreateRef when the branch is new. Both attempts are surfaced
// in the error so the caller can tell whether GitHub rejected the SHA (which
// suggests the fork-network object store is not shared with this repo, and
// the user needs to push the branch manually) or something else went wrong.
func (c *Client) EnsureRef(branchName, sha string) error {
	refName := "refs/heads/" + branchName
	ref := &github.Reference{
		Ref:    github.String(refName),
		Object: &github.GitObject{SHA: github.String(sha)},
	}

	_, _, updateErr := c.client.Git.UpdateRef(c.ctx, c.owner, c.repo, ref, true)
	if updateErr == nil {
		glog.V(2).Infof("EnsureRef updated %s in %s/%s to %s", refName, c.owner, c.repo, sha)
		return nil
	}

	_, _, createErr := c.client.Git.CreateRef(c.ctx, c.owner, c.repo, ref)
	if createErr == nil {
		glog.V(2).Infof("EnsureRef created %s in %s/%s at %s", refName, c.owner, c.repo, sha)
		return nil
	}

	return fmt.Errorf("EnsureRef: failed to update or create %s in %s/%s at %s (update err: %v; create err: %v)",
		refName, c.owner, c.repo, sha, updateErr, createErr)
}
