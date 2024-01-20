package util

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// GetProcesssGroupID purports to get a process group Id common to all
// subprocesses of its pid argument.
//
// There should be a better way to do this.
//
// Goal is to be able to support killing any subprocesses created by
// RunInSubShell.  At the moment, its up to authors to clean up after
// themselves.
func GetProcesssGroupID(pid int) (int, error) {
	//  /bin/ps -o pid,pgid,rgid,ppid,cmd
	//  /bin/ps -o pgid=12492 --no-headers
	cmdOut, execErr := exec.Command(
		"/bin/ps", "--pid", strconv.Itoa(pid), "-o", "pgid", "--no-headers").Output()
	groupID := strings.TrimSpace(string(cmdOut))
	if execErr != nil || len(groupID) < 1 {
		return 0, errors.New(
			"Unable to yank groupID from ps command: " + groupID + " " + execErr.Error())
	}
	pgid, convErr := strconv.Atoi(groupID)
	if convErr != nil {
		return 0, convErr
	}
	return pgid, nil
}

// Check reports the error fatally if it's non-nil.
func Check(msg string, err error) {
	if err != nil {
		glog.Fatal(errors.Wrap(err, msg))
	}
}

// An attempt to kill any and all child processes.
func killProcesssGroup(pgid int) {
	killer := exec.Command("/bin/kill", "-TERM", "--", fmt.Sprintf("-%v", pgid))
	killer.Start()
}

// Convert tabs, newlines, etc. to normal blanks.
func convertBadWhiteSpaceToBlanks(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case 0x000A, 0x000B, 0x000C, 0x000D, 0x0085, 0x2028, 0x2029:
			return ' '
		default:
			return r
		}
	}, s)
}

var leading = regexp.MustCompile("^[0-9]+_")

// DropLeadingNumbers drops leading numbers and underscores.
func DropLeadingNumbers(s string) string {
	r := leading.FindStringIndex(s)
	if r == nil {
		return s
	}
	return s[r[1]:]
}

// SampleString converts a long multi-line string to a short one-line sample.
func SampleString(incoming string, max int) string {
	s := len(incoming)
	if s > max {
		s = max
	}
	return convertBadWhiteSpaceToBlanks(strings.TrimSpace(incoming[:s]))
}

// Spaces returns a string of length n with only spaces.
func Spaces(n int) string {
	if n < 1 {
		return ""
	}
	return fmt.Sprintf("%"+strconv.Itoa(n)+"s", " ")
}
