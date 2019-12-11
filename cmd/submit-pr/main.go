package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/bretmckee/git-tools/pkg/repo/repodata"
	"github.com/golang/glog"
)

func submitPR(r *repodata.RepoData, dryRun, force bool, baseBranch string, number int) error {
	const retrySeconds = 60
	pr, ok := r.PrByNumber[number]
	if !ok {
		return fmt.Errorf("submitPR: %d was not found", number)
	}
	if pr.GetMerged() {
		glog.Warningf("PR %d is already merged.", number)
		return nil
	}
	if prRef := pr.GetBase().GetRef(); prRef != baseBranch {
		err := fmt.Errorf("pr base ref (%q) does not match base branch ref (%q):", prRef, baseBranch)
		if !force {
			return err
		}
		glog.Warningf("because force was specified, ignoring error %v", err)
	}
	bb, err := r.Branch(baseBranch)
	if err != nil {
		return fmt.Errorf("failed to get base branch %q: %v", baseBranch, err)
	}
	if prSHA, bbSHA := pr.GetBase().GetSHA(), bb.GetCommit().GetSHA(); prSHA != bbSHA {
		err := fmt.Errorf("pr base SHA (%q) does not match base branch SHA(%q):", prSHA, bbSHA)
		if !force {
			return err
		}
		glog.Warningf("because force was specified, ignoring error %v", err)
	}
	ref := pr.GetHead().GetRef()
	status, err := r.CombinedStatus(ref)
	if err != nil {
		return fmt.Errorf("submitPR: failed to get statuses:%v", err)
	}
	for status.GetState() == "pending" {
		// TODO(bretmckee): Consider an argument to terminate this loop after a
		// timeout.
		glog.Warning("pr %d status is pending: waiting %d seconds", number, retrySeconds)
		time.Sleep(time.Second * retrySeconds)
	}
	if state := status.GetState(); state == "failure" {
		err := fmt.Errorf("pr %d cannot be submitted because it has status %s", number, state)
		if !force {
			return err
		}
		glog.Warningf("because force was specified, ignoring error %v", err)
	}
	if dryRun {
		glog.Warningf("skipping submission of %d because a dry run was requested", number)
		return nil
	}
	if _, err := r.MergePullRequest(number, pr.GetBase().GetSHA(), ""); err != nil {
		return fmt.Errorf("failed to submit PR %d", number)
	}
	glog.Infof("Successfully submitted %d", number)
	return nil
}

func main() {
	var (
		baseBranch  = flag.String("base", "master", "Base branch")
		dryRun      = flag.Bool("dry-run", false, "Dry Run mode -- no pull requests will be created")
		force       = flag.Bool("force", false, "Submit even if not fully approved.")
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

	if err := submitPR(r, *dryRun, *force, *baseBranch, *pr); err != nil {
		glog.Exitf("submitPR failed: %v", err)
	}
}
