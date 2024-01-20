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

	"github.com/monopole/mdrip/tobeinternal/base"

	"github.com/golang/glog"
)

const (
	usageText = `
Extracts code blocks from the given markdown files for further processing.

Modes:

 --mode print  (the default)

   Print extracted code blocks to stdout.

   Use

      eval "$(mdrip file.md)"

   to run in current terminal, impacting your environment.

   Use

      mdrip file.md | source /dev/stdin

   to run in an ephemeral shell that exits with extracted code status.

 --mode test

   To assure that the code blocks in markdown files continue to work,
   a test suite should assert that the following exits with status 0:

     mdrip --mode test --label foo /path/to/markdown

   This extracts code blocks with the label '@foo' from markdown files
   in and below the given directory, and runs them in an mdrip subshell,
   leaving the executing shell unchanged.

   The stdout and stderr of the subprocess are captured, and used to
   report output from failing blocks, facilitating error diagnosis.

   In any other mode, mdrip exits with non-zero status only when used
   incorrectly, e.g. file not found, bad flags, etc.
   In --mode test, mdrip exits with the status of any failing code block.

 --mode demo

   Starts a web server (see --port and --hostname flag) to offer a
   rendered version of the markdown facilitating execution of
   command blocks.

   Key or mouse events copy code blocks to the user's clipboard
   and, if tmux is running, "paste" them to the active tmux window.

 --mode tmux

   Only useful if both a local tmux instance is running, and an mdrip
   is running remotely in '--mode demo'.

   In this mode the first argument to mdrip, normally treated as a
   markdown filename, is treated as a URL.  mdrip attempts to open a
   websocket to that URL.  Discover the URL to use from the help panel
   of the mdrip running in --mode demo.

   When a user clicks on a code block served by mdrip (in --mode demo)
   an attempt is made to find a websocket associated with the user's
   web session.

   If a socket is found, the code block is sent to the socket.  Upon
   receipt, mdrip (in --mode tmux) sends the block to local tmux as if
   the user had typed it.
`
)

// ModeType distinguishes the primary modes of execution in mdrip main.go.
// These could be separate programs, but don't want to require multiple downloads.
type ModeType int

const (
	modeUnknown ModeType = iota
	// ModePrint - extract code to stdout.
	ModePrint
	// ModeTest - run extracted code in a subshell, reporting errors.
	ModeTest
	// ModeDemo - render markdown in a webserver.
	ModeDemo
	// ModeTmux - run a tiny server that connects tmux to an mdrip in ModeDemo.
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

	blockTimeOut = flag.Duration("blockTimeOut", 1*time.Minute,
		`In --mode test, the max amount of time to wait for a command block to exit.`)

	ignoreTestFailure = flag.Bool("ignoreTestFailure", false,
		`In --mode test, exit with success regardless of extracted code failure.`)
)

// Config holds configuration for an instance of mdrip.
type Config struct {
	label   base.Label
	mode    ModeType
	dataSet *base.DataSet
}

func determineMode() ModeType {
	if len(*mode) == 0 {
		return ModePrint
	}
	if len(*mode) < 3 {
		return modeUnknown
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

// BlockTimeOut is the duration to give a block to run before considering it dead.
func (c *Config) BlockTimeOut() time.Duration {
	return *blockTimeOut
}

// Preambled is a count of blocks to run in the current shell
// before staring a subshell, to impact the env of the current shell.
func (c *Config) Preambled() int {
	return *preambled
}

// HostAndPort for the server when in ModeDemo.
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

// Mode returns the mode of the mdrip instance.
func (c *Config) Mode() ModeType {
	return c.mode
}

// IgnoreTestFailure means don't exit with error if a test fails in ModeTest.
func (c *Config) IgnoreTestFailure() bool {
	return *ignoreTestFailure
}

// Label to use when extracting code blocks for testing or printing.
func (c *Config) Label() base.Label {
	return c.label
}

// DataSet holds the source of data parsed from the mdrip command line.
func (c *Config) DataSet() *base.DataSet {
	return c.dataSet
}

// DefaultConfig is a config for tests.
func DefaultConfig() *Config {
	ds, _ := base.NewDataSet([]string{"foo"})
	return &Config{base.WildCardLabel, ModePrint, ds}
}

// GetConfig parses configuration from command line args.
func GetConfig() (*Config, error) {
	flag.Usage = Usage
	flag.Parse()
	dataSource, err := base.NewDataSet(flag.Args())
	if err != nil {
		return nil, err
	}
	desiredMode := determineMode()
	if desiredMode == modeUnknown {
		return nil, errors.New(`specify print, test, demo or tmux as the mode`)
	}
	if *ignoreTestFailure && desiredMode != ModeTest {
		return nil, errors.New(`makes no sense to specify --ignoreTestFailure without --mode test`)
	}
	return &Config{determineLabel(), desiredMode, dataSource}, nil
}

// Usage prints a usage message to stdErr.
func Usage() {
	fmt.Fprintf(os.Stderr, "\nUsage:  %s {fileName}...\n", os.Args[0])
	fmt.Fprint(os.Stderr, usageText)
	fmt.Fprint(os.Stderr, "\n\nFlags:\n\n")
	flag.PrintDefaults()
}
