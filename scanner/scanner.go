package scanner

import (
	"bufio"
	"github.com/golang/glog"
	"io"
	"time"
)

// Special strings that might appear in shell output, signalling
// things to the stream processors.
const MsgHappy = "MDRIP_HAPPY_Completed_command_block"
const MsgError = "MDRIP_ERROR_Problem_while_executing_command_block"
const MsgTimeout = "MDRIP_TIMEOUT_Command_block_did_not_finish_in_allotted_time"

// BuffScanner returns a channel to which it will write lines of text.
//
// The text is harvested from an io stream, which will be read until
// the io stream hits EOF or otherwise closes - at which point the
// returned channel is closed.
//
// If the io stream blocks for longer than the given wait time, the
// function will send a special line of text to the channel and close
// it.
func BuffScanner(wait time.Duration, label string, stream io.ReadCloser) <-chan string {
	chLine := make(chan string, 1)

	xScanner := func() <-chan string {
		chBuffLine := make(chan string, 1)
		go func() {
			defer close(chBuffLine)
			scanner := bufio.NewScanner(stream)
			if glog.V(2) {
				glog.Infof("stream: %s - starting up", label)
			}

			for scanner.Scan() {
				if glog.V(2) {
					glog.Infof("stream: %s - calling Text", label)
				}
				line := scanner.Text()
				if glog.V(2) {
					glog.Infof("stream: %s - got \"%s\"", label, line)
				}
				chBuffLine <- line
				if glog.V(2) {
					glog.Infof("stream: %s - handed \"%s\" to channel", label, line)
				}
			}
			if glog.V(2) {
				glog.Infof("stream: %s - exited Scan loop", label)
			}

			if err := scanner.Err(); err != nil {
				if glog.V(2) {
					glog.Infof("stream: %s - error : %s", label, err.Error())
				}
				chBuffLine <- MsgError + " : " + err.Error()
			}
			if glog.V(2) {
				glog.Infof("stream: %s - completely done", label)
			}
		}()
		return chBuffLine
	}

	chBuffLine := xScanner()

	go func() {
		defer close(chLine)
		for {
			if glog.V(2) {
				glog.Infof("buffScanner: %s - top of loop", label)
			}
			select {
			case line, ok := <-chBuffLine:
				if ok {
					if glog.V(2) {
						glog.Infof("buffScanner: %s - got line, sending on", label)
					}
					chLine <- line
					if glog.V(2) {
						glog.Infof("buffScanner: %s - sent line", label)
					}
				} else {
					if glog.V(2) {
						glog.Infof("buffScanner: %s - done reading the stream", label)
					}
					chBuffLine = nil
					return
				}
			case <-time.After(wait):
				chLine <- MsgTimeout
				if glog.V(2) {
					glog.Infof("buffScanner: %s - timed out", label)
				}
				return
			}
		}
	}()

	return chLine
}
