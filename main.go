package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/monopole/mdrip/config"
	"github.com/monopole/mdrip/program"
	"github.com/monopole/mdrip/tmux"
)

func main() {
	c := config.GetConfig()
	// A program has a timeout and a name.
	p := program.NewProgram(c.BlockTimeOut(), c.ScriptName(), c.FileNames())

	switch c.Mode() {
	case config.ModeTmux:
		w := ioutil.Discard
		if tmux.IsProgramInstalled(tmux.ProgramName) {
			t := tmux.NewTmux(tmux.ProgramName)
			err := t.Refresh()
			if err != nil {
				log.Fatal(err)
			}
			w = t
		}
		p.Serve(w, c.HostAndPort())
	case config.ModeTest:
		p.Reload()
		if r := p.RunInSubShell(); r.Problem() != nil {
			r.Print(c.ScriptName())
			if !c.IgnoreTestFailure() {
				log.Fatal(r.Problem())
			}
		}
	default:
		p.Reload()
		if c.Preambled() > 0 {
			p.PrintPreambled(os.Stdout, c.Preambled())
		} else {
			p.PrintNormal(os.Stdout)
		}
	}
}
