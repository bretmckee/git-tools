package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bretmckee/git-tools/pkg/repo/client"
	"github.com/bretmckee/git-tools/pkg/urls"
	"github.com/golang/glog"
)

func rebasePRs(upstream *client.Client, dryRun bool, number int) error {
	closedPR, err := upstream.PullRequest(number)
	if err != nil {
		return fmt.Errorf("PR %d could not be read: %v", number, err)
	}
	if !closedPR.GetMerged() {
		return fmt.Errorf("PR %d has not been merged", number)
	}
	ref := closedPR.GetHead().GetRef()
	newBase := closedPR.GetBase().GetRef()
	prs, err := upstream.PullRequests()
	if err != nil {
		return fmt.Errorf("unable to get pull requests: %v", err)
	}
	for _, pr := range prs {
		if pr.GetBase().GetRef() == ref {
			if dryRun {
				glog.Infof("PR %d matched branch %s, not changing base to %s because of dry run flag", pr.GetNumber(), ref, newBase)
				continue
			}
			glog.Infof("PR %d matched branch %s, changing base to %s", pr.GetNumber(), ref, newBase)
			if err := upstream.ChangePullRequestBase(pr.GetNumber(), newBase); err != nil {
				return fmt.Errorf("failed to change base: %v", err)
			}
		}
	}

	return nil
}

func main() {
	var (
		dryRun        = flag.Bool("dry-run", false, "Dry Run mode -- no pull requests will be created")
		baseURL       = flag.String("url", "", "GitHub Base URL")
		login         = flag.String("login", "", "Login of the user to submit for.")
		number        = flag.Int("pr", 0, "id of the closed pull request to rebase around")
		sourceOwner   = flag.String("source-owner", "", "Name of the owner (user or org) of the repo to create the commit in.")
		sourceRepo    = flag.String("source-repo", "", "Name of repo to create the commit in.")
		token         = flag.String("token", "", "github auth token to use (also checks environment GITHUB_TOKEN")
		uploadURL     = flag.String("upload", "", "GitHub Upload URL")
		upstreamOwner = flag.String("upstream-owner", "", "Owner of the upstream repo where PRs live. Defaults to --source-owner (same-repo mode).")
		upstreamRepo  = flag.String("upstream-repo", "", "Name of the upstream repo where PRs live. Defaults to --source-repo (same-repo mode).")
	)
	flag.Parse()
	if *token == "" {
		*token = os.Getenv("GITHUB_TOKEN")
	}
	if *token == "" {
		glog.Exit("Unauthorized: No token present")
	}
	if *sourceOwner == "" || *sourceRepo == "" || *login == "" {
		glog.Exitf("A non-empty value must be specified for the flags `-source-owner (=%q)`, `-source-repo (=%q)` and `-login (=%q)`", *sourceOwner, *sourceRepo, *login)
	}
	if *number <= 0 {
		glog.Exit("An positive integer value must be specified for `-pr`")
	}
	if *upstreamOwner == "" {
		*upstreamOwner = *sourceOwner
	}
	if *upstreamRepo == "" {
		*upstreamRepo = *sourceRepo
	}

	b, u, err := urls.Get(*baseURL, *uploadURL)
	if err != nil {
		glog.Exitf("failed to get URLs: %v", err)
	}

	_, upstream, err := client.CreatePair(b, u, *sourceOwner, *sourceRepo, *upstreamOwner, *upstreamRepo, *login, *token)
	if err != nil {
		glog.Exitf("failed to create clients: %v", err)
	}

	if err := rebasePRs(upstream, *dryRun, *number); err != nil {
		glog.Exitf("rebasePRs failed: %v", err)
	}
}
