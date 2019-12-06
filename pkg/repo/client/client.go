package client

import (
	"context"

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

func Create(owner, repo, login, token string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		owner:  owner,
		repo:   repo,
		login:  login,
		ctx:    ctx,
		client: github.NewClient(tc),
	}
}
