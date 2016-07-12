package main

import (
	"fmt"
	"github.com/monopole/mdrip/config"
	"github.com/monopole/mdrip/model"
	"github.com/monopole/mdrip/util"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	c := config.GetConfig()
	program := model.NewProgram()

	for _, fileName := range c.FileNames {
		contents, err := ioutil.ReadFile(string(fileName))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read %q\n", fileName)
			config.Usage()
			os.Exit(2)
		}
		m := util.Parse(string(contents))
		script, ok := m[c.ScriptName]
		if !ok {
			fmt.Fprintf(os.Stderr, "No block labelled %q in file %q.\n", c.ScriptName, fileName)
			os.Exit(3)
		}
		program.Add(model.NewScript(fileName, script))
	}

	if !c.Subshell {
		if c.Preambled >= 0 {
			fmt.Printf("Yo\n")
			program.DumpPreambled(os.Stdout, c.ScriptName, c.Preambled)
		} else {
			fmt.Printf("Beans\n")
			program.DumpNormal(os.Stdout, c.ScriptName)
		}
		return
	}

	result := util.RunInSubShell(program, c.BlockTimeOut)
	if result.Problem() != nil {
		result.Dump(c.ScriptName)
		if !c.Succeed {
			log.Fatal(result.Problem())
		}
	}
}
