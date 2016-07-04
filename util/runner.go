package util

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"github.com/monopole/mdrip/model"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

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
				if glog.V(2) {
					glog.Info("accumulateOutput %s: Timeout return.", prefix)
				}
				out <- model.NewFailureOutput(accum.String())
				return
			}
			if strings.HasPrefix(line, MsgError) {
				accum.WriteString(line + "\n")
				if glog.V(2) {
					glog.Info("accumulateOutput %s: Error return.", prefix)
				}
				out <- model.NewFailureOutput(accum.String())
				return
			}
			if strings.HasPrefix(line, MsgHappy) {
				if glog.V(2) {
					glog.Info("accumulateOutput %s: %s", prefix, line)
				}
				out <- model.NewSuccessOutput(accum.String())
				accum.Reset()
			} else {
				if glog.V(2) {
					glog.Info("accumulateOutput %s: Accumulating [%s]", prefix, line)
				}
				accum.WriteString(line + "\n")
			}
		}

		if glog.V(2) {
			glog.Info("accumulateOutput %s: <--- This channel has closed.", prefix)
		}
		trailing := strings.TrimSpace(accum.String())
		if len(trailing) > 0 {
			if glog.V(2) {
				glog.Info(
					"accumulateOutput %s: Erroneous (missing-happy) output [%s]", prefix, accum.String())
			}
			out <- model.NewFailureOutput(accum.String())
		} else {
			if glog.V(2) {
				glog.Info("accumulateOutput %s: Nothing trailing.", prefix)
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

	chOut := BuffScanner(blockTimeout, "stdout", stdOut)
	chErr := BuffScanner(1*time.Minute, "stderr", stdErr)

	chAccOut := accumulateOutput("stdOut", chOut)
	chAccErr := accumulateOutput("stdErr", chErr)

	errResult = model.NewScriptResult()
	for _, bucket := range scriptBuckets {
		numBlocks := len(bucket.Script())
		for i, block := range bucket.Script() {
			glog.Info("Running %s (%d/%d) from %s\n",
				block.Name(), i+1, numBlocks, bucket.FileName())
			if glog.V(2) {
				glog.Info("userBehavior: sending \"%s\"", block.Code())
			}

			result := <-chAccOut

			if result == nil || !result.Succeeded() {
				// A nil result means stdout has closed early because a
				// sub-subprocess failed.
				if result == nil {
					if glog.V(2) {
						glog.Info("userBehavior: stdout Result == nil.")
					}
					// Perhaps chErr <- MsgError + " : early termination; stdout has closed."
				} else {
					if glog.V(2) {
						glog.Info("userBehavior: stdout Result: %s", result.Output())
					}
					errResult.SetOutput(result.Output()).SetMessage(result.Output())
				}
				errResult.SetFileName(bucket.FileName()).SetIndex(i).SetBlock(block)
				fillErrResult(chAccErr, errResult)
				return
			}
		}
	}
	glog.Info("All done, no errors triggered.\n")
	return
}

// fillErrResult fills an instance of ScriptResult.
func fillErrResult(chAccErr <-chan *model.BlockOutput, errResult *model.ScriptResult) {
	result := <-chAccErr
	if result == nil {
		if glog.V(2) {
			glog.Info("userBehavior: stderr Result == nil.")
		}
		errResult.SetProblem(errors.New("unknown"))
		return
	}
	errResult.SetProblem(errors.New(result.Output())).SetMessage(result.Output())
	if glog.V(2) {
		glog.Info("userBehavior: stderr Result: %s", result.Output())
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
		for _, block := range bucket.Script() {
			checkWrite(block.Code().String(), scriptFile)
			checkWrite("\n", scriptFile)
			checkWrite("echo "+MsgHappy+" "+block.Name().String()+"\n", scriptFile)
		}
	}
	if glog.V(2) {
		glog.Info("RunInSubShell: running commands from %s", scriptFile.Name())
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
	if glog.V(2) {
		glog.Info("RunInSubShell: pid = %d", pid)
	}
	pgid, err := getProcesssGroupId(pid)
	if err == nil {
		if glog.V(2) {
			glog.Info("RunInSubShell:  pgid = %d", pgid)
		}
	}

	result = userBehavior(scriptBuckets, blockTimeout, stdOut, stdErr)

	if glog.V(2) {
		glog.Info("RunInSubShell:  Waiting for shell to end.")
	}
	waitError := shell.Wait()
	if result.Problem() == nil {
		result.SetProblem(waitError)
	}
	if glog.V(2) {
		glog.Info("RunInSubShell:  Shell done.")
	}

	// killProcesssGroup(pgid)
	return
}
