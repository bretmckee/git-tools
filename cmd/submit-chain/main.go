package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/google/go-github/v28/github"
	"github.com/kr/pretty"
	"golang.org/x/oauth2"
)

func (c *Connection) processChain(prs map[string]*github.PullRequest, chain Chain) error {
	glog.Infof("processChain begins for %v", chain)
	mergedPrevious := false
	for _, sha := range chain.SHAs {
		pr, ok := prs[sha]
		if !ok {
			return fmt.Errorf("Unable to locate pr for %q", sha)
		}
		id := int(*pr.Number)
		pr, err := c.GetPr(id)
		if err != nil {
			return fmt.Errorf("Unable to get pr %d for %q", id)
		}
		// If we merged the previous PR, we have to change the base of this one.
		if mergedPrevious {
			if prs[sha], err = c.updateBase(id); err != nil {
				return fmt.Errorf("Failed to update base for %d: %v", id, err)
			}
			mergedPrevious = false
			glog.Warning("Short circuit returning, Mergable=%v", prs[sha].Mergeable)
			return nil
		}
		if pr.Mergeable == nil {
			glog.Infof("processChain ends for %s which has nil Mergeable", sha)
			return nil
		}
		if !*pr.Mergeable {
			glog.Infof("processChain ends for %s because Mergeable is false", sha)
			return nil
		}
		glog.Infof("merging branch %d", id)
		if prs[sha], err = c.mergePR(id, sha, *pr.Body); err != nil {
			return fmt.Errorf("Unable to merge %d: %v", id, err)
		}
		mergedPrevious = true
	}
	return nil
}

func (c *Connection) process(prs []*github.PullRequest, login string) error {
	byId := make(map[string]*github.PullRequest)
	for _, pr := range prs {
		if *pr.User.Login == login {
			byId[*pr.Head.SHA] = pr
		}
	}

	chains, err := buildChains(byId)
	if err != nil {
		return fmt.Errorf("buildChains failed: %v", err)
	}

	for _, chain := range chains {
		if err := c.processChain(byId, chain); err != nil {
			return fmt.Errorf("process Chain failed: %v", err)
		}
	}

	return nil
}

func main() {
	var (
		sourceOwner = flag.String("source-owner", "", "Name of the owner (user or org) of the repo to create the commit in.")
		sourceRepo  = flag.String("source-repo", "", "Name of repo to create the commit in.")
		login       = flag.String("login", "", "Login of the user to submit for.")
	)
	flag.Parse()
	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token == "" {
		glog.Fatal("Unauthorized: No token present")
	}
	if *sourceOwner == "" || *sourceRepo == "" || *login == "" {
		glog.Fatal("You need to specify a non-empty value for the flags `-user`, `-source-owner` and `-source-repo`")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	c := newConnection(*sourceOwner, *sourceRepo, github.NewClient(tc))

	branches, err := c.ListBranches()
	if err != nil {
		glog.Fatalf("list branches failed: %v", err)
	}

	for i, b := range branches {
		glog.V(2).Infof("Branch %d: %# v\n", i, pretty.Formatter(*b))
	}
	return

	prs, err := c.ListPRs()
	if err != nil {
		glog.Fatalf("list pull requests failed: %v", err)
	}

	if err := c.process(prs, *login); err != nil {
		glog.Fatalf("process failed: %v", err)
	}
}
