package main

import (
	"os"

	"github.com/monopole/mdrip/v2/internal/commands/generatetestdata"
	"github.com/monopole/mdrip/v2/internal/commands/print"
	"github.com/monopole/mdrip/v2/internal/commands/raw"
	"github.com/monopole/mdrip/v2/internal/commands/serve"
	"github.com/monopole/mdrip/v2/internal/commands/test"
	"github.com/monopole/mdrip/v2/internal/commands/version"
	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/parsren/usegold"
	"github.com/monopole/mdrip/v2/internal/provenance"
	"github.com/monopole/mdrip/v2/internal/utils"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const (
	shortHelp = "Extract and manipulate code blocks from markdown."
)

func newCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   utils.PgmName + " {path}",
		Short: shortHelp,
		Long:  shortHelp + " (" + provenance.GetProvenance().Version + ")",
	}
	ldr := loader.New(
		afero.NewOsFs(), loader.IsMarkDownFile, loader.InNotIgnorableFolder)
	p := usegold.NewGParser()
	c.AddCommand(
		print.NewCommand(ldr, p),
		raw.NewCommand(),
		serve.NewCommand(ldr, p),
		test.NewCommand(ldr, p),
		version.NewCommand(),
		generatetestdata.NewCommand(),
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
