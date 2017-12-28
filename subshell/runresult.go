package subshell

import (
	"fmt"
	"os"
	"strings"

	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/program"
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

func (x BlockOutput) Completed() bool {
	return x.completed == yep
}

func (x BlockOutput) Output() string {
	return x.output
}

func NewIncompleteOutput(output string) *BlockOutput {
	return &BlockOutput{nope, output}
}

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

func NewRunResult(out, err *BlockOutput) *RunResult {
	return &RunResult{
		out, err, "", -1,
		program.NewEmptyBlockPgm(),
		nil}
}

// One of those "This should never happen" things.
func (x *RunResult) HasProgrammerError() bool {
	return (x.stdOut == nil && x.stdErr != nil) || (x.stdOut != nil && x.stdErr == nil)
}

// Output on stderr doesn't mean there was a failure
func (x *RunResult) Completed() bool {
	return (x.stdOut == nil || x.stdOut.Completed()) &&
		(x.stdErr == nil || x.stdErr.Completed())
}

func (x *RunResult) StdOut() string {
	if x.stdOut == nil {
		return "[stdout empty]"
	}
	return x.stdOut.output
}

func (x *RunResult) StdErr() string {
	if x.stdErr == nil {
		return "[stderr empty]"
	}
	return x.stdErr.output
}

func (x *RunResult) SetFileName(n base.FilePath) *RunResult {
	x.fileName = n
	return x
}

func (x *RunResult) FileName() base.FilePath {
	return x.fileName
}

func (x *RunResult) SetError(e error) *RunResult {
	x.anErr = e
	return x
}

func (x *RunResult) Error() error {
	return x.anErr
}

func (x *RunResult) SetIndex(i int) *RunResult {
	x.index = i
	return x
}

func (x *RunResult) Index() int {
	return x.index
}

func (x *RunResult) SetBlock(b *program.BlockPgm) *RunResult {
	x.block = b
	return x
}

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
