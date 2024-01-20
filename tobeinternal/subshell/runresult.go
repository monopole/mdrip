package subshell

import (
	"fmt"
	"os"
	"strings"

	"github.com/monopole/mdrip/tobeinternal/base"
	"github.com/monopole/mdrip/tobeinternal/program"
)

type status int

const (
	yep status = iota
	nope
)

// BlockOutput pairs success status (yes or no) with the output
// collected from a stream (i.e. stderr or stdout) as a result of
// executing a command block (or as much as could be executed before
// it failed).
//
// Output can appear on stderr without necessarily being associated
// with shell failure, so it's collected even in successful runs.
type BlockOutput struct {
	completed status
	output    string
}

// Completed is true if the stream was processed without error.  Does not
// mean that the shell completed without error, only means there was no
// piping problem or unexpected early termination.
func (x BlockOutput) Completed() bool {
	return x.completed == yep
}

// Output returns text accumulated from a stream.
func (x BlockOutput) Output() string {
	return x.output
}

// NewIncompleteOutput returns a BlockOutput configured to signal incompletion.
func NewIncompleteOutput(output string) *BlockOutput {
	return &BlockOutput{nope, output}
}

// NewCompleteOutput returns a BlockOutput configured to signal completion.
func NewCompleteOutput(output string) *BlockOutput {
	return &BlockOutput{yep, output}
}

// RunResult pairs BlockOutput with meta data about shell execution.
type RunResult struct {
	stdOut   *BlockOutput      // stdout from block execution
	stdErr   *BlockOutput      // stdErr from block execution
	fileName base.FilePath     // File in which the error occurred.
	index    int               // Index of command block with error.
	block    *program.BlockPgm // The command block with the error.
	anErr    error             // Shell error, if any.
}

// NewRunResult is a ctor for RunResult.
func NewRunResult(out, err *BlockOutput) *RunResult {
	return &RunResult{
		out, err, "", -1,
		program.NewEmptyBlockPgm(),
		nil}
}

// HasProgrammerError is one of those "This should never happen" things.
func (x *RunResult) HasProgrammerError() bool {
	return (x.stdOut == nil && x.stdErr != nil) || (x.stdOut != nil && x.stdErr == nil)
}

// Completed means the shell run completed, but implies nothing about exit code.
func (x *RunResult) Completed() bool {
	return (x.stdOut == nil || x.stdOut.Completed()) &&
		(x.stdErr == nil || x.stdErr.Completed())
}

// StdOut returns the accumulation from stdout.
func (x *RunResult) StdOut() string {
	if x.stdOut == nil {
		return ""
	}
	return x.stdOut.output
}

// StdErr returns the accumulation from stderr.
func (x *RunResult) StdErr() string {
	if x.stdErr == nil {
		return ""
	}
	return x.stdErr.output
}

// SetFileName sets the filename.
func (x *RunResult) SetFileName(n base.FilePath) *RunResult {
	x.fileName = n
	return x
}

// FileName returns the filename.
func (x *RunResult) FileName() base.FilePath {
	return x.fileName
}

// SetError sets the shell error.
func (x *RunResult) SetError(e error) *RunResult {
	x.anErr = e
	return x
}

// Error gets the shell error.
func (x *RunResult) Error() error {
	return x.anErr
}

// SetIndex sets the index of the failing block.
func (x *RunResult) SetIndex(i int) *RunResult {
	x.index = i
	return x
}

// Index gets the index of the failing block.
func (x *RunResult) Index() int {
	return x.index
}

// SetBlock sets the contents of the failing block.
func (x *RunResult) SetBlock(b *program.BlockPgm) *RunResult {
	x.block = b
	return x
}

// Print reports the result to stderr.
func (x *RunResult) Print(selectedLabel base.Label) {
	delim := strings.Repeat("-", 70) + "\n"
	fmt.Fprintf(os.Stderr, delim)
	x.block.Print(os.Stderr, "Error", x.index+1, selectedLabel, x.fileName)
	fmt.Fprintf(os.Stderr, delim)
	printCapturedOutput("stdOut", delim, x.StdOut())
	printCapturedOutput("stdErr", delim, x.StdErr())
}

func printCapturedOutput(name, delim, output string) {
	fmt.Fprintf(os.Stderr, "\n%s capture:\n", name)
	fmt.Fprintf(os.Stderr, delim)
	fmt.Fprintf(os.Stderr, output)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, delim)
}
