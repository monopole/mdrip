package list

import (
	"log/slog"
	"os"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/parsren"
	"github.com/spf13/cobra"
)

const (
	cmdName   = "list"
	shortHelp = "List titles and index numbers of code blocks below the given path"
)

func NewCommand(ldr *loader.FsLoader, p parsren.MdParserRenderer) *cobra.Command {
	c := &cobra.Command{
		Use:   cmdName + " [{path}]",
		Short: shortHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			fld, err := ldr.LoadTrees(args)
			if err != nil {
				return err
			}
			if fld == nil {
				slog.Warn("No markdown found.")
				return nil
			}
			fld.Accept(p)
			loader.PrintTitles(os.Stdout, p.Filter(parsren.AllBlocks))
			return nil
		},
		SilenceUsage: true,
	}
	return c
}
