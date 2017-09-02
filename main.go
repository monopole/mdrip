package main

import (
	"github.com/monopole/mdrip/config"
	"github.com/monopole/mdrip/program"
	"github.com/monopole/mdrip/subshell"
	"github.com/monopole/mdrip/tmux"
	"github.com/monopole/mdrip/webserver"
	"io"
	"io/ioutil"
	"log"
	"os"
)

// make a tmux writer.  If no tmux, return a discarder.
func makeWriter() io.Writer {
	if up, err := tmux.IsTmuxUp(tmux.Path); !up {
		log.Print(err)
		log.Print("Will run anyway, discarding scripts.")
		return ioutil.Discard
	}
	return tmux.NewTmuxByName(tmux.Path)
}

func main() {
	c := config.GetConfig()
	p := program.NewProgram(c.ScriptName(), c.FileNames())

	switch c.Mode() {
	case config.ModeTmux:
		webserver.NewWebserver(p).Serve(makeWriter(), c.HostAndPort())
	case config.ModeTest:
		p.Reload()
		s := subshell.NewSubshell(c.BlockTimeOut(), p)
		if r := s.Run(); r.Problem() != nil {
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
