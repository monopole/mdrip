// Package config validates flag values and command line arguments and
// converts them to read-only type-safe values for mdrip.
package config

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/monopole/mdrip/model"
	"os"
	"strconv"
	"time"
	"unicode"
)

const (
	usageText = `
Extracts code blocks from the given markdown files for further processing.


E.g. if the markdown file contains

  Blah blah blah.
  Blah blah blah.

  <!-- @goHome @foo -->
  '''
  cd $HOME
  '''

  Blah blah blah.
  Blah blah blah.

  <!-- @platitude @apple -->
  '''
  echo "an apple a day keeps the doctor away"
  '''

  Blah blah blah.
  Blah blah blah.

  <!-- @reportNearbyStar @foo @bar -->
  '''
  echo "Proxima Centauri"
  '''

  Blah blah blah.
  Blah blah blah.

then the command 'mdrip --label foo {fileName}' emits:

  cd $HOME
  echo "Proxima Centauri"

while the command 'mdrip --label platitude {fileName}' emits:

  echo "an apple a day keeps the doctor away"


Modes:

 --mode print  (the default)

   Print extracted script to stdout.

   Use 
      eval "$(mdrip file.md)"
   to run in current terminal, impacting your environment.

   Use
      mdrip file.md | source /dev/stdin
   to run in a piped shell that exits with extracted code status.

 --mode web

   Starts a web server at http://localhost:8000 to offer a UX
   facilitating execution of command blocks in an existing tmux
   session.

   Change port using --port flag.

 --mode test

   Use this flag for markdown-based feature tests.

   Suppose one has a tutorial consisting of command line instructions
   in a markdown file.

   To assure that those instructions continue to work, some test suite
   can assert that the following command exits with status 0:

     mdrip --mode test /path/to/tutorial.md

   This runs extracted blocks in an mdrip subshell, leaving the
   executing shell unchanged.

   In this mode, mdrip captures the stdout and stderr of the
   subprocess, reporting only blocks that fail, facilitating error
   diagnosis.

   Normally, mdrip exits with non-zero status only when used
   incorrectly, e.g. file not found, bad flags, etc.  In in test mode,
   mdrip will exit with the status of any failing code block.
`
)

type ModeType int

const (
	ModePrint ModeType = iota
	ModeWeb
	ModeTest
)

var (
	mode = flag.String("mode", "print",
		`Mode is print, test or web.`)

	label = flag.String("label", "",
		`Using "--label foo" means extract only blocks annotated with "<!-- @foo -->".`)

	preambled = flag.Int("preambled", 0,
		`In --mode print, run the first {n} blocks in the current shell, and the rest in a trapped subshell.`)

	useHostname = flag.Bool("useHostname", false,
		`In --mode web, use the hostname utility to specify where to serve, else implicitly use localhost.`)

	port = flag.Int("port", 8000,
		`In --mode web, use given port for the local web server.`)

	blockTimeOut = flag.Duration("blockTimeOut", 7*time.Second,
		`In --mode test, the max amount of time to wait for a command block to exit.`)

	ignoreTestFailure = flag.Bool("ignoreTestFailure", false,
		`In --mode test, exit with success regardless of extracted code failure.`)
)

// A forgiving interpretation of mode argument.
func determineMode() ModeType {
	if len(*mode) == 0 {
		return ModePrint
	}
	switch unicode.ToLower([]rune(*mode)[0]) {
	case 't':
		return ModeTest
	case 'w':
		return ModeWeb
	default:
		return ModePrint
	}
}

func determineLabel() model.Label {
	if len(*label) == 0 {
		return model.AnyLabel
	}
	return model.Label(*label)
}

func determineFiles() []model.FileName {
	f := make([]model.FileName, flag.NArg())
	for i, n := range flag.Args() {
		f[i] = model.FileName(n)
	}
	return f
}

type Config struct {
	scriptName model.Label
	mode       ModeType
	FileNames  []model.FileName
}

func (c *Config) BlockTimeOut() time.Duration {
	return *blockTimeOut
}

func (c *Config) Preambled() int {
	return *preambled
}

func (c *Config) HostAndPort() string {
	hostname := "localhost"
	if *useHostname {
		var err error
		hostname, err = os.Hostname()
		if err != nil {
			glog.Fatalf("Trouble with hostname: %v", err)
		}
	}
	return hostname + ":" + strconv.Itoa(*port)
}

func (c *Config) Mode() ModeType {
	return c.mode
}

func (c *Config) IgnoreTestFailure() bool {
	return *ignoreTestFailure
}

func (c *Config) ScriptName() model.Label {
	return c.scriptName
}

func GetConfig() *Config {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Must specify a file name.")
		// TODO if file is --, read from stdin.
		usage()
		os.Exit(1)
	}

	desiredMode := determineMode()

	if *ignoreTestFailure && desiredMode != ModeTest {
		fmt.Fprintln(os.Stderr,
			`Makes no sense to specify --ignoreTestFailure without --mode test.`)
		usage()
		os.Exit(1)
	}

	return &Config{determineLabel(), desiredMode, determineFiles()}
}

func usage() {
	fmt.Fprintf(os.Stderr, "\nUsage:  %s {fileName}...\n", os.Args[0])
	fmt.Fprint(os.Stderr, usageText)
	fmt.Fprint(os.Stderr, "\n\nFlags:\n\n")
	flag.PrintDefaults()
}
