package tmux

// go test -v github.com/monopole/mdrip/tmux --alsologtostderr

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
	_, err := IsTmuxUp(badName)
	if err == nil {
		t.Errorf("Should fail using a nonsensical name like \"%s\".", badName)
	}
}

func TestAssureNotUp(t *testing.T) {
	if !IsProgramInstalled(Path) {
		t.Skip(skipNoTmux)
	}
	up, _ := IsTmuxUp(Path)
	if up {
		t.Error("tmux must not be up during test runs")
	}
}

func TestStartAndStopTmuxSession(t *testing.T) {
	if !IsProgramInstalled(Path) {
		t.Skip(skipNoTmux)
	}
	x := NewTmuxByName(Path)
	if x.IsRunning() {
		t.Skip(skipAlreadyRunning)
	}
	var out string
	err := x.Start()
	if err != nil {
		t.Errorf("unable to start session: %s", err)
	}
	if !x.IsRunning() {
		t.Errorf("tmux should appear as running: %s", err)
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
