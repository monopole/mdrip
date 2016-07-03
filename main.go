package main

import (
	"flag"
	"fmt"
	"github.com/monopole/mdrip/model"
	"github.com/monopole/mdrip/util"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var blockTimeOut = flag.Duration("blockTimeOut", 7*time.Second,
	"The max amount of time to wait for a command block to exit.")

// emitStraightScript simply prints the contents of scriptBuckets.
func emitStraightScript(w io.Writer, label model.Label, scriptBuckets []*model.ScriptBucket) {
	for _, bucket := range scriptBuckets {
		bucket.Dump(w, label, 0)
	}
	fmt.Fprintf(w, "echo \" \"\n")
	fmt.Fprintf(w, "echo \"All done.  No errors.\"\n")
}

// emitPreambledScript emits the first script normally, then emit it
// again, as well as the the remaining scripts, so that they run in a
// subshell.
//
// This allows the aggregrate script to be structured as 1) a preamble
// initialization script that impacts the environment of the active
// shell, followed by 2) a script that executes as a subshell that
// exits on error.  An exit in (2) won't cause the active shell (most
// likely a terminal) to close.
//
// The first script must be able to complete without exit on error
// because its not running as a subshell.  So it should just set
// environment variables and/or define shell funtions.
func emitPreambledScript(w io.Writer, label model.Label, scriptBuckets []*model.ScriptBucket, n int) {
	scriptBuckets[0].Dump(w, label, n)
	delim := "HANDLED_SCRIPT"
	fmt.Fprintf(w, " bash -euo pipefail <<'%s'\n", delim)
	fmt.Fprintf(w, "function handledTrouble() {\n")
	fmt.Fprintf(w, "  echo \" \"\n")
	fmt.Fprintf(w, "  echo \"Unable to continue!\"\n")
	fmt.Fprintf(w, "  exit 1\n")
	fmt.Fprintf(w, "}\n")
	fmt.Fprintf(w, "trap handledTrouble INT TERM\n")
	emitStraightScript(w, label, scriptBuckets)
	fmt.Fprintf(w, "%s\n", delim)
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

func main() {
	flag.Usage = usage
	preambled := flag.Int("preambled", -1,
		"Place all scripts in a subshell, preambled by the first {n} blocks in the first script.")
	subshell := flag.Bool("subshell", false,
		"Run extracted blocks in subshell (leaves your env vars and pwd unchanged).")
	swallow := flag.Bool("swallow", false,
		"Swallow errors from subshell (non-zero exit only on problems in driver code).")
	flag.Parse()
	if *swallow && !*subshell {
		fmt.Fprintf(os.Stderr, "Makes no sense to specify --swallow but not --subshell.\n")
		usage()
		os.Exit(1)
	}
	if flag.NArg() < 2 {
		usage()
		os.Exit(1)
	}
	label := model.Label(flag.Arg(0))
	scriptBuckets := make([]*model.ScriptBucket, flag.NArg()-1)

	for i := 1; i < flag.NArg(); i++ {
		fileName := flag.Arg(i)
		contents, err := ioutil.ReadFile(fileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read %q\n", fileName)
			usage()
			os.Exit(2)
		}
		m := util.Parse(string(contents))
		script, ok := m[label]
		if !ok {
			fmt.Fprintf(os.Stderr, "No block labelled %q in file %q.\n", label, fileName)
			os.Exit(3)
		}
		scriptBuckets[i-1] = model.NewScriptBucket(fileName, script)
	}

	if len(scriptBuckets) < 1 {
		return
	}

	if !*subshell {
		if *preambled >= 0 {
			emitPreambledScript(os.Stdout, label, scriptBuckets, *preambled)
		} else {
			emitStraightScript(os.Stdout, label, scriptBuckets)
		}
		return
	}

	result := util.RunInSubShell(scriptBuckets, *blockTimeOut)
	if result.GetProblem() != nil {
		util.Complain(result, label)
		if !*swallow {
			log.Fatal(result.GetProblem())
		}
	}
}
