package model

import (
	"fmt"
	"io"
	"strings"
)

// ParsedFile associates a file's name with its parsed content.
type  ParsedFile struct {
	filePath  FilePath
	blocks   []*CommandBlock
}

const (
	TmplNameParsedFile = "parsedFile"
	TmplBodyParsedFile = `
{{define "` + TmplNameParsedFile + `"}}
<!-- <h2>mdrip {{.Path}}</h2> -->
{{range $i, $b := .Blocks}}
  <div class="commandBlock" data-id="{{$i}}">
  {{ template "` + tmplNameCommandBlock + `" $b }}
  </div>
{{end}}
{{end}}
`
)

func NewParsedFile(fileName FilePath, blocks []*CommandBlock) *ParsedFile {
	return &ParsedFile{fileName, blocks}
}

func (s *ParsedFile) Path() FilePath {
	return s.filePath
}

func (s *ParsedFile) Blocks() []*CommandBlock {
	return s.blocks
}

// Print sends contents to the given Writer.
//
// If n <= 0, print everything, else only print the first n blocks.
//
// n is a count not an index, so to print only the first two blocks,
// pass n==2, not n==1.
func (s *ParsedFile) Print(w io.Writer, label Label, n int) {
	fmt.Fprintf(w, "#\n# ParsedFile @%s from %s \n#\n", label, s.filePath)
	delimFmt := "#" + strings.Repeat("-", 70) + "#  %s %d of %d\n"
	blockCount := len(s.blocks)
	for i, block := range s.blocks {
		if n > 0 && i >= n {
			break
		}
		fmt.Fprintf(w, delimFmt, "Start", i+1, blockCount)
		block.Print(w, "#", i+1, label, s.filePath)
		fmt.Fprintf(w, delimFmt, "End", i+1, blockCount)
		fmt.Fprintln(w)
	}
}
