package tmux

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/gorilla/websocket"
)

type Tmux struct {
	path   string
	paneId string
}

const (
	Path        = "/usr/bin/tmux"
	SessionName = "mdrip"
)

func NewTmux(programName string) *Tmux {
	return &Tmux{programName, "0"}
}

func IsProgramInstalled(programName string) bool {
	_, err := exec.LookPath(programName)
	return err == nil
}

func (t Tmux) IsUp() bool {
	if _, err := exec.LookPath(t.path); err != nil {
		return false
	}
	cmd := exec.Command(t.path, "info")
	if _, err := cmd.CombinedOutput(); err != nil {
		return false
	}
	return true
}

func closeSocket(c *websocket.Conn, done chan struct{}) {
	defer c.Close()
	// Send a close frame, wait for the other side to close the connection.
	err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		glog.Error("write close:", err)
		return
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
}

func (t Tmux) Adapt(addr string) {
	done := make(chan struct{})

	glog.Info("connecting to ", addr)

	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		glog.Fatal("dial: ", err)
	}
	defer closeSocket(c, done)

	messages := make(chan []byte)

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				glog.Info("error on read:", err)
				return
			}
			messages <- message
		}
	}()

	err = c.WriteMessage(
		websocket.TextMessage, []byte("greetings from mdrip --mode tmux"))
	if err != nil {
		glog.Error("trouble saying hello:", err)
		return
	}

	for {
		select {
		case m := <-messages:
			glog.Info("received: ", string(m))
			t.Write(m)
			// TODO: Cancel previous timeout, start new one ??
		case <-done:
			return
		case <-time.After(10 * time.Minute):
			// *Always* stop after this super-timeout.
			return
		}
	}
}

// Write bytes to a tmux session for interpretation as shell commands.
//
// Uses this kludge:
//
//  writes bytes to a temp file,
//
//  tells tmux to load that file into a tmux paste buffer,
//
//  then tells tmux to 'paste' it into a session for what looks a lot
//  like use-behavior.  yay tmux.
//
// TODO: look for a better tmux api (dbus?)
func (t Tmux) Write(bytes []byte) (n int, err error) {
	tmpFile, err := ioutil.TempFile("", "mdrip-block-")
	check("create temp file", err)
	check("chmod temp file", os.Chmod(tmpFile.Name(), 0644))
	defer func() {
		glog.Info("Used temp file ", tmpFile.Name())
		check("delete temp file", os.Remove(tmpFile.Name()))
	}()

	n, err = tmpFile.Write(bytes)
	if err != nil {
		glog.Fatalf("Could not write %d bytes: %v", len(bytes), err)
	}
	if n != len(bytes) {
		glog.Fatalf("Expected to write %d bytes, wrote %d", len(bytes), n)
	}

	cmd := exec.Command(t.path, "load-buffer", tmpFile.Name())
	out, err := cmd.Output()
	if err == nil {
		cmd = exec.Command(t.path, "paste-buffer", "-t", t.paneId)
		out, err = cmd.Output()
	}

	if err != nil {
		glog.Info("cmd = ", cmd.Args)
		glog.Info("out = ", out)
	}
	return len(bytes), err
}

func (t Tmux) start() error {
	cmd := exec.Command(t.path, "new", "-s", SessionName, "-d")
	out, err := cmd.Output()
	glog.Info("Starting ", out)
	return err
}

func (t Tmux) stop() error {
	cmd := exec.Command(t.path, "kill-session", "-t", SessionName)
	out, err := cmd.Output()
	glog.Info("Stopping ", out)
	return err
}

func (t Tmux) listSessions() (string, error) {
	cmd := exec.Command(t.path, "list-sessions")
	raw, err := cmd.Output()
	glog.Info("List ", string(raw))
	return string(raw), err
}

// check reports the error fatally if it's non-nil.
func check(msg string, err error) {
	if err != nil {
		fmt.Printf("Problem with %s: %v\n", msg, err)
		glog.Fatal(err)
	}
}
