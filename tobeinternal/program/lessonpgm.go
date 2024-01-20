package program

import (
	"fmt"
	"io"
	"strings"

	"github.com/monopole/mdrip/tobeinternal/base"
)

// LessonPgm has a one to one correspondence to a file.
type LessonPgm struct {
	path   base.FilePath
	blocks []*BlockPgm
}

// NewLessonPgm is a ctor.
func NewLessonPgm(p base.FilePath, blocks []*BlockPgm) *LessonPgm {
	return &LessonPgm{p, blocks}
}

// Name of the LessonPgm.
func (l *LessonPgm) Name() string { return l.path.Base() }

// Path of the file that holds the raw markdown for the lesson.
func (l *LessonPgm) Path() base.FilePath { return l.path }

// Blocks is all the code blocks extracted from the markdown.
func (l *LessonPgm) Blocks() []*BlockPgm { return l.blocks }

// Print sends contents to the given Writer.
//
// If n <= 0, print everything, else only print the first n blocks.
//
// n is a count not an index, so to print only the first two blocks,
// pass n==2, not n==1.
func (l *LessonPgm) Print(w io.Writer, label base.Label, n int) {
	fmt.Fprintf(w, "#\n# Script @%s from %s \n#\n", label, l.path)
	delimFmt := "#" + strings.Repeat("-", 70) + "#  %s %d of %d\n"
	for i, block := range l.blocks {
		if n > 0 && i >= n {
			break
		}
		fmt.Fprintf(w, delimFmt, "Start", i+1, len(l.blocks))
		block.Print(w, "#", i+1, label, l.path)
		fmt.Fprintf(w, delimFmt, "End", i+1, len(l.blocks))
		fmt.Fprintln(w)
	}
}
