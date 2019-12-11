package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bretmckee/git-tools/pkg/repo/repodata"
	"github.com/golang/glog"
)

func rebasePRs(r *repodata.RepoData, dryRun bool, number int) error {
	closedPR, ok := r.PrByNumber[number]
	if !ok {
		return fmt.Errorf("PR %d was not found", number)
	}
	if !*closedPR.Merged {
		return fmt.Errorf("PR %d has not been merged", number)
	}
	ref := *closedPR.Head.Ref
	newBase := *closedPR.Base.Ref
	for _, pr := range r.PrByNumber {
		if *pr.Base.Ref == ref {
			glog.Infof("PR %d matched branch %s, changing base to %s", *pr.Number, ref, newBase)
			if dryRun {
				continue
			}
			if err := r.ChangePullRequestBase(number, newBase); err != nil {
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
		pr          = flag.Int("pr", 0, "id of the closed pull request to rebase around")
		sourceOwner = flag.String("source-owner", "", "Name of the owner (user or org) of the repo to create the commit in.")
		sourceRepo  = flag.String("source-repo", "", "Name of repo to create the commit in.")
		token       = flag.String("token", "", "github auth token to use (also checks environment GITHUB_AUTH_TOKEN")
	)
	flag.Parse()
	if *token == "" {
		*token = os.Getenv("GITHUB_AUTH_TOKEN")
	}
	if *token == "" {
		glog.Exit("Unauthorized: No token present")
	}
	if *sourceOwner == "" || *sourceRepo == "" || *login == "" {
		glog.Exitf("A non-empty value must be specified for the flags `-source-owner (=%q)`, `-source-repo (=%q)` and `-login (=%q)`", *sourceOwner, *sourceRepo, *login)
	}
	if *pr <= 0 {
		glog.Exit("An positive integer value must be specified for `-pr`")
	}

	r, err := repodata.Create(*sourceOwner, *sourceRepo, *login, *token)
	if err != nil {
		glog.Exitf("failed to create repodata: %v", err)
	}

	if err := rebasePRs(r, *dryRun, *pr); err != nil {
		glog.Exitf("rebasePRs failed: %v", err)
	}
}
