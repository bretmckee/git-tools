package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/google/go-github/v28/github"
	"github.com/kr/pretty"
	"golang.org/x/oauth2"
)

const MaxChainLength = 100

type Connection struct {
	ctx    context.Context
	owner  string
	repo   string
	login  string
	client *github.Client
}

func newConnection(owner, repo string, client *github.Client) *Connection {
	return &Connection{
		ctx:    context.Background(),
		owner:  owner,
		repo:   repo,
		client: client}
}

func (c *Connection) ListBranches() ([]*github.Branch, error) {
	o := &github.ListOptions{}
	branches, resp, err := c.client.Repositories.ListBranches(c.ctx, c.owner, c.repo, o)
	if err != nil {
		return nil, fmt.Errorf("Failed to list branches: %v", err)
	}
	if glog.V(2) {
		for i, b := range branches {
			glog.V(2).Infof("branch %d: %# v\n", i, pretty.Formatter(*b))
		}
		glog.V(2).Infof("resp=%# v\n", resp)
	}
	return branches, nil
}

func (c *Connection) GetBranch(name string) (*github.Branch, error) {
	b, _, err := c.client.Repositories.GetBranch(c.ctx, c.owner, c.repo, name)
	if err != nil {
		return nil, fmt.Errorf("get of branch %q failed: %v", name, err)
	}
	glog.V(2).Infof("Get of Branch %q: %# v\n", name, pretty.Formatter(*b))
	return b, nil
}

func (c *Connection) ListPRs() ([]*github.PullRequest, error) {
	o := &github.PullRequestListOptions{}
	prs, _, err := c.client.PullRequests.List(c.ctx, c.owner, c.repo, o)
	if err != nil {
		return nil, fmt.Errorf("Failed to list pull requests: %v", err)
	}
	return prs, nil
}

func (c *Connection) GetPr(id int) (*github.PullRequest, error) {
	pr, _, err := c.client.PullRequests.Get(c.ctx, c.owner, c.repo, id)
	if err != nil {
		return nil, fmt.Errorf("Get of PR %d failed: %v", id, err)
	}
	glog.V(2).Infof("PR %d: %# v\n", id, pretty.Formatter(*pr))
	return pr, nil
}

func (c *Connection) GetCommit(sha string) (*github.Commit, error) {
	commit, _, err := c.client.Git.GetCommit(c.ctx, c.owner, c.repo, sha)
	if err != nil {
		return nil, fmt.Errorf("Get of commit %q failed: %v", sha, err)
	}
	glog.V(2).Infof("Commit %q: %# v\n", sha, pretty.Formatter(*commit))
	return commit, nil
}

func (c *Connection) GetCommitChain(pos, end string) ([]*github.Commit, error) {
	var chain []*github.Commit

	for pos != end && len(chain) < MaxChainLength {
		commit, err := c.GetCommit(pos)
		if err != nil {
			return nil, fmt.Errorf("GetCommitChain failed to get commit: %v", err)
		}
		if parents := len(commit.Parents); parents != 1 {
			return nil, fmt.Errorf("GetCommitChain: commit %s has %d parents", pos, parents)
		}
		chain = append(chain, commit)
		pos = *commit.Parents[0].SHA
	}

	return chain, nil
}

type Chain struct {
	SHAs []string
}

func (c *Connection) updateBase(id int) (*github.PullRequest, error) {
	glog.Warningf("updateBase returning unchanged pr %d", id)
	pr, err := c.GetPr(id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get pr after updating base for %d: %v", id, err)
	}
	return pr, nil
}

