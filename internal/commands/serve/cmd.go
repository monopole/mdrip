package serve

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/parsren"
	"github.com/monopole/mdrip/v2/internal/tmux"
	"github.com/monopole/mdrip/v2/internal/utils"
	"github.com/monopole/mdrip/v2/internal/web/server"
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
			slog.Error("trouble with hostname", "err", err)
		}
	}
	return hostname + ":" + strconv.Itoa(fl.port)
}

func makeTitle(t string, args []string) string {
	if len(t) > 0 {
		return t
	}
	if len(args) > 0 {
		return strings.Join(args, ",")
	}
	return "{test data}"
}

func NewCommand(ldr *loader.FsLoader, p parsren.MdParserRenderer) *cobra.Command {
	flags := myFlags{}
	c := &cobra.Command{
		Use:     cmdName,
		Short:   "Serve a markdown / code-running application",
		Example: utils.PgmName + " " + cmdName + " {path/to/folder}",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				// Serving is more restrictive than testing because the
				// file path being rendered is shown in the URL
				return fmt.Errorf(
					"specify a single relative path so that URL paths work correctly")
			}
			if len(args) == 0 {
				args = []string{string(loader.CurrentDir)}
			}
			dl := server.NewDataLoader(
				ldr, args, p, makeTitle(flags.title, args))
			// Heat up the cache, and see if the args are okay.
			if err := dl.LoadAndRender(); err != nil {
				return fmt.Errorf("data loader fail; %w", err)
			}
			s, err := server.NewServer(dl, getCommandRunner())
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

func getCommandRunner() io.Writer {
	tx, err := tmux.NewTmux(tmux.PgmName)
	if err != nil || tx == nil {
		slog.Warn(tmux.PgmName+" not available", "err", err)
		return &fakeTmux{}
	}
	if !tx.IsUp() {
		slog.Warn(tmux.PgmName + " executable present, but not running")
		return &fakeTmux{}
	}
	return tx
}

type fakeTmux struct{}

func (tx *fakeTmux) Write(bytes []byte) (int, error) {
	slog.Debug("Would run", "codeSnip", utils.Summarize(bytes))
	return 0, nil
}
