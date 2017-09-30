package main

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/monopole/mdrip/config"
	"github.com/monopole/mdrip/lexer"
	"github.com/monopole/mdrip/program"
	"github.com/monopole/mdrip/subshell"
	"github.com/monopole/mdrip/tmux"
	"github.com/monopole/mdrip/webserver"
)

func main() {
	c := config.GetConfig()
	switch c.Mode() {
	case config.ModeTmux:
		t := tmux.NewTmux(tmux.Path)
		if !t.IsUp() {
			glog.Fatal(tmux.Path, " not running")
		}
		// Treat the first arg as a host address argument.
		t.Adapt(c.DataSource().FirstArg())
	case config.ModeWeb:
		s, err := webserver.NewServer(c.DataSource())
		if err != nil {
			fmt.Println(err)
			return
		}
		s.Serve(c.HostAndPort())
	case config.ModeTest:
		t, err := lexer.NewLoader(c.DataSource()).Load()
		if err != nil {
			fmt.Println(err)
			return
		}
		p := program.NewProgramFromTutorial(c.Label(), t)
		s := subshell.NewSubshell(c.BlockTimeOut(), p)
		if r := s.Run(); r.Problem() != nil {
			r.Print(c.Label())
			if !c.IgnoreTestFailure() {
				glog.Fatal(r.Problem())
			}
		}
	default:
		t, err := lexer.NewLoader(c.DataSource()).Load()
		if err != nil {
			fmt.Println(err)
			return
		}
		p := program.NewProgramFromTutorial(c.Label(), t)
		if c.Preambled() > 0 {
			p.PrintPreambled(os.Stdout, c.Preambled())
		} else {
			p.PrintNormal(os.Stdout)
		}
	}
}
