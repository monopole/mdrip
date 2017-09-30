package program

import (
	"fmt"
	"io"

	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
)

// Program is a list of Lessons and a label.
// Each LessonPgm represents a file, so a Program is a collection of N files.
// Every BlockTut in every lesson is known to have the given label.
type Program struct {
	label   base.Label
	lessons []*LessonPgm
}

func (p *Program) Lessons() []*LessonPgm { return p.lessons }
func (p *Program) Label() base.Label     { return p.label }

func NewProgram(lessons []*LessonPgm) *Program {
	return &Program{base.WildCardLabel, lessons}
}

// Build program from blocks extracted from a tutorial.
func NewProgramFromTutorial(l base.Label, t model.Tutorial) *Program {
	v := NewLessonPgmExtractor(l)
	t.Accept(v)
	return &Program{l, v.Lessons()}
}

// PrintNormal simply prints the contents of a program.
func (p Program) PrintNormal(w io.Writer) {
	for _, s := range p.lessons {
		s.Print(w, p.label, 0)
	}
	fmt.Fprintf(w, "echo \" \"\n")
	fmt.Fprintf(w, "echo \"All done.  No errors.\"\n")
}

// PrintPreambled emits the first n blocks of a file normally, then
// emits the n blocks _again_, as well as all the remaining blocks
// from remaining files, so that they run in a subshell with signal
// handling.
//
// This allows the aggregate command sequence (series of command blocks) to be
// structured as 1) a preamble initialization that impacts the
// environment of the active shell, followed by 2) everything
// else executing in a subshell that exits on error.  That way, an exit
// in (2) won't cause the active shell to close.  This is annoying
// if one is running the sequence in a terminal.
//
// It's up to the markdown author to assure that the n blocks can
// always complete without exit on error because they will run in the
// existing terminal.  These blocks should just set environment
// variables and/or define shell functions.
//
// The goal is to let the user both modify their existing terminal
// environment, and run remaining code in a trapped subshell, and
// survive any errors in that subshell with a modified environment.
func (p Program) PrintPreambled(w io.Writer, n int) {
	// Write the first n blocks of the first file normally.
	p.lessons[0].Print(w, p.label, n)
	// Followed by everything appearing in a bash subshell.
	hereDocName := "HANDLED_SCRIPT"
	fmt.Fprintf(w, " bash -euo pipefail <<'%s'\n", hereDocName)
	fmt.Fprintf(w, "function handledTrouble() {\n")
	fmt.Fprintf(w, "  echo \" \"\n")
	fmt.Fprintf(w, "  echo \"Unable to continue!\"\n")
	fmt.Fprintf(w, "  exit 1\n")
	fmt.Fprintf(w, "}\n")
	fmt.Fprintf(w, "trap handledTrouble INT TERM\n")
	p.PrintNormal(w)
	fmt.Fprintf(w, "%s\n", hereDocName)
}
