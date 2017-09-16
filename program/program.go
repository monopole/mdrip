package program

import (
	"fmt"
	"io"

	"github.com/golang/glog"
	"github.com/monopole/mdrip/model"
)

// Program is a list of ParsedFiles.
type Program struct {
	label       model.Label
	fileNames   []model.FilePath
	tutorial Tutorial
	ParsedFiles []*model.ParsedFile
}

const (
	TmplNameProgram = "program"
	TmplBodyProgram = `
{{define "` + TmplNameProgram + `"}}
{{range $i, $s := .AllParsedFiles}}
  <div data-id="{{$i}}">
  {{ template "` + model.TmplNameParsedFile + `" $s }}
  </div>
{{end}}
{{end}}
`
)

func NewProgram(label model.Label, fileNames []model.FilePath) *Program {
	return &Program{label, fileNames, nil, []*model.ParsedFile{}}
}

// Build program code from blocks extracted from markdown files.
func (p *Program) GetTutorial() Tutorial {
	return p.tutorial
}

// Build program code from blocks extracted from markdown files.
func (p *Program) Reload() {
	p.ParsedFiles = []*model.ParsedFile{}
	var err error
		p.tutorial, err = LoadMany(p.fileNames)
	if err != nil {
		glog.Warning("Trouble reading files.")
		return
	}
	v := NewTutorialParser(p.label)
	p.tutorial.Accept(v)
	p.ParsedFiles = v.Files()
}

// Check dies if program is empty.
func (p *Program) DieIfEmpty() {
	if p.ParsedFileCount() < 1 {
		if p.label.IsAny() {
			glog.Fatal("No blocks found in the given files.")
		} else {
			glog.Fatalf("No blocks labelled %q found in the given files.", p.label)
		}
	}
}

func (p *Program) Add(s *model.ParsedFile) *Program {
	p.ParsedFiles = append(p.ParsedFiles, s)
	return p
}

// Exported only for the template.
func (p *Program) AllParsedFiles() []*model.ParsedFile {
	return p.ParsedFiles
}

func (p *Program) ParsedFileCount() int {
	return len(p.ParsedFiles)
}

// PrintNormal simply prints the contents of a program.
func (p Program) PrintNormal(w io.Writer) {
	for _, s := range p.ParsedFiles {
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
	p.ParsedFiles[0].Print(w, p.label, n)
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
