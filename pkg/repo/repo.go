package repo

import (
	"github.com/google/go-github/v28/github"
)

type Repo interface {
	// Branches returns a slice which contains all the branches for the
	// repository.  Note that not all fields in the individual elements may be
	// filled it. If complete data is required for a branch, Branch should be
	// called.
	Branches() ([]*github.Branch, error)

	// Branch returns full information for branch `name`.
	Branch(name string) (*github.Branch, error)

	// PullRequests returns a slice which contains all the pull requests for the
	// repository.  Note that not all fields in the individual elements may be
	// filled it. If complete data is required for a pull request, PullRequest
	// should be called.
	PullRequests() ([]*github.PullRequest, error)

	// PullRequest returns full information for pull request `num`.
	PullRequest(num int) (*github.PullRequest, error)

	// MergePullRequest
	MergePullRequest(num int, sha, msg string) (*github.PullRequest, error)

	//ChangePullRequestBase changes the base of pull request `num` to be `ref`.
	ChangePullRequestBase(num int, ref string) error

	// CreatePullRequest creates a new pull request.
	CreatePullRequest(npr *github.NewPullRequest) (*github.PullRequest, error)

	// Commit returns the full information for the commit with SHA `sha`.
	Commit(sha string) (*github.Commit, error)

	// Statuses returns the statues for commit ref.
	CombinedStatus(ref string) (*github.CombinedStatus, error)
}
