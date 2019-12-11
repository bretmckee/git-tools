package client

import (
	"fmt"

	"github.com/google/go-github/v28/github"
)

func (c *Client) CombinedStatus(ref string) (*github.CombinedStatus, error) {
	o := &github.ListOptions{}
	s, _, err := c.client.Repositories.GetCombinedStatus(c.ctx, c.owner, c.repo, ref, o)
	if err != nil {
		return nil, fmt.Errorf("Failed to get statuses for %q: %v", ref, err)
	}
	return s, nil
}
