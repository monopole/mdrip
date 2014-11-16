package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func dump(fileName, label string, scripts []string) {
	fmt.Printf("#\n# Script @%s from %s\n#\n", label, fileName)
	delimFmt := "#" + strings.Repeat("-", 70) + "#  %s %d\n"
	for i, script := range scripts {
		fmt.Printf(delimFmt, "Start", i+1)
		fmt.Print(script)
		fmt.Printf(delimFmt, "End", i+1)
		fmt.Println()
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "\nUsage:  %s {fileName} {label}\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr,
		`
Reads a markdown file, extracts scripts with a given @label,
and either runs them in a subshell or emits them to stdout.

If the markdown file contains

  Blah blah blah.
  <!-- @foo -->
  '''
  cd $HOME
  '''
  Blah blah blah.
  <!-- @bar @apple -->
  '''
  echo "I am script bar"
  '''
  Blah blah blah.
  <!-- @foo @baz -->
  '''
  echo "I am script foo"
  '''
  Blah blah blah.

then the command '{thisProgram} {fileName} foo' emits: 

  cd $HOME
  echo "I am script foo."

Pipe output to 'source /dev/stdin' to run it directly.

Use --subshell to run it in a subshell leaving your current shell env
vars and pwd unchanged (the scripts can, however, do anything to your
computer, file system, etc.).
`)
}

func main() {
	flag.Usage = usage
	subshell := flag.Bool("subshell", false, "Run extracted scripts in subshell (leaves your env vars and pwd unchanged).")
	swallow := flag.Bool("swallow", false, "Swallow errors from subshell (non-zero exit only on problems in driver code).")
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
	fileName := flag.Arg(0)
	label := flag.Arg(1)
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read %q\n", fileName)
		usage()
		os.Exit(2)
	}

	m := Parse(string(contents))
	scripts, ok := m[label]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unable to find a script labelled %q.\n", label)
		os.Exit(3)
	}
	if !*subshell {
		dump(fileName, label, scripts)
		return
	}
	result := RunInSubShell(scripts)
	if result.err != nil {
		Complain(result, label, fileName)
		if !*swallow {
			log.Fatal(result.err)
		}
	}
}
