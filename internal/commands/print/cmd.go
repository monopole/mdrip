package print

import (
	"fmt"
	"log/slog"
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
	upTo  int
	debug bool
}

const shortHelp = "Print code blocks below the given path as a shell script"

func NewCommand(ldr *loader.FsLoader, p parsren.MdParserRenderer) *cobra.Command {
	flags := myFlags{}
	c := &cobra.Command{
		Use:   cmdName + " {path}",
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
			if len(args) < 1 {
				return fmt.Errorf("specify a path")
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
			blocks := p.Filter(filter)
			if flags.upTo > len(blocks) {
				return fmt.Errorf("only %d blocks passed the filter", len(blocks))
			}
			if flags.upTo > 0 {
				blocks = blocks[:flags.upTo]
			}
			loader.PrintBlocks(os.Stdout, blocks)
			return nil
		},
	}
	c.Flags().IntVar(
		&flags.upTo,
		"upto",
		0,
		"Print blocks up to and including the given index; omit remaining blocks. Use 'list' command to see indices.")
	c.Flags().StringVar(
		&flags.label,
		"label",
		"",
		"Print only the code blocks that have this label")
	if utils.AllowDebug {
		c.Flags().BoolVar(
			&flags.debug,
			"debug",
			false,
			"Use hard-coded markdown test data instead of reading from current directory")
	}
	return c
}
