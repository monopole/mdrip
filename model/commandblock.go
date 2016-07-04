package model

import (
	"fmt"
	"io"
)

// Labels are applied to code blocks to identify them and allow the
// blocks to be grouped into categories, e.g. tests or tutorials.
type Label string

func (l Label) String() string {
	return string(l)
}

// RawCode contains the actual sequence of shell commands (including
// stuff like HERE documents) to run as an opaque block.  If all
// commands succeed, the block succeeds, else the block fails.
type RawCode string

func (c RawCode) String() string {
	return string(c)
}

// CommandBlock groups Code with its labels.
type CommandBlock struct {
	labels []Label
	code   RawCode
}

func NewCommandBlock(labels []Label, code string) *CommandBlock {
	if len(labels) < 1 {
		// Assure at least one label.
		labels = []Label{Label("unknown")}
	}
	return &CommandBlock{labels, RawCode(code)}
}

// GetName returns the name of the block.  It's always the first
// label, and construction assures there will be at least one.
func (x CommandBlock) Name() Label {
	return x.labels[0]
}

func (x CommandBlock) Labels() []Label {
	return x.labels
}

func (x CommandBlock) Code() RawCode {
	return x.code
}

func (x CommandBlock) Dump(
	w io.Writer, prefix string, n int, label Label, fileName string) {
	fmt.Fprintf(w, "echo \"%s @%s (block #%d in %s) of %s\"\n\n",
		prefix, x.Name(), n, label, fileName)
	fmt.Fprint(w, x.Code())
}
