package util

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// getProcessGroupId purports to get a process group Id common to all
// subprocesses of its pid argument.
//
// There should be a better way to do this.
//
// Goal is to be able to support killing any subprocesses created by
// RunInSubShell.  At the moment, its up to command script authors to
// clean up after themselves.
func getProcesssGroupId(pid int) (int, error) {
	//  /bin/ps -o pid,pgid,rgid,ppid,cmd
	//  /bin/ps -o pgid=12492 --no-headers
	cmdOut, execErr := exec.Command(
		"/bin/ps", "--pid", strconv.Itoa(pid), "-o", "pgid", "--no-headers").Output()
	groupId := strings.TrimSpace(string(cmdOut))
	if execErr != nil || len(groupId) < 1 {
		return 0, errors.New(
			"Unable to yank groupId from ps command: " + groupId + " " + execErr.Error())
	}
	pgid, convErr := strconv.Atoi(groupId)
	if convErr != nil {
		return 0, convErr
	}
	return pgid, nil
}

// An attempt to kill any and all child processes.
func killProcesssGroup(pgid int) {
	killer := exec.Command("/bin/kill", "-TERM", "--", fmt.Sprintf("-%v", pgid))
	killer.Start()
}
