package main

import (
	"os"

	"github.com/monopole/mdrip/v2/internal/commands/demo"
	"github.com/monopole/mdrip/v2/internal/commands/gentestdata"
	"github.com/monopole/mdrip/v2/internal/commands/print"
	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/parsren/usegold"
	"github.com/monopole/mdrip/v2/internal/utils"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const (
	shortHelp = "Extract and manipulate code blocks from a markdown tree."
)

func newCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   utils.PgmName + " {path}",
		Short: shortHelp,
		Long:  shortHelp + " (" + utils.Version + ")",
	}
	ldr := loader.New(afero.NewOsFs())
	p := usegold.NewGParser()
	c.AddCommand(
		demo.NewCommand(ldr, p),
		gentestdata.NewCommand(),
		print.NewCommand(ldr, p),
		// "test" disabled until the UX improves - pipe "print" into "bash -e" instead.
		// test.NewCommand(ldr, p),
		// "tmux" websocket service disabled until a reasonable use case found.
		// the concept works fine on localhost without a websocket.
		// tmux.NewCommand(ldr),
	)
	return c
}

func main() {
	if err := newCommand().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
