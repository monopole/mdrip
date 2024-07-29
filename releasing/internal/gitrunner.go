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
	rn *runner
	// regExpTag matches a valid tag.
	regExpTag *regexp.Regexp
}

func NewGitRunner(doIt behavior, d time.Duration) *GitRunner {
	return &GitRunner{
		rn:        newRunner("git", "", doIt, d),
		regExpTag: regexp.MustCompile(`^v[0-9].`),
	}
}

func (git *GitRunner) AssureCleanWorkspace() error {
	git.rn.comment("assuring a clean workspace")
	if err := git.rn.run(noHarmDone, "status"); err != nil {
		return err
	}
	if !strings.Contains(git.rn.Out(), "nothing to commitHash, working tree clean") {
		return fmt.Errorf("the workspace isn't clean")
	}
	return nil
}

func (git *GitRunner) GetLatestTag() (string, error) {
	git.rn.comment("getting latest tag")
	if err := git.rn.run(noHarmDone, "describe", "--tags", "--abbrev=0"); err != nil {
		return "", err
	}
	if !git.regExpTag.Match([]byte(git.rn.Out())) {
		return "", fmt.Errorf(
			"purported tag %q doesn't match re %q",
			git.rn.Out(), git.regExpTag.String())
	}
	return strings.TrimSpace(git.rn.Out()), nil
}

func (git *GitRunner) GetCommitAtTag(tag string) (string, error) {
	git.rn.comment("getting commitHash at tag " + tag)
	if err := git.rn.run(noHarmDone, "show-ref", "-s", "refs/tags/"+tag); err != nil {
		return "", err
	}
	return strings.TrimSpace(git.rn.Out()), nil
}

func (git *GitRunner) GetLatestCommit() (string, error) {
	git.rn.comment("getting latest commitHash")
	if err := git.rn.run(noHarmDone, "rev-parse", "--verify", "HEAD"); err != nil {
		return "", err
	}
	return strings.TrimSpace(git.rn.Out()), nil

}
