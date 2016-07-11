package config

import (
	"flag"
	"fmt"
	"github.com/monopole/mdrip/model"
	"os"
	"time"
)

var blockTimeOut = flag.Duration("blockTimeOut", 7*time.Second,
	"The max amount of time to wait for a command block to exit.")

var preambled = flag.Int("preambled", -1,
	"Place all scripts in a subshell, preambled by the first {n} blocks in the first script.")

var subshell = flag.Bool("subshell", false,
	"Run extracted blocks in subshell (leaves your env vars and pwd unchanged).")

var swallow = flag.Bool("swallow", false,
	"Swallow errors from subshell (non-zero exit only on problems in driver code).")

type Config struct {
	Preambled    int
	Subshell     bool
	Swallow      bool
	BlockTimeOut time.Duration
	ScriptName   model.Label
	FileNames    []model.FileName
}

func GetConfig() *Config {

	flag.Usage = Usage
	flag.Parse()

	if *swallow && !*subshell {
		fmt.Fprintf(os.Stderr, "Makes no sense to specify --swallow but not --subshell.\n")
		Usage()
		os.Exit(1)
	}

	// Must have a label, followed by at least one file name.
	if flag.NArg() < 2 {
		Usage()
		os.Exit(1)
	}

	c := &Config{}
	c.Subshell = *subshell
	c.Preambled = *preambled
	c.Swallow = *swallow
	c.ScriptName = model.Label(flag.Arg(0))
	c.FileNames = make([]model.FileName, flag.NArg()-1)

	for i := 1; i < flag.NArg(); i++ {
		c.FileNames[i-1] = model.FileName(flag.Arg(i))
	}

	return c

}

func Usage() {
	fmt.Fprintf(os.Stderr, "\nUsage:  %s {label} {fileName}...\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr,
		`
Reads markdown files, extracts code blocks with a given @label, and
either runs them in a subshell or emits them to stdout.

If the markdown file contains

  Blah blah blah.
  Beand
  <!-- @goHome @foo -->
  '''
  cd $HOME
  '''
  Blah blah blah.
  <!-- @echoApple @apple -->
  '''
  echo "an apple a day keeps the doctor away"
  '''
  Blah blah blah.
  <!-- @echoCloseStar @foo @baz -->
  '''
  echo "Proxima Centauri"
  '''
  Blah blah blah.

then the command '{this} foo {fileName}' emits:

  cd $HOME
  echo "Proxima Centauri"

Pipe output to 'source /dev/stdin' to run it directly.

Use --subshell to run the blocks in a subshell leaving your current
shell env vars and pwd unchanged.  The code blocks can, however, do
anything to your computer that you can.
`)
}
