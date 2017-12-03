// Package config validates flag values and command line arguments and
// converts them to read-only type-safe values for mdrip.
package config

import (
	"errors"
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

Modes:

 --mode print  (the default)

   Print extracted code to stdout. Use
      eval "$(mdrip file.md)"
   to run in current terminal, impacting your environment. Use
      mdrip file.md | source /dev/stdin
   to run in an ephemeral shell that exits with extracted code status.

 --mode test

   Use this flag for markdown-based feature tests.

   To assure that the code blocks in a markdown file continue to work,
   some test suite can assert that this command exits with status 0:

     mdrip --mode test /path/to/tutorial

   This extracts code blocks from markdown on that path, and runs them
   in an mdrip subshell, leaving the executing shell unchanged.
   mdrip captures the stdout and stderr of the subprocess, and reports
   output from failing blocks, facilitating error diagnosis.

   Normally, mdrip exits with non-zero status only when used
   incorrectly, e.g. file not found, bad flags, etc.  In in test mode,
   mdrip exits with the status of any failing code block.

 --mode demo

   Starts a web server at http://localhost:8000 to offer a rendered
   version of the markdown facilitating execution of command blocks.

   Clicked command blocks are automatically copied to the user's clipboard
   and, if tmux is running, "pasted" to the active tmux window.
   See also flags --hostname and --port.

 --mode tmux

   Only useful if both a local tmux instance is running, and an mdrip
   is running remotely (not locally) in '--mode demo'.

   In this mode the first argument to mdrip, normally treated as a
   markdown filename, is treated as a URL.  mdrip attempts to open a
   websocket to that URL.  Discover the URL from mdrip's demo mode help
   button.

   Meanwhile, when a web user clicks on a code block served by mdrip
   (in --mode demo) an attempt is made to find a websocket associated
   with the user's web session.

   If a socket is found, the code block is sent to the socket.  Upon
   receipt, mdrip (in --mode tmux) sends the block to local tmux as if
   the user had typed it.

   This results in 'one click' behavior that's surprisingly handy.
`
)

type ModeType int

const (
	ModeUnknown ModeType = iota
	ModePrint
	ModeTest
	ModeDemo
	ModeTmux
)

var (
	mode = flag.String("mode", "print",
		`Mode is print, test, demo or tmux.`)

	label = flag.String("label", "",
		`Using "--label foo" means extract only blocks annotated with "<!-- @foo -->".`)

	preambled = flag.Int("preambled", 0,
		`In --mode print, run the first {n} blocks in the current shell, and the rest in a trapped subshell.`)

	useHostname = flag.Bool("useHostname", false,
		`In --mode demo, use the hostname utility to specify where to serve, else implicitly use localhost.`)

	port = flag.Int("port", 8000,
		`In --mode demo, expose HTTP at the given port.`)

	blockTimeOut = flag.Duration("blockTimeOut", 7*time.Second,
		`In --mode test, the max amount of time to wait for a command block to exit.`)

	ignoreTestFailure = flag.Bool("ignoreTestFailure", false,
		`In --mode test, exit with success regardless of extracted code failure.`)
)

type Config struct {
	label      base.Label
	mode       ModeType
	dataSource *base.DataSet
}

func determineMode() ModeType {
	if len(*mode) == 0 {
		return ModePrint
	}
	if len(*mode) < 3 {
		return ModeUnknown
	}
	// Use 3rd letter since test and tmux have `t` as char 1,
	// and test and demo have `e` as char 2.
	switch unicode.ToLower([]rune(*mode)[2]) {
	case 's': // test
		return ModeTest
	case 'm': // demo
		return ModeDemo
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

func (c *Config) DataSet() *base.DataSet {
	return c.dataSource
}

// nonsense for tests - need something better.
func DefaultConfig() *Config {
	ds, _ := base.NewDataSet([]string{"foo"})
	return &Config{base.WildCardLabel, ModePrint, ds}
}

func GetConfig() (*Config, error) {
	flag.Usage = Usage
	flag.Parse()
	dataSource, err := base.NewDataSet(flag.Args())
	if err != nil {
		return nil, err
	}
	desiredMode := determineMode()
	if desiredMode == ModeUnknown {
		return nil, errors.New(`For mode, specify print, test, demo or tmux.`)
	}
	if *ignoreTestFailure && desiredMode != ModeTest {
		return nil, errors.New(`Makes no sense to specify --ignoreTestFailure without --mode test.`)
	}
	return &Config{determineLabel(), desiredMode, dataSource}, nil
}

func Usage() {
	fmt.Fprintf(os.Stderr, "\nUsage:  %s {fileName}...\n", os.Args[0])
	fmt.Fprint(os.Stderr, usageText)
	fmt.Fprint(os.Stderr, "\n\nFlags:\n\n")
	flag.PrintDefaults()
}
