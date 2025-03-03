package loader

import (
	"fmt"
	"github.com/monopole/mdrip/v2/internal/loader/lexer"
	"io"
)

// CodeBlock groups code from a FencedCodeBlock with a set of labels.
type CodeBlock struct {
	// Labels on a block.  This is a list, rather than a set, because
	// the first label is considered the name of the block.
	labels LabelList
	name   string
	code   string
	index  int
	parent *MyFile
}

func NewCodeBlock(
	fi *MyFile, code string, index int, labels ...Label) *CodeBlock {
	b := &CodeBlock{code: code, index: index, parent: fi}
	b.AddLabels(labels)
	b.name = computeName(code, index, labels...)
	return b
}

func computeName(code string, index int, labels ...Label) string {
	candidates := removeSpecialLabels(labels)
	if len(candidates) > 0 {
		return string(candidates[0])
	}
	return fmt.Sprintf("%s%02d", lexer.MakeIdentifier(code, 3), index)
}

func removeSpecialLabels(l []Label) (result []Label) {
	for i := range l {
		if l[i] != SleepLabel && l[i] != SkipLabel {
			result = append(result, l[i])
		}
	}
	return
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
	cb.name = computeName(cb.code, cb.index, cb.labels...)
}

func (cb *CodeBlock) Code() string {
	return cb.code
}

// HasLabel is true if the block has the given label argument.
func (cb *CodeBlock) HasLabel(label Label) bool {
	return cb.labels.Contains(label)
}

// Name returns the name of the code block.
func (cb *CodeBlock) Name() string {
	return cb.name
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
