package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func check(msg string, err error) {
	if err != nil {
		fmt.Printf("Problem with %s\n", msg, err)
		log.Fatal(err)
	}
}

type textBucket struct {
	success bool
	output  string
}

// Special strings that might appear in shell output.
const msgHappy = "MDRIP HAPPY!"
const msgTimeout = "MDRIP TIMEOUT!"
const msgError = "MDRIP ERROR!"

// An uninterruptable stream scanner.
// If Scan hangs for some reason, so will this.
func scanBuffer(stream io.ReadCloser) chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		errScanner := bufio.NewScanner(stream)
		for errScanner.Scan() {
			out <- errScanner.Text()
		}
		if err := errScanner.Err(); err != nil {
			out <- msgError + " : " + err.Error()
		}
	}()
	return out
}

// Accumulates strings off a channel until the channel closes, taking
// action if the strings match a particular pattern.  On the happy
// path, strings are accumulated and every so often sent on to another
// channel with a success == true flag attached.  On a sad path, an
// accumulation of strings is sent with a success == false flag
// attached, and the function exits early, before it's input channel
// closes.
func accumulateOutput(prefix string, in <-chan string) <-chan *textBucket {
	out := make(chan *textBucket)
	var accum bytes.Buffer
	go func() {
		defer close(out)
		for line := range in {
			if *debug {
				fmt.Printf("DEBUG: %s: %s\n", prefix, line)
			}
			if strings.HasPrefix(line, msgTimeout) {
				accum.WriteString("\n" + line + "\n")
				accum.WriteString("A subprocess might still be running.\n")
				out <- &textBucket{false, accum.String()}
				if *debug {
					fmt.Printf("DEBUG: %s: Timeout return.\n", prefix)
				}
				return
			}
			if strings.HasPrefix(line, msgError) {
				accum.WriteString(line + "\n")
				out <- &textBucket{false, accum.String()}
				if *debug {
					fmt.Printf("DEBUG: %s Error return.\n", prefix)
				}
				return
			}
			if strings.HasPrefix(line, msgHappy) {
				out <- &textBucket{true, accum.String()}
				accum.Reset()
			} else {
				accum.WriteString(line + "\n")
			}
		}
		trailing := strings.TrimSpace(accum.String())
		if len(trailing) > 0 {
			// Should only be true if the loop above terminated on an error.
			out <- &textBucket{false, accum.String()}
		}
	}()
	if *debug {
		fmt.Printf("DEBUG: %s Done with accumulateOutput.\n", prefix)
	}
	return out
}

func supplyTimeout(ch1, ch2 chan string, label string, doneCh <-chan bool) {
	stepSize := 100 * time.Millisecond
	limit := 50
	if *debug {
		totalTime := time.Duration(limit) * stepSize
		fmt.Printf("DEBUG: Timeout countdown of %v (step %s) starting for %s\n", totalTime, stepSize, label)
	}
	for i := 0; i < limit; i++ {
		time.Sleep(stepSize)
		select {
		case <-doneCh:
			if *debug {
				fmt.Printf("DEBUG: Timeout cancelled for %s\n", label)
			}
			return
		default:
		}
	}
	if *debug {
		fmt.Printf("DEBUG: Timeout expired on %s, punching in face.\n", label)
	}
	ch1 <- msgTimeout
	ch2 <- msgTimeout
}

type ErrorBucket struct {
	textBucket
	fileName string
	index    int
	block    *codeBlock
	err      error
	message  string
}

type ScriptBucket struct {
	fileName string
	script   []*codeBlock
}

var emptyArray []string = []string{}
var emptyCodeBlock *codeBlock = &codeBlock{emptyArray, ""}

