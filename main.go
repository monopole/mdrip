package main

import (
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
			log.Fatal("Unable to read %q\n", fileName)
		}
		m := lexer.Parse(string(contents))
		if script, ok := m[c.ScriptName]; ok {
			p.Add(model.NewScript(fileName, script))
		}
	}

	if p.ScriptCount() < 1 {
		log.Fatal("Found no blocks labelled \"%q\" in the given files.", c.ScriptName)
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
