package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/monopole/mdrip/config"
	"github.com/monopole/mdrip/lexer"
	"github.com/monopole/mdrip/model"
	"github.com/monopole/mdrip/tmux"
)

func main() {
	c := config.GetConfig()
	// A program has a timeout and a name.
	p := model.NewProgram(c.BlockTimeOut(), c.ScriptName())

	// Build program code from blocks extracted from markdown files.
	for _, fileName := range c.FileNames {
		contents, err := ioutil.ReadFile(string(fileName))
		if err != nil {
			log.Printf("Unable to read file \"%s\".", fileName)
		}
		m := lexer.Parse(string(contents))
		if blocks, ok := m[c.ScriptName()]; ok {
			p.Add(model.NewScript(fileName, blocks))
		}
	}

	if p.ScriptCount() < 1 {
		if c.ScriptName().IsAny() {
			log.Fatal("No blocks found in the given files.")
		} else {
			log.Fatalf("No blocks labelled %q found in the given files.", c.ScriptName())
		}
	}

	switch c.Mode() {
	case config.ModeTmux:
		t := tmux.NewTmux(tmux.ProgramName)
		err := t.Refresh()
		if err != nil {
			log.Fatal(err)
		}
		p.Serve(t, c.HostAndPort())
	case config.ModeTest:
		if r := p.RunInSubShell(); r.Problem() != nil {
			r.Print(c.ScriptName())
			if !c.IgnoreTestFailure() {
				log.Fatal(r.Problem())
			}
		}
	default:
		if c.Preambled() > 0 {
			p.PrintPreambled(os.Stdout, c.Preambled())
		} else {
			p.PrintNormal(os.Stdout)
		}
	}
}
