package client

import (
	"context"
	"fmt"

	"github.com/bretmckee/git-tools/pkg/repo"
	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
)

type Client struct {
	owner  string
	repo   string
	login  string
	ctx    context.Context
	client *github.Client
}

var _ repo.Repo = (*Client)(nil)

func Create(baseURL, uploadURL, owner, repo, login, token string) (*Client, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	client, err := github.NewEnterpriseClient(baseURL, uploadURL, tc)
	if err != nil {
		return nil, fmt.Errorf("failed to create github client: %v", err)
	}

	return &Client{
		owner:  owner,
		repo:   repo,
		login:  login,
		ctx:    ctx,
		client: client,
	}, nil
}

func (c *Client) Owner() string {
	return c.owner
}

func (c *Client) Repo() string {
	return c.repo
}

// When fork and upstream refer to the same owner/repo, both returned pointers
// refer to the same underlying Client so callers can detect same-repo mode via
// pointer equality (fork == upstream).
func CreatePair(baseURL, uploadURL, sourceOwner, sourceRepo, upstreamOwner, upstreamRepo, login, token string) (fork *Client, upstream *Client, err error) {
	fork, err = Create(baseURL, uploadURL, sourceOwner, sourceRepo, login, token)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create fork client: %v", err)
	}
	if sourceOwner == upstreamOwner && sourceRepo == upstreamRepo {
		return fork, fork, nil
	}
	upstream, err = Create(baseURL, uploadURL, upstreamOwner, upstreamRepo, login, token)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create upstream client: %v", err)
	}
	return fork, upstream, nil
}
