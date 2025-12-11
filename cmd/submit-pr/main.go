package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/bretmckee/git-tools/pkg/repo/client"
	"github.com/bretmckee/git-tools/pkg/urls"
	"github.com/golang/glog"
	"github.com/google/go-github/v28/github"
	"github.com/kr/pretty"
)

const (
	maxCommitChainLength = 20
)

var errAlreadyMerged = fmt.Errorf("PR is already merged")

type sleepFunc func(time.Duration)

func stripDirectiveMarker(message, directive string) string {
	// Normalize line endings to handle both Unix (LF) and Windows (CRLF)
	normalized := strings.ReplaceAll(message, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")

	// Remove trailing whitespace to find the actual end
	trimmed := strings.TrimRight(normalized, " \t\n")

	// Check if the last non-empty line is the directive
	directivePattern := fmt.Sprintf(`\n%s:\s*[/A-Za-z0-9_.-]+$`, regexp.QuoteMeta(directive))
	re := regexp.MustCompile(directivePattern)

	if re.MatchString(trimmed) {
		// Remove the directive line
		trimmed = re.ReplaceAllString(trimmed, "")

		// Now check if the new last line is only underscores
		trimmed = strings.TrimRight(trimmed, " \t\n")
		underscorePattern := regexp.MustCompile(`(^|\n)\s*_+\s*$`)
		if underscorePattern.MatchString(trimmed) {
			trimmed = underscorePattern.ReplaceAllString(trimmed, "$1")
		}
	}

	// Clean up whitespace
	cleaned := strings.TrimSpace(trimmed)
	cleaned = regexp.MustCompile(`\n\n\n+`).ReplaceAllString(cleaned, "\n\n")

	return cleaned
}

func validatePRAndBranches(c *client.Client, pr *github.PullRequest, baseBranch string, force bool) error {
	if pr.GetMerged() {
		glog.Warningf("PR %d is already merged.", pr.GetNumber())
		return errAlreadyMerged
	}

	if prRef := pr.GetBase().GetRef(); prRef != baseBranch {
		err := fmt.Errorf("pr base ref (%q) does not match base branch ref (%q)", prRef, baseBranch)
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
		err := fmt.Errorf("pr base SHA (%q) does not match base branch SHA (%q)", prSHA, bbSHA)
		if !force {
			return err
		}
		glog.Warningf("because force was specified, ignoring error %v", err)
	}

	return nil
}

func waitForChecks(c *client.Client, ref string, number int, force bool, sleepFn sleepFunc) error {
	const retrySeconds = 60

	if sleepFn == nil {
		sleepFn = time.Sleep
	}

	for {
		state, err := c.AggregatedStatus(ref)
		if err != nil {
			return fmt.Errorf("failed to get status: %v", err)
		}

		if state == "failure" {
			err := fmt.Errorf("pr %d cannot be submitted because it has status %s", number, state)
			if !force {
				return err
			}
			glog.Warningf("because force was specified, ignoring error %v", err)
			return nil
		}

		if state != "pending" {
			return nil
		}

		if force {
			glog.Warningf("PR is pending, but not waiting because force was specified")
			return nil
		}

		glog.Warningf("pr %d status is pending: waiting %d seconds", number, retrySeconds)
		sleepFn(time.Second * retrySeconds)
	}
}

func submitMsg(c *client.Client, prBody string, first, last string, directive string) (string, error) {
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

		cleanedMsg := stripDirectiveMarker(commit.GetMessage(), directive)
		msg = "* " + cleanedMsg + "\n\n" + msg
		glog.V(2).Infof("after commit %s, msg=[%v]", pos, msg)
		pos = *commit.Parents[0].SHA
		if l >= maxCommitChainLength {
			return "", fmt.Errorf("submitMsg: max chain length (%d) exceeded", maxCommitChainLength)
		}
	}

	// If there are fewer than two commits, just use the pr body as the message.
	if l < 2 {
		return stripDirectiveMarker(prBody, directive), nil
	}
	return msg, nil
}

func submitPR(c *client.Client, dryRun, force bool, baseBranch string, number int, method, directive string) error {
	pr, err := c.PullRequest(number)
	if err != nil {
		return fmt.Errorf("submitPR: failed to get %d: %v", number, err)
	}

	if err := validatePRAndBranches(c, pr, baseBranch, force); err != nil {
		if err == errAlreadyMerged {
			return nil
		}
		return err
	}

	if err := waitForChecks(c, pr.GetHead().GetRef(), number, force, nil); err != nil {
		return err
	}

	msg, err := submitMsg(c, *pr.Body, pr.GetHead().GetSHA(), pr.GetBase().GetSHA(), directive)
	if err != nil {
		return fmt.Errorf("submitPR failed to build submitMsg: %v", err)
	}

	if dryRun {
		glog.Warningf("skipping submission of %d because a dry run was requested, msg=[%s]", number, msg)
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
		baseURL     = flag.String("url", "", "GitHub Base URL")
		directive   = flag.String("directive", "bretmckee-branch", "Directive marker to remove from commit messages")
		dryRun      = flag.Bool("dry-run", false, "Dry Run mode -- no pull requests will be created")
		force       = flag.Bool("force", false, "Submit even if not fully approved.")
		login       = flag.String("login", "", "Login of the user to submit for.")
		method      = flag.String("method", "squash", "github merge method -- [merge|rebase|squash]")
		pr          = flag.Int("pr", 0, "id of the pull request to submit")
		sourceOwner = flag.String("source-owner", "", "Name of the owner (user or org) of the repo to create the commit in.")
		sourceRepo  = flag.String("source-repo", "", "Name of repo to create the commit in.")
		token       = flag.String("token", "", "github auth token to use (also checks environment GITHUB_TOKEN")
		uploadURL   = flag.String("upload", "", "GitHub Upload URL")
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

	b, u, err := urls.Get(*baseURL, *uploadURL)
	if err != nil {
		glog.Exitf("failed to get URLs: %v", err)
	}

	c, err := client.Create(b, u, *sourceOwner, *sourceRepo, *login, *token)
	if err != nil {
		glog.Exitf("failed to create client: %v", err)
	}

	if err := submitPR(c, *dryRun, *force, *baseBranch, *pr, *method, *directive); err != nil {
		glog.Exitf("submitPR failed: %v", err)
	}
}
