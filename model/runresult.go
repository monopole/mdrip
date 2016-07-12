package model

import (
	"fmt"
	"os"
	"strings"
)

type status int

const (
	yep status = iota
	nope
)

// BlockOutput pairs status (success or failure) with the output
// collected from a stream (i.e. stderr or stdout) as a result of
// executing all or part of a command block.
//
// Output can appear on stderr without neccessarily being associated
// with shell failure.
type BlockOutput struct {
	success status
	output  string
}

func (x BlockOutput) Succeeded() bool {
	return x.success == yep
}

func (x BlockOutput) Output() string {
	return x.output
}

func NewFailureOutput(output string) *BlockOutput {
	return &BlockOutput{nope, output}
}

func NewSuccessOutput(output string) *BlockOutput {
	return &BlockOutput{yep, output}
}

// RunResult pairs BlockOutput with meta data about shell execution.
type RunResult struct {
	BlockOutput
	fileName FileName      // File in which the error occurred.
	index    int           // Command block index.
	block    *CommandBlock // Content of actual command block.
	problem  error         // Error, if any.
	message  string        // Detailed error message, if any.
}

func NewRunResult() *RunResult {
	noLabels := []Label{}
	blockOutput := NewFailureOutput("")
	return &RunResult{*blockOutput, "", -1, NewCommandBlock(noLabels, ""), nil, ""}
}

// For tests.
func NoCommandsRunResult(
	blockOutput *BlockOutput, fileName FileName, index int, message string) *RunResult {
	noLabels := []Label{}
	return &RunResult{
		*blockOutput, fileName, index,
		NewCommandBlock(noLabels, ""), nil, message}
}

func (x *RunResult) FileName() FileName {
	return x.fileName
}

func (x *RunResult) Problem() error {
	return x.problem
}

func (x *RunResult) SetProblem(e error) *RunResult {
	x.problem = e
	return x
}

func (x *RunResult) Message() string {
	return x.message
}

func (x *RunResult) SetMessage(m string) *RunResult {
	x.message = m
	return x
}

func (x *RunResult) SetOutput(m string) *RunResult {
	x.output = m
	return x
}

func (x *RunResult) Index() int {
	return x.index
}

func (x *RunResult) SetIndex(i int) *RunResult {
	x.index = i
	return x
}

func (x *RunResult) SetBlock(b *CommandBlock) *RunResult {
	x.block = b
	return x
}

func (x *RunResult) SetFileName(n FileName) *RunResult {
	x.fileName = n
	return x
}

// Complain spits the contents of a RunResult to stderr.
func (x *RunResult) Dump(selectedLabel Label) {
	delim := strings.Repeat("-", 70) + "\n"
	fmt.Fprintf(os.Stderr, delim)
	x.block.Dump(os.Stderr, "Error", x.index+1, selectedLabel, x.fileName)
	fmt.Fprintf(os.Stderr, delim)
	dumpCapturedOutput("Stdout", delim, x.output)
	if len(x.message) > 0 {
		dumpCapturedOutput("Stderr", delim, x.message)
	}
}

func dumpCapturedOutput(name, delim, output string) {
	fmt.Fprintf(os.Stderr, "\n%s capture:\n", name)
	fmt.Fprintf(os.Stderr, delim)
	fmt.Fprintf(os.Stderr, output)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, delim)
}
