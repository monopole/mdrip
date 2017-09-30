// Package config validates flag values and command line arguments and
// converts them to read-only type-safe values for mdrip.
package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
	"unicode"

	"github.com/golang/glog"
	"github.com/monopole/mdrip/base"
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

   Print extracted code to stdout.

   Use
      eval "$(mdrip file.md)"
   to run in current terminal, impacting your environment.

   Use
      mdrip file.md | source /dev/stdin
   to run in a piped shell that exits with extracted code status.
   Does not impact your current shell.

 --mode web

   Starts a web server at http://localhost:8000 to offer a rendered
   version of the markdown facilitating execution of command blocks.

   Change port using --port flag.  See also flag --hostname.

 --mode tmux

   Only useful if both a local tmux instance is running, and somewhere
   on the net mdrip is running in '--mode web'.

   In this mode the first argument to mdrip, normally treated as a
   markdown filename, is treated as a URL.  mdrip attempts to open a
   websocket to that URL.

   Meanwhile, when a web user clicks on a code block served by mdrip
   (in --mode web) an attempt is made to find a websocket associated
   with the user's web session.

   If a socket is found, the code block is sent to the socket.  Upon
   receipt, mdrip (in --mode tmux) sends the block to local tmux as if
   the user had typed it.

   This results in 'one click' behavior that's surprisingly handy.

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
	ModeUnknown ModeType = iota
	ModePrint
	ModeTest
	ModeWeb
	ModeTmux
)

var (
	mode = flag.String("mode", "print",
		`Mode is print, test, web or tmux.`)

	label = flag.String("label", "",
		`Using "--label foo" means extract only blocks annotated with "<!-- @foo -->".`)

	preambled = flag.Int("preambled", 0,
		`In --mode print, run the first {n} blocks in the current shell, and the rest in a trapped subshell.`)

	useHostname = flag.Bool("useHostname", false,
		`In --mode web, use the hostname utility to specify where to serve, else implicitly use localhost.`)

	port = flag.Int("port", 8000,
		`In --mode web, expose HTTP at the given port.`)

	blockTimeOut = flag.Duration("blockTimeOut", 7*time.Second,
		`In --mode test, the max amount of time to wait for a command block to exit.`)

	ignoreTestFailure = flag.Bool("ignoreTestFailure", false,
		`In --mode test, exit with success regardless of extracted code failure.`)
)

type Config struct {
	label      base.Label
	mode       ModeType
	dataSource *base.DataSource
}

// A forgiving interpretation of mode argument.
func determineMode() ModeType {
	if len(*mode) == 0 {
		return ModePrint
	}
	if len(*mode) < 3 {
		return ModeUnknown
	}
	// Use 3rd letter since test and tmux have `t` as char 1,
	// and test and web have `e` as char 2.
	switch unicode.ToLower([]rune(*mode)[2]) {
	case 's': // test
		return ModeTest
	case 'b': // web
		return ModeWeb
	case 'u': // tmux
		return ModeTmux
	default:
		return ModePrint
	}
}

func determineLabel() base.Label {
	if len(*label) == 0 {
		return base.WildCardLabel
	}
	return base.Label(*label)
}

func (c *Config) BlockTimeOut() time.Duration {
	return *blockTimeOut
}

func (c *Config) Preambled() int {
	return *preambled
}

func (c *Config) HostAndPort() string {
	hostname := "" // docker breaks if one uses localhost here
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

func (c *Config) Label() base.Label {
	return c.label
}

func (c *Config) DataSource() *base.DataSource {
	return c.dataSource
}

func GetConfig() *Config {
	flag.Usage = usage
	flag.Parse()

	dataSource, err := base.NewDataSource(flag.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		// TODO: if arg is --, read from stdin?
		usage()
		os.Exit(1)
	}

	desiredMode := determineMode()
	if desiredMode == ModeUnknown {
		fmt.Fprintln(os.Stderr, `For mode, specify print, test, web or tmux.`)
		usage()
		os.Exit(1)
	}

	if *ignoreTestFailure && desiredMode != ModeTest {
		fmt.Fprintln(os.Stderr,
			`Makes no sense to specify --ignoreTestFailure without --mode test.`)
		usage()
		os.Exit(1)
	}

	return &Config{determineLabel(), desiredMode, dataSource}
}

func usage() {
	fmt.Fprintf(os.Stderr, "\nUsage:  %s {fileName}...\n", os.Args[0])
	fmt.Fprint(os.Stderr, usageText)
	fmt.Fprint(os.Stderr, "\n\nFlags:\n\n")
	flag.PrintDefaults()
}
