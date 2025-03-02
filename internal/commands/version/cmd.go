package version

import (
	"fmt"

	"github.com/monopole/mdrip/v2/internal/provenance"
	"github.com/monopole/mdrip/v2/internal/utils"
	"github.com/spf13/cobra"
)

const cmdName = "version"

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:     cmdName,
		Short:   "Show program version",
		Example: utils.PgmName + " " + cmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("%v\n", provenance.GetProvenance())
			return nil
		},
	}
}
