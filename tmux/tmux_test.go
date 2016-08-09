package tmux

// go test -v github.com/monopole/mdrip/tmux --alsologtostderr

import (
	"strings"
	"testing"
)

const (
	badName     = "nonsensicalFakeHopeNotInstalledPgmName"
	sessionName = "tmuxTestSessionThatShouldNotSurviveTest"
)

func TestBadName(t *testing.T) {
	x := NewTmux(badName)
	err := x.Refresh()
	if err == nil {
		t.Errorf("Should fail using a nonsensical name like \"%s\".", badName)
	}
}

func TestTmuxInstalled(t *testing.T) {
	x := NewTmux(ProgramName)
	err := x.Refresh()
	if err != nil {
		t.Errorf("\"%s\" not installed?", ProgramName)
	}
}

func TestStartAndStopTmuxSession(t *testing.T) {
	x := NewTmux(ProgramName)
	var out string
	err := x.Start()
	if err != nil {
		t.Errorf("unable to start session: %s", err)
	}
	out, err = x.ListSessions()
	if err != nil {
		t.Errorf("unable to list session: %s", err)
	}
	if !strings.Contains(out, SessionName+":") {
		t.Errorf("Expected %s:, got %s", SessionName, out)
	}
	err = x.Stop()
	if err != nil {
		t.Errorf("unable to stop session: %s", err)
	}
}
