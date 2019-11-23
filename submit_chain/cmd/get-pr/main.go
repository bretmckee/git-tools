package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v28/github"
	"github.com/kr/pretty"
	"golang.org/x/oauth2"
)

func main() {
	var (
		sourceOwner = flag.String("source-owner", "", "Name of the owner (user or org) of the repo to create the commit in.")
		sourceRepo  = flag.String("source-repo", "", "Name of repo to create the commit in.")
		login       = flag.String("login", "", "Login of the user to submit for.")
	)
	flag.Parse()
	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token == "" {
		log.Fatal("Unauthorized: No token present")
	}
	if *sourceOwner == "" || *sourceRepo == "" || *login == "" {
		log.Fatal("You need to specify a non-empty value for the flags `-user`, `-source-owner` and `-source-repo`")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	for _, id := range []int{49} {
		pr, _, err := client.PullRequests.Get(ctx, *sourceOwner, *sourceRepo, id)
		if err != nil {
			log.Fatal("Get failed: %v", err)
		}
		fmt.Printf("%# v\n", pretty.Formatter(*pr))
	}
}
