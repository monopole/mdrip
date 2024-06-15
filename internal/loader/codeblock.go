package loader

import (
	"fmt"
	"io"
)

// CodeBlock groups an ast.FencedCodeBlock with a set of labels.
type CodeBlock struct {
	// Labels on a block.  This is a list, rather than a set, because
	// the _first_ label is considered the name of the block.
	labels   LabelList
	code     string
	language string
	index    int
	parent   *MyFile
}

func NewCodeBlock(
	fi *MyFile, code string, index int, language string, labels ...Label) *CodeBlock {
	b := &CodeBlock{code: code, index: index, language: language, parent: fi}
	b.AddLabels(labels)
	return b
}

func (cb *CodeBlock) printHeader(wr io.Writer, i int, content []byte) {
	_, _ = fmt.Fprintf(wr, "%3d. %v\n", i, cb.labels)
}

// Equals is true if the block have the same content,
// ignoring the parent.
func (cb *CodeBlock) Equals(other *CodeBlock) bool {
	return cb.code == other.code &&
		cb.language == other.language &&
		cb.labels.Equals(other.labels)
}

func (cb *CodeBlock) AddLabels(labels []Label) {
	cb.labels = append(cb.labels, labels...)
}

func (cb *CodeBlock) Code() string {
	return cb.code
}

// HasLabel is true if the block has the given label argument.
func (cb *CodeBlock) HasLabel(label Label) bool {
	if label == WildCardLabel {
		return true
	}
	return cb.labels.Contains(label)
}

// FirstLabel attempts to return the first human-authored label,
// and failing that makes up a label using the index.
func (cb *CodeBlock) FirstLabel() Label {
	for _, l := range cb.labels {
		if l != WildCardLabel && l != AnonLabel {
			return l
		}
	}
	return Label(fmt.Sprintf("codeBlock%03d", cb.index))
}

func (cb *CodeBlock) Dump(wr io.Writer) {
	if len(cb.labels) > 0 {
		_, _ = fmt.Fprintf(wr, "# labels: ")
		for _, l := range cb.labels {
			_, _ = fmt.Fprint(wr, " ", l)
		}
		_, _ = fmt.Fprintln(wr)
	}
	_, _ = fmt.Fprintf(wr, "#   lang: %s\n", cb.language)
	_, _ = fmt.Fprint(wr, cb.code)
}

func DumpBlocks(wr io.Writer, blocks []*CodeBlock) {
	for i, b := range blocks {
		_, _ = fmt.Fprintf(wr, "# ----- BLOCK%3d -----------------\n", i)
		b.Dump(wr)
	}
}
