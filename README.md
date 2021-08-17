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
* Run `git config --global alias.pb push-branches` to add the pb alias to git.
* [Create a Personal Access Token](
  https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token)
  and ensure that it is in the GITHUB_TOKEN environment variable (maybe via
  .profile?)

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
To Be Written.

### Submit a PR based on a commit message
To Be Written.
