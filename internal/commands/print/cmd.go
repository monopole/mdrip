package print

import (
	"fmt"
	"os"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/parsren"
	"github.com/monopole/mdrip/v2/internal/utils"
	"github.com/spf13/cobra"
)

const (
	cmdName = "print"
)

type myFlags struct {
	label string
	debug bool
}

func NewCommand(ldr *loader.FsLoader, p parsren.MdParserRenderer) *cobra.Command {
	flags := myFlags{}
	c := &cobra.Command{
		Use:     cmdName,
		Short:   "Prints an extracted shell script",
		Example: utils.PgmName + " " + cmdName + " {path/to/folder}",
		RunE: func(cmd *cobra.Command, args []string) error {
			fld, err := ldr.LoadTrees(args)
			if err != nil {
				return err
			}
			if fld == nil {
				fmt.Println("No markdown found.")
				return nil
			}
			if flags.debug {
				loader.NewVisitorDump(os.Stdout).VisitFolder(fld)
			}
			fld.Accept(p)
			loader.DumpBlocks(os.Stdout, p.FilteredBlocks(loader.WildCardLabel))
			return nil
		},
		SilenceUsage: true,
	}
	c.Flags().StringVar(
		&flags.label,
		"label",
		"",
		"Extract only code blocks with this label.")
	c.Flags().BoolVar(
		&flags.debug,
		"debug",
		false,
		"Use hard coded markdown test data instead of reading from current directory.")

	return c
}
