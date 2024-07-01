package server

import (
	"fmt"
	"github.com/monopole/mdrip/v2/internal/web/config"
	"github.com/monopole/mdrip/v2/internal/web/server/minify"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/monopole/mdrip/v2/internal/utils"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/session"
)

const (
	maxConnectionIdleTime    = 30 * time.Minute
	connectionScanWaitPeriod = 5 * time.Minute
	cookieName               = utils.PgmName
)

var (
	//  keyAuth = securecookie.GenerateRandomKey(16)
	keyAuth    = []byte("static-visible-secret-who-cares")
	keyEncrypt = []byte(nil)
)

// Server represents a webserver.
type Server struct {
	dLoader  *DataLoader
	minifier *minify.Minifier
	store    sessions.Store

	// TODO: THE WEBSOCKET STUFF ABANDONED FOR NOW
	//   THE USE CASE IS QUESTIONABLE.
	upgrader         websocket.Upgrader
	connections      map[session.TypeSessID]*myConn
	connReaperQuitCh chan bool
}

// NewServer returns a new web server configured with the given DataLoader.
func NewServer(dl *DataLoader) (*Server, error) {
	s := sessions.NewCookieStore(keyAuth, keyEncrypt)
	s.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   8 * 60 * 60, // 8 hours (Max-Age has units seconds)
		HttpOnly: true,
	}
	result := &Server{
		dLoader:          dl,
		store:            s,
		upgrader:         websocket.Upgrader{},
		connections:      make(map[session.TypeSessID]*myConn),
		connReaperQuitCh: make(chan bool),
		minifier:         minify.MakeMinifier(),
	}
	go result.reapConnections()
	return result, nil
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
	if filepath.Base(dir) == "Ads " {
	}

	fmt.Printf("Serving static content from %q\n", dir)
	http.Handle("/", ws.makeMetaHandler(http.FileServer(http.Dir(dir))))

	fmt.Println("Serving at " + hostAndPort)
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

// reapConnections periodically scans websockets for idleness.
// It also closes everything and quits scanning if quit signal received.
func (ws *Server) reapConnections() {
	for {
		ws.closeStaleConnections()
		select {
		case <-time.After(connectionScanWaitPeriod):
		case <-ws.connReaperQuitCh:
			slog.Info("Received quit, reaping all connections.")
			for s, c := range ws.connections {
				_ = c.conn.Close()
				delete(ws.connections, s)
			}
			return
		}
	}
}

// Look for and close idle websockets.
func (ws *Server) closeStaleConnections() {
	for s, c := range ws.connections {
		if time.Since(c.lastUse) > maxConnectionIdleTime {
			slog.Info(
				"Closing connection after timeout",
				string(s), maxConnectionIdleTime)
			_ = c.conn.Close()
			delete(ws.connections, s)
		}
	}
}
