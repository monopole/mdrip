package webserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	htmlTmpl "html/template"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	textTmpl "text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/parsren"
	"github.com/monopole/mdrip/v2/internal/tmux"
	"github.com/monopole/mdrip/v2/internal/utils"
	"github.com/monopole/mdrip/v2/internal/webapp"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/mdrip"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/session"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
)

const (
	maxConnectionIdleTime    = 30 * time.Minute
	connectionScanWaitPeriod = 5 * time.Minute
	cookieName               = utils.PgmName
	minifyJs                 = true
	minifyCss                = true
)

var (
	//  keyAuth = securecookie.GenerateRandomKey(16)
	keyAuth    = []byte("static-visible-secret-who-cares")
	keyEncrypt = []byte(nil)
)

// Server represents a webserver.
type Server struct {
	dLoader *DataLoader
	store   sessions.Store

	// TODO: THE WEBSOCKET STUFF ABANDONED FOR NOW
	//   THE USE CASE IS QUESTIONABLE.
	upgrader         websocket.Upgrader
	connections      map[session.TypeSessID]*myConn
	connReaperQuitCh chan bool
	minifier         *minify.M
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
		minifier:         minify.New(),
	}
	result.minifier.AddFunc(webapp.MimeJs, js.Minify)
	result.minifier.AddFunc(webapp.MimeCss, css.Minify)
	go result.reapConnections()
	return result, nil
}

// Serve offers an HTTP service.
func (ws *Server) Serve(hostAndPort string) (err error) {
	r := mux.NewRouter()
	r.HandleFunc("/favicon.ico", ws.handleFavicon)
	r.HandleFunc("/_/image", ws.handleImage)
	r.HandleFunc("/_/q", ws.handleQuit)
	r.HandleFunc("/_/d", ws.handleDebugPage)
	r.HandleFunc("/_/r", ws.handleReload)
	//r.HandleFunc("/_/ws", ws.openWebSocket)
	r.HandleFunc(session.PathGetJs, ws.handleGetJs)
	r.HandleFunc(session.PathGetCss, ws.handleGetCss)
	r.HandleFunc(session.PathGetLabelsForFile, ws.handleGetLabelsForFile)
	r.HandleFunc(session.PathGetHtmlForFile, ws.handleGetHtmlForFile)
	r.HandleFunc(session.PathRunBlock, ws.handleRunCodeBlock)
	r.HandleFunc(session.PathSave, ws.handleSaveSession)
	r.PathPrefix("/").HandlerFunc(ws.handleRenderWebApp)
	fmt.Println("Serving at " + hostAndPort)
	if err = http.ListenAndServe(hostAndPort, r); err != nil {
		slog.Error("unable to start server", err)
	}
	return err
}

func (ws *Server) handleSaveSession(w http.ResponseWriter, r *http.Request) {
	slog.Info("Saving session", "req", r.URL)
	s, err := ws.store.Get(r, cookieName)
	if err != nil {
		write500(w, err)
		return
	}
	s.Values[session.KeyIsNavOn] = getBoolParam(session.KeyIsNavOn, r, false)
	s.Values[session.KeyIsTitleOn] = getBoolParam(session.KeyIsTitleOn, r, false)
	s.Values[session.KeyMdFileIndex] = getIntParam(session.KeyMdFileIndex, r, 0)
	s.Values[session.KeyBlockIndex] = getIntParam(session.KeyBlockIndex, r, 0)
	if err = s.Save(r, w); err != nil {
		slog.Error("Unable to save session: %v", err)
	}
	_, _ = fmt.Fprintln(w, "Ok")
	slog.Info("Saved session.")
}

func (ws *Server) handleGetHtmlForFile(wr http.ResponseWriter, req *http.Request) {
	slog.Info("handleGetHtmlForFile ", "req", req.URL)
	f, err := ws.getRenderedMdFile(req)
	if err != nil {
		write500(wr, fmt.Errorf("handleGetHtmlForFile render; %w", err))
		return
	}
	_, err = wr.Write([]byte(f.Html))
	if err != nil {
		write500(wr, fmt.Errorf("handleGetHtmlForFile write; %w", err))
		return
	}
	slog.Info("handleGetHtmlForFile success")
}

func (ws *Server) handleGetLabelsForFile(wr http.ResponseWriter, req *http.Request) {
	slog.Info("handleGetLabelsForFile ", "req", req.URL)
	f, err := ws.getRenderedMdFile(req)
	if err != nil {
		write500(wr, fmt.Errorf("handleGetLabelsForFile render; %w", err))
		return
	}
	var jsn []byte
	jsn, err = json.Marshal(loader.NewLabelList(f.Blocks))
	if err != nil {
		write500(wr, fmt.Errorf("handleGetLabelsForFile marshal; %w", err))
		return
	}
	if _, err = wr.Write(jsn); err != nil {
		write500(wr, fmt.Errorf("handleGetLabelsForFile write; %w", err))
		return
	}
	slog.Info("handleGetLabelsForFile success")
}

