package tmux

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/golang/glog"
)

type Tmux struct {
	paneId string
}

const (
	programName = "tmux"
)

func NewTmux() *Tmux {
	return &Tmux{"0"}
}

func (t Tmux) Initialize() {
	_, err := exec.LookPath(programName)
	if err != nil {
		fmt.Printf("Unable to find %s: %v\n", programName, err)
		glog.Fatal(err)
	}
	fmt.Println("Be sure tmux is running.")
	fmt.Printf("Sending commands to tmux pane Id %s.\n", t.paneId)
}

// Write bytes to a tmux session for interpretation / execution as shell commands.
//
// Uses a kludge to write the block to a temp file, then tell tmux to
// load that file into a tmux paste buffer then 'paste' it into a
// session for what looks a lot like an intuitive user-directed
// action.
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

	cmd := exec.Command(programName, "load-buffer", tmpFile.Name())
	out, err := cmd.Output()
	if err == nil {
		cmd = exec.Command(programName, "paste-buffer", "-t", t.paneId)
		out, err = cmd.Output()
	}

	if err != nil {
		glog.Info("cmd = ", cmd.Args)
		glog.Info("out = ", out)
	}
	return len(bytes), err
}

// check reports the error fatally if it's non-nil.
func check(msg string, err error) {
	if err != nil {
		fmt.Printf("Problem with %s: %v\n", msg, err)
		glog.Fatal(err)
	}
}
