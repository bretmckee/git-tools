package main

import (
	"bufio"
	"context"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v28/github"
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

	listOpts := &github.PullRequestListOptions{}
	prs, resp, err := client.PullRequests.List(ctx, *sourceOwner, *sourceRepo, listOpts)
	if err != nil {
		log.Fatalf("Failed to list pull requests: %v", err)
	}

	w := bufio.NewWriter(os.Stdout)
	enc := gob.NewEncoder(w)
	if err := enc.Encode(prs); err != nil {
		log.Fatalf("failed to encode prs: %v", err)
	}

	w.Flush()

	log.Printf("len(prs)=%d", len(prs))
	for _, pr := range prs {
		log.Printf("pr=%v\n\n", pr)
	}
	//var filtered []*github.PullRequest
	//for _, pr := range prs {
	//		if *pr.User.Login == *login {
	//			filtered = append(filtered, pr)
	//		}
	//	}
	//	for _, pr := range filtered {
	//		fmt.Printf("pr=%v\n\n", pr)
	//	}
	fmt.Printf("resp=%+v\n", resp)
}
