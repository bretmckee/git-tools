package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/bretmckee/git-tools/pkg/repo/repodata"
	"github.com/golang/glog"
	"github.com/kr/pretty"
)

func createPR(r *repodata.RepoData, branch, base, oldest, newest string) error {
	glog.Infof("Creating PR for branch %s based on %s, oldest=%s, newest=%s", branch, base, oldest, newest)
	o, err := r.Commit(oldest)
	if err != nil {
		return fmt.Errorf("CreatePR failed to get oldest commit %s: %v", oldest, err)
	}
	glog.Infof("Oldest Commit: %# v\n", pretty.Formatter(*o))
	values := regexp.MustCompile("[\n]+").Split(*o.Message, 2)
	if err := r.CreatePullRequest(values[0], branch, base, values[1], false, true); err != nil {
		return fmt.Errorf("createPR failed to pr for %s:%v", branch, err)
	}
	//glog.Info("title=%q body=%q", title, body)
	return nil
}

// createPRs createa any needed Pull Requests for commits in the range
// baseBranch...tipBranch.
func createPRs(r *repodata.RepoData, tipBranch, baseBranch string) error {
	b, err := r.Branch(tipBranch)
	if err != nil {
		return fmt.Errorf("failed to get tip branch %q: %v", tipBranch, err)
	}
	if b.Commit.SHA == nil {
		return fmt.Errorf("Branch Commit SHA is nil: %# v\n", pretty.Formatter(*b))
	}
	glog.V(2).Infof("Branch %# v\n", pretty.Formatter(*b))

	bb, err := r.Branch(baseBranch)
	if err != nil {
		return fmt.Errorf("failed to get base branch %q: %v", baseBranch, err)
	}
	if bb.Commit.SHA == nil {
		return fmt.Errorf("Branch Commit SHA is nil: %# v\n", pretty.Formatter(*bb))
	}
	glog.V(2).Infof("Base Branch %# v\n", pretty.Formatter(*bb))
	if err := r.LoadData(); err != nil {
		return fmt.Errorf("failed to load data: %v", err)
	}

	chain, err := r.CommitChain(*b.Commit.SHA, *bb.Commit.SHA)
	if err != nil {
		return fmt.Errorf("Get commit chain failed: %v", err)
	}
	prev := ""
	base := baseBranch
	for _, commit := range chain {
		if prev == "" {
			prev = commit
		}
		branch, ok := r.BranchBySHA[commit]
		if !ok {
			// This commit does not represent a branch
			continue
		}
		if err := createPR(r, *branch.Name, base, prev, commit); err != nil {
			return fmt.Errorf("failed to create pr: %v", err)
		}
		prev = ""
		base = *branch.Name
	}

	return nil
}

func main() {
	var (
		baseBranch  = flag.String("base", "master", "Base Branch")
		branch      = flag.String("branch", "", "Starting Branch")
		login       = flag.String("login", "", "Login of the user to submit for.")
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
		glog.Exitf("You need to specify a non-empty value for the flags `-source-owner (=%q)`, `-source-repo (=%q)` and `-login (=%q)`", *sourceOwner, *sourceRepo, *login)
	}
	if *branch == "" || *baseBranch == "" {
		glog.Exit("Both branch and base must not be specified")
	}

	r := repodata.Create(*sourceOwner, *sourceRepo, *login, *token)

	if err := createPRs(r, *branch, *baseBranch); err != nil {
		glog.Exitf("submitPRs failed: %v", err)
	}

}
