package model

import (
	"fmt"
	"io"
	"strings"
)

// An ordered list of command blocks from the given file path.
type Script struct {
	path   FilePath
	blocks []*CommandBlock
}

func (s Script) Path() FilePath                            { return s.path }
func (s Script) Blocks() []*CommandBlock                   { return s.blocks }
func NewScript(p FilePath, blocks []*CommandBlock) *Script { return &Script{p, blocks} }

const (
	TmplNameScript = "parsedFile"
	TmplBodyScript = `
{{define "` + TmplNameScript + `"}}
<!-- <h2>mdrip {{.Path}}</h2> -->
{{range $i, $b := .Blocks}}
  <div class="commandBlock" data-id="{{$i}}">
  {{ template "` + TmplNameCommandBlock + `" $b }}
  </div>
{{end}}
{{end}}
`
)

// Print sends contents to the given Writer.
//
// If n <= 0, print everything, else only print the first n blocks.
//
// n is a count not an index, so to print only the first two blocks,
// pass n==2, not n==1.
func (s *Script) Print(w io.Writer, label Label, n int) {
	fmt.Fprintf(w, "#\n# Script @%s from %s \n#\n", label, s.path)
	delimFmt := "#" + strings.Repeat("-", 70) + "#  %s %d of %d\n"
	blockCount := len(s.blocks)
	for i, block := range s.blocks {
		if n > 0 && i >= n {
			break
		}
		fmt.Fprintf(w, delimFmt, "Start", i+1, blockCount)
		block.Print(w, "#", i+1, label, s.path)
		fmt.Fprintf(w, delimFmt, "End", i+1, blockCount)
		fmt.Fprintln(w)
	}
}
