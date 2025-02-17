package test

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/parsren"
	"github.com/monopole/shexec"
	"github.com/monopole/shexec/channeler"
	"github.com/spf13/cobra"
)

const (
	cmdName          = "test"
	durationStartup  = 10 * time.Second
	durationShutdown = 3 * time.Second
	rumple           = "rumpleStiltSkin"

	colReset = "\033[0m"
	colRed   = "\033[31m"
	colCyan  = "\033[36m"
	colWhite = "\033[97m"
)

type myFlags struct {
	label        string
	blockTimeOut time.Duration
}

const shortHelp = "Tests the code blocks extracted from markdown"

func NewCommand(ldr *loader.FsLoader, p parsren.MdParserRenderer) *cobra.Command {
	flags := myFlags{}
	c := &cobra.Command{
		Use:   cmdName + " [{path}]",
		Short: shortHelp,
		Long: shortHelp + `

This command silently runs extracted code blocks, optionally selected by label.

Any block labelled with @` + string(loader.SkipLabel) + ` will be ignored.

The command fails (non-zero exit code) if an extracted code block fails.

Output is constrained to show only the content of the failing code block
and its output and error streams.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fld, err := ldr.LoadTrees(args)
			if err != nil {
				return err
			}
			fld.Accept(p)
			label := loader.WildCardLabel
			if flags.label != "" {
				label = loader.Label(flags.label)
			}
			return runTheBlocks(
				p.Filter(func(b *loader.CodeBlock) bool {
					return b.HasLabel(label) && !b.HasLabel(loader.SkipLabel)
				}),
				flags.blockTimeOut)
		},
		SilenceUsage: true,
	}
	c.Flags().StringVar(
		&flags.label,
		"label",
		"",
		"Extract only code blocks with this label.")
	c.Flags().DurationVar(
		&flags.blockTimeOut,
		"block-time-out",
		30*time.Second,
		"The max amount of time to wait for a command block to exit.")

	return c
}

func runTheBlocks(blocks []*loader.CodeBlock, timeout time.Duration) error {
	const (
		unlikelyWordOut = rumple + "Out"
		unlikelyWordErr = rumple + "Err"
	)
	sh := shexec.NewShell(shexec.Parameters{
		Params: channeler.Params{Path: "/bin/bash", Args: []string{"-e"}},
		SentinelOut: shexec.Sentinel{
			C: "echo " + unlikelyWordOut,
			V: unlikelyWordOut,
		},
		SentinelErr: shexec.Sentinel{
			C: "echo " + unlikelyWordErr + " 1>&2",
			V: unlikelyWordErr,
		},
	})
	if err := sh.Start(durationStartup); err != nil {
		return err
	}
	for _, b := range blocks {
		c := shexec.NewRecallCommander(b.Code())
		if err := sh.Run(timeout, c); err != nil {
			return reportError(err, b, c)
		}
	}
	return sh.Stop(durationShutdown, "")
}

func reportError(
	_ error, b *loader.CodeBlock, c *shexec.RecallCommander) error {
	// TODO: Get a better error from the infrastructure for reporting.
	//  Right now it's something like "sentinel not found".
	//  Capture exit code from subprocess and report that instead.

	_, _ = fmt.Fprintf(os.Stderr, "Block '%s':\n", b.FirstLabel())
	_, _ = fmt.Fprint(os.Stderr, colCyan)
	for _, line := range strings.Split(b.Code(), "\n") {
		if len(line) > 0 {
			_, _ = fmt.Fprintln(os.Stderr, " ", line)
		}
	}
	_, _ = fmt.Fprint(os.Stderr, colReset)
	printStream("stdout", c.DataOut(), colWhite)
	printStream("stderr", c.DataErr(), colRed)
	return fmt.Errorf("code block %q failed", b.FirstLabel())
}

func printStream(kind string, lines []string, color string) {
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
