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