func (ws *Server) getRenderedMdFile(req *http.Request) (*parsren.RenderedMdFile, error) {
	mdFileIndex := getIntParam(session.KeyMdFileIndex, req, -1)
	files := ws.dLoader.RenderedFiles()
	if mdFileIndex < 0 || mdFileIndex > len(files) {
		return nil, fmt.Errorf(
			"mdFileIndex==%d out of range 0..%d", mdFileIndex, len(files))
	}
	return files[mdFileIndex], nil
}

func (ws *Server) handleGetJs(wr http.ResponseWriter, req *http.Request) {
	slog.Info("handleGetJs ", "req", req.URL)
	var (
		err  error
		tmpl *textTmpl.Template
	)
	// Parsing the javascript as 'html' replaces "i < 2" with "i &lt; 2" and you
	// spend an hour tracking down why.  Parse as 'text' instead.  And no this isn't
	// solvable with template.Js, because we're _inflating_ a template full of Js,
	// not _injecting_ known Js into some template.
	tmpl, err = common.ParseAsTextTemplate(mdrip.AsTmplJs())
	if err != nil {
		write500(wr, fmt.Errorf("tmpl js parse fail; %w", err))
		return
	}
	wr.Header().Set("Content-Type", webapp.MimeJs)
	if minifyJs {
		if err = ws.minify(wr, webapp.MimeJs, tmpl, mdrip.TmplNameJs); err != nil {
			write500(wr, err)
			return
		}
		slog.Info("handleGetJs minified success")
		return
	}
	err = tmpl.ExecuteTemplate(wr, mdrip.TmplNameJs, mdrip.MakeBaseParams())
	if err != nil {
		write500(wr, fmt.Errorf("tmpl js inflate fail; %w", err))
		return
	}
	slog.Info("handleGetJs success")
}

func (ws *Server) handleGetCss(wr http.ResponseWriter, req *http.Request) {
	slog.Info("handleGetCss ", "req", req.URL)
	var (
		err  error
		tmpl *textTmpl.Template
	)
	tmpl, err = common.ParseAsTextTemplate(mdrip.AsTmplCss())
	if err != nil {
		write500(wr, fmt.Errorf("tmpl css parse fail; %w", err))
		return
	}
	wr.Header().Set("Content-Type", webapp.MimeCss)
	if minifyCss {
		if err = ws.minify(wr, webapp.MimeCss, tmpl, mdrip.TmplNameCss); err != nil {
			write500(wr, err)
			return
		}
		slog.Info("handleGetCss minified success")
		return
	}
	err = tmpl.ExecuteTemplate(wr, mdrip.TmplNameCss, mdrip.MakeBaseParams())
	if err != nil {
		write500(wr, fmt.Errorf("tmpl css inflate fail; %w", err))
		return
	}
	slog.Info("handleGetCss success")
}

func (ws *Server) minify(
	wr http.ResponseWriter, mimeType string, tmpl *textTmpl.Template, tmplName string) error {
	// There's probably some man-in-the-middle way to do this to skip using "buff" and "ugly".
	var (
		buff bytes.Buffer
		ugly []byte
	)
	err := tmpl.ExecuteTemplate(&buff, tmplName, mdrip.MakeBaseParams())
	if err != nil {
		return fmt.Errorf("tmpl %s inflate fail; %w", mimeType, err)
	}
	ugly, err = ws.minifier.Bytes(mimeType, buff.Bytes())
	if err != nil {
		return fmt.Errorf("%s minification fail; %w", mimeType, err)
	}
	if _, err = wr.Write(ugly); err != nil {
		return fmt.Errorf("write of %s failed; %w", mimeType, err)
	}
	return nil
}

// reload performs a data reload.
func (ws *Server) reload(wr http.ResponseWriter, req *http.Request) error {
	mySess, _ := ws.store.Get(req, cookieName)
	_ = mySess.Save(req, wr)
	ws.dLoader.makeLastLoadTimeVeryOld()
	return ws.dLoader.LoadAndRender()
}

