package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func dump(label string, scriptBuckets []*ScriptBucket) {
	dashes := strings.Repeat("-", 70)
	for _, bucket := range scriptBuckets {

		fmt.Printf("#\n# Script @%s from %q \n#\n", label, bucket.fileName)
		delimFmt := "#" + dashes + "#  %s %d\n"
		for i, block := range bucket.script {
			allLabels := strings.Join(block.labels, " ")
			fmt.Printf(delimFmt, "Start", i+1)
			fmt.Printf("echo \"Block %d (%s) %s\"\n####\n", i+1, allLabels, bucket.fileName)
			fmt.Print(block.codeText)
			fmt.Printf(delimFmt, "End", i+1)
			fmt.Println()
		}
	}
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
  <!-- @foo -->
  '''
  cd $HOME
  '''
  Blah blah blah.
  <!-- @bar @apple -->
  '''
  echo "I am block bar"
  '''
  Blah blah blah.
  <!-- @foo @baz -->
  '''
  echo "I am block foo"
  '''
  Blah blah blah.

then the command '{this} foo {fileName}' emits: 

  cd $HOME
  echo "I am block foo."

Pipe output to 'source /dev/stdin' to run it directly.

Use --subshell to run the blocks in a subshell leaving your current
shell env vars and pwd unchanged.  The code blocks can, however, do
anything to your computer that you can.
`)
}

func main() {
	flag.Usage = usage
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
	label := flag.Arg(0)
	scriptBuckets := make([]*ScriptBucket, flag.NArg()-1)

	for i := 1; i < flag.NArg(); i++ {
		fileName := flag.Arg(i)
		contents, err := ioutil.ReadFile(fileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read %q\n", fileName)
			usage()
			os.Exit(2)
		}
		m := Parse(string(contents))
		script, ok := m[label]
		if !ok {
			fmt.Fprintf(os.Stderr, "Unable to find a block labelled %q in file %q.\n", label, fileName)
			os.Exit(3)
		}
		scriptBuckets[i-1] = &ScriptBucket{fileName, script}
	}

	if !*subshell {
		dump(label, scriptBuckets)
		return
	}

	result := RunInSubShell(scriptBuckets)
	if result.err != nil {
		Complain(result, label)
		if !*swallow {
			log.Fatal(result.err)
		}
	}
}
