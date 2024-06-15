package loader

import (
	"fmt"
	"testing"
)

func TestSplitDomainAndRepo(t *testing.T) {
	type testC struct {
		spec   string
		domain string
	}
	for _, tc := range []testC{
		{
			spec:   "gH:REPO",
			domain: "git@github.com:",
		},
		{
			spec:   "git@github.com:REPO",
			domain: "git@github.com:",
		},
		{
			spec:   "https://github.com/REPO",
			domain: "https://github.com/",
		},
		{
			spec:   "htTps://github.com:REPO",
			domain: "https://github.com:",
		},
		{
			spec:   "hTtps://github.tesla.com:REPO",
			domain: "https://github.tesla.com:",
		},
		{
			spec:   "https://github.tesla.com/REPO",
			domain: "https://github.tesla.com/",
		},
	} {
		t.Run(tc.spec, func(t *testing.T) {
			d, r := splitDomainAndRemainder(tc.spec)
			if d != tc.domain || r != "REPO" {
				t.Errorf("\n"+
					"    spec %q\n"+
					"  domain %q\n"+
					"    repo %q", tc.spec, d, r)
			}
		})
	}
}

func TestSplitRepoAndPath(t *testing.T) {
	const repoName = "monopole/mdrip"
	for _, pathName := range []string{
		"",
		"README.md",
		"foo/index.md",
		"more/than/one/blahBlah.md",
	} {
		for _, tstFmt := range []string{
			"%s",
			"%s.git",
		} {
			arg := makeTheTestArgument(repoName, pathName, tstFmt)
			repo, path, err := splitRepoAndPath(arg)
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

func makeTheTestArgument(repoName string, pathName string, tstFmt string) string {
	repoSpec := fmt.Sprintf(tstFmt, repoName)
	if pathName == "" {
		return repoSpec
	}
	return repoSpec + string(RootSlash) + pathName
}
