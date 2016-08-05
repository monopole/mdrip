package model

import (
	"fmt"
	"io"
	"strings"
)

// script associates a list of CommandBlocks with the name of the
// file they came from.
type script struct {
	fileName FileName
	blocks   []*CommandBlock
}

const (
	tmplNameScript = "script"
	tmplBodyScript = `
{{define "` + tmplNameScript + `"}}
<h1>mdrip {{.FileName}}</h1>
{{range $i, $b := .Blocks}}
  <div class="commandBlock" data-id="{{$i}}">
  {{ template "` + tmplNameCommandBlock + `" $b }}
  </div>
{{end}}
{{end}}
`
)

func NewScript(fileName FileName, blocks []*CommandBlock) *script {
	return &script{fileName, blocks}
}

func (s script) FileName() FileName {
	return s.fileName
}

func (s script) Blocks() []*CommandBlock {
	return s.blocks
}

// Print sends contents to the given Writer.
//
// If n <= 0, print everything, else only print the first n blocks.
//
// n is a count not an index, so to print only the first two blocks,
// pass n==2, not n==1.
func (s script) Print(w io.Writer, label Label, n int) {
	fmt.Fprintf(w, "#\n# Script @%s from %s \n#\n", label, s.FileName())
	delimFmt := "#" + strings.Repeat("-", 70) + "#  %s %d of %d\n"
	blockCount := len(s.blocks)
	for i, block := range s.blocks {
		if n > 0 && i >= n {
			break
		}
		fmt.Fprintf(w, delimFmt, "Start", i+1, blockCount)
		block.Print(w, "#", i+1, label, s.FileName())
		fmt.Fprintf(w, delimFmt, "End", i+1, blockCount)
		fmt.Fprintln(w)
	}
}
