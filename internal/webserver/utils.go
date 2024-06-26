package webserver

import (
	"errors"
	"fmt"
	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/tmux"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/monopole/mdrip/v2/internal/parsren"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/session"
)

func (ws *Server) getRenderedMdFile(req *http.Request) (*parsren.RenderedMdFile, error) {
	mdFileIndex := getIntParam(session.KeyMdFileIndex, req, -1)
	files := ws.dLoader.RenderedFiles()
	if mdFileIndex < 0 || mdFileIndex > len(files) {
		return nil, fmt.Errorf(
			"mdFileIndex==%d out of range 0..%d", mdFileIndex, len(files))
	}
	return files[mdFileIndex], nil
}

// reload performs a data reload.
func (ws *Server) reload(wr http.ResponseWriter, req *http.Request) error {
	mySess, _ := ws.store.Get(req, cookieName)
	_ = mySess.Save(req, wr)
	ws.dLoader.makeLastLoadTimeVeryOld()
	return ws.dLoader.LoadAndRender()
}

func getIntParam(n string, r *http.Request, d int) int {
	v, err := strconv.Atoi(r.URL.Query().Get(n))
	if err != nil {
		return d
	}
	return v
}

func getBoolParam(n string, r *http.Request, d bool) bool {
	v, err := strconv.ParseBool(r.URL.Query().Get(n))
	if err != nil {
		return d
	}
	return v
}

func write500(w http.ResponseWriter, e error) {
	slog.Error(e.Error())
	http.Error(w, e.Error(), http.StatusInternalServerError)
}

func inRange(wr http.ResponseWriter, name string, arg, n int) bool {
	if arg >= 0 || arg < n {
		return true
	}
	http.Error(wr,
		fmt.Sprintf("%s %d out of range 0-%d",
			name, arg, n-1), http.StatusBadRequest)
	return false
}

func (ws *Server) attemptTmuxWrite(b *loader.CodeBlock) error {
	// For debugging add: b.Dump(os.Stderr)
	tx := tmux.NewTmux(tmux.Path)
	if !tx.IsUp() {
		return errors.New("no local tmux to write to")
	}
	_, err := tx.Write([]byte(b.Code()))
	return err
}
