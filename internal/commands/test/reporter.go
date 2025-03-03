package test

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/shexec"
)

const (
	colReset = "\033[0m"
	colRed   = "\033[31m"
	colGreen = "\033[32m"
	colCyan  = "\033[36m"
	colGray  = "\033[90m"
	colWhite = "\033[97m"
)

type reporter struct {
	f     string
	count int
	size  int
	quiet bool
}

func makeReporter(quiet bool, blocks []*loader.CodeBlock) *reporter {
	maxPathLen, maxBlockNameLen := fieldSizes(blocks)
	countWidth := len(strconv.Itoa(len(blocks)))
	return &reporter{
		f: fmt.Sprintf("%%%dd/%%%dd  %%%ds  %%-%ds  ",
			countWidth, countWidth, maxPathLen, maxBlockNameLen),
		quiet: quiet,
		size:  len(blocks),
	}
}

func fieldSizes(
	blocks []*loader.CodeBlock) (maxPathLen, maxBlockNameLen int) {
	for _, b := range blocks {
		if len(b.Path()) > maxPathLen {
			maxPathLen = len(b.Path())
		}
		if len(b.Name()) > maxBlockNameLen {
			maxBlockNameLen = len(b.Name())
		}
	}
	return
}

func (r *reporter) header(b *loader.CodeBlock) {
	if r.quiet {
		return
	}
	r.count++
	fmt.Printf(r.f, r.count, r.size, b.Path(), b.Name())
}

func (r *reporter) skip() {
	if r.quiet {
		return
	}
	fmt.Print(colGray)
	fmt.Print("SKIP")
	fmt.Print(colReset)
	fmt.Println()
}

func (r *reporter) pass() {
	if r.quiet {
		return
	}
	fmt.Print(colGreen)
	fmt.Print("PASS")
	fmt.Print(colReset)
	fmt.Println()
}

func (r *reporter) fail(
	_ error, b *loader.CodeBlock, c *shexec.RecallCommander) {
	// TODO: Get a better error from the infrastructure for reporting.
	//  Right now it's something like "sentinel not found".
	//  Capture exit code from subprocess and report that instead.

	if !r.quiet {
		fmt.Print(colRed)
		fmt.Print("FAIL")
		fmt.Print(colReset)
		fmt.Println()
	}

	_, _ = fmt.Fprintf(os.Stderr, "%s %s:\n", b.Path(), b.Name())
	_, _ = fmt.Fprint(os.Stderr, colCyan)
	for _, line := range strings.Split(b.Code(), "\n") {
		if len(line) > 0 {
			_, _ = fmt.Fprintln(os.Stderr, " ", line)
		}
	}
	_, _ = fmt.Fprint(os.Stderr, colReset)
	dumpCapture("stdout", c.DataOut(), colWhite)
	dumpCapture("stderr", c.DataErr(), colRed)
}

func dumpCapture(kind string, lines []string, color string) {
	_, _ = fmt.Fprint(os.Stderr, kind, ":")
	if len(lines) == 0 {
		_, _ = fmt.Fprintln(os.Stderr, " <empty>")
		return
	}
	_, _ = fmt.Fprintln(os.Stderr)
	_, _ = fmt.Fprint(os.Stderr, color)
	for _, line := range lines {
		_, _ = fmt.Fprintf(os.Stderr, "  %s\n", line)
	}
	_, _ = fmt.Fprint(os.Stderr, colReset)
}
