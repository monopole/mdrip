package test

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/TwiN/go-color"
	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/parsren"
	"github.com/monopole/mdrip/v2/internal/utils"
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
	label        string
	blockTimeOut time.Duration
}

func NewCommand(ldr *loader.FsLoader, p parsren.MdParserRenderer) *cobra.Command {
	flags := myFlags{}
	c := &cobra.Command{
		Use:   cmdName,
		Short: "Tests an extracted shell script",
		Long: `Tests an extracted shell script.
This is experimental, to see if we can get a better exerience
than simply piping into "bash -e".
`,
		Example: utils.PgmName + " " + cmdName + " {path/to/folder}",
		RunE: func(cmd *cobra.Command, args []string) error {
			fld, err := ldr.LoadTrees(args)
			if err != nil {
				return err
			}
			label := loader.WildCardLabel
			if flags.label != "" {
				label = loader.Label(flags.label)
			}
			fld.Accept(p)
			return runTheBlocks(p.FilteredBlocks(label), flags.blockTimeOut)
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
	_, _ = fmt.Fprintf(
		os.Stderr, "Block '%s':\n", b.FirstLabel())
	fmt.Fprint(os.Stderr, color.Cyan)
	for _, line := range strings.Split(b.Code(), "\n") {
		if len(line) > 0 {
			_, _ = fmt.Fprintln(os.Stderr, " ", line)
		}
	}
	fmt.Fprint(os.Stderr, color.Reset)
	_, _ = fmt.Fprintln(os.Stderr, "Output streams:")
	for _, line := range c.DataOut() {
		_, _ = fmt.Fprintf(
			os.Stderr, "  out: %s%s%s\n", color.White, line, color.Reset)
	}
	for _, line := range c.DataErr() {
		_, _ = fmt.Fprintf(
			os.Stderr, "  err: %s%s%s\n", color.Red, line, color.Reset)
	}
	return fmt.Errorf("code block %q failed", b.FirstLabel())
}
