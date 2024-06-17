package test

import (
	"fmt"
	"log/slog"
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
	debugging        = true
)

type myFlags struct {
	label        string
	blockTimeOut time.Duration
}

func NewCommand(ldr *loader.FsLoader, p parsren.MdParserRenderer) *cobra.Command {
	flags := myFlags{}
	c := &cobra.Command{
		Use:     cmdName,
		Short:   "Tests an extracted shell script",
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
	const unlikelyWord = "rumplestilskin"
	sh := shexec.NewShell(shexec.Parameters{
		Params: channeler.Params{Path: "/bin/bash", Args: []string{"-e"}},
		SentinelOut: shexec.Sentinel{
			C: "echo " + unlikelyWord,
			V: unlikelyWord,
		},
		// TODO: try:
		//  SentinelErr: shexec.Sentinel{
		//	  C: unlikelyWord,
		//	  V: `unrecognized command: "` + unlikelyWord + `"`,
		//  },
		//SentinelErr: shexec.Sentinel{
		//	C: unlikelyWord + "Err",
		//	V: unlikelyWord + `Err: command not found`,
		//},
	})
	if err := sh.Start(durationStartup); err != nil {
		return err
	}
	for _, b := range blocks {
		if debugging {
			fmt.Println("==== running " + string(b.FirstLabel()))
		}
		c := shexec.NewRecallCommander(b.Code())
		// TODO: there's a race condition in that when a bad command hits,
		// and the process dies, the error from that bad command somehow
		// slips in behind the error encountered when the shexec infra tries
		// to write the next command.  This means that the error reported
		// is the one from the infra, not the one from the process - and
		// we want to see the latter.
		// c := &shexec.PassThruCommander{C: b.Code()}
		if debugging {
			fmt.Println("------------------------------")
			fmt.Print(b.Code())
			fmt.Println("------------------------------")
		}
		if err := sh.Run(timeout, c); err != nil {
			if debugging {
				fmt.Println("returning from command with err = ", err.Error())
				fmt.Println("stdErr=", c.DataErr())
			}
			return err
		}
		if debugging {
			fmt.Println("no error, going for next command")
		}
	}
	if err := sh.Stop(durationShutdown, ""); err != nil {
		return err
	}
	slog.Info("All done.")
	return nil
}
