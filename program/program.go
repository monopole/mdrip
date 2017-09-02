package program

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/golang/glog"
	"github.com/monopole/mdrip/lexer"
	"github.com/monopole/mdrip/model"
)

// Program is a list of scripts, each from their own file.
type Program struct {
	label     model.Label
	fileNames []model.FileName
	Scripts   []*model.Script
}

const (
	TmplNameProgram = "program"
	TmplBodyProgram = `
{{define "` + TmplNameProgram + `"}}
{{range $i, $s := .AllScripts}}
  <div data-id="{{$i}}">
  {{ template "` + model.TmplNameScript + `" $s }}
  </div>
{{end}}
{{end}}
`
)

func NewProgram(label model.Label, fileNames []model.FileName) *Program {
	return &Program{label, fileNames, []*model.Script{}}

}

// Build program code from blocks extracted from markdown files.
func (p *Program) Reload() {
	p.Scripts = []*model.Script{}
	for _, fileName := range p.fileNames {
		contents, err := ioutil.ReadFile(string(fileName))
		if err != nil {
			glog.Warning("Unable to read file \"%s\".", fileName)
		}
		m := lexer.Parse(string(contents))
		if blocks, ok := m[p.label]; ok {
			p.Add(model.NewScript(fileName, blocks))
		}
	}
}

// Check dies if program is empty.
func (p *Program) DieIfEmpty() {
	if p.ScriptCount() < 1 {
		if p.label.IsAny() {
			glog.Fatal("No blocks found in the given files.")
		} else {
			glog.Fatalf("No blocks labelled %q found in the given files.", p.label)
		}
	}
}

func (p *Program) Add(s *model.Script) *Program {
	p.Scripts = append(p.Scripts, s)
	return p
}

// Exported only for the template.
func (p *Program) AllScripts() []*model.Script {
	return p.Scripts
}

func (p *Program) ScriptCount() int {
	return len(p.Scripts)
}

// PrintNormal simply prints the contents of a program.
func (p Program) PrintNormal(w io.Writer) {
	for _, s := range p.Scripts {
		s.Print(w, p.label, 0)
	}
	fmt.Fprintf(w, "echo \" \"\n")
	fmt.Fprintf(w, "echo \"All done.  No errors.\"\n")
}

// PrintPreambled emits the first n blocks of a script normally, then
// emits the n blocks _again_, as well as all the remaining scripts,
// so that they run in a subshell with signal handling.
//
// This allows the aggregrate script to be structured as 1) a preamble
// initialization script that impacts the environment of the active
// shell, followed by 2) a script that executes as a subshell that
// exits on error.  An exit in (2) won't cause the active shell
// to close (annoying if it is a terminal).
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
	// Write the first n blocks if the first script normally.
	p.Scripts[0].Print(w, p.label, n)
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
