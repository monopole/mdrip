package tmux

import (
	"strings"
	"testing"
)

const (
	badName            = "nonsensicalFakeHopeNotInstalledPgmName"
	skipAlreadyRunning = "skipping since tmux already running"
	skipNoTmux         = "skipping since tmux not found"
)

func TestBadName(t *testing.T) {
	if IsProgramInstalled(badName) {
		t.Errorf("Should fail using a nonsensical name like \"%s\".", badName)
	}
}

func TestStartAndStopTmuxSession(t *testing.T) {
	if !IsProgramInstalled(Path) {
		t.Skip(skipNoTmux)
	}
	x := NewTmux(Path)
	if x.IsUp() {
		t.Skip(skipAlreadyRunning)
	}
	var out string
	err := x.start()
	if err != nil {
		t.Errorf("unable to start session: %s", err)
	}
	if !x.IsUp() {
		t.Errorf("tmux should appear as running")
	}
	out, err = x.listSessions()
	if err != nil {
		t.Errorf("unable to list session: %s", err)
	}
	if !strings.Contains(out, SessionName+":") {
		t.Errorf("Expected %s:, got %s", SessionName, out)
	}
	err = x.stop()
	if err != nil {
		t.Errorf("unable to stop session: %s", err)
	}
}
