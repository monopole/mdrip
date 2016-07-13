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

	// Build the program from blocks extracted from markdown files.
	for _, fileName := range c.FileNames {
		contents, err := ioutil.ReadFile(string(fileName))
		if err != nil {
			log.Print("Unable to read file %q.", fileName)
		}
		m := lexer.Parse(string(contents))
		if blocks, ok := m[c.ScriptName]; ok {
			p.Add(model.NewScript(fileName, blocks))
		}
	}

	if p.ScriptCount() < 1 {
		log.Fatal("Found no blocks labelled \"%q\" in the given files.", c.ScriptName)
	}

	// Either run or print the program.
	if c.RunInSubshell {
		if r := p.RunInSubShell(c.BlockTimeOut); r.Problem() != nil {
			r.Print(c.ScriptName)
			if c.FailWithSubshell {
				log.Fatal(r.Problem())
			}
		}
	} else {
		if c.Preambled >= 0 {
			p.PrintPreambled(os.Stdout, c.ScriptName, c.Preambled)
		} else {
			p.PrintNormal(os.Stdout, c.ScriptName)
		}
	}
}
