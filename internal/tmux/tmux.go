package tmux

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/monopole/mdrip/v2/internal/utils"
)

// Tmux holds information about a tmux process (https://github.com/tmux/tmux).
type Tmux struct {
	path   string
	paneID string
}

var _ io.Writer = &Tmux{}

const (
	// PgmName is the name of the tmux executable.
	PgmName = "tmux"
	// SessionName is the string to use when naming a tmux session.
	SessionName = utils.PgmName
)

// NewTmux is a ctor.
func NewTmux(programName string) (*Tmux, error) {
	p, err := exec.LookPath(programName)
	if err != nil {
		return nil, err
	}
	return &Tmux{p, "0"}, nil
}

// IsUp true if tmux appears to be running.
func (tx Tmux) IsUp() bool {
	cmd := exec.Command(tx.path, "info")
	bs, err := cmd.CombinedOutput()
	if err == nil {
		return true
	}
	out := string(bs)
	slog.Info("info output", "out", out)
	// The following might not be reliable.  See
	// https://github.com/tmuxinator/tmuxinator/issues/536
	return strings.TrimSpace(out) == "no current client"
}

// Write bytes to a tmux session for interpretation as shell commands.
// Uses this kludge:
// * writes bytes to a temp file,
// * tells tmux to load that file into a tmux paste buffer,
// * then tells tmux to 'paste' it into a session.
// * yay tmux!
// TODO: look for a better tmux api (dbus?)
func (tx Tmux) Write(bytes []byte) (n int, err error) {
	var tmpFile *os.File
	tmpFile, err = os.CreateTemp("", SessionName+"-block-")
	check("create temp file", err)
	check("chmod temp file", os.Chmod(tmpFile.Name(), 0644))
	defer func() {
		slog.Info("Using", "tmpFile", tmpFile.Name())
		check("delete temp file", os.Remove(tmpFile.Name()))
	}()

	n, err = tmpFile.Write(bytes)
	if err != nil {
		slog.Error("Could not write tmp file", "err", err)
	}
	if n != len(bytes) {
		slog.Error(
			"Could not write bytes", "n", n, "len(bytes)", len(bytes))
	}
	var out []byte

	cmd := exec.Command(tx.path, "load-buffer", tmpFile.Name())
	out, err = cmd.Output()
	if err == nil {
		cmd = exec.Command(tx.path, "paste-buffer", "-t", tx.paneID)
		out, err = cmd.Output()
	}
	if err != nil {
		slog.Error("failed cmd", "args", cmd.Args, "out", out)
	}
	return len(bytes), err
}

func (tx Tmux) Start() error {
	cmd := exec.Command(tx.path, "new-session", "-s", SessionName, "-d")
	out, err := cmd.Output()
	slog.Info("start", "out", out)
	slog.Info("start", "err", err)
	return err
}

func (tx Tmux) Stop() error {
	cmd := exec.Command(tx.path, "kill-session", "-t", SessionName)
	out, err := cmd.Output()
	slog.Info("stop", "out", out)
	return err
}

func (tx Tmux) ListSessions() (string, error) {
	cmd := exec.Command(tx.path, "list-sessions")
	raw, err := cmd.Output()
	slog.Info("List", "raw", string(raw))
	return string(raw), err
}

// check reports the error fatally if it's non-nil.
func check(msg string, err error) {
	if err != nil {
		panic(fmt.Errorf("%s; %w", msg, err))
	}
}
