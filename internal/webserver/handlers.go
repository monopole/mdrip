package webserver

import (
	"encoding/json"
	"fmt"
	"github.com/monopole/mdrip/v2/internal/utils"
	htmlTmpl "html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/webapp"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/mdrip"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/session"
	"github.com/monopole/mdrip/v2/internal/webserver/minify"
)

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

func (ws *Server) handleGetJs(wr http.ResponseWriter, req *http.Request) {
	slog.Info("handleGetJs", "req", req.URL)
	ws.minifier.Write(wr, &minify.Args{
		MimeType: webapp.MimeJs,
		Tmpl: minify.TmplArgs{
			Name: mdrip.TmplNameJs,
			Body: mdrip.AsTmplJs(),
			Params: mdrip.MakeBaseParams(
				ws.dLoader.appState.Facts.MaxNavWordLength),
		},
	})
}

func (ws *Server) handleGetCss(wr http.ResponseWriter, req *http.Request) {
	slog.Info("handleGetCss", "req", req.URL)
	ws.minifier.Write(wr, &minify.Args{
		MimeType: webapp.MimeCss,
		Tmpl: minify.TmplArgs{
			Name: mdrip.TmplNameCss,
			Body: mdrip.AsTmplCss(),
			Params: mdrip.MakeBaseParams(
				ws.dLoader.appState.Facts.MaxNavWordLength),
		},
	})
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

func (ws *Server) handleQuit(w http.ResponseWriter, _ *http.Request) {
	slog.Info("Received quit.")
	close(ws.connReaperQuitCh)
	_, _ = fmt.Fprint(w, "\nbye bye\n")
	go func() {
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}()
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
