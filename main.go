package main

import (
	"github.com/golang/glog"
	"os"

	"fmt"
	"github.com/monopole/mdrip/config"
	"github.com/monopole/mdrip/program"
	"github.com/monopole/mdrip/subshell"
	"github.com/monopole/mdrip/tmux"
	"github.com/monopole/mdrip/webserver"
)

func realMain(c *config.Config) {
	p := program.NewProgram(c.Label(), c.FileNames())

	switch c.Mode() {
	case config.ModeTmux:
		t := tmux.NewTmux(tmux.Path)
		if !t.IsUp() {
			glog.Fatal(tmux.Path, " not running")
		}
		// Steal the first fileName as a host address argument.
		t.Adapt(string(c.FileNames()[0]))
	case config.ModeWeb:
		webserver.NewWebserver(p).Serve(c.HostAndPort())
	case config.ModeTest:
		p.Reload()
		s := subshell.NewSubshell(c.BlockTimeOut(), p)
		if r := s.Run(); r.Problem() != nil {
			r.Print(c.Label())
			if !c.IgnoreTestFailure() {
				glog.Fatal(r.Problem())
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

func testLoader(c *config.Config) {
	t, err := program.LoadMany(c.FileNames())
	if err != nil {
		fmt.Println(err)
		return
	}
	t.Accept(program.NewTutorialPrinter(os.Stdout))
}

func main() {
	c := config.GetConfig()
	testLoader(c)
	realMain(c)
}
