package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/bretmckee/git-tools/pkg/repo/client"
	"github.com/golang/glog"
	"github.com/google/go-github/v28/github"
	"github.com/kr/pretty"
)

const (
	maxCommitChainLength = 20
)

func submitMsg(c *client.Client, prBody string, first, last string) (string, error) {
	msg := ""
	l := 0
	glog.V(2).Infof("submitMsg begins first=%s, last=%s", first, last)
	for pos := first; pos != last; l++ {
		commit, err := c.Commit(pos)
		if err != nil {
			return "", fmt.Errorf("submitMsg: failed to retrieve commit: %v", err)
		}
		glog.V(2).Infof("submitMsg processes commit: %s", pretty.Sprintf("%# v", commit))
		if parents := len(commit.Parents); parents != 1 {
			return "", fmt.Errorf("submitMsg: commit %s has %d parents", pos, parents)
		}

		msg = "* " + commit.GetMessage() + "\n\n" + msg
		glog.V(2).Infof("after commit %s, msg=[%v]", pos, msg)
		pos = *commit.Parents[0].SHA
		if l >= maxCommitChainLength {
			return "", fmt.Errorf("submitMsg: max chain length (%d) exceeded", maxCommitChainLength)
		}
	}

	// If there are fewer than two commits, just use the pr body as the message.
	if l < 2 {
		return prBody, nil
	}
	return msg, nil
}

func submitPR(c *client.Client, dryRun, force bool, baseBranch string, number int, method string) error {
	const retrySeconds = 60
	pr, err := c.PullRequest(number)
	if err != nil {
		return fmt.Errorf("submitPR: failed to get %d: %v", number, err)
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
	bb, err := c.Branch(baseBranch)
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
	var status *github.CombinedStatus
	for {
		// TODO(bretmckee): Consider an argument to terminate this loop after a
		// timeout.
		status, err = c.CombinedStatus(ref)
		if err != nil {
			return fmt.Errorf("submitPR: failed to get combined status: %v", err)
		}
		if status.GetState() != "pending" {
			break
		}
		if force {
			glog.Warningf("bPR is pending, but not waiting because force was specified")
			break
		}
		glog.Warningf("pr %d status is pending: waiting %d seconds", number, retrySeconds)
		time.Sleep(time.Second * retrySeconds)
	}
	if state := status.GetState(); state == "failure" {
		err := fmt.Errorf("pr %d cannot be submitted because it has status %s", number, state)
		if !force {
			return err
		}
		glog.Warningf("because force was specified, ignoring error %v", err)
	}
	// TODO(bretmckee): Consider adding a way to specify a message.
	msg, err := submitMsg(c, *pr.Body, pr.GetHead().GetSHA(), pr.GetBase().GetSHA())
	if err != nil {
		return fmt.Errorf("submitPR failed to build summitMsg: %v", err)
	}
	if dryRun {
		glog.Warningf("skipping submission of %d because a dry run was requested", number)
		return nil
	}
	if _, err := c.MergePullRequest(number, pr.GetHead().GetSHA(), method, msg); err != nil {
		return fmt.Errorf("failed to submit PR %d: %v", number, err)
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
		method      = flag.String("method", "squash", "github merge method -- [merge|rebase|squash]")
		pr          = flag.Int("pr", 0, "id of the pull request to submit")
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
	if *pr <= 0 {
		glog.Exit("An positive integer value must be specified for `-pr`")
	}

	c := client.Create(*sourceOwner, *sourceRepo, *login, *token)
	if err := submitPR(c, *dryRun, *force, *baseBranch, *pr, *method); err != nil {
		glog.Exitf("submitPR failed: %v", err)
	}
}
