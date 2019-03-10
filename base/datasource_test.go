package base

import (
	"fmt"
	"testing"
)

var repoNames = []string{"monopole/mdrip", "kubernetes/website"}

var paths = []string{"", "README.md", "foo/index.md"}

var extractFmts = []string{
	"gh:%s",
	"GH:%s",
	"gitHub.com/%s",
	"https://github.com/%s",
	"hTTps://github.com/%s",
	"git@gitHUB.com:%s.git",
	"github.com:%s",
}

func TestExtractGithubRepoName(t *testing.T) {
	for _, repoName := range repoNames {
		for _, pathName := range paths {
			for _, extractFmt := range extractFmts {
				spec := repoName
				if len(pathName) > 0 {
					spec += "/" + pathName
				}
				input := fmt.Sprintf(extractFmt, spec)
				if !smellsLikeGithubCloneArg(input) {
					t.Errorf("Should smell like github arg: %s\n", input)
					continue
				}
				repo, path, err := extractGithubRepoName(input)
				if err != nil {
					t.Errorf("input='%s', err=%v", input, err)
				}
				if repo != repoName {
					t.Errorf("\n"+
						"       from %s\n"+
						"    gotRepo %s\n"+
						"desiredRepo %s\n", input, repo, repoName)
				}
				if path != pathName {
					t.Errorf("\n"+
						"       from %s\n"+
						"    gotPath %s\n"+
						"desiredPath %s\n", input, path, pathName)
				}
			}
		}
	}
}
