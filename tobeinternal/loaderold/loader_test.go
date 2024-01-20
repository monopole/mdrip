package loaderold

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/monopole/mdrip/tobeinternal/base"
	"github.com/monopole/mdrip/tobeinternal/model"
)

// only run locally, not on travis
func TestReload(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loader-test-")
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
	err = os.WriteFile(tmpDir+"/foo.md", out.Bytes(), 0644)
	if err != nil {
		t.Errorf("Trouble writing to " + tmpDir)
		return
	}
	ds, err := base.NewDataSet([]string{tmpDir})
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
	ds, err := base.NewDataSet([]string{"git@github.com:monopole/mdrip.git"})
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
