package loader

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/monopole/mdrip/v2/internal/loader/lexer"
)

// CodeBlock groups code from a FencedCodeBlock with a set of labels.
type CodeBlock struct {
	// Labels on a block.  This is a list, rather than a set, because
	// the first label might become the name of the block.
	labels     LabelList
	titleWords []string
	code       string
	index      int
	parent     *MyFile
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

const (
	maxWordsInId = 4
	maxWordSize  = 5
)

// ResetTitle sets the title words for the block.
func (cb *CodeBlock) ResetTitle(disAmbig map[string]int) {
	var normal []string
	var special []string
	for _, l := range cb.labels {
		if l.IsSpecial() {
			special = append(special, string(l))
		} else {
			normal = append(normal, string(l))
		}
	}
	var first string
	if len(normal) > 0 {
		first = normal[0]
		normal = normal[1:]
	} else {
		first = lexer.MakeIdentifier(cb.code, maxWordsInId, maxWordSize)
	}
	if disAmbig != nil {
		c := disAmbig[first]
		c++
		disAmbig[first] = c
		if c > 1 {
			first += strconv.Itoa(c)
		}
	}
	cb.titleWords = append(append([]string{first}, normal...), special...)
}

// UniqName returns the name of the code block.
func (cb *CodeBlock) UniqName() string {
	return cb.titleWords[0]
}

// Title returns the title of the code block.
func (cb *CodeBlock) Title() string {
	return strings.Join(cb.titleWords, " ")
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
