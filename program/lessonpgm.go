package program

import (
	"fmt"
	"github.com/monopole/mdrip/model"
	"io"
	"strings"
)

// A LessonPgm has a one to one correspondence to a file.
type LessonPgm struct {
	path   model.FilePath
	blocks []*BlockPgm
}

func NewLessonPgm(p model.FilePath, blocks []*BlockPgm) *LessonPgm {
	return &LessonPgm{p, blocks}
}

func (l *LessonPgm) Name() string         { return l.path.Base() }
func (l *LessonPgm) Path() model.FilePath { return l.path }
func (l *LessonPgm) Blocks() []*BlockPgm  { return l.blocks }

// Print sends contents to the given Writer.
//
// If n <= 0, print everything, else only print the first n blocks.
//
// n is a count not an index, so to print only the first two blocks,
// pass n==2, not n==1.
func (l *LessonPgm) Print(w io.Writer, label model.Label, n int) {
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
