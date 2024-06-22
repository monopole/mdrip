package serve

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/parsren"
	"github.com/monopole/mdrip/v2/internal/utils"
	"github.com/monopole/mdrip/v2/internal/webserver"
	"github.com/spf13/cobra"
)

const cmdName = "serve"

type myFlags struct {
	port        int
	title       string
	useHostName bool
}

// hostAndPort for the server.
func (fl *myFlags) hostAndPort() string {
	hostname := "" // docker breaks if one uses localhost here
	if fl.useHostName {
		var err error
		hostname, err = os.Hostname()
		if err != nil {
			slog.Error("Trouble with hostname: %v", err)
		}
	}
	return hostname + ":" + strconv.Itoa(fl.port)
}

func makeTitle(t string, args []string) string {
	if len(t) > 0 {
		return t
	}
	title := "markdown from "
	if len(args) > 0 {
		return title + strings.Join(args, ",")
	}
	return title + "test data"
}

func NewCommand(ldr *loader.FsLoader, p parsren.MdParserRenderer) *cobra.Command {
	flags := myFlags{}
	c := &cobra.Command{
		Use:     cmdName,
		Short:   "Serves a markdown / code-running application at a particular port on localhost.",
		Example: utils.PgmName + " " + cmdName + " {path/to/folder}",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				args = []string{string(loader.CurrentDir)}
			}
			dl := webserver.NewDataLoader(
				ldr, args, p, makeTitle(flags.title, args))
			// Heat up the cache, and see if the args are okay.
			if err := dl.LoadAndRender(); err != nil {
				return fmt.Errorf("data loader fail; %w", err)
			}
			s, err := webserver.NewServer(dl)
			if err != nil {
				return err
			}
			return s.Serve(flags.hostAndPort())
		},
	}
	// TODO: pull title from the first header of the first markdown file?
	c.Flags().StringVar(
		&flags.title,
		"title",
		"",
		"Text to use as a title for the webpage.")
	c.Flags().IntVar(
		&flags.port,
		"port",
		8080,
		"Port at which to serve HTTP requests for the demo.")
	c.Flags().BoolVar(
		&flags.useHostName,
		"use-host-name",
		false,
		"Use the 'hostname' utility to specify where to serve, else implicitly use 'localhost'.")
	return c
}
