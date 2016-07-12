package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/monopole/mdrip/config"
	"github.com/monopole/mdrip/lexer"
	"github.com/monopole/mdrip/model"
)

func main() {
	c := config.GetConfig()
	p := model.NewProgram()

	for _, fileName := range c.FileNames {
		contents, err := ioutil.ReadFile(string(fileName))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read %q\n", fileName)
			os.Exit(2)
		}
		m := lexer.Parse(string(contents))
		script, ok := m[c.ScriptName]
		if !ok {
			fmt.Fprintf(os.Stderr,
				"No block labelled %q in file %q.\n", c.ScriptName, fileName)
			os.Exit(3)
		}
		p.Add(model.NewScript(fileName, script))
	}

	if c.Subshell {
		r := p.RunInSubShell(c.BlockTimeOut)
		if r.Problem() != nil {
			r.Dump(c.ScriptName)
			if c.FailWithSubshell {
				log.Fatal(r.Problem())
			}
		}
	} else {
		if c.Preambled >= 0 {
			p.DumpPreambled(os.Stdout, c.ScriptName, c.Preambled)
		} else {
			p.DumpNormal(os.Stdout, c.ScriptName)
		}
	}
}
