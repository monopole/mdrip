package test

import (
	"fmt"
	"os"
	"time"

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
	debugging        = false
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
		unlikelyWordOut = "rumplestilskinOut"
		unlikelyWordErr = "rumplestilskinErr"
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
		if debugging {
			fmt.Println("==== running " + string(b.FirstLabel()))
		}
		c := shexec.NewRecallCommander(b.Code())
		if debugging {
			fmt.Println("------------------------------")
			fmt.Print(b.Code())
			fmt.Println("------------------------------")
		}
		if err := sh.Run(timeout, c); err != nil {
			if debugging {
				fmt.Println("returning from command with err = ", err.Error())
			}
			for _, line := range c.DataErr() {
				fmt.Fprintln(os.Stderr, line)
			}
			return fmt.Errorf("failure in code block %q", b.FirstLabel())
		}
		if debugging {
			fmt.Println("no error, going for next command")
		}
	}
	return sh.Stop(durationShutdown, "")
}