// handleRenderWebApp sends a full "single-page" web app.
// The app does XHRs as you click around or use keys.
func (ws *Server) handleRenderWebApp(wr http.ResponseWriter, req *http.Request) {
	slog.Info("Rendering web app", "req", req.URL)
	var err error
	mySess, _ := ws.store.Get(req, cookieName)
	session.AssureDefaults(mySess)
	if err = mySess.Save(req, wr); err != nil {
		write500(wr, fmt.Errorf("session save fail; %w", err))
		return
	}
	if err = ws.dLoader.LoadAndRender(); err != nil {
		write500(wr, fmt.Errorf("data loader fail; %w", err))
		return
	}
	var tmpl *htmlTmpl.Template
	tmpl, err = common.ParseAsHtmlTemplate(webapp.AsTmpl())
	if err != nil {
		write500(wr, fmt.Errorf("template parsing fail; %w", err))
		return
	}
	ws.dLoader.appState.SetInitialFileIndex(req.URL.Path)
	err = tmpl.ExecuteTemplate(
		wr, webapp.TmplName,
		mdrip.MakeParams(ws.dLoader.navLeftRoot, ws.dLoader.appState))
	if err != nil {
		write500(wr, fmt.Errorf("template rendering failure; %w", err))
		return
	}
}

func (ws *Server) handleFavicon(w http.ResponseWriter, _ *http.Request) {
	Lissajous(w, 7, 3, 1)
}

func (ws *Server) handleImage(w http.ResponseWriter, r *http.Request) {
	mySess, _ := ws.store.Get(r, cookieName)
	_ = mySess.Save(r, w)
	Lissajous(w,
		getIntParam("s", r, 300),
		getIntParam("c", r, 30),
		getIntParam("n", r, 100))
}

// handleReload forces a data reload.
func (ws *Server) handleReload(wr http.ResponseWriter, req *http.Request) {
	slog.Info("Handling data reload", "url", req.URL)
	if err := ws.reload(wr, req); err != nil {
		write500(wr, fmt.Errorf("handleReload; %w", err))
		return
	}
}

// handleDebugPage forces a data reload and shows a debug page.
func (ws *Server) handleDebugPage(wr http.ResponseWriter, req *http.Request) {
	slog.Info("Rendering debug page", "url", req.URL)
	if err := ws.reload(wr, req); err != nil {
		write500(wr, fmt.Errorf("handleDebugPage; %w", err))
		return
	}
	ws.dLoader.folder.Accept(loader.NewVisitorDump(wr))
	loader.DumpBlocks(wr, ws.dLoader.FilteredBlocks())
}

func write500(w http.ResponseWriter, e error) {
	slog.Error(e.Error())
	http.Error(w, e.Error(), http.StatusInternalServerError)
}

func (ws *Server) handleQuit(w http.ResponseWriter, _ *http.Request) {
	slog.Info("Received quit.")
	close(ws.connReaperQuitCh)
	_, _ = fmt.Fprint(w, "\nbye bye\n")
	go func() {
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}()
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

func inRange(wr http.ResponseWriter, name string, arg, n int) bool {
	if arg >= 0 || arg < n {
		return true
	}
	http.Error(wr,
		fmt.Sprintf("%s %d out of range 0-%d",
			name, arg, n-1), http.StatusBadRequest)
	return false
}

func (ws *Server) handleRunCodeBlock(wr http.ResponseWriter, req *http.Request) {
	slog.Info(" ")
	slog.Info("Running code block", "url", req.URL)
	arg := req.URL.Query().Get(session.KeyMdSessID)
	if len(arg) == 0 {
		http.Error(wr, "No session id for block runner", http.StatusBadRequest)
		return
	}
	sessID := session.TypeSessID(arg)
	mdFileIndex := getIntParam(session.KeyMdFileIndex, req, -1)
	blockIndex := getIntParam(session.KeyBlockIndex, req, -1)
	slog.Info("args:",
		session.KeyMdSessID, sessID,
		session.KeyMdFileIndex, mdFileIndex,
		session.KeyBlockIndex, blockIndex,
	)

	if !inRange(
		wr, session.KeyMdFileIndex,
		mdFileIndex, len(ws.dLoader.RenderedFiles())) {
		return
	}
	mdFile := ws.dLoader.RenderedFiles()[mdFileIndex]

	if !inRange(wr, session.KeyBlockIndex, blockIndex, len(mdFile.Blocks)) {
		return
	}
	block := mdFile.Blocks[blockIndex]
	slog.Info("Will attempt to run", "codeSnip",
		utils.SampleString(block.Code(), 80))

	var err error

	// TODO: THE WEBSOCKET STUFF ABANDONED FOR NOW
	//   THE USE CASE IS QUESTIONABLE.
	//   FAILING OVER TO DIRECT TMUX WRITE

	c := ws.connections[sessID]
	if c == nil {
		slog.Info("no socket for", "sessID", sessID)
	} else {
		_, err = c.Write([]byte(block.Code()))
		if err != nil {
			slog.Warn("socket write failed", "err", err)
			delete(ws.connections, sessID)
		}
	}
	if c == nil || err != nil {
		slog.Info("no socket, attempting direct tmux paste")
		err = ws.attemptTmuxWrite(block)
		if err != nil {
			slog.Warn("tmux write failed", "err", err)
		}
	}
	_, _ = fmt.Fprintln(wr, "Ok")
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
