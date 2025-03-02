package raw

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/utils"
	"github.com/spf13/cobra"
)

const cmdName = "raw"

type myFlags struct {
	port        int
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

func NewCommand() *cobra.Command {
	flags := myFlags{}
	c := &cobra.Command{
		Use:     cmdName,
		Short:   "Serve raw markdown from the given path (debugging)",
		Example: utils.PgmName + " " + cmdName + " {path/to/folder}",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("specify a single relative path")
			}
			if len(args) == 0 {
				args = []string{string(loader.CurrentDir)}
			}
			return doIt(args[0], flags.hostAndPort())
		},
	}
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

func doIt(dir string, hostAndPort string) error {
	slog.Debug("Serving from " + dir)
	return http.ListenAndServe(
		hostAndPort, logUrl(http.FileServer(http.Dir(dir))))
}

func logUrl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		slog.Debug(req.URL.String())
		next.ServeHTTP(w, req)
	})
}
