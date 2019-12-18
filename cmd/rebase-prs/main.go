package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bretmckee/git-tools/pkg/repo/client"
	"github.com/golang/glog"
)

func rebasePRs(c *client.Client, dryRun bool, number int) error {
	closedPR, err := c.PullRequest(number)
	if err != nil {
		return fmt.Errorf("PR %d could not be read: %v", number, err)
	}
	if !closedPR.GetMerged() {
		return fmt.Errorf("PR %d has not been merged", number)
	}
	ref := closedPR.GetHead().GetRef()
	newBase := closedPR.GetBase().GetRef()
	prs, err := c.PullRequests()
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
			if err := c.ChangePullRequestBase(pr.GetNumber(), newBase); err != nil {
				return fmt.Errorf("failed to change base: %v", err)
			}
		}
	}

	return nil
}

func main() {
	var (
		dryRun      = flag.Bool("dry-run", false, "Dry Run mode -- no pull requests will be created")
		login       = flag.String("login", "", "Login of the user to submit for.")
		number      = flag.Int("pr", 0, "id of the closed pull request to rebase around")
		sourceOwner = flag.String("source-owner", "", "Name of the owner (user or org) of the repo to create the commit in.")
		sourceRepo  = flag.String("source-repo", "", "Name of repo to create the commit in.")
		token       = flag.String("token", "", "github auth token to use (also checks environment GITHUB_TOKEN")
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

	c := client.Create(*sourceOwner, *sourceRepo, *login, *token)

	if err := rebasePRs(c, *dryRun, *number); err != nil {
		glog.Exitf("rebasePRs failed: %v", err)
	}
}
