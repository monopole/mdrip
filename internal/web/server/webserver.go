package server

import (
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/monopole/mdrip/v2/internal/utils"
	"github.com/monopole/mdrip/v2/internal/web/config"
	"github.com/monopole/mdrip/v2/internal/web/server/minify"
)

const (
	cookieName = utils.PgmName
)

var (
	//  keyAuth = securecookie.GenerateRandomKey(16)
	keyAuth    = []byte("static-visible-secret-who-cares")
	keyEncrypt = []byte(nil)
)

// Server represents a webserver.
type Server struct {
	// dLoader loads markdown to serve.
	dLoader *DataLoader
	// minifier minifies generates html, js and css before serving it.
	minifier *minify.Minifier
	// store manages cookie state - experimental, not sure that
	// it's useful to store app state.  FWIW, it attempts to put you on the same
	// codeblock if you reload (start a new session).
	store sessions.Store
	// codeWriter accepts codeblocks for execution or simply printing.
	codeWriter io.Writer
}

// NewServer returns a new web server.
func NewServer(dl *DataLoader, r io.Writer) (*Server, error) {
	s := sessions.NewCookieStore(keyAuth, keyEncrypt)
	s.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   8 * 60 * 60, // 8 hours (Max-Age has units seconds)
		HttpOnly: true,
	}
	return &Server{
		dLoader:    dl,
		store:      s,
		minifier:   minify.MakeMinifier(),
		codeWriter: r,
	}, nil
}

// Serve offers an HTTP service.
func (ws *Server) Serve(hostAndPort string) (err error) {
	http.HandleFunc("/favicon.ico", ws.handleFavicon)
	http.HandleFunc(config.Dynamic(config.RouteImage), ws.handleImage)
	http.HandleFunc(config.Dynamic(config.RouteQuit), ws.handleQuit)
	http.HandleFunc(config.Dynamic(config.RouteDebug), ws.handleDebugPage)
	http.HandleFunc(config.Dynamic(config.RouteReload), ws.handleReload)
	// http.Handle(session.Dynamic(session.RouteWebSocket), ws.openWebSocket)
	http.HandleFunc(config.Dynamic(config.RouteJs), ws.handleGetJs)
	http.HandleFunc(config.Dynamic(config.RouteCss), ws.handleGetCss)
	http.HandleFunc(config.Dynamic(config.RouteLabelsForFile), ws.handleGetLabelsForFile)
	http.HandleFunc(config.Dynamic(config.RouteHtmlForFile), ws.handleGetHtmlForFile)
	http.HandleFunc(config.Dynamic(config.RouteRunBlock), ws.handleRunCodeBlock)
	http.HandleFunc(config.Dynamic(config.RouteSave), ws.handleSaveSession)

	// In server mode, the dLoader.paths slice has exactly one entry.
	dir := strings.TrimSuffix(ws.dLoader.paths[0], "/")
	slog.Info("Serving static content from ", "dir", dir)
	http.Handle("/", ws.makeMetaHandler(http.FileServer(http.Dir(dir))))

	slog.Info("Serving at " + hostAndPort)
	if err = http.ListenAndServe(hostAndPort, nil); err != nil {
		slog.Error("unable to start server", err)
	}
	return err
}

func (ws *Server) makeMetaHandler(fsHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/") ||
			// trigger markdown rendering
			strings.HasSuffix(req.URL.Path, ".md") {
			ws.handleRenderWebApp(w, req)
			return
		}
		// just serve a file.
		fsHandler.ServeHTTP(w, req)
	})
}