package loader

import (
	"os"
	"testing"

	"bytes"
	"fmt"
	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
	"io/ioutil"
	"path/filepath"
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
					spec = filepath.Join(spec, pathName)
				}
				input := fmt.Sprintf(extractFmt, spec)
				if !smellsLikeGithubCloneArg(input) {
					t.Errorf("Should smell like github arg: %s\n", input)
					continue
				}
				repo, path, err := extractGithubRepoName(input)
				if err != nil {
					t.Errorf("problem %v", err)
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
			got := buildGithubCloneArg(repoName)
			want := "https://github.com/" + repoName + ".git"
			if got != want {
				t.Errorf("\n"+
					" got %s\n"+
					"want %s\n", got, want)
			}
		}
	}
}

// only run locally, not on travis
func xTestReload(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "loader-test-")
	if err != nil {
		t.Errorf("Trouble creating temp dir")
		return
	}

	var out bytes.Buffer
	fmt.Fprintln(&out, "hello there")
	fmt.Fprintln(&out, "<!-- @beans -->")
	fmt.Fprintln(&out, "```")
	fmt.Fprintln(&out, "echo face")
	fmt.Fprintln(&out, "```")
	err = ioutil.WriteFile(tmpDir+"/foo.md", out.Bytes(), 0644)
	if err != nil {
		t.Errorf("Trouble writing to " + tmpDir)
		return
	}
	ds, err := base.NewDataSource([]string{tmpDir})
	if err != nil {
		t.Errorf("Trouble making datasource")
		return
	}
	tut, err := NewLoader(ds).Load()
	if err != nil {
		t.Errorf("Unable to load tutorial: %v", err)
		return
	}
	printer := model.NewTutorialTxtPrinter(os.Stdout)
	tut.Accept(printer)
}

// only run locally, not on travis
func xTestLoadTutorialFromGitHub(t *testing.T) {
	ds, err := base.NewDataSource([]string{"git@github.com:monopole/mdrip.git"})
	if err != nil {
		t.Errorf("Trouble making datasource")
		return
	}
	tut, err := NewLoader(ds).Load()
	if err != nil {
		t.Errorf("Error reading from github: %v", err)
		return
	}
	printer := model.NewTutorialTxtPrinter(os.Stdout)
	tut.Accept(printer)
}
