package loader

import (
	"os"
	"testing"

	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
)

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