func (c *Connection) mergePR(id int, sha, msg string) (*github.PullRequest, error) {
	o := &github.PullRequestOptions{
		SHA:         sha,
		MergeMethod: "rebase",
	}
	res, resp, err := c.client.PullRequests.Merge(c.ctx, c.owner, c.repo, id, msg, o)
	if err != nil {
		return nil, fmt.Errorf("github merge of %d failed: %v", id, err)
	}
	glog.V(2).Infof("merge PR %d res: %# v\n", id, pretty.Formatter(*res))
	glog.V(2).Infof("merge PR %d resp: %# v\n", id, pretty.Formatter(*resp))
	pr, err := c.GetPr(id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get pr after updating base for %d: %v", id, err)
	}
	return pr, nil
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

func (c *Connection) processChain(prs map[string]*github.PullRequest, chain Chain) error {
	glog.Infof("processChain begins for %v", chain)
	mergedPrevious := false
	for _, sha := range chain.SHAs {
		pr, ok := prs[sha]
		if !ok {
			return fmt.Errorf("Unable to locate pr for %q", sha)
		}
		id := int(*pr.Number)
		pr, err := c.GetPr(id)
		if err != nil {
			return fmt.Errorf("Unable to get pr %d for %q", id)
		}
		// If we merged the previous PR, we have to change the base of this one.
		if mergedPrevious {
			if prs[sha], err = c.updateBase(id); err != nil {
				return fmt.Errorf("Failed to update base for %d: %v", id, err)
			}
			mergedPrevious = false
			glog.Warning("Short circuit returning, Mergable=%v", prs[sha].Mergeable)
			return nil
		}
		if pr.Mergeable == nil {
			glog.Infof("processChain ends for %s which has nil Mergeable", sha)
			return nil
		}
		if !*pr.Mergeable {
			glog.Infof("processChain ends for %s because Mergeable is false", sha)
			return nil
		}
		glog.Infof("merging branch %d", id)
		if prs[sha], err = c.mergePR(id, sha, *pr.Body); err != nil {
			return fmt.Errorf("Unable to merge %d: %v", id, err)
		}
		mergedPrevious = true
	}
	return nil
}

func (c *Connection) process(prs []*github.PullRequest, login string) error {
	byId := make(map[string]*github.PullRequest)
	for _, pr := range prs {
		if *pr.User.Login == login {
			byId[*pr.Head.SHA] = pr
		}
	}

	chains, err := buildChains(byId)
	if err != nil {
		return fmt.Errorf("buildChains failed: %v", err)
	}

	for _, chain := range chains {
		if err := c.processChain(byId, chain); err != nil {
			return fmt.Errorf("process Chain failed: %v", err)
		}
	}

	return nil
}

func (c *Connection) submitPRs(tipBranch, baseBranch string) error {
	b, err := c.GetBranch(tipBranch)
	if err != nil {
		return fmt.Errorf("failed to get tip branch %q: %v", tipBranch, err)
	}
	if b.Commit.SHA == nil {
		return fmt.Errorf("Branch Commit SHA is nil: %# v\n", pretty.Formatter(*b))
	}
	glog.V(2).Infof("Branch %# v\n", pretty.Formatter(*b))

	bb, err := c.GetBranch(baseBranch)
	if err != nil {
		return fmt.Errorf("failed to get base branch %q: %v", baseBranch, err)
	}
	if bb.Commit.SHA == nil {
		return fmt.Errorf("Branch Commit SHA is nil: %# v\n", pretty.Formatter(*bb))
	}
	glog.V(2).Infof("Base Branch %# v\n", pretty.Formatter(*bb))

	branches, err := c.ListBranches()
	if err != nil {
		return fmt.Errorf("list branches failed: %v", err)
	}

	bySHA := make(map[string]*github.Branch)
	bySHA[*b.Commit.SHA] = b
	bySHA[*bb.Commit.SHA] = bb

	for _, b := range branches {
		if _, ok := bySHA[*b.Commit.SHA]; ok {
			glog.Infof("skipping duplicate branch %s for %s", *b.Name, *b.Commit.SHA)
			continue
		}
		bySHA[*b.Commit.SHA] = b
	}

	chain, err := c.GetCommitChain(*b.Commit.SHA, *bb.Commit.SHA)
	if err != nil {
		return fmt.Errorf("get commit chain failed: %v", err)
	}
	prev := ""
	for _, commit := range chain {
		branch, ok := bySHA[*commit.SHA]
		if !ok {
			// This commit does not represent a branch
			continue
		}
		if prev != "" {
		}
		glog.Infof("%s is branch %s", *commit.SHA, *branch.Name)
	}

	return nil
}

func main() {
	var (
		sourceOwner = flag.String("source-owner", "", "Name of the owner (user or org) of the repo to create the commit in.")
		sourceRepo  = flag.String("source-repo", "", "Name of repo to create the commit in.")
		login       = flag.String("login", "", "Login of the user to submit for.")
		baseBranch  = flag.String("base", "master", "Base Branch")
		branch      = flag.String("branch", "", "Starting Branch")
	)
	flag.Parse()
	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token == "" {
		glog.Fatal("Unauthorized: No token present")
	}
	if *sourceOwner == "" || *sourceRepo == "" || *login == "" {
		glog.Fatal("You need to specify a non-empty value for the flags `-user`, `-source-owner` and `-source-repo`")
	}
	if *branch == "" || *baseBranch == "" {
		glog.Fatal("Both branch and base must not be empty.")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	c := newConnection(*sourceOwner, *sourceRepo, github.NewClient(tc))

	if err := c.submitPRs(*branch, *baseBranch); err != nil {
		glog.Fatalf("submitPRs failed: %v", err)
	}

}
