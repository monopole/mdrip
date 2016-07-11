package model

import (
	"fmt"
	"io"
	"strings"
)

// A collection of ScriptBucket
type Program struct {
	scripts []*ScriptBucket
}

func NewProgram() *Program {
	return &Program{}
}

func (p *Program) Add(s *ScriptBucket) *Program {
	p.scripts = append(p.scripts, s)
	return p
}

func (p *Program) Scripts() []*ScriptBucket {
	return p.scripts
}

// emitStraightScript simply prints the contents of scriptBuckets.
func (p Program) DumpNormal(w io.Writer, label Label) {
	for _, bucket := range p.scripts {
		bucket.Dump(w, label, 0)
	}
	fmt.Fprintf(w, "echo \" \"\n")
	fmt.Fprintf(w, "echo \"All done.  No errors.\"\n")
}

// emitPreambledScript emits the first script normally, then emit it
// again, as well as the the remaining scripts, so that they run in a
// subshell.
//
// This allows the aggregrate script to be structured as 1) a preamble
// initialization script that impacts the environment of the active
// shell, followed by 2) a script that executes as a subshell that
// exits on error.  An exit in (2) won't cause the active shell (most
// likely a terminal) to close.
//
// The first script must be able to complete without exit on error
// because its not running as a subshell.  So it should just set
// environment variables and/or define shell funtions.
func (p Program) DumpPreambled(w io.Writer, label Label, n int) {
	p.Scripts()[0].Dump(w, label, n)
	delim := "HANDLED_SCRIPT"
	fmt.Fprintf(w, " bash -euo pipefail <<'%s'\n", delim)
	fmt.Fprintf(w, "function handledTrouble() {\n")
	fmt.Fprintf(w, "  echo \" \"\n")
	fmt.Fprintf(w, "  echo \"Unable to continue!\"\n")
	fmt.Fprintf(w, "  exit 1\n")
	fmt.Fprintf(w, "}\n")
	fmt.Fprintf(w, "trap handledTrouble INT TERM\n")
	p.DumpNormal(w, label)
	fmt.Fprintf(w, "%s\n", delim)
}

// ScriptBucket associates a list of CommandBlocks with the name of the
// file they came from.
type ScriptBucket struct {
	fileName FileName
	script   []*CommandBlock
}

func NewScriptBucket(fileName FileName, script []*CommandBlock) *ScriptBucket {
	return &ScriptBucket{fileName, script}
}

func (b ScriptBucket) FileName() FileName {
	return b.fileName
}

func (b ScriptBucket) Script() []*CommandBlock {
	return b.script
}

// Dump sends contents to the given Writer.
//
// If n <= 0, dump everything, else only dump the first n blocks.  n
// is a count not an index.  If you want the first two blocks dumped,
// pass n==2, not n==1.
func (bucket ScriptBucket) Dump(w io.Writer, label Label, n int) {
	fmt.Fprintf(w, "#\n# Script @%s from %s \n#\n", label, bucket.FileName())
	delimFmt := "#" + strings.Repeat("-", 70) + "#  %s %d of %d\n"
	blockCount := len(bucket.Script())
	for i, block := range bucket.Script() {
		if n > 0 && i >= n {
			break
		}
		fmt.Fprintf(w, delimFmt, "Start", i+1, blockCount)
		block.Dump(w, "#", i+1, label, bucket.FileName())
		fmt.Fprintf(w, delimFmt, "End", i+1, blockCount)
		fmt.Fprintln(w)
	}
}
