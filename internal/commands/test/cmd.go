package test

import (
	"fmt"
	"log/slog"
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
			blocks := p.FilteredBlocks(label)
			if debugging {
				loader.DumpBlocks(os.Stdout, blocks)
				// fld.Accept(loader.NewVisitorDump(os.Stdout))
			}
			return runTheBlocks(blocks, flags.blockTimeOut)
		},
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
		Params: channeler.Params{Path: "/bin/sh"},
		SentinelOut: shexec.Sentinel{
			C: "echo " + unlikelyWord,
			V: unlikelyWord,
		},
		// TODO: try:
		//  SentinelErr: shexec.Sentinel{
		//	  C: unlikelyWord,
		//	  V: `unrecognized command: "` + unlikelyWord + `"`,
		//  },
	})
	if err := sh.Start(durationStartup); err != nil {
		return err
	}
	for _, b := range blocks {
		slog.Info("running", "command", b.FirstLabel())
		c := shexec.NewLabellingCommander(b.Code())
		// TODO: try: c := &shexec.PassThruCommander{C: blocks[i].Code()}
		fmt.Println(b.Code())
		if err := sh.Run(timeout, c); err != nil {
			fmt.Println("err = ", err.Error())
			return err
		}
		fmt.Println("no error")
	}
	if err := sh.Stop(durationShutdown, ""); err != nil {
		return err
	}
	slog.Info("All done.")
	return nil
}
