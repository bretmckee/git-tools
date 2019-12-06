package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bretmckee/git-tools/pkg/repo"
	"github.com/bretmckee/git-tools/pkg/repo/client"
	"github.com/golang/glog"
	"github.com/google/go-github/v28/github"
	"github.com/kr/pretty"
)

type RepoData struct {
	repo.Repo
	branchBySHA map[string]*github.Branch
	prByNumber  map[int]*github.PullRequest
	chain       []string
}

const MaxChainLength = 100

func (r *RepoData) GetCommitChain(pos, end string) error {
	var chain []string

	glog.Infof("getCommitChain begins pos=%s end=%s", pos, end)
	for pos != end && len(chain) < MaxChainLength {
		glog.V(2).Infof("pos=%s", pos)
		commit, err := r.Commit(pos)
		if err != nil {
			return fmt.Errorf("GetCommitChain failed to get commit: %v", err)
		}
		if parents := len(commit.Parents); parents != 1 {
			return fmt.Errorf("GetCommitChain: commit %s has %d parents", pos, parents)
		}
		chain = append(chain, pos)
		pos = *commit.Parents[0].SHA
	}

	for left, right := 0, len(chain)-1; left < right; left, right = left+1, right-1 {
		chain[left], chain[right] = chain[right], chain[left]
	}
	r.chain = chain
	glog.Infof("chain=%v", chain)
	return nil
}

type Chain struct {
	SHAs []string
}

func buildChains(prs map[string]*github.PullRequest) ([]Chain, error) {
	var starts []string
	deps := make(map[string]string)
	for _, pr := range prs {
		if *pr.Base.Ref == "master" {
			starts = append(starts, *pr.Head.SHA)
			continue
		}
		if id, ok := deps[*pr.Base.SHA]; ok {
			return nil, fmt.Errorf("only linear dependencies are supported, but %s and %s depend on %s", id, *pr.Head.SHA, id)
		}
		deps[*pr.Base.SHA] = *pr.Head.SHA
	}
	glog.Infof("starts=%v", starts)
	glog.Infof("deps=%v", deps)
	var chains []Chain
	for _, id := range starts {
		var chain Chain
		for ; id != ""; id = deps[id] {
			glog.Infof("appending %s", id)
			chain.SHAs = append(chain.SHAs, id)
		}
		glog.Infof("chain: %v", chain)
		chains = append(chains, chain)
	}
	return chains, nil
}

//XXfunc processChain(c repo.Repo, prs map[string]*github.PullRequest, chain Chain) error {
//XX	glog.Infof("processChain begins for %v", chain)
//XX	mergedPrevious := false
//XX	for _, sha := range chain.SHAs {
//XX		pr, ok := prs[sha]
//XX		if !ok {
//XX			return fmt.Errorf("Unable to locate pr for %q", sha)
//XX		}
//XX		id := int(*pr.Number)
//XX		pr, err := c.PullRequest(id)
//XX		if err != nil {
//XX			return fmt.Errorf("Unable to get pr %d for %q", id)
//XX		}
//XX		// If we merged the previous PR, we have to change the base of this one.
//XX		if mergedPrevious {
//XX			if err := c.ChangePullRequestBase(id); err != nil {
//XX				return fmt.Errorf("Failed to update base for %d: %v", id, err)
//XX			}
//XX			mergedPrevious = false
//XX			glog.Warning("Short circuit returning, Mergable=%v", prs[sha].Mergeable)
//XX			return nil
//XX		}
//XX		if pr.Mergeable == nil {
//XX			glog.Infof("processChain ends for %s which has nil Mergeable", sha)
//XX			return nil
//XX		}
//XX		if !*pr.Mergeable {
//XX			glog.Infof("processChain ends for %s because Mergeable is false", sha)
//XX			return nil
//XX		}
//XX		glog.Infof("merging branch %d", id)
//XX		if prs[sha], err = c.MergePullRequest(id, sha, *pr.Body); err != nil {
//XX			return fmt.Errorf("Unable to merge %d: %v", id, err)
//XX		}
//XX		mergedPrevious = true
//XX	}
//XX	return nil
//XX}

func (r *RepoData) loadBranches() error {
	branches, err := r.Branches()
	if err != nil {
		return fmt.Errorf("list branches failed: %v", err)
	}
	bySHA := make(map[string]*github.Branch)
	for _, b := range branches {
		if _, ok := bySHA[*b.Commit.SHA]; ok {
			glog.Infof("skipping duplicate branch %s for %s", *b.Name, *b.Commit.SHA)
			continue
		}
		glog.V(2).Infof("adding branch %s: %# v", *b.Name, pretty.Formatter(*b))
		bySHA[*b.Commit.SHA] = b
	}
	r.branchBySHA = bySHA
	return nil
}

func (r *RepoData) loadPRs() error {
	prs, err := r.PullRequests()
	if err != nil {
		return fmt.Errorf("unable to get pull requests: %v", err)
	}
	glog.Infof("got %d pull requests", len(prs))
	// TODO(bretmckee): Make this by SHA of pull request branch
	prByNumber := make(map[int]*github.PullRequest)
	for _, pr := range prs {
		id := *pr.Number
		if prByNumber[id], err = r.PullRequest(id); err != nil {
			return fmt.Errorf("unable to fetch full PR %d: %v", id, err)
		}
		glog.Infof("adding pr %d: %# v", id, pretty.Formatter(prByNumber[id]))
	}
	return nil
}

func (r *RepoData) loadData(tip, base string) error {
	if err := r.loadBranches(); err != nil {
		return fmt.Errorf("createPRs failed to load branches: %v", err)
	}
	if err := r.loadPRs(); err != nil {
		return fmt.Errorf("createPRs failed to load PRs: %v", err)
	}
	if err := r.GetCommitChain(tip, base); err != nil {
		return fmt.Errorf("get commit chain failed: %v", err)
	}
	return nil
}

// createPRs createa any needed Pull Requests for commits in the range
// baseBranch...tipBranch.
func (r *RepoData) createPRs(tipBranch, baseBranch string) error {
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
	if err := r.loadData(*b.Commit.SHA, *bb.Commit.SHA); err != nil {
		return fmt.Errorf("failed to load data: %v", err)
	}
	return nil
	//XX	chain := ""
	//XX	prev := ""
	//XX	for _, commit := range chain {
	//XX		branch, ok := r.branchBySHA[*commit.SHA]
	//XX		if !ok {
	//XX			// This commit does not represent a branch
	//XX			continue
	//XX		}
	//XX		if prev != "" {
	//XX		}
	//XX		glog.Infof("%s is branch %s", *commit.SHA, *branch.Name)
	//XX	}

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

	r := RepoData{Repo: client.Create(*sourceOwner, *sourceRepo, *login, *token)}

	if err := r.createPRs(*branch, *baseBranch); err != nil {
		glog.Exitf("submitPRs failed: %v", err)
	}

}
