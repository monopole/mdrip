package tmux_test

import (
	. "github.com/monopole/mdrip/v2/internal/tmux"
	"strings"
	"testing"
)

const (
	badName            = "nonsensicalFakeHopeNotInstalledPgmName"
	skipAlreadyRunning = "skipping since tmux already running"
	skipNoTmux         = "skipping since tmux not found"
)

func TestBadName(t *testing.T) {
	if _, err := NewTmux(badName); err == nil {
		t.Errorf(
			"should fail using a nonsensical name like %q", badName)
	}
}

func TestStartAndStopTmuxSession(t *testing.T) {
	x, err := NewTmux(PgmName)
	if err != nil {
		t.Skip(skipNoTmux)
	}
	if x.IsUp() {
		t.Skip(skipAlreadyRunning)
	}
	var out string
	if err = x.Start(); err != nil {
		t.Errorf("unable to start session: %s", err)
	}
	if !x.IsUp() {
		t.Errorf("tmux should appear as running")
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
