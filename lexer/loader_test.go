package lexer

import (
	"os"
	"testing"

	"bytes"
	"fmt"
	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
	"io/ioutil"
)

var repoNames = []string{"monopole/mdrip", "kubernetes/kubernetes.github.io"}

var extractFmts = []string{
	"gh:%s",
	"GH:%s",
	"github.com/%s",
	"https://github.com/%s",
	"git@github.com:%s.git",
	"github.com:%s",
}

func TestExtractGithubRepoName(t *testing.T) {
	for _, repoName := range repoNames {
		for _, extractFmt := range extractFmts {
			input := fmt.Sprintf(extractFmt, repoName)
			if !smellsLikeGithubCloneArg(input) {
				t.Errorf("Should smell like github arg: %s\n", input)
				continue
			}
			got := extractGithubRepoName(input)
			if got != repoName {
				t.Errorf("\n"+
					"       from %s\n"+
					"        got %s\n"+
					"desiredRepo %s\n", input, got, repoName)
			}
		}
		got := buildGithubCloneArg(repoName)
		want := "git@github.com:" + repoName + ".git"
		if got != want {
			t.Errorf("\n"+
				" got %s\n"+
				"want %s\n", got, want)
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
	defer os.RemoveAll(tmpDir)

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
