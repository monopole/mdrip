package model

import (
	"fmt"
	"io"
	"strings"
)

// Script associates a list of CommandBlocks with the name of the
// file they came from.
type Script struct {
	fileName FileName
	blocks   []*CommandBlock
}

func NewScript(fileName FileName, script []*CommandBlock) *Script {
	return &Script{fileName, script}
}

func (b Script) FileName() FileName {
	return b.fileName
}

func (b Script) Blocks() []*CommandBlock {
	return b.blocks
}

// Dump sends contents to the given Writer.
//
// If n <= 0, dump everything, else only dump the first n blocks.
//
// n is a count not an index.
//
// If you want the first two blocks dumped, pass n==2, not n==1.
func (s Script) Dump(w io.Writer, label Label, n int) {
	fmt.Fprintf(w, "#\n# Script @%s from %s \n#\n", label, s.FileName())
	delimFmt := "#" + strings.Repeat("-", 70) + "#  %s %d of %d\n"
	blockCount := len(s.blocks)
	for i, block := range s.blocks {
		if n > 0 && i >= n {
			break
		}
		fmt.Fprintf(w, delimFmt, "Start", i+1, blockCount)
		block.Dump(w, "#", i+1, label, s.FileName())
		fmt.Fprintf(w, delimFmt, "End", i+1, blockCount)
		fmt.Fprintln(w)
	}
}
