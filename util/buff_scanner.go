package util

import (
	"bufio"
	"fmt"
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
func BuffScanner(wait time.Duration, label string, stream io.ReadCloser, debug bool) <-chan string {
	chLine := make(chan string, 1)

	xScanner := func() <-chan string {
		chBuffLine := make(chan string, 1)
		go func() {
			defer close(chBuffLine)
			scanner := bufio.NewScanner(stream)
			if debug {
				fmt.Printf("DEBUG: xScanner: %s - starting up\n", label)
			}
			for scanner.Scan() {
				if debug {
					fmt.Printf("DEBUG: xScanner: %s - calling Text\n", label)
				}
				line := scanner.Text()
				if debug {
					fmt.Printf("DEBUG: xScanner: %s - got \"%s\"\n", label, line)
				}
				chBuffLine <- line
				if debug {
					fmt.Printf("DEBUG: xScanner: %s - handed \"%s\" to channel\n", label, line)
				}
			}

			if debug {
				fmt.Printf("DEBUG: xScanner: %s - exitted Scan loop\n", label)
			}
			if err := scanner.Err(); err != nil {
				if debug {
					fmt.Printf("DEBUG: xScanner: %s - error : %s\n", label, err.Error())
				}
				chBuffLine <- MsgError + " : " + err.Error()
			}
			if debug {
				fmt.Printf("DEBUG: xScanner: %s - completely done\n", label)
			}
		}()
		return chBuffLine
	}

	chBuffLine := xScanner()

	go func() {
		defer close(chLine)
		for {
			if debug {
				fmt.Printf("DEBUG: buffScanner: %s - top of loop\n", label)
			}
			select {
			case line, ok := <-chBuffLine:
				if ok {
					if debug {
						fmt.Printf("DEBUG: buffScanner: %s - got line, sending on\n", label)
					}
					chLine <- line
					if debug {
						fmt.Printf("DEBUG: buffScanner: %s - sent line\n", label)
					}
				} else {
					if debug {
						fmt.Printf("DEBUG: buffScanner: %s - done reading the stream\n", label)
					}
					chBuffLine = nil
					return
				}
			case <-time.After(wait):
				chLine <- MsgTimeout
				if debug {
					fmt.Printf("DEBUG: buffScanner: %s - timed out\n", label)
				}
				return
			}
		}
	}()

	return chLine
}
