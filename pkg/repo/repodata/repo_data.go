package repodata

import (
	"fmt"

	"github.com/bretmckee/git-tools/pkg/repo/client"
	"github.com/golang/glog"
	"github.com/google/go-github/v28/github"
	"github.com/kr/pretty"
)

type RepoData struct {
	Fork        *client.Client
	Upstream    *client.Client
	BranchBySHA map[string][]*github.Branch
	PrBySHA     map[string]*github.PullRequest
	PrByNumber  map[int]*github.PullRequest
}

func Create(baseURL, uploadURL, sourceOwner, sourceRepo, upstreamOwner, upstreamRepo, login, token string) (*RepoData, error) {
	fork, upstream, err := client.CreatePair(baseURL, uploadURL, sourceOwner, sourceRepo, upstreamOwner, upstreamRepo, login, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create clients: %v", err)
	}
	r := &RepoData{
		Fork:     fork,
		Upstream: upstream,
	}
	if err := r.LoadData(); err != nil {
		return nil, fmt.Errorf("failed to load data: %v", err)
	}
	return r, nil
}

// IsFork reports whether the tool is operating in fork mode (source and
// upstream differ). Same-repo mode is signalled by fork and upstream being the
// same underlying *client.Client (see client.CreatePair).
func (r *RepoData) IsFork() bool {
	return r.Fork != r.Upstream
}

// HeadRefFor returns the value to use for NewPullRequest.Head. In same-repo
// mode it is the bare branch name; in fork mode it is "login:branch" as
// required by GitHub for cross-fork PRs.
func (r *RepoData) HeadRefFor(branch string) string {
	if !r.IsFork() {
		return branch
	}
	return fmt.Sprintf("%s:%s", r.Fork.Owner(), branch)
}

// SyncAnchor keeps the upstream anchor branch pointing at the fork's SHA so
// cross-fork stacked PRs can reference it as their base. No-op in same-repo
// mode. Relies on GitHub's fork network sharing the git object store: a SHA
// pushed only to the fork can be referenced by a ref in upstream.
func (r *RepoData) SyncAnchor(branchName, sha string) error {
	if !r.IsFork() {
		return nil
	}
	if err := r.Upstream.EnsureRef(branchName, sha); err != nil {
		return fmt.Errorf("SyncAnchor: %v", err)
	}
	return nil
}

func (r *RepoData) Branch(name string) (*github.Branch, error) {
	return r.Fork.Branch(name)
}

func (r *RepoData) Commit(sha string) (*github.Commit, error) {
	return r.Fork.Commit(sha)
}

func (r *RepoData) CreatePullRequest(npr *github.NewPullRequest) (*github.PullRequest, error) {
	return r.Upstream.CreatePullRequest(npr)
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

func (r *RepoData) loadBranches() error {
	branches, err := r.Fork.Branches()
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
	prs, err := r.Upstream.PullRequests()
	if err != nil {
		return fmt.Errorf("unable to get pull requests: %v", err)
	}
	glog.V(2).Infof("got %d pull requests", len(prs))
	r.PrBySHA = make(map[string]*github.PullRequest)
	r.PrByNumber = make(map[int]*github.PullRequest)
	for _, pr := range prs {
		sha := *pr.Head.SHA
		id := *pr.Number
		fullPR, err := r.Upstream.PullRequest(id)
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
