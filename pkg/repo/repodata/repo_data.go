package repodata

import (
	"fmt"

	"github.com/bretmckee/git-tools/pkg/repo"
	"github.com/bretmckee/git-tools/pkg/repo/client"
	"github.com/golang/glog"
	"github.com/google/go-github/v28/github"
	"github.com/kr/pretty"
)

type RepoData struct {
	repo.Repo
	BranchBySHA map[string][]*github.Branch
	PrBySHA     map[string]*github.PullRequest
	PrByNumber  map[int]*github.PullRequest
}

func Create(baseURL, uploadURL, sourceOwner, sourceRepo, login, token string) (*RepoData, error) {
	c, err := client.Create(baseURL, uploadURL, sourceOwner, sourceRepo, login, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}
	r := &RepoData{
		Repo: c,
	}
	if err := r.LoadData(); err != nil {
		return nil, fmt.Errorf("failed to load data: %v", err)
	}
	return r, nil
}

const MaxChainLength = 150

func (r *RepoData) CommitChain(pos, end string) ([]string, error) {
	var chain []string

	glog.V(2).Infof("GetCommitChain begins pos=%s end=%s", pos, end)
	for pos != end && len(chain) < MaxChainLength {
		glog.V(2).Infof("pos=%s", pos)
		commit, err := r.Commit(pos)
		if err != nil {
			return nil, fmt.Errorf("GetCommitChain failed to get commit: %v", err)
		}
		if parents := len(commit.Parents); parents != 1 {
			return nil, fmt.Errorf("GetCommitChain: commit %s has %d parents", pos, parents)
		}
		chain = append(chain, pos)
		pos = *commit.Parents[0].SHA
	}

	for left, right := 0, len(chain)-1; left < right; left, right = left+1, right-1 {
		chain[left], chain[right] = chain[right], chain[left]
	}
	return chain, nil
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
	glog.V(2).Infof("starts=%v", starts)
	glog.V(2).Infof("deps=%v", deps)
	var chains []Chain
	for _, id := range starts {
		var chain Chain
		for ; id != ""; id = deps[id] {
			glog.V(2).Infof("appending %s", id)
			chain.SHAs = append(chain.SHAs, id)
		}
		glog.V(2).Infof("chain: %v", chain)
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
	r.BranchBySHA = make(map[string][]*github.Branch)
	for _, b := range branches {
		glog.V(2).Infof("adding branch %s: %# v", *b.Name, pretty.Formatter(*b))
		r.BranchBySHA[*b.Commit.SHA] = append(r.BranchBySHA[*b.Commit.SHA], b)
	}
	return nil
}

func (r *RepoData) loadPRs() error {
	prs, err := r.PullRequests()
	if err != nil {
		return fmt.Errorf("unable to get pull requests: %v", err)
	}
	glog.V(2).Infof("got %d pull requests", len(prs))
	// TODO(bretmckee): Make this by SHA of pull request branch
	r.PrBySHA = make(map[string]*github.PullRequest)
	r.PrByNumber = make(map[int]*github.PullRequest)
	for _, pr := range prs {
		sha := *pr.Head.SHA
		id := *pr.Number
		fullPR, err := r.PullRequest(id)
		if err != nil {
			return fmt.Errorf("unable to fetch full PR %d (sha %s): %v", id, sha, err)
		}
		glog.V(2).Infof("adding pr %d: %# v", id, pretty.Formatter(fullPR))
		r.PrBySHA[sha] = fullPR
		r.PrByNumber[id] = fullPR
	}
	return nil
}

func (r *RepoData) LoadData() error {
	if err := r.loadBranches(); err != nil {
		return fmt.Errorf("createPRs failed to load branches: %v", err)
	}
	if err := r.loadPRs(); err != nil {
		return fmt.Errorf("createPRs failed to load PRs: %v", err)
	}
	return nil
}
