package tmux

import (
	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/utils"
	"github.com/spf13/cobra"
)

const cmdName = "tmux"

func NewCommand(ldr *loader.FsLoader) *cobra.Command {
	return &cobra.Command{
		Use:     cmdName,
		Short:   "Opens a websocket to a given URI, and forwards incoming messages to a local tmux instance.",
		Example: utils.PgmName + " " + cmdName + " {URI}",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
