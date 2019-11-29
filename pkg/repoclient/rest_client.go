package client

import (
	"context"

	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
)

type RESTClient struct {
	owner  string
	repo   string
	login  string
	ctx    context.Context
	client *github.Client
}

func Create(owner, repo, login, token string) *RESTClient {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	return &RESTClient{
		owner:  owner,
		repo:   repo,
		login:  login,
		ctx:    ctx,
		client: github.NewClient(tc),
	}
}
