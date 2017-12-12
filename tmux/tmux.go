package tmux

import (
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"github.com/monopole/mdrip/util"
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
		glog.Info("Unable to find tmux")
		return false
	}
	cmd := exec.Command(t.path, "info")
	if o, err := cmd.CombinedOutput(); err != nil {
		glog.Info("Unable to run tmux: ", err)
		glog.Info("info output: ", string(o))
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
		glog.Error("write close:", err)
		return
	}
	select {
	case <-done:
		glog.Info("closing socket per done signal")
	case <-time.After(60 * time.Second):
		glog.Info("closing socket per timeout")
	}
}

func (t Tmux) Adapt(addr string) {
	done := make(chan struct{})

	glog.Info("connecting to ", addr)

	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		glog.Fatal("dial: ", err)
	}
	glog.Info("dial succeeded")
	defer closeSocket(c, done)

	messages := make(chan []byte)

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				glog.Infof("error on read: %v", err)
				glog.Infof("message with error: %v", message)
				return
			}
			glog.Info("message received")
			messages <- message
		}
	}()

	err = c.WriteMessage(
		websocket.TextMessage, []byte("greetings from mdrip --mode tmux"))
	if err != nil {
		glog.Error("trouble saying hello:", err)
		return
	}
	glog.Info("sent hello message")

	for {
		select {
		case m := <-messages:
			n := string(m)
			if len(n) > 40 {
				n = n[:40] + "..."
			}
			glog.Info("received for execution: ", n)
			t.Write(m)
			glog.Info("sent for execution")
			// TODO: Cancel previous timeout, start new one ??
		case <-done:
			glog.Info("done signal found")
			return
		case <-time.After(10 * time.Minute):
			glog.Info("backstop timeout expired")
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
	util.Check("create temp file", err)
	util.Check("chmod temp file", os.Chmod(tmpFile.Name(), 0644))
	defer func() {
		glog.Info("Used temp file ", tmpFile.Name())
		util.Check("delete temp file", os.Remove(tmpFile.Name()))
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
	glog.Info("Err: ", err)
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
