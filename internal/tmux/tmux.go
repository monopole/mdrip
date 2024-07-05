package tmux

import (
	"github.com/monopole/mdrip/v2/internal/utils"
	"log/slog"
	"os"
	"os/exec"
)

// Tmux holds information about a tmux process (https://github.com/tmux/tmux).
type Tmux struct {
	path   string
	paneID string
}

const (
	// Path is the default path to the tmux executable on disk.
	Path = "/usr/bin/tmux"
	// SessionName is the string to use when naming a tmux session.
	SessionName = "mdrip"
)

// NewTmux is a ctor.
func NewTmux(programName string) *Tmux {
	return &Tmux{programName, "0"}
}

// IsProgramInstalled checks for tmux.
func IsProgramInstalled(programName string) bool {
	_, err := exec.LookPath(programName)
	return err == nil
}

// IsUp true if tmux appears to be running.
func (tx Tmux) IsUp() bool {
	if _, err := exec.LookPath(tx.path); err != nil {
		slog.Info("Unable to find tmux")
		return false
	}
	cmd := exec.Command(tx.path, "info")
	if o, err := cmd.CombinedOutput(); err != nil {
		slog.Error("Unable to run tmux", "err", err)
		slog.Info("info output", "out", string(o))
		// This isn't right.  See
		// https://github.com/tmuxinator/tmuxinator/issues/536
		x := string(o)
		if x[0:len(x)-1] == "no current client" {
			return true
		}
		return false
	}
	return true
}

// Write bytes to a tmux session for interpretation as shell commands.
//
// Uses this kludge:
//
//		writes bytes to a temp file,
//
//		tells tmux to load that file into a tmux paste buffer,
//
//		then tells tmux to 'paste' it into a session.
//
//	 yay tmux!
//
// TODO: look for a better tmux api (dbus?)
func (tx Tmux) Write(bytes []byte) (n int, err error) {
	var tmpFile *os.File
	tmpFile, err = os.CreateTemp("", SessionName+"-block-")
	utils.Check("create temp file", err)
	utils.Check("chmod temp file", os.Chmod(tmpFile.Name(), 0644))
	defer func() {
		slog.Info("Using", "tmpFile", tmpFile.Name())
		utils.Check("delete temp file", os.Remove(tmpFile.Name()))
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

func (tx Tmux) start() error {
	cmd := exec.Command(tx.path, "new-session", "-s", SessionName, "-d")
	out, err := cmd.Output()
	slog.Info("start", "out", out)
	slog.Info("start", "err", err)
	return err
}

func (tx Tmux) stop() error {
	cmd := exec.Command(tx.path, "kill-session", "-t", SessionName)
	out, err := cmd.Output()
	slog.Info("stop", "out", out)
	return err
}

func (tx Tmux) listSessions() (string, error) {
	cmd := exec.Command(tx.path, "list-sessions")
	raw, err := cmd.Output()
	slog.Info("List", "raw", string(raw))
	return string(raw), err
}
