// Package config validates flag values and command line arguments and
// converts them to read-only type-safe values for mdrip.
package config

import (
	"flag"
	"fmt"
	"github.com/monopole/mdrip/model"
	"os"
	"time"
)

var (
	blockTimeOut = flag.Duration("blockTimeOut", 7*time.Second,
		"The max amount of time to wait for a command block to exit.")

	preambled = flag.Int("preambled", 0,
		"Place all scripts in a subshell, "+
			"preambled by the first {n} blocks in the first script.")

	port = flag.Int("port", 0,
		"Start a web server on given port.")

	runInSubshell = flag.Bool("subshell", false,
		"Run extracted blocks in subshell (leaves your env vars and pwd unchanged).")

	failWithSubshell = flag.Bool("failWithSubshell", false,
		"Fail if the subshell fails (normally only fails on a usage error). Only makes sense with --subshell.")
)

type Config struct {
	scriptName model.Label
	FileNames  []model.FileName
}

func (c *Config) BlockTimeOut() time.Duration {
	return *blockTimeOut
}

func (c *Config) Preambled() int {
	return *preambled
}

func (c *Config) Port() int {
	return *port
}

func (c *Config) RunInSubshell() bool {
	return *runInSubshell
}

func (c *Config) FailWithSubshell() bool {
	return *failWithSubshell
}

func (c *Config) ScriptName() model.Label {
	return c.scriptName
}

func GetConfig() *Config {
	flag.Usage = usage
	flag.Parse()

	if *failWithSubshell && !*runInSubshell {
		fmt.Fprintln(os.Stderr,
			"Makes no sense to specify --failWithSubshell but not --subshell.")
		usage()
		os.Exit(1)
	}

	if *port > 0 && *runInSubshell {
		fmt.Fprintln(os.Stderr,
			"Cannot specify both --port and --subshell; they are two different modes of operation.")
		usage()
		os.Exit(1)
	}

	if flag.NArg() < 2 {
		fmt.Fprintln(os.Stderr,
			"Must have a label, followed by at least one file name.")
		usage()
		os.Exit(1)
	}

	c := &Config{
		model.Label(flag.Arg(0)),
		make([]model.FileName, flag.NArg()-1)}

	for i := 1; i < flag.NArg(); i++ {
		c.FileNames[i-1] = model.FileName(flag.Arg(i))
	}

	return c
}

func usage() {
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
