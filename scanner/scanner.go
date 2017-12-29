package scanner

import (
	"bufio"
	"github.com/golang/glog"
	"io"
	"time"
)

// Special strings that might appear in shell output, signalling
// things to the stream processors.

// MsgHappy indicates no-anomaly completion of a command block.
const MsgHappy = "MDRIP_HAPPY_Completed_command_block"

// MsgError indicates something went wrong in processing the command block.
// This is not necessarily the indication of a command failure, and may only
// indicate a logic or stream error in mdrip.
const MsgError = "MDRIP_ERROR_Problem_while_processing_command_block"

// MsgTimeout indicates the block didn't complete fast enough.
const MsgTimeout = "MDRIP_TIMEOUT_Command_block_did_not_finish_in_allotted_time"

// convertStreamToLineChannel returns a string channel to which it writes _lines_.
// The lines are pulled from an IO stream.
//
// Basically, this function just converts the problem of reading lines of text
// from a stream to one of reading strings from a channel. Stream errors are
// converted to special messages on the channel.
//
// When the stream ends, the channel closes.
func convertStreamToLineChannel(label string, stream io.ReadCloser) <-chan string {
	ch := make(chan string, 1)
	go func() {
		defer close(ch)
		scanner := bufio.NewScanner(stream)
		if glog.V(2) {
			glog.Infof("stream: %s - starting scan loop", label)
		}
		for scanner.Scan() {
			line := scanner.Text()
			if glog.V(2) {
				glog.Infof("stream %s: forwarding \"%s\"", label, line)
			}
			ch <- line
		}
		if err := scanner.Err(); err != nil {
			if glog.V(2) {
				glog.Infof("stream %s: error in scanner : %s", label, err.Error())
			}
			ch <- MsgError + " : " + err.Error()
			return
		}
		if glog.V(2) {
			glog.Infof("stream %s: done reading", label)
		}
	}()
	return ch
}

// BuffScanner returns a channel to which it will write lines of text.
//
// The text is harvested from an io stream. If the io stream blocks for longer
// than the given wait time, BuffScanner will send a special line of text to
// the channel and close it.
func BuffScanner(wait time.Duration, label string, stream io.ReadCloser) <-chan string {
	chIn := convertStreamToLineChannel(label, stream)
	chOut := make(chan string, 1)
	go func() {
		defer close(chOut)
		for {
			select {
			case line, ok := <-chIn:
				if ok {
					chOut <- line
				} else {
					if glog.V(2) {
						glog.Infof("buffScanner: %s closed normally", label)
					}
					chIn = nil
					return
				}
			case <-time.After(wait):
				chOut <- MsgTimeout
				if glog.V(2) {
					glog.Infof("buffScanner: %s timed out after %v!", label, wait)
				}
				return
			}
		}
	}()
	return chOut
}
