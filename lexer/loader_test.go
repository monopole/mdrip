package lexer

import (
	"os"
	"testing"

	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
)

func TestReload(t *testing.T) {
	_, err := LoadTutorialFromPaths([]base.FilePath{})
	if err == nil {
		t.Errorf("Reading an empty array should not work.")
	}
}

func TestLoadTutorialFromGitHub(t *testing.T) {
	tut, err := LoadTutorialFromGitHub("git@github.com:monopole/mdrip.git")
	if err != nil {
		t.Errorf("Error reading from github: %v", err)
		return
	}
	printer := model.NewTutorialTxtPrinter(os.Stdout)
	tut.Accept(printer)
}
