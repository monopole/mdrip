package subshell

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/monopole/mdrip/scanner"
	"github.com/monopole/mdrip/tutorial"
	"github.com/monopole/mdrip/util"
)

// Subshell can run a program
type Subshell struct {
	blockTimeout time.Duration
	program      *tutorial.Program
}

func NewSubshell(timeout time.Duration, p *tutorial.Program) *Subshell {
	return &Subshell{timeout, p}
}

// Check reports the error fatally if it's non-nil.
func check(msg string, err error) {
	if err != nil {
		fmt.Printf("Problem with %s: %v\n", msg, err)
		glog.Fatal(err)
	}
}

// userBehavior acts like a command line user.
//
// TODO(monopole): update the comments, as this function no longer writes to stdin.
// See https://github.com/monopole/mdrip/commit/a7be6a6fb62ccf8dfe1c2906515ce3e83d0400d7
//
// It writes command blocks to shell, then waits after each block to
// see if the block worked.  If the block appeared to complete without
// error, the routine sends the next block, else it exits early.
func (s *Subshell) userBehavior(stdOut, stdErr io.ReadCloser) (errResult *RunResult) {

	chOut := scanner.BuffScanner(s.blockTimeout, "stdout", stdOut)
	chErr := scanner.BuffScanner(1*time.Minute, "stderr", stdErr)

	chAccOut := accumulateOutput("stdOut", chOut)
	chAccErr := accumulateOutput("stdErr", chErr)

	errResult = NewRunResult()
	for _, lesson := range s.program.Lessons() {
		numBlocks := len(lesson.OnlyBlocksWithLabel(s.program.Label()))
		for i, block := range lesson.OnlyBlocksWithLabel(s.program.Label()) {
			glog.Info("Running %s (%d/%d) from %s\n",
				block.Name(), i+1, numBlocks, lesson.Path())
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
					// Perhaps chErr <- scanner.MsgError +
					//   " : early termination; stdout has closed."
				} else {
					if glog.V(2) {
						glog.Info("userBehavior: stdout Result: %s", result.Output())
					}
					errResult.SetOutput(result.Output()).SetMessage(result.Output())
				}
				errResult.SetFileName(lesson.Path()).SetIndex(i).SetBlock(block)
				fillErrResult(chAccErr, errResult)
				return
			}
		}
	}
	glog.Info("All done, no errors triggered.\n")
	return
}

// fillErrResult fills an instance of RunResult.
func fillErrResult(chAccErr <-chan *BlockOutput, errResult *RunResult) {
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
func (s *Subshell) Run() (result *RunResult) {
	// Write program to a file to be executed.
	tmpFile, err := ioutil.TempFile("", "mdrip-file-")
	check("create temp file", err)
	check("chmod temp file", os.Chmod(tmpFile.Name(), 0744))
	for _, file := range s.program.Lessons() {
		for _, block := range file.OnlyBlocksWithLabel(s.program.Label()) {
			write(tmpFile, block.Code().String())
			write(tmpFile, "\n")
			write(tmpFile, "echo "+scanner.MsgHappy+" "+block.Name()+"\n")
		}
	}
	if glog.V(2) {
		glog.Info("RunInSubShell: running commands from %s", tmpFile.Name())
	}
	defer func() {
		check("delete temp file", os.Remove(tmpFile.Name()))
	}()

	// Adding "-e" to force the subshell to die on any error.
	shell := exec.Command("bash", "-e", tmpFile.Name())

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
	pgid, err := util.GetProcesssGroupId(pid)
	if err == nil {
		if glog.V(2) {
			glog.Info("RunInSubShell:  pgid = %d", pgid)
		}
	}

	result = s.userBehavior(stdOut, stdErr)

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

func write(writer io.Writer, output string) {
	n, err := writer.Write([]byte(output))
	if err != nil {
		glog.Fatalf("Could not write %d bytes: %v", len(output), err)
	}
	if n != len(output) {
		glog.Fatalf("Expected to write %d bytes, wrote %d", len(output), n)
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
func accumulateOutput(prefix string, in <-chan string) <-chan *BlockOutput {
	out := make(chan *BlockOutput)
	var accum bytes.Buffer
	go func() {
		defer close(out)
		for line := range in {
			if strings.HasPrefix(line, scanner.MsgTimeout) {
				accum.WriteString("\n" + line + "\n")
				accum.WriteString("A subprocess might still be running.\n")
				if glog.V(2) {
					glog.Info("accumulateOutput %s: Timeout return.", prefix)
				}
				out <- NewFailureOutput(accum.String())
				return
			}
			if strings.HasPrefix(line, scanner.MsgError) {
				accum.WriteString(line + "\n")
				if glog.V(2) {
					glog.Info("accumulateOutput %s: Error return.", prefix)
				}
				out <- NewFailureOutput(accum.String())
				return
			}
			if strings.HasPrefix(line, scanner.MsgHappy) {
				if glog.V(2) {
					glog.Info("accumulateOutput %s: %s", prefix, line)
				}
				out <- NewSuccessOutput(accum.String())
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
					"accumulateOutput %s: Erroneous (missing-happy) output [%s]",
					prefix, accum.String())
			}
			out <- NewFailureOutput(accum.String())
		} else {
			if glog.V(2) {
				glog.Info("accumulateOutput %s: Nothing trailing.", prefix)
			}
		}
	}()
	return out
}
