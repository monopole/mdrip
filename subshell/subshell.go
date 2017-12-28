package subshell

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/monopole/mdrip/program"
	"github.com/monopole/mdrip/scanner"
	"github.com/monopole/mdrip/util"
)

// Subshell can run a program
type Subshell struct {
	blockTimeout time.Duration
	program      *program.Program
}

const cleanup = false

func NewSubshell(timeout time.Duration, p *program.Program) *Subshell {
	return &Subshell{timeout, p}
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
					glog.Infof("accum from %s: Timeout return.", prefix)
				}
				out <- NewFailureOutput(accum.String())
				return
			}
			if strings.HasPrefix(line, scanner.MsgError) {
				accum.WriteString(line + "\n")
				if glog.V(2) {
					glog.Infof("accum from %s: Error return.", prefix)
				}
				out <- NewFailureOutput(accum.String())
				return
			}
			if strings.HasPrefix(line, scanner.MsgHappy) {
				if glog.V(2) {
					glog.Infof("accum from %s: %s", prefix, line)
				}
				out <- NewSuccessOutput(accum.String())
				accum.Reset()
			} else {
				// Normal accumulation.
				if glog.V(2) {
					glog.Infof("accum from %s: [%s]", prefix, line)
				}
				accum.WriteString(line + "\n")
			}
		}

		if glog.V(2) {
			glog.Infof("accum from %s: <--- stream done.", prefix)
		}
		trailing := strings.TrimSpace(accum.String())
		if len(trailing) > 0 {
			if glog.V(2) {
				glog.Infof(
					"accum from %s: Extra output after happy echo [%s]",
					prefix, accum.String())
			}
			out <- NewFailureOutput(accum.String())
		}
	}()
	return out
}

func makeAccumulator(
	wait time.Duration, name string, stream io.ReadCloser) <-chan *BlockOutput {
	return accumulateOutput(name, scanner.BuffScanner(wait, name, stream))
}

// userBehavior acts like a command line user.
//
// TODO(monopole): update the comments, as this function no longer writes to stdin.
// See https://github.com/monopole/mdrip/commit/a7be6a6fb62ccf8dfe1c2906515ce3e83d0400d7
//
// It writes command blocks to shell, then waits after each block to
// see if the block worked.  If the block appeared to complete without
// error, the routine sends the next block, else it exits early.
func (s *Subshell) userBehavior(stdOut, stdErr io.ReadCloser) *RunResult {

	chAccOut := makeAccumulator(s.blockTimeout, "stdOut", stdOut)
	chAccErr := makeAccumulator(s.blockTimeout, "stdErr", stdErr)

	for _, lesson := range s.program.Lessons() {
		numBlocks := len(lesson.Blocks())
		for i, block := range lesson.Blocks() {
			glog.Infof("Expecting output of %s (%d/%d) from %s\n",
				block.Name(), i+1, numBlocks, lesson.Path())
			if glog.V(2) {
				glog.Infof("\n%s", block.Code())
			}

			outResult := <-chAccOut

			// Often a command will send output to stdErr, even if there was no error.
			// It's just an I/O stream, and output there doesn't mean that the command
			// that wrote to it exited with a non-zero status.
			// So we have to intentionally write a happy message to stderr after every
			// command and expect to absorb it here.  That way stderr output from successful
			// commands is absorbed and discarded, so that when a real error happens, the
			// contents of stdErr are only from that command, not from some far earlier
			// command.
			errResult := <-chAccErr

			if outResult == nil || !outResult.Succeeded() {
				var finalResult *RunResult
				// A nil result means stdout has closed early because a
				// sub-subprocess failed.
				if outResult == nil {
					finalResult = NewRunResult(NewFailureOutput(""))
					if glog.V(2) {
						glog.Info("userBehavior: stdout Result == nil.")
					}
					// Perhaps chErr <- scanner.MsgError +
					//   " : early termination; stdout has closed."
				} else {
					finalResult = NewRunResult(NewFailureOutput(outResult.Output()))
					if glog.V(2) {
						glog.Infof("userBehavior: stdout Result: %s", outResult.Output())
					}
					finalResult.SetMessage(outResult.Output())
				}
				finalResult.SetFileName(lesson.Path()).SetIndex(i).SetBlock(block)
				addErrResult(errResult, finalResult)
				return finalResult
			}
			// "This should never happen." :P
			if errResult == nil || !errResult.Succeeded() {
				finalResult := NewRunResult(NewFailureOutput(""))
				addErrResult(errResult, finalResult)
				return finalResult
			}
		}
	}
	glog.Info("All done, no errors triggered.")
	return NewRunResult(NewSuccessOutput(""))
}

func addErrResult(errResult *BlockOutput, finalResult *RunResult) {
	if errResult == nil {
		if glog.V(2) {
			glog.Info("userBehavior: stderr Result == nil.")
		}
		finalResult.SetProblem(errors.New("unknown"))
		return
	}
	finalResult.SetProblem(errors.New(errResult.Output())).SetMessage(errResult.Output())
	if glog.V(2) {
		glog.Infof("userBehavior: stderr Result: %s", errResult.Output())
	}
}

func writeString(writer io.Writer, output string) {
	n, err := writer.Write([]byte(output))
	if err != nil {
		glog.Fatalf("Could not write %d bytes: %v", len(output), err)
	}
	if n != len(output) {
		glog.Fatalf("Expected to write %d bytes, wrote %d", len(output), n)
	}
}

// Write program to disk
func writeFile(lessons []*program.LessonPgm) *os.File {
	f, err := ioutil.TempFile("", "mdrip-file-")
	util.Check("create temp file", err)
	util.Check("chmod temp file", os.Chmod(f.Name(), 0744))
	for _, lesson := range lessons {
		for _, block := range lesson.Blocks() {
			writeString(f, block.Code().String())
			writeString(f, "echo "+scanner.MsgHappy+" "+block.Name()+"\n")
			// Also bookend stderr.
			writeString(f, "echo "+scanner.MsgHappy+" "+block.Name()+" 1>&2\n\n")
		}
	}
	if glog.V(2) {
		glog.Infof("Run: running commands from %s", f.Name())
	}
	return f
}

// Run runs command blocks in a subprocess, stopping and
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
	tmpFile := writeFile(s.program.Lessons())
	if cleanup {
		defer func() {
			util.Check("delete temp file", os.Remove(tmpFile.Name()))
		}()
	}

	// Adding "-e" so that the subshell exists on error.
	shell := exec.Command("bash", "-e", tmpFile.Name())

	//stdIn, err := shell.StdinPipe()
	//util.Check("in pipe", err)
	//util.Check("close shell's stdin", stdIn.Close())

	stdOut, err := shell.StdoutPipe()
	util.Check("out pipe", err)

	stdErr, err := shell.StderrPipe()
	util.Check("err pipe", err)

	err = shell.Start()
	util.Check("shell start", err)

	pid := shell.Process.Pid
	if glog.V(2) {
		glog.Infof("Run: pid = %d", pid)
	}
	pgid, err := util.GetProcesssGroupId(pid)
	if err == nil {
		if glog.V(2) {
			glog.Infof("Run:  pgid = %d", pgid)
		}
	}

	result = s.userBehavior(stdOut, stdErr)

	if glog.V(2) {
		glog.Info("Run:  Waiting for shell to end.")
	}
	waitError := shell.Wait()
	if result.Problem() == nil {
		result.SetProblem(waitError)
	}
	if glog.V(2) {
		glog.Info("Run:  Shell done.")
	}

	// killProcesssGroup(pgid)
	return
}
