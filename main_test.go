package main

import (
	"testing"
	"github.com/monopole/mdrip/config"
)

// TODO: real tests
// e.g. make a tmpdir, put a script in there, have script write file
// then run in test mode and confirm the file was written via test execution
// and has the correct contents.  requires a filesystem currently.
func TestTrueMain(t *testing.T) {
    err := trueMain(config.DefaultConfig())
	if err == nil {
		t.Errorf("Expected error")
		return
	}
}
