package subshell

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	program2 "github.com/monopole/mdrip/tobeinternal/program"
	"github.com/monopole/mdrip/tobeinternal/scanner"
	"github.com/monopole/mdrip/tobeinternal/util"

	"github.com/golang/glog"
)

// Subshell can run a program
type Subshell struct {
	blockTimeout time.Duration
	program      *program2.Program
}

// NewSubshell returns a shell loaded with a program and block timeout ready to run.
func NewSubshell(timeout time.Duration, p *program2.Program) *Subshell {
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
// out with a completion flag attached.  This continues until the
// input channel closes.
//
// On a sad path, an accumulation of strings is sent with an incompletion
// flag, and the function exits early, before it's input channel closes.
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
					glog.Infof("accum %s: Timeout return.", prefix)
				}
				out <- NewIncompleteOutput(accum.String())
				return
			}
			if strings.HasPrefix(line, scanner.MsgError) {
				accum.WriteString(line + "\n")
				if glog.V(2) {
					glog.Infof("accum %s: Error return.", prefix)
				}
				out <- NewIncompleteOutput(accum.String())
				return
			}
			if strings.HasPrefix(line, scanner.MsgHappy) {
				if glog.V(2) {
					glog.Infof("accum %s: %s", prefix, line)
				}
				out <- NewCompleteOutput(accum.String())
				accum.Reset()
			} else {
				// Normal accumulation.
				if glog.V(2) {
					glog.Infof("accum %s: [%s]", prefix, line)
				}
				accum.WriteString(line + "\n")
			}
		}
		if glog.V(2) {
			glog.Infof("accum %s closed.", prefix)
		}
		trailing := strings.TrimSpace(accum.String())
		if len(trailing) > 0 {
			if glog.V(2) {
				glog.Infof(
					"accum from %s: Extra output after happy echo [%s]",
					prefix, trailing)
			}
			out <- NewIncompleteOutput(trailing)
		}
	}()
	return out
}

