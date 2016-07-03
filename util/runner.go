package util

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/monopole/mdrip/model"
	"io"
	"io/ioutil"
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

func checkWrite(output string, writer io.Writer) {
	n, err := writer.Write([]byte(output))
	if err != nil {
		log.Fatalf("Could not write %d bytes: %v", len(output), err)
	} else if n != len(output) {
		log.Fatalf("Expected to write %d bytes, wrote %d", len(output), n)
	}
}

// check reports the error fatally if it's non-nil.
func check(msg string, err error) {
	if err != nil {
		fmt.Printf("Problem with %s: %v\n", msg, err)
		log.Fatal(err)
	}
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
func accumulateOutput(prefix string, in <-chan string) <-chan *model.BlockOutput {
	out := make(chan *model.BlockOutput)
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
				out <- model.NewFailureOutput(accum.String())
				return
			}
			if strings.HasPrefix(line, MsgError) {
				accum.WriteString(line + "\n")
				if *debug {
					fmt.Printf("DEBUG: accumulateOutput %s: Error return.\n", prefix)
				}
				out <- model.NewFailureOutput(accum.String())
				return
			}
			if strings.HasPrefix(line, MsgHappy) {
				if *debug {
					fmt.Printf("DEBUG: accumulateOutput %s: %s\n", prefix, line)
				}
				out <- model.NewSuccessOutput(accum.String())
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
			out <- model.NewFailureOutput(accum.String())
		} else {
			if *debug {
				fmt.Printf("DEBUG: accumulateOutput %s: Nothing trailing.\n", prefix)
			}
		}
	}()
	return out
}

// userBehavior acts like a command line user.
// TODO(monopole): update the comments, as this function no longer writes to stdin.
// See https://github.com/monopole/mdrip/commit/a7be6a6fb62ccf8dfe1c2906515ce3e83d0400d7
//
// It writes command blocks to shell, then waits after  each block to
// see if the block worked.  If the block appeared to complete without
// error, the routine sends the next block, else it exits early.
func userBehavior(scriptBuckets []*model.ScriptBucket, blockTimeout time.Duration,
	stdOut, stdErr io.ReadCloser) (errResult *model.ScriptResult) {

	chOut := BuffScanner(blockTimeout, "stdout", stdOut, *debug)
	chErr := BuffScanner(1*time.Minute, "stderr", stdErr, *debug)

	chAccOut := accumulateOutput("stdOut", chOut)
	chAccErr := accumulateOutput("stdErr", chErr)

	errResult = model.NewScriptResult()
	for _, bucket := range scriptBuckets {
		numBlocks := len(bucket.GetScript())
		for i, block := range bucket.GetScript() {
			fmt.Printf("Running %s (%d/%d) from %s\n",
				block.GetName(), i+1, numBlocks, bucket.GetFileName())
			if *debug {
				fmt.Printf("DEBUG: userBehavior: sending \"%s\"\n", block.GetCode())
			}

			result := <-chAccOut

			if result == nil || !result.Succeeded() {
				// A nil result means stdout has closed early because a
				// sub-subprocess failed.
				if result == nil {
					if *debug {
						fmt.Printf("DEBUG: userBehavior: stdout Result == nil.\n")
					}
					// Perhaps chErr <- MsgError + " : early termination; stdout has closed."
				} else {
					if *debug {
						fmt.Printf("DEBUG: userBehavior: stdout Result: %s\n", result.GetOutput())
					}
					errResult.SetOutput(result.GetOutput()).SetMessage(result.GetOutput())
				}
				errResult.SetFileName(bucket.GetFileName()).SetIndex(i).SetBlock(block)
				fillErrResult(chAccErr, errResult)
				return
			}
		}
	}
	fmt.Printf("All done, no errors triggered.\n")
	return
}

// fillErrResult fills an instance of ScriptResult.
func fillErrResult(chAccErr <-chan *model.BlockOutput, errResult *model.ScriptResult) {
	result := <-chAccErr
	if result == nil {
		if *debug {
			fmt.Printf("DEBUG: userBehavior: stderr Result == nil.\n")
		}
		errResult.SetProblem(errors.New("unknown"))
		return
	}
	errResult.SetProblem(errors.New(result.output)).SetMessage(result.GetOutput())
	if *debug {
		fmt.Printf("DEBUG: userBehavior: stderr Result: %s\n", result.GetOutput())
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
func RunInSubShell(scriptBuckets []*model.ScriptBucket, blockTimeout time.Duration) (
	result *model.ScriptResult) {
	// Write script buckets to a file to be executed.
	scriptFile, err := ioutil.TempFile("", "mdrip-script-")
	check("create temp file", err)
	check("chmod temp file", os.Chmod(scriptFile.Name(), 0744))
	for _, bucket := range scriptBuckets {
		for _, block := range bucket.GetScript() {
			checkWrite(block.GetCode().String(), scriptFile)
			checkWrite("\n", scriptFile)
			checkWrite("echo "+MsgHappy+" "+block.GetName().String()+"\n", scriptFile)
		}
	}
	if *debug {
		fmt.Printf("DEBUG: RunInSubShell: running commands from %s\n", scriptFile.Name())
	}
	defer func() {
		check("delete temp file", os.Remove(scriptFile.Name()))
	}()

	// Adding "-e" to force the subshell to die on any error.
	shell := exec.Command("bash", "-e", scriptFile.Name())

	stdIn, err := shell.StdinPipe()
	check("in pipe", err)
	check("close shell's stdin", stdIn.Close())

	stdOut, err := shell.StdoutPipe()
	check("out pipe", err)

	stdErr, err := shell.StderrPipe()
	check("err pipe", err)

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

	result = userBehavior(scriptBuckets, blockTimeout, stdOut, stdErr)

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
