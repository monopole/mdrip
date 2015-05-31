package util

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Should switch to a logging package that supports levels,
// e.g. https://github.com/golang/glog
var debug = flag.Bool("debug", false,
	"If true, dump more information during run.")

// check reports the error fatally if its non-nil.
func check(msg string, err error) {
	if err != nil {
		fmt.Printf("Problem with %s\n", msg, err)
		log.Fatal(err)
	}
}

// blockOutput pairs the output collected from a stream (i.e. stderr
// or stdout) as a result of executing all or part of a command block
// with a bool indicating if the output is associated with shell
// success or shell failure.  Output can appear on stderr without
// neccessarily being associated with shell failure.
type blockOutput struct {
	success bool
	output  string
}

// accumulateOutput returns a channel to which it writes objects that
// contain what purport to be the entire output of one command block.
//
// To do so, it accumulates strings off a channel representing command
// block output until the channel closes, or until a string arrives
// that matches a particular pattern.
//
// On the happy path, strings are accumulated and every so often sent
// out with a success == true flag attached.  This continues until the
// input channel closes.
//
// On a sad path, an accumulation of strings is sent with a success ==
// false flag attached, and the function exits early, before it's
// input channel closes.
func accumulateOutput(prefix string, in <-chan string) <-chan *blockOutput {
	out := make(chan *blockOutput)
	var accum bytes.Buffer
	go func() {
		defer close(out)
		for line := range in {
			if strings.HasPrefix(line, MsgTimeout) {
				accum.WriteString("\n" + line + "\n")
				accum.WriteString("A subprocess might still be running.\n")
				if *debug {
					fmt.Printf("DEBUG: accumulateOutput %s: Timeout return.\n", prefix)
				}
				out <- &blockOutput{false, accum.String()}
				return
			}
			if strings.HasPrefix(line, MsgError) {
				accum.WriteString(line + "\n")
				if *debug {
					fmt.Printf("DEBUG: accumulateOutput %s: Error return.\n", prefix)
				}
				out <- &blockOutput{false, accum.String()}
				return
			}
			if strings.HasPrefix(line, MsgHappy) {
				if *debug {
					fmt.Printf("DEBUG: accumulateOutput %s: %s\n", prefix, line)
				}
				out <- &blockOutput{true, accum.String()}
				accum.Reset()
			} else {
				if *debug {
					fmt.Printf("DEBUG: accumulateOutput %s: Accumulating [%s]\n", prefix, line)
				}
				accum.WriteString(line + "\n")
			}
		}

		if *debug {
			fmt.Printf("DEBUG: accumulateOutput %s: <--- This channel has closed.\n", prefix)
		}
		trailing := strings.TrimSpace(accum.String())
		if len(trailing) > 0 {
			if *debug {
				fmt.Printf(
					"DEBUG: accumulateOutput %s: Erroneous (missing-happy) output [%s]\n", prefix, accum.String())
			}
			out <- &blockOutput{false, accum.String()}
		} else {
			if *debug {
				fmt.Printf("DEBUG: accumulateOutput %s: Nothing trailing.\n", prefix)
			}
		}
	}()
	return out
}

// ScriptResult pairs blockOutput with meta data about shell execution.
type ScriptResult struct {
	blockOutput
	fileName string        // File in which the error occurred.
	index    int           // Command block index.
	block    *CommandBlock // Content of actual command block.
	problem  error         // Error, if any.
	message  string        // Detailed error message, if any.
}

func (x ScriptResult) GetFileName() string {
	return x.fileName
}

func (x ScriptResult) GetProblem() error {
	return x.problem
}

// ScriptBucket associates a list of commandBlocks with the name of the
// file they came from.
type ScriptBucket struct {
	fileName string
	script   []*CommandBlock
}

func (x ScriptBucket) GetFileName() string {
	return x.fileName
}
func (x ScriptBucket) GetScript() []*CommandBlock {
	return x.script
}
func NewScriptBucket(fileName string, script []*CommandBlock) *ScriptBucket {
	return &ScriptBucket{fileName, script}
}

