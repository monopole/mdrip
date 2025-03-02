package loader

import (
	"fmt"
	"io"
)

// CodeBlock groups code from a FencedCodeBlock with a set of labels.
type CodeBlock struct {
	// Labels on a block.  This is a list, rather than a set, because
	// the _first_ label is considered the name of the block.
	labels LabelList
	code   string
	index  int
	parent *MyFile
}

func NewCodeBlock(
	fi *MyFile, code string, index int, labels ...Label) *CodeBlock {
	b := &CodeBlock{code: code, index: index, parent: fi}
	b.AddLabels(labels)
	return b
}

// Path is the path to the file holding the block.
func (cb *CodeBlock) Path() FilePath {
	return cb.parent.Path()
}

// Equals is true if the block have the same content,
// ignoring the parent.
func (cb *CodeBlock) Equals(other *CodeBlock) bool {
	return cb.code == other.code && cb.labels.Equals(other.labels)
}

func (cb *CodeBlock) AddLabels(labels []Label) {
	cb.labels = append(cb.labels, labels...)
}

func (cb *CodeBlock) Code() string {
	return cb.code
}

// HasLabel is true if the block has the given label argument.
func (cb *CodeBlock) HasLabel(label Label) bool {
	return cb.labels.Contains(label)
}

// FirstLabel attempts to return the first human-authored label,
// and failing that makes up a label using the index.
func (cb *CodeBlock) FirstLabel() Label {
	if len(cb.labels) > 0 {
		return cb.labels[0]
	}
	return Label(fmt.Sprintf("codeBlock%03d", cb.index))
}

func (cb *CodeBlock) Dump(wr io.Writer, index int) {
	_, _ = fmt.Fprintf(wr, "# ----- BLOCK%4d: ", index)
	for _, l := range cb.labels {
		_, _ = fmt.Fprint(wr, " ", l)
	}
	_, _ = fmt.Fprintln(wr)
	_, _ = fmt.Fprint(wr, cb.code)
}

func DumpBlocks(wr io.Writer, blocks []*CodeBlock) {
	for i, b := range blocks {
		b.Dump(wr, i)
	}
}
