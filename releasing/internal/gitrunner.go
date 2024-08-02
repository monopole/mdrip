package internal

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	mainBranch = "master"
)

// GitRunner runs some git commands.
type GitRunner struct {
	rn *MyRunner
	// regExpTag matches a valid tag.
	regExpTag *regexp.Regexp
}

func NewGitRunner(doIt behavior, d time.Duration) *GitRunner {
	return &GitRunner{
		rn:        NewMyRunner("git", "", doIt, d),
		regExpTag: regexp.MustCompile(`^v[0-9].`),
	}
}

func (git *GitRunner) AssureCleanWorkspace() error {
	git.rn.comment("assuring a clean workspace")
	if err := git.rn.run(NoHarmDone, "status"); err != nil {
		return err
	}
	if !strings.Contains(git.rn.Out(), "nothing to commit, working tree clean") {
		return fmt.Errorf("the workspace isn't clean")
	}
	return nil
}

func (git *GitRunner) GetLatestTag() (string, error) {
	git.rn.comment("getting latest tag")
	if err := git.rn.run(NoHarmDone, "describe", "--tags", "--abbrev=0"); err != nil {
		return "", err
	}
	if !git.regExpTag.Match([]byte(git.rn.Out())) {
		return "", fmt.Errorf(
			"purported tag %q doesn't match re %q",
			git.rn.Out(), git.regExpTag.String())
	}
	return strings.TrimSpace(git.rn.Out()), nil
}

func (git *GitRunner) GetTagAtCommit(hash string) (string, error) {
	git.rn.comment("getting tag closest to hash " + hash)
	if err := git.rn.run(NoHarmDone, "describe", "--exact-match", hash); err != nil {
		return "", fmt.Errorf("%s %w", git.rn.Out(), err)
	}
	return strings.TrimSpace(git.rn.Out()), nil
}

func (git *GitRunner) GetCommitHashOfHead() (string, error) {
	git.rn.comment("getting the commit hash of HEAD")
	if err := git.rn.run(NoHarmDone, "rev-parse", "--verify", "HEAD"); err != nil {
		return "", fmt.Errorf("%s %w", git.rn.Out(), err)
	}
	return strings.TrimSpace(git.rn.Out()), nil
}
