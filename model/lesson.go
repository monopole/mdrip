package model

import (
	"fmt"
	"io"
	"strings"
)

// A Lesson has a one to one correspondence to a file.
// It must have a name, and should have blocks.
type Lesson struct {
	path   FilePath
	blocks []*CommandBlock
}

func NewLesson(p FilePath, blocks []*CommandBlock) *Lesson {
	return &Lesson{p, blocks}
}

func NewLessonFromModelBlocks(p FilePath, blocks []*LabelledBlock) *Lesson {
	result := make([]*CommandBlock, len(blocks))
	for i, b := range blocks {
		result[i] = NewCommandBlock(b.Labels(), b.Prose(), b.Code())
	}
	return NewLesson(p, result)
}

func (l *Lesson) Accept(v TutVisitor)     { v.VisitLesson(l) }
func (l *Lesson) Name() string            { return l.path.Base() }
func (l *Lesson) Path() FilePath    { return l.path }
func (l *Lesson) Blocks() []*CommandBlock { return l.blocks }
func (l *Lesson) Children() []Tutorial {
	result := []Tutorial{}
	for _, b := range l.blocks {
		result = append(result, b)
	}
	return result
}

func (l *Lesson) GetBlocksWithLabel(label Label) []*CommandBlock {
	result := []*CommandBlock{}
	for _, b := range l.blocks {
		if b.HasLabel(label) {
			result = append(result, b)
		}
	}
	return result
}

// Print sends contents to the given Writer.
//
// If n <= 0, print everything, else only print the first n blocks.
//
// n is a count not an index, so to print only the first two blocks,
// pass n==2, not n==1.
func (l *Lesson) Print(w io.Writer, label Label, n int) {
	fmt.Fprintf(w, "#\n# Script @%s from %s \n#\n", label, l.path)
	delimFmt := "#" + strings.Repeat("-", 70) + "#  %s %d of %d\n"
	blocks := l.GetBlocksWithLabel(label)
	for i, block := range blocks {
		if n > 0 && i >= n {
			break
		}
		fmt.Fprintf(w, delimFmt, "Start", i+1, len(blocks))
		block.Print(w, "#", i+1, label, l.path)
		fmt.Fprintf(w, delimFmt, "End", i+1, len(blocks))
		fmt.Fprintln(w)
	}
}
