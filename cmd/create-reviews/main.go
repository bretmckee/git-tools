package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/bretmckee/git-tools/pkg/repo/repodata"
	"github.com/golang/glog"
	"github.com/google/go-github/v28/github"
	"github.com/kr/pretty"
)

func createPR(r *repodata.RepoData, branch, base, oldest, newest string, dryRun bool) error {
	o, err := r.Commit(oldest)
	if err != nil {
		return fmt.Errorf("CreatePR failed to get oldest commit %s: %v", oldest, err)
	}
	glog.V(2).Infof("Oldest Commit: %# v\n", pretty.Formatter(*o))
	values := regexp.MustCompile("[\n]+").Split(*o.Message, 2)
	if len(values) != 2 {
		return fmt.Errorf("CreatePR failed to split the message for commit %s (it probably does not have a body)", oldest)
	}

	npr := &github.NewPullRequest{
		Title:               github.String(values[0]),
		Head:                github.String(branch),
		Base:                github.String(base),
		Body:                github.String(values[1]),
		MaintainerCanModify: github.Bool(false),
		Draft:               github.Bool(true),
	}
	glog.V(2).Infof("npr=%# v", pretty.Formatter(*npr))
	if dryRun {
		glog.Infof("dryrun skipping: Creating PR for branch %s based on %s, oldest=%s, newest=%s", branch, base, oldest, newest)
		return nil
	}
	glog.V(2).Infof("Creating PR for branch %s based on %s, oldest=%s, newest=%s", branch, base, oldest, newest)
	pr, err := r.CreatePullRequest(npr)
	if err != nil {
		return fmt.Errorf("createPR failed to pr for %s:%v", branch, err)
	}
	glog.Infof("Created PR %d for branch %s", *pr.Number, branch)
	return nil
}

func findBranch(baseBranch string, branches []*github.Branch) (*github.Branch, error) {
	switch l := len(branches); l {
	case 1:
		return branches[0], nil
	case 2:
		for _, br := range branches {
			if *br.Name != baseBranch {
				return br, nil
			}
		}
		return nil, fmt.Errorf("findBranch: failed to find non-base branch for %s", branches[0].Commit.SHA)
	default:
		return nil, fmt.Errorf("findbranch: commit %s has invalid number of branches (%d), expect 1 or 2", branches[0].Commit.SHA, l)
	}
}

// createPRs createa any needed Pull Requests for commits in the range
// baseBranch...tipBranch.
func createPRs(r *repodata.RepoData, tipBranch, baseBranch string, maxCreates int, includeBranch, dryRun bool) error {
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

	chain, err := r.CommitChain(b.GetCommit().GetSHA(), *bb.Commit.SHA)
	if err != nil {
		return fmt.Errorf("Get commit chain failed: %v", err)
	}
	created := 0
	prev := ""
	base := baseBranch
	for _, commit := range chain {
		glog.V(2).Infof("examining commit %s", commit)
		if commit == *b.Commit.SHA && !includeBranch {
			glog.V(2).Infof("skipping tip branch %s", commit)
			return nil
		}
		if prev == "" {
			glog.V(2).Infof("setting prev to commit %s", commit)
			prev = commit
		}
		branches, ok := r.BranchBySHA[commit]
		if !ok {
			// This commit does not represent a branch
			glog.V(2).Infof("commit %s does not represent a branch", commit)
			continue
		}
		branch, err := findBranch(baseBranch, branches)
		if err != nil {
			return fmt.Errorf("failed to find branch: %v", err)
		}
		// We are at a commit that needs a PR. Create one unless there already is
		// one.
		if pr, ok := r.PrBySHA[*branch.Commit.SHA]; ok {
			glog.V(2).Infof("branch %s (sha %s) already has PR %d", *branch.Name, *branch.Commit.SHA, *pr.Number)
			base = *branch.Name
			prev = ""
			continue
		}
		if created >= maxCreates {
			return fmt.Errorf("maximum number of pull requests (%d) created, skipping creation for branch %s", maxCreates, *branch.Name)
		}
		if err := createPR(r, *branch.Name, base, prev, commit, dryRun); err != nil {
			return fmt.Errorf("failed to create pr: %v", err)
		}
		created += 1
		base = *branch.Name
		prev = ""
	}

	return nil
}

func main() {
	var (
		baseBranch    = flag.String("base", "master", "Base Branch")
		branch        = flag.String("branch", "", "Starting Branch")
		dryRun        = flag.Bool("dry-run", false, "Dry Run mode -- no pull requests will be created")
		includeBranch = flag.Bool("include-branch", false, "Create a PR for --branch")
		login         = flag.String("login", "", "Login of the user to create for.")
		maxCreates    = flag.Int("max-creates", 10, "Maximum number of pull requests to create")
		sourceOwner   = flag.String("source-owner", "", "Name of the owner (user or org) of the repo to create the commit in.")
		sourceRepo    = flag.String("source-repo", "", "Name of repo to create the commit in.")
		token         = flag.String("token", "", "github auth token to use (also checks environment GITHUB_TOKEN")
	)
	flag.Parse()
	if *token == "" {
		*token = os.Getenv("GITHUB_TOKEN")
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

	r, err := repodata.Create(*sourceOwner, *sourceRepo, *login, *token)
	if err != nil {
		glog.Exitf("failed to create repodata: %v", err)
	}

	if err := createPRs(r, *branch, *baseBranch, *maxCreates, *includeBranch, *dryRun); err != nil {
		glog.Exitf("createPRs failed: %v", err)
	}

}