// processShellOutput associates shell output with command blocks
// (both stderr and stdout).  It assume that the shell program has
// been seeded with echo MsgHappy statements.
//
// It loops over the blocks, trying to pull output off,
// expecting successful block output to include MsgHappy
// on both stderr and stdout.
//
// The expectation is that if a block fails (triggering -e)
// or if it times out, MsgHappy will not be seen.
//
// If a shell exits successfully, this method should visit
// all of its blocks, and return nil.  If this method doesn't
// visit all of its blocks, the shell should have exited
// with an error.
func processShellOutput(
	lessons []*program2.LessonPgm,
	chAccOut, chAccErr <-chan *BlockOutput) *RunResult {
	var prevOut, prevErr *BlockOutput
	for _, lesson := range lessons {
		numBlocks := len(lesson.Blocks())
		for i, block := range lesson.Blocks() {
			glog.Infof("Expecting output of %s (%d/%d) from %s\n",
				block.Name(), i+1, numBlocks, lesson.Path())
			if glog.V(2) {
				glog.Infof("\n%s", block.Code())
			}
			outBlock := <-chAccOut
			errBlock := <-chAccErr
			// These can be nil if there was absolutely no output, either because
			// there were no commands, or only commands with no output, e.g. /bin/false.
			if outBlock == nil || !outBlock.Completed() ||
				errBlock == nil || !errBlock.Completed() {
				return NewRunResult(
					outBlock, errBlock).SetFileName(lesson.Path()).SetIndex(i).SetBlock(block)
			}
			prevOut = outBlock
			prevErr = errBlock
		}
	}
	glog.Info("All done, no errors triggered.")
	return NewRunResult(prevOut, prevErr)
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

// Write command blocks to disk to create a bash script.
// Start the script with exit on error options like -e.
// After each block, add two echo commands, one for stderr, the other for stdout.
// These echos will be used to associate output on either stream with the
// command that produced it.
// Doing this instead of writing to the shell's stdinpipe because of
// https://github.com/monopole/mdrip/commit/a7be6a6fb62ccf8dfe1c2906515ce3e83d0400d7
func writeFile(lessons []*program2.LessonPgm) *os.File {
	f, err := ioutil.TempFile("", "mdrip-file-")
	util.Check("create temp file", err)
	util.Check("chmod temp file", os.Chmod(f.Name(), 0744))
	writeString(f, "set -e\n")
	writeString(f, "set -u\n")
	writeString(f, "set -o pipefail\n")
	for _, lesson := range lessons {
		for _, block := range lesson.Blocks() {
			writeString(f, block.Code().String())
			writeString(f, "echo "+scanner.MsgHappy+" "+block.Name()+"\n")
			writeString(f, "echo "+scanner.MsgHappy+" "+block.Name()+" 1>&2\n\n")
		}
	}
	if glog.V(2) {
		glog.Infof("Run: running commands from %s", f.Name())
	}
	return f
}

func makeAccumulator(
	wait time.Duration, name string, stream io.ReadCloser) <-chan *BlockOutput {
	return accumulateOutput(name, scanner.BuffScanner(wait, name, stream))
}

// politeWait waits for shell to end, and return its exit error.
// It doesn't wait long, because we presume the shell is either already done,
// or is hung, and we've already burned s.blockTimeout waiting for it.
func politeWait(shell *exec.Cmd) (err error) {
	done := make(chan error, 1)
	err = nil
	go func() {
		if glog.V(2) {
			glog.Info("Run:  Waiting for shell to end.")
		}
		done <- shell.Wait()
		glog.Info("Run:  Wait completed.")
	}()
	select {
	case <-time.After(2 * time.Second):
		glog.Infof("Run:  killing the shell after a polite wait")
		err = shell.Process.Kill()
		if err == nil {
			err = errors.New("shell timed out")
		} // else pass along the error from Kill.
	case err = <-done:
		if err != nil {
			glog.Infof("Run:  Shell failed with error %v.", err)
		}
	}
	return
}

// Run runs command blocks in a subprocess, stopping and
// reporting on any error.
//
// Command blocks are strings presumably holding code from some shell
// language.  The strings may be more complex than single commands
// delimitted by linefeeds - e.g. blocks that operate on HERE
// documents, or multi-line commands using line continuation via '\',
// quotes or curly brackets.
//
// This function itself is not a shell interpreter, so it cannot know
// if one line of text from a command block is an individual command
// or part of something else.
//
// Error reporting works by discarding output from command blocks that
// succeeded, and only reporting the contents of stdout and stderr
// when the subprocess exits on error.
func (s *Subshell) Run() (result *RunResult) {
	tmpFile := writeFile(s.program.Lessons())
	defer func() {
		// Windows has trouble with processes hanging on to temp files.
		attempts := 6
		var err error
		for i := 0; i < attempts; i++ {
			if i > 0 {
				time.Sleep(1 * time.Second)
			}
			err = os.Remove(tmpFile.Name())
			if err == nil {
				return
			}
		}
		msg := fmt.Sprintf(
			"After %d attempts, unable to delete %s, error=%v",
			attempts, tmpFile.Name(), err)
		if runtime.GOOS == "windows" {
			// Something wrong with end of process detection
			// or release on windows.  Just log and return.
			log.Print(msg)
			return
		}
		// Hold other OS's to higher standard.
		log.Fatal(msg)
	}()

	shell := exec.Command("bash", tmpFile.Name())

	stdIn, err := shell.StdinPipe()
	util.Check("in pipe", err)
	util.Check("close shell's stdin", stdIn.Close())

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
	pgid, err := util.GetProcesssGroupID(pid)
	if err == nil {
		if glog.V(2) {
			glog.Infof("Run:  pgid = %d", pgid)
		}
	}

	result = processShellOutput(
		s.program.Lessons(),
		makeAccumulator(s.blockTimeout, "stdOut", stdOut),
		makeAccumulator(s.blockTimeout, "stdErr", stdErr))

	// At this point, we've either successfully accounted for output
	// from all command blocks, or something timed out, or something went
	// wrong in the plumbing.  If the shell is still running, it
	// should be killed.

	err = politeWait(shell)
	if glog.V(2) {
		glog.Info("Run:  Shell done.")
	}
	if err == nil {
		if result.HasProgrammerError() {
			err = errors.New("unexpected programmer error - need code fix")
		} else if !result.Completed() {
			err = errors.New("problem processing stdout and/or stderr")
		}
	}
	result.SetError(err)
	// killProcesssGroup(pgid)?
	return
}
