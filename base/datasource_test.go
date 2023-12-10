package base

import (
	"fmt"
	"testing"
)

func TestExtractGithubRepoName(t *testing.T) {
	for _, repoName := range []string{
		"monopole/mdrip",
		"kubernetes/website",
	} {
		for _, pathName := range []string{
			"",
			"README.md",
			"foo/index.md",
			"more/than/one/blahBlah.md",
		} {
			for _, tstFmt := range []string{
				"gh:%s",
				"GH:%s",
				"gitHub.com/%s",
				"https://github.com/%s",
				"hTTps://github.com/%s",
				"git@gitHUB.com:%s.git",
				"github.com:%s",
			} {
				arg := makeTheTestArgument(repoName, pathName, tstFmt)
				if !SmellsLikeGithubCloneArg(arg) {
					t.Errorf("Should smell like github arg: %s\n", arg)
					continue
				}
				repo, path, err := ExtractGithubRepoName(arg)
				if err != nil {
					t.Errorf("input='%s', err=%v", arg, err)
				}
				if repo != repoName {
					t.Errorf("\n"+
						"       from %s\n"+
						"    gotRepo %s\n"+
						"desiredRepo %s\n", arg, repo, repoName)
				}
				if path != pathName {
					t.Errorf("\n"+
						"       from %s\n"+
						"    gotPath %s\n"+
						"desiredPath %s\n", arg, path, pathName)
				}
			}
		}
	}
}

func makeTheTestArgument(repoName string, pathName string, extractFmt string) string {
	spec := repoName
	if len(pathName) > 0 {
		spec += "/" + pathName
	}
	return fmt.Sprintf(extractFmt, spec)
}
