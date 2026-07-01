# Bret's stacked changes Github workflow tools

This repository contains my Github utilities. I use them to improve my workflow
when I have multiple stacked commits.

The workflow makes heavy use of `git rebase -i`, and if you are not familiar with
it you probably should become so before attempting to use these tools.

## Overview

The workflow these tools supports involves a few steps:
* Write the code.
* Use git `rebase -i` to re-arrange the commits into right order and pieces for
  the PRs you want to submit.
* Use git `rebase -i` to annotate which commits should have their own PRs.
* Run git push-branches (a.k.a git pb) to create and push branches for those
  commits.
* Look at GitHub to make sure that they are right.
* Run create-reviews to create reviews for the desired PRs
* In response to reviews:
 * Use `git rebase -i` to make any changes required. The commit messages for these
   should not be annotated unless you want a separate PR.
 * Run git pb again to update the PRs on GitHub.
* When the oldest PR is approved, run `submit-prs` to submit it (or a sequence).

## Installation
After cloning this repository, you need to:
* Build the executables with make.
* Arrange for the scripts in the scripts/ directory to be in your path. I do
  this by symlinking them into ~/bin.
* Modify scripts/git-push-branches by changing DIRECTIVE and BRANCH_PREFIX to
  include what you want them to be (you probably don't want "bretmckee" in them).
  If you leave BRANCH_PREFIX at its default it will be `users/${LOGIN}/`, so
  setting `LOGIN=<your-github-username>` in your environment is enough for
  most people.
* Run `git config --global alias.pb push-branches` to add the pb alias to git.
* [Create a Personal Access Token](
  https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token)
  and ensure that it is in the GITHUB_TOKEN environment variable (maybe via
  .profile?)
* For fork-based workflows: adding the upstream repo as a second git remote
  (e.g. `git remote add upstream <upstream-url>`) is optional but recommended
  so `submit-prs` can fetch the latest base branch after a merge. `git pb`
  itself does not need it - branches only get pushed to your fork.

## Using the scripts

### Create a branch based on a commit message
For your first experiment, I recommend you
* Create a github repo to experiment on, allowing Github to create a README.md
  file.
* Create a development branch, change README.md, and commit the change,
  ending with a string that matches the DIRECTIVE you set above. I like to include
  a line with two underscores before it to set the text apart, so mine might
  look like:
```
Update README.md

Add some more information to the read me.
__
bretmckee-branch: update-readme
```
* Push the new branch to git with `git pb`
* Look at the branch with Github to make sure it was properly created.

### Create a PR based on a commit message
## For github.com accounts

## For Enterprise accounts
```bash
create-reviews --logtostderr --draft=false --source-owner=bretmckee --login=bretmckee \
--base=main --branch=bretmckee/develop \
--source-repo=$(basename $(git rev-parse \
--show-toplevel)) -token "${GITHUB_TOKEN}"
```

## For fork-based accounts
When commits live in a fork and PRs are created against a different upstream
repository, pass `--upstream-owner` (and optionally `--upstream-repo` when the
upstream repo has a different name than the fork):
```bash
create-reviews --logtostderr --draft=false \
--source-owner=bretmckee --login=bretmckee \
--source-repo=$(basename $(git rev-parse --show-toplevel)) \
--upstream-owner=some-org \
--base=main --branch=users/bretmckee/develop \
-token "${GITHUB_TOKEN}"
```
When either `--upstream-*` flag differs from its `--source-*` counterpart,
`create-reviews` operates in fork mode: it lists branches on the fork, lists
and creates PRs in the upstream, and sets the PR head to
`${LOGIN}:${branch}` as GitHub requires for cross-fork PRs. When both flags
match (or are omitted), behaviour is identical to same-repo mode.

For stacked PRs, each PR after the first uses the previous stack branch as
its base, which GitHub requires to exist in the upstream repo. `create-reviews`
handles this automatically via the GitHub API: for every stack branch it sees
in the fork, it creates or force-updates a matching ref in the upstream
pointing at the same SHA. This relies on GitHub's fork network sharing the
git object store (true for real forks; if it fails on your setup the error
message surfaces both the update and create attempts so you can diagnose).
No dual push, no extra env var - `--upstream-owner` is all you need.

### Submit a PR based on a commit message
## For github.com accounts
## For Enterprise accounts
```bash
REMOTE="<remote>" SOURCE_OWNER="<owner>" URL="https://<url>/api/v3" TOKEN="${GITHUB_ENTERPRISE_TOKEN}" submit-prs 13
```

## For fork-based accounts
```bash
REMOTE="origin" UPSTREAM_REMOTE="upstream" \
SOURCE_OWNER="bretmckee" UPSTREAM_OWNER="some-org" \
TOKEN="${GITHUB_TOKEN}" submit-prs 13
```
`UPSTREAM_OWNER` (and `UPSTREAM_REPO` when the upstream repo name differs
from the fork) is forwarded to `submit-pr` and `rebase-prs` so PRs are queried
and merged in the upstream. `UPSTREAM_REMOTE` is optional and only used for
the `git fetch` of the latest base branch after each merge - if you omit it,
the base is fetched from `${REMOTE}` (your fork), which is fine as long as
your fork's default branch is synced with upstream.
