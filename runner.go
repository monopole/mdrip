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
	chOut, chErr chan *textBucket) (errResult *ErrorBucket) {
	emptyArray := []string{}
	errResult = &ErrorBucket{textBucket{false, ""}, "", -1, &codeBlock{emptyArray, ""}, nil, ""}
	for _, bucket := range scriptBuckets {
		for i, block := range bucket.script {
			_, err := stdIn.Write([]byte(block.codeText))
			check("write script", err)
			_, err = stdIn.Write([]byte("\necho " + signal + "\n"))
			check("write signal", err)
			result := <-chOut
			if result == nil || !result.success {
				errResult.fileName = bucket.fileName
				errResult.index = i
				errResult.block = block
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
	}
	_, err := stdIn.Write([]byte("exit\n"))
	check("trouble ending shell", err)
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
	fmt.Fprintf(os.Stderr, "Error in block %d from label %q of file %q:\n",
		result.index+1, label, result.fileName)
	fmt.Fprintf(os.Stderr, delim)
	fmt.Fprintf(os.Stderr, string(result.block.codeText))
	fmt.Fprintf(os.Stderr, delim)
	complain("Stdout", delim, result.output)
	if len(result.message) > 0 {
		complain("Stderr", delim, result.message)
	}
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

	chOut := make(chan *textBucket)
	go scanOut(stdOut, chOut)

	chErr := make(chan *textBucket)
	go scanErr(stdErr, chErr)

	result = userBehavior(stdIn, scriptBuckets, chOut, chErr)
	result.err = shell.Wait()
	return
}
