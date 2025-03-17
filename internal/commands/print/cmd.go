package print

import (
	"log/slog"
	"os"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/parsren"
	"github.com/monopole/mdrip/v2/internal/utils"
	"github.com/spf13/cobra"
	"errors"
)

const (
	cmdName = "print"
)

type myFlags struct {
	label string
	debug bool
}

const shortHelp = "Print extracted code blocks as a shell script"

func NewCommand(ldr *loader.FsLoader, p parsren.MdParserRenderer) *cobra.Command {
	flags := myFlags{}
	c := &cobra.Command{
		Use:   cmdName + " [{path}]",
		Short: shortHelp,
		Long: shortHelp + `

Any block labelled with @` + string(loader.SkipLabel) + ` will be ignored.

To have the effect of a test, pipe the output of this
command into a shell, e.g.

  ` + utils.PgmName + ` ` + cmdName + ` --label foo . | /bin/bash -e

The entire pipe succeeds only if all the extracted blocks succeed.

If your intention is to test, the command '` + utils.PgmName + ` test' yields
cleaner output, showing only the failing block and its output streams.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("path is required, if you really want to do this directory, provide './'")
			}
			
			fld, err := ldr.LoadTrees(args)
			if err != nil {
				return err
			}
			if fld == nil {
				slog.Warn("No markdown found.")
				return nil
			}
			if flags.debug {
				loader.NewVisitorDump(os.Stdout).VisitFolder(fld)
			}
			fld.Accept(p)
			filter := parsren.AllBlocksButSkip
			if flags.label != "" {
				filter = func(b *loader.CodeBlock) bool {
					return b.HasLabel(loader.Label(flags.label)) &&
						!b.HasLabel(loader.SkipLabel)
				}
			}
			loader.DumpBlocks(os.Stdout, p.Filter(filter))
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
