package internal

import (
	"time"
)

// GithubRunner runs some "gh" commands.
type GithubRunner struct {
	rn *MyRunner
}

func NewGithubRunner(doIt behavior, dirSrc string, d time.Duration) *GithubRunner {
	return &GithubRunner{
		rn: NewMyRunner("gh", dirSrc, doIt, d),
	}
}

func (gh *GithubRunner) Release(tag string, assets []string) error {
	gh.rn.comment("releasing at tag " + tag)
	return gh.rn.run(
		UndoIsHard,
		append([]string{
			"release",
			"create",
			tag,
			"--verify-tag",
			"--draft",
			"--generate-notes", //automatically create title and notes.
			// "--title", tag,
			// Here's an example of generating release notes:
			// https://github.com/kubernetes-sigs/kustomize/blob/master/releasing/compile-changelog.sh
			//	"--notes-file", "foo.md",
		}, assets...)...)
}

func (gh *GithubRunner) Out() string {
	return gh.rn.Out()
}
