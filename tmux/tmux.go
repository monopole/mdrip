package tmux

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/golang/glog"
)

type Tmux struct {
	programName string
	paneId      string
	okay        bool
}

const (
	ProgramName = "tmux"
	SessionName = "mdrip"
)

func NewTmux(programName string) *Tmux {
	return &Tmux{programName, "0", true}
}

func (t Tmux) Ok() bool {
	return t.okay
}

func (t Tmux) Refresh() error {
	_, err := exec.LookPath(t.programName)
	if err != nil {
		fmt.Printf("Unable to find %s: %v\n", t.programName, err)
		t.okay = false
		return err
	}
	fmt.Printf("Be sure %s is running.\n", t.programName)
	fmt.Printf("Sending commands to %s pane Id %s.\n", t.programName, t.paneId)
	return nil
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

	cmd := exec.Command(t.programName, "load-buffer", tmpFile.Name())
	out, err := cmd.Output()
	if err == nil {
		cmd = exec.Command(t.programName, "paste-buffer", "-t", t.paneId)
		out, err = cmd.Output()
	}

	if err != nil {
		glog.Info("cmd = ", cmd.Args)
		glog.Info("out = ", out)
	}
	return len(bytes), err
}

func (t Tmux) Start() error {
	cmd := exec.Command(t.programName, "new", "-s", SessionName, "-d")
	out, err := cmd.Output()
	glog.Info("Starting ", out)
	return err
}

func (t Tmux) Stop() error {
	cmd := exec.Command(t.programName, "kill-session", "-t", SessionName)
	out, err := cmd.Output()
	glog.Info("Stopping ", out)
	return err
}

func (t Tmux) ListSessions() (string, error) {
	cmd := exec.Command(t.programName, "list-sessions")
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