// Writes command blocks to shell.  Attempts to wait after each block
// to see if the block worked.  If the block appeared to complete
// without error, the routine sends the next block, else it exits
// early.
func userBehavior(stdIn io.Writer, scriptBuckets []*ScriptBucket,
	chOut, chErr chan string) (errResult *ErrorBucket) {
	emptyArray := []string{}
	chAccErr := accumulateOutput("stdErr", chErr)
	chAccOut := accumulateOutput("stdOut", chOut)
	errResult = &ErrorBucket{textBucket{false, ""}, "", -1, &codeBlock{emptyArray, ""}, nil, ""}
	for _, bucket := range scriptBuckets {
		for i, block := range bucket.script {
			fmt.Printf("Running %s (%d/%d) from %s\n",
				block.labels[0], i+1, len(bucket.script), bucket.fileName)
			_, err := stdIn.Write([]byte(block.codeText))
			check("write script", err)
			_, err = stdIn.Write([]byte("\necho " + msgHappy + "\n"))
			check("write msgHappy", err)
			doneCh := make(chan bool, 1)
			go supplyTimeout(chOut, chErr, block.labels[0], doneCh)
			// The following won't block, because a timeout will happen.
			result := <-chAccOut
			doneCh <- true
			if result == nil || !result.success {
				errResult.fileName = bucket.fileName
				errResult.index = i
				errResult.block = block
				if result != nil {
					errResult.output = result.output
					if *debug {
						fmt.Printf("DEBUG: stdout Result: %s\n", result.output)
					}
				} else {
					if *debug {
						fmt.Printf("DEBUG: stdout Result == nil.\n")
					}
				}
				result = <-chAccErr
				if result != nil {
					errResult.err = errors.New(result.output)
					errResult.message = result.output
					if *debug {
						fmt.Printf("DEBUG: stderr Result: %s\n", result.output)
					}
				} else {
					errResult.err = errors.New("unknown")
					if *debug {
						fmt.Printf("DEBUG: stderr Result == nil.\n")
					}
				}
				if *debug {
					fmt.Printf("DEBUG: exitting subshell.\n")
				}
				// The shell is likely stalled, but send an exit anyway.
				exitShell(stdIn)
				return
			}
		}
	}
	exitShell(stdIn)
	fmt.Printf("All done, no errors triggered.\n")
	return
}

func exitShell(stdIn io.Writer) {
	_, err := stdIn.Write([]byte("exit\n"))
	check("trouble ending shell", err)
}

func dumpCapturedOutput(name, delim, output string) {
	fmt.Fprintf(os.Stderr, "\n%s capture:\n", name)
	fmt.Fprintf(os.Stderr, delim)
	fmt.Fprintf(os.Stderr, output)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, delim)
}

func Complain(result *ErrorBucket, label string) {
	delim := strings.Repeat("-", 70) + "\n"
	fmt.Fprintf(os.Stderr, "Error in block '%s' (#%d of script '%s') in %s:\n",
		result.block.labels[0], result.index+1, label, result.fileName)
	fmt.Fprintf(os.Stderr, delim)
	fmt.Fprintf(os.Stderr, string(result.block.codeText))
	fmt.Fprintf(os.Stderr, delim)
	dumpCapturedOutput("Stdout", delim, result.output)
	if len(result.message) > 0 {
		dumpCapturedOutput("Stderr", delim, result.message)
	}
}

// To support killing on error any subprocesses created by
// RunInSubShell, find the process group, the idea being to kill all
// processes in that group if the shell exits abnormally.
//
// There should be a better way to do this.
func getProcesssGroupId(pid int) (int, error) {
	//  /bin/ps -o pid,pgid,rgid,ppid,cmd
	//  /bin/ps -o pgid=12492 --no-headers
	cmdOut, execErr := exec.Command("/bin/ps", "--pid", strconv.Itoa(pid), "-o", "pgid", "--no-headers").Output()
	groupId := strings.TrimSpace(string(cmdOut))
	if execErr != nil || len(groupId) < 1 {
		return 0, errors.New("Unable to yank groupId from ps command: " + groupId + " " + execErr.Error())
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

// Runs code blocks in a subprocess, reporting failure if any.
//
// Code blocks are strings holding opaque shell code.  The strings may
// be more complex than single commands delimitted by linefeeds -
// e.g. blocks that operate on HERE documents, or multi-line commands
// using line continuation via '\', quotes or curly brackets.  This
// code is not a shell interpreter, so the 'atom' of success or
// failure is an entire code block.  The atom won't be a 'line' since
// lines won't be parsed or known.
func RunInSubShell(scriptBuckets []*ScriptBucket) (result *ErrorBucket) {
	// Add "-e" to have shell die on any error.
	shell := exec.Command("bash", "-e")

	stdOut, err := shell.StdoutPipe()
	check("out pipe", err)

	stdErr, err := shell.StderrPipe()
	check("err pipe", err)

	stdIn, err := shell.StdinPipe()
	check("in pipe", err)

	err = shell.Start()
	check("shell start", err)

	chOut := scanBuffer(stdOut)
	chErr := scanBuffer(stdErr)

	pid := shell.Process.Pid
	if *debug {
		fmt.Printf("DEBUG: pid = %d\n", pid)
	}
	pgid, err := getProcesssGroupId(pid)
	if err == nil {
		if *debug {
			fmt.Printf("DEBUG: pgid = %d\n", pgid)
		}
	}

	result = userBehavior(stdIn, scriptBuckets, chOut, chErr)

	if *debug {
		fmt.Printf("DEBUG: Waiting for shell to end.\n")
	}
	waitError := shell.Wait()
	if result.err == nil {
		result.err = waitError
	}
	if *debug {
		fmt.Printf("DEBUG: Shell done.\n")
	}

	// killProcesssGroup(pgid)
	return
}
