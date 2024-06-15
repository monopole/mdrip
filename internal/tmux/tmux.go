package tmux

import (
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/gorilla/websocket"
	"github.com/monopole/mdrip/v2/internal/utils"
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

func closeSocket(c *websocket.Conn, done chan struct{}) {
	defer c.Close()
	// Send a close frame, wait for the other side to close the connection.
	err := c.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		slog.Error("write close:", err)
		return
	}
	select {
	case <-done:
		slog.Info("closing socket per done signal")
	case <-time.After(60 * time.Second):
		slog.Info("closing socket per timeout")
	}
}

// Adapt opens a websocket to the given address, and sends what it gets to tmux.
// TODO: THIS STUFF ABANDONED FOR NOW AS THE USE CASE IS QUESTIONABLE.
func (tx Tmux) Adapt(addr string) {
	done := make(chan struct{})

	slog.Info("connecting", "addr", addr)

	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		slog.Error("dial: ", err)
		panic(err)
	}
	slog.Info("dial succeeded")
	defer closeSocket(c, done)

	messages := make(chan []byte)

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				slog.Info("error on read", "err", err)
				slog.Info("message with error", "message", message)
				return
			}
			slog.Info("message received")
			messages <- message
		}
	}()

	err = c.WriteMessage(
		websocket.TextMessage,
		[]byte("greetings from "+SessionName+" --mode tmux"))
	if err != nil {
		slog.Error("trouble saying hello:", err)
		return
	}
	slog.Info("sent hello message")

	for {
		select {
		case m := <-messages:
			n := string(m)
			if len(n) > 40 {
				n = n[:40] + "..."
			}
			slog.Info("received for execution", "n", n)
			tx.Write(m)
			slog.Info("sent for execution")
			// TODO: Cancel previous timeout, start new one ??
		case <-done:
			slog.Info("done signal found")
			return
		case <-time.After(10 * time.Minute):
			slog.Info("backstop timeout expired")
			return
		}
	}
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
