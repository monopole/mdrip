package model

import (
	"fmt"
	"io"
)

// Program is a list of Scripts, each from their own file.
type Program struct {
	scripts []*Script
}

func NewProgram() *Program {
	return &Program{}
}

func (p *Program) Add(s *Script) *Program {
	p.scripts = append(p.scripts, s)
	return p
}

func (p *Program) Scripts() []*Script {
	return p.scripts
}

// DumpNormal simply prints the contents of a program.
func (p Program) DumpNormal(w io.Writer, label Label) {
	for _, s := range p.scripts {
		s.Dump(w, label, 0)
	}
	fmt.Fprintf(w, "echo \" \"\n")
	fmt.Fprintf(w, "echo \"All done.  No errors.\"\n")
}

// DumpPreambled emits the first n blocks of a script normally, then
// emits the n blocks _again_, as well as the the remaining scripts,
// so that they run in a subshell.
//
// This allows the aggregrate script to be structured as 1) a preamble
// initialization script that impacts the environment of the active
// shell, followed by 2) a script that executes as a subshell that
// exits on error.  An exit in (2) won't cause the active shell (most
// likely a terminal) to close.
//
// The first script must be able to complete without exit on error
// because its not running as a subshell.  So it should just set
// environment variables and/or define shell functions.
//
// The goal is to let the user both modify their existing terminal
// environment, and run remaining code in a trapped subshell, and
// survive any errors in that subshell with a modified environment.
func (p Program) DumpPreambled(w io.Writer, label Label, n int) {
	// Write the first n blocks normally
	p.Scripts()[0].Dump(w, label, n)
	// Followed by everything appearing in a bash subshell.
	hereDocName := "HANDLED_SCRIPT"
	fmt.Fprintf(w, " bash -euo pipefail <<'%s'\n", hereDocName)
	fmt.Fprintf(w, "function handledTrouble() {\n")
	fmt.Fprintf(w, "  echo \" \"\n")
	fmt.Fprintf(w, "  echo \"Unable to continue!\"\n")
	fmt.Fprintf(w, "  exit 1\n")
	fmt.Fprintf(w, "}\n")
	fmt.Fprintf(w, "trap handledTrouble INT TERM\n")
	p.DumpNormal(w, label)
	fmt.Fprintf(w, "%s\n", hereDocName)
}
