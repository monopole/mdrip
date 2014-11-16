package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
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

const signal = "CytochromeCPathIntegralLonelyMountainCosmicBackground"

// scanErr watches and parses stdErr.
func scanErr(stdErr io.ReadCloser, results chan *textBucket) {
	defer close(results)
	var accum bytes.Buffer
	errScanner := bufio.NewScanner(stdErr)
	for errScanner.Scan() {
		line := errScanner.Text()
		accum.WriteString(line + "\n")
	}
	trailing := strings.TrimSpace(accum.String())
	if len(trailing) > 0 {
		results <- &textBucket{false, trailing}
	}
}

// scanOut watches and parses stdOut.
func scanOut(stdOut io.ReadCloser, results chan *textBucket) {
	defer close(results)
	var accum bytes.Buffer
	outScanner := bufio.NewScanner(stdOut)
	for outScanner.Scan() {
		line := outScanner.Text()
		if strings.HasPrefix(line, signal) {
			results <- &textBucket{true, accum.String()}
			accum.Reset()
		} else {
			accum.WriteString(line + "\n")
		}
	}
	trailing := strings.TrimSpace(accum.String())
	err := outScanner.Err()
	if err != nil || len(trailing) > 0 {
		results <- &textBucket{false, trailing}
	}
}

type ErrorBucket struct {
	textBucket
	index   int
	script  string
	err     error
	message string
}

// userBehavior writes N scripts to shell and looks at the result.
func userBehavior(stdIn io.Writer, scripts []string, chOut, chErr chan *textBucket) (errResult *ErrorBucket) {
	errResult = &ErrorBucket{textBucket{false, ""}, -1, "", nil, ""}
	for i, script := range scripts {
		_, err := stdIn.Write([]byte(script))
		check("write script", err)
		_, err = stdIn.Write([]byte("\necho " + signal + "\n"))
		check("write signal", err)
		result := <-chOut
		if result == nil || !result.success {
			errResult.index = i
			errResult.script = script
			if result != nil {
				errResult.output = result.output
			}
			result = <-chErr
			if result != nil {
				errResult.message = result.output
			}
			return
		}
	}
	_, err := stdIn.Write([]byte("exit\n"))
	check("trouble ending shell", err)
	return
}

func Complain(result *ErrorBucket, label, fileName string) {
	delim := strings.Repeat("-", 70) + "\n"
	fmt.Fprintf(os.Stderr, "Error in script %d from thread label %q of file %q:\n", result.index+1, label, fileName)
	fmt.Fprintf(os.Stderr, delim)
	fmt.Fprintf(os.Stderr, result.script)
	fmt.Fprintf(os.Stderr, delim)
	fmt.Fprintf(os.Stderr, "\nStdout capture:\n")
	fmt.Fprintf(os.Stderr, delim)
	fmt.Fprintf(os.Stderr, result.output)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, delim)
	if len(result.message) > 0 {
		fmt.Fprintf(os.Stderr, "\nStderr capture:\n")
		fmt.Fprintf(os.Stderr, delim)
		fmt.Fprintf(os.Stderr, result.message)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, delim)
	}
}

// Runs N scripts in subprocess, reporting failure if any.
//
// The 'scripts' are strings holding opaque shell code.  The strings
// may be more complex than single commands delimitted by linefeeds -
// e.g. scripts that operate on HERE documents, or multi-line commands
// using line continuation via '\', quotes or curly brackets.  This
// code is not a shell interpreter, so the 'atom' of success or
// failure is an entire script (any one of the N scripts).  The atom
// won't be a 'line' since lines won't be parsed or known.
func RunInSubShell(scripts []string) (result *ErrorBucket) {
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

	chOut := make(chan *textBucket)
	go scanOut(stdOut, chOut)

	chErr := make(chan *textBucket)
	go scanErr(stdErr, chErr)

	result = userBehavior(stdIn, scripts, chOut, chErr)
	result.err = shell.Wait()
	return
}