// userBehavior acts like a command line user.
//
// It writes command blocks to shell, then waits after  each block to
// see if the block worked.  If the block appeared to complete without
// error, the routine sends the next block, else it exits early.
func userBehavior(stdIn io.Writer, scriptBuckets []*ScriptBucket, blockTimeout time.Duration,
	stdOut, stdErr io.ReadCloser) (errResult *ScriptResult) {
	emptyArray := []string{}

	chOut := BuffScanner(blockTimeout, "stdout", stdOut, *debug)
	chErr := BuffScanner(1*time.Minute, "stderr", stdErr, *debug)

	chAccOut := accumulateOutput("stdOut", chOut)
	chAccErr := accumulateOutput("stdErr", chErr)

	errResult = &ScriptResult{blockOutput{false, ""}, "", -1, &CommandBlock{emptyArray, ""}, nil, ""}
	for _, bucket := range scriptBuckets {
		for i, block := range bucket.script {
			blockName := block.labels[0]
			fmt.Printf("Running %s (%d/%d) from %s\n",
				blockName, i+1, len(bucket.script), bucket.fileName)
			if *debug {
				fmt.Printf("DEBUG: userBehavior: sending \"%s\"\n", block.codeText)
			}
			_, err := stdIn.Write([]byte(block.codeText))
			check("write script", err)
			if *debug {
				fmt.Printf("DEBUG: userBehavior: sending happy\n")
			}
			_, err = stdIn.Write([]byte("\necho " + MsgHappy + " " + blockName + "\n"))
			check("write msgHappy", err)

			result := <-chAccOut

			if result == nil || !result.success {
				// A nil result means stdout has closed early because a
				// sub-subprocess failed.
				if result == nil {
					if *debug {
						fmt.Printf("DEBUG: userBehavior: stdout Result == nil.\n")
						// fmt.Printf("DEBUG: userBehavior: sending warning to stdErr\n")
					}
					//					chErr <- MsgError + " : early termination; stdout has closed."
				} else {
					if *debug {
						fmt.Printf("DEBUG: userBehavior: stdout Result: %s\n", result.output)
					}
					// Shell may still be alive despite a failure (e.g. an mdrip
					// imposed timeout).  Maybe send exit.
					exitShell(stdIn)
					errResult.output = result.output
					errResult.message = result.output
				}
				errResult.fileName = bucket.fileName
				errResult.index = i
				errResult.block = block
				fillErrResult(chAccErr, errResult)
				return
			}
		}
	}
	exitShell(stdIn)
	fmt.Printf("All done, no errors triggered.\n")
	return
}

// fillErrResult fills an instance of ScriptResult.
func fillErrResult(chAccErr <-chan *blockOutput, errResult *ScriptResult) {
	result := <-chAccErr
	if result == nil {
		if *debug {
			fmt.Printf("DEBUG: userBehavior: stderr Result == nil.\n")
		}
		errResult.problem = errors.New("unknown")
		return
	}
	errResult.problem = errors.New(result.output)
	errResult.message = result.output
	if *debug {
		fmt.Printf("DEBUG: userBehavior: stderr Result: %s\n", result.output)
	}
}

func exitShell(stdIn io.Writer) {
	if *debug {
		fmt.Printf("DEBUG: userBehavior: exiting subshell.\n")
	}
	stdIn.Write([]byte("exit\n"))
	// Don't check for error - it either works, or we'll have
	// already reported a failed shell.
}

func dumpCapturedOutput(name, delim, output string) {
	fmt.Fprintf(os.Stderr, "\n%s capture:\n", name)
	fmt.Fprintf(os.Stderr, delim)
	fmt.Fprintf(os.Stderr, output)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, delim)
}

// Complain spits the contents of a ScriptResult to stderr.
func Complain(result *ScriptResult, label string) {
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

// RunInSubShell runs command blocks in a subprocess, stopping and
// reporting on any error.  The subprocess runs with the -e flag, so
// it will abort if any sub-subprocess (any command) fails.
//
// Command blocks are strings presumably holding code from some shell
// language.  The strings may be more complex than single commands
// delimitted by linefeeds - e.g. blocks that operate on HERE
// documents, or multi-line commands using line continuation via '\',
// quotes or curly brackets.
//
// This function itself is not a shell interpreter, so it has no idea
// if one line of text from a command block is an individual command
// or part of something else.
//
// Error reporting works by discarding output from command blocks that
// succeeded, and only reporting the contents of stdout and stderr
// when the subprocess exits on error.
func RunInSubShell(scriptBuckets []*ScriptBucket, blockTimeout time.Duration) (
	result *ScriptResult) {
	// Adding "-e" to force the subshell to die on any error.
	shell := exec.Command("bash", "-e")

	stdOut, err := shell.StdoutPipe()
	check("out pipe", err)

	stdErr, err := shell.StderrPipe()
	check("err pipe", err)

	stdIn, err := shell.StdinPipe()
	check("in pipe", err)

	err = shell.Start()
	check("shell start", err)

	pid := shell.Process.Pid
	if *debug {
		fmt.Printf("DEBUG: RunInSubShell: pid = %d\n", pid)
	}
	pgid, err := getProcesssGroupId(pid)
	if err == nil {
		if *debug {
			fmt.Printf("DEBUG: RunInSubShell:  pgid = %d\n", pgid)
		}
	}

	result = userBehavior(stdIn, scriptBuckets, blockTimeout, stdOut, stdErr)

	if *debug {
		fmt.Printf("DEBUG: RunInSubShell:  Waiting for shell to end.\n")
	}
	waitError := shell.Wait()
	if result.problem == nil {
		result.problem = waitError
	}
	if *debug {
		fmt.Printf("DEBUG: RunInSubShell:  Shell done.\n")
	}

	// killProcesssGroup(pgid)
	return
}
