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
		errScanner := bufio.NewScanner(stream)
		for errScanner.Scan() {
			out <- errScanner.Text()
		}
		if err := errScanner.Err(); err != nil {
			out <- msgError + " : " + err.Error()
		}
		close(out)
	}()
	return out
}

// Accumulates strings off a channel, taking action if the strings
// match a particular pattern.
func accumulateOutput(in <-chan string) <-chan *textBucket {
	out := make(chan *textBucket)
	var accum bytes.Buffer
	go func() {
		defer close(out)
		for line := range in {
			if strings.HasPrefix(line, msgTimeout) {
				accum.WriteString("\n" + line + "\n")
				accum.WriteString("A subprocess might still be running.\n")
				out <- &textBucket{false, accum.String()}
				return
			}
			if strings.HasPrefix(line, msgError) {
				accum.WriteString(line + "\n")
				out <- &textBucket{false, accum.String()}
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
	return out
}

func supplyTimeout(ch1, ch2 chan string, doneCh <-chan bool) {
	time.Sleep(5 * time.Second)
	select {
	case <-doneCh:
		// The timeout has been cancelled.
		return
	default:
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

// userBehavior writes N scripts to shell and looks at the result.
func userBehavior(stdIn io.Writer, scriptBuckets []*ScriptBucket,
	chOut, chErr chan string) (errResult *ErrorBucket) {
	emptyArray := []string{}
	chAccErr := accumulateOutput(chErr)
	chAccOut := accumulateOutput(chOut)
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
			go supplyTimeout(chOut, chErr, doneCh)
			// The following won't block, because a timeout will happen.
			result := <-chAccOut
			doneCh <- true
			if result == nil || !result.success {
				errResult.fileName = bucket.fileName
				errResult.index = i
				errResult.block = block
				if result != nil {
					errResult.output = result.output
				}
				result = <-chAccErr
				if result != nil {
					errResult.message = result.output
				}
				return
			}
		}
	}
	_, err := stdIn.Write([]byte("exit\n"))
	check("trouble ending shell", err)
	fmt.Printf("All done, no errors triggered.\n")
	return
}

func complain(name, delim, output string) {
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
	complain("Stdout", delim, result.output)
	if len(result.message) > 0 {
		complain("Stderr", delim, result.message)
	}
}

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
	fmt.Printf("pid = %d\n", pid)
	pgid, err := getProcesssGroupId(pid)
	if err == nil {
		fmt.Printf("pgid = %d\n", pgid)
	}

	result = userBehavior(stdIn, scriptBuckets, chOut, chErr)

	result.err = shell.Wait()

	// killProcesssGroup(pgid)
	return
}
