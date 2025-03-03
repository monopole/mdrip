package test

import (
	"fmt"
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
)

type myFlags struct {
	quiet        bool
	label        string
	blockTimeOut time.Duration
}

const shortHelp = "Test code blocks extracted from markdown"

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
		RunE: func(_ *cobra.Command, args []string) error {
			fld, err := ldr.LoadTrees(args)
			if err != nil {
				return err
			}
			fld.Accept(p)
			filter := parsren.AllBlocks
			if flags.label != "" {
				filter = func(b *loader.CodeBlock) bool {
					return b.HasLabel(loader.Label(flags.label))
				}
			}
			return runTheBlocks(
				p.Filter(filter), flags.quiet, flags.blockTimeOut)
		},
		SilenceUsage: true,
	}
	c.Flags().StringVar(
		&flags.label,
		"label",
		"",
		"Extract only code blocks with this label.")
	c.Flags().BoolVar(
		&flags.quiet,
		"quiet",
		false,
		"Suppress printing of code block names during test.")
	c.Flags().DurationVar(
		&flags.blockTimeOut,
		"block-time-out",
		30*time.Second,
		"The max amount of time to wait for a command block to exit.")

	return c
}

func runTheBlocks(
	blocks []*loader.CodeBlock, quiet bool, timeout time.Duration) error {
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
	r := makeReporter(quiet, blocks)
	for _, b := range blocks {
		r.header(b)
		if b.HasLabel(loader.SkipLabel) {
			r.skip()
			continue
		}
		c := shexec.NewRecallCommander(b.Code())
		if err := sh.Run(timeout, c); err != nil {
			r.fail(err, b, c)
			return fmt.Errorf("code block %q failed", b.Name())
		}
		r.pass()
	}
	return sh.Stop(durationShutdown, "")
}
