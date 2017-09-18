package model

import (
	"fmt"
	"io"
)

// Program is a list of Scripts with a common label.
// Each Script came from a file, so a Program is
// a collection of N files.
type Program struct {
	label   Label
	scripts []*Script
}

func (p *Program) Scripts() []*Script                { return p.scripts }
func (p *Program) Label() Label                      { return p.label }
func NewProgram(l Label, scripts []*Script) *Program { return &Program{l, scripts} }

const (
	TmplNameProgram = "program"
	TmplBodyProgram = `
{{define "` + TmplNameProgram + `"}}
{{range $i, $s := .Scripts}}
  <div data-id="{{$i}}">
  {{ template "` + TmplNameScript + `" $s }}
  </div>
{{end}}
{{end}}
`
)

// PrintNormal simply prints the contents of a program.
func (p Program) PrintNormal(w io.Writer) {
	for _, s := range p.scripts {
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
// This allows the aggregate program (series of blocks) to be
// structured as 1) a preamble initialization that impacts the
// environment of the active shell, followed by 2) everything
// else executing in a subshell that exits on error.  An exit
// in (2) won't cause the active shell to close - very annoying
// if one is running in a terminal.
//
// It's up to the markdown author to assure that the n blocks can
// always complete without exit on error because they will run in the
// existing terminal.  Hence these blocks should just set environment
// variables and/or define shell functions.
//
// The goal is to let the user both modify their existing terminal
// environment, and run remaining code in a trapped subshell, and
// survive any errors in that subshell with a modified environment.
func (p Program) PrintPreambled(w io.Writer, n int) {
	// Write the first n blocks of the first file normally.
	p.scripts[0].Print(w, p.label, n)
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
