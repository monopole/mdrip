package server

import (
	"encoding/json"
	"fmt"
	htmlTmpl "html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/web/app"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/common"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/mdrip"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/session"
	"github.com/monopole/mdrip/v2/internal/web/config"
	"github.com/monopole/mdrip/v2/internal/web/server/minify"
)

// handleRenderWebApp sends a full "single-page" web app.
// The app does XHRs as you click around or use keys.
func (ws *Server) handleRenderWebApp(wr http.ResponseWriter, req *http.Request) {
	slog.Debug("Rendering web app", "req", req.URL)
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
	tmpl, err = common.ParseAsHtmlTemplate(app.AsTmpl())
	if err != nil {
		write500(wr, fmt.Errorf("template parsing fail; %w", err))
		return
	}
	ws.dLoader.appState.SetInitialFileIndex(req.URL.Path)
	err = tmpl.ExecuteTemplate(
		wr, app.TmplName,
		mdrip.MakeParams(ws.dLoader.navLeftRoot, ws.dLoader.appState))
	if err != nil {
		write500(wr, fmt.Errorf("template rendering failure; %w", err))
		return
	}
}

func (ws *Server) handleSaveSession(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Saving session", "req", r.URL)
	s, err := ws.store.Get(r, cookieName)
	if err != nil {
		write500(w, err)
		return
	}
	s.Values[config.KeyIsNavOn] = getBoolParam(config.KeyIsNavOn, r, false)
	s.Values[config.KeyIsTitleOn] = getBoolParam(config.KeyIsTitleOn, r, false)
	s.Values[config.KeyMdFileIndex] = getIntParam(config.KeyMdFileIndex, r, 0)
	s.Values[config.KeyBlockIndex] = getIntParam(config.KeyBlockIndex, r, 0)
	if err = s.Save(r, w); err != nil {
		slog.Error("unable to save session", "err", err)
	}
	_, _ = fmt.Fprintln(w, "Ok")
	slog.Debug("Saved session.")
}

func (ws *Server) handleGetHtmlForFile(wr http.ResponseWriter, req *http.Request) {
	slog.Debug("handleGetHtmlForFile ", "req", req.URL)
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
	slog.Debug("handleGetHtmlForFile success")
}

func (ws *Server) handleGetLabelsForFile(wr http.ResponseWriter, req *http.Request) {
	slog.Debug("handleGetLabelsForFile ", "req", req.URL)
	f, err := ws.getRenderedMdFile(req)
	if err != nil {
		write500(wr, fmt.Errorf("handleGetLabelsForFile render; %w", err))
		return
	}
	var jsn []byte
	jsn, err = json.Marshal(loader.NewBlockNameList(f.Blocks))
	if err != nil {
		write500(wr, fmt.Errorf("handleGetLabelsForFile marshal; %w", err))
		return
	}
	if _, err = wr.Write(jsn); err != nil {
		write500(wr, fmt.Errorf("handleGetLabelsForFile write; %w", err))
		return
	}
	slog.Debug("handleGetLabelsForFile success")
}

func (ws *Server) handleGetJs(wr http.ResponseWriter, req *http.Request) {
	slog.Debug("handleGetJs", "req", req.URL)
	ws.minifier.Write(wr, &minify.Args{
		MimeType: app.MimeJs,
		Tmpl: minify.TmplArgs{
			Name: mdrip.TmplNameJs,
			Body: mdrip.AsTmplJs(),
			Params: mdrip.MakeBaseParams(
				ws.dLoader.appState.Facts.MaxNavWordLength),
		},
	})
}

func (ws *Server) handleGetCss(wr http.ResponseWriter, req *http.Request) {
	slog.Debug("handleGetCss", "req", req.URL)
	ws.minifier.Write(wr, &minify.Args{
		MimeType: app.MimeCss,
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

func (ws *Server) handleLissajous(w http.ResponseWriter, r *http.Request) {
	mySess, _ := ws.store.Get(r, cookieName)
	_ = mySess.Save(r, w)
	Lissajous(w,
		getIntParam("s", r, 300),
		getIntParam("c", r, 30),
		getIntParam("n", r, 100))
}

// handleReload forces a data reload.
func (ws *Server) handleReload(wr http.ResponseWriter, req *http.Request) {
	slog.Debug("Handling data reload", "url", req.URL)
	if err := ws.reload(wr, req); err != nil {
		write500(wr, fmt.Errorf("handleReload; %w", err))
		return
	}
}

// handleDebugPage forces a data reload and shows a debug page.
func (ws *Server) handleDebugPage(wr http.ResponseWriter, req *http.Request) {
	slog.Debug("Rendering debug page", "url", req.URL)
	if err := ws.reload(wr, req); err != nil {
		write500(wr, fmt.Errorf("handleDebugPage; %w", err))
		return
	}
	ws.dLoader.folder.Accept(loader.NewVisitorDump(wr))
	loader.PrintBlocks(wr, ws.dLoader.AllBlocks())
}

func (ws *Server) handleQuit(w http.ResponseWriter, _ *http.Request) {
	slog.Debug("Received quit.")
	_, _ = fmt.Fprint(w, "\nbye bye\n")
	go func() {
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}()
}

func (ws *Server) handleRunCodeBlock(wr http.ResponseWriter, req *http.Request) {
	slog.Debug(" ")
	slog.Debug("Running code block", "url", req.URL)
	arg := req.URL.Query().Get(config.KeyMdSessID)
	if len(arg) == 0 {
		http.Error(wr, "No session id for block codeWriter", http.StatusBadRequest)
		return
	}
	sessID := session.TypeSessID(arg)
	mdFileIndex := getIntParam(config.KeyMdFileIndex, req, -1)
	blockIndex := getIntParam(config.KeyBlockIndex, req, -1)
	slog.Debug("args:",
		config.KeyMdSessID, sessID,
		config.KeyMdFileIndex, mdFileIndex,
		config.KeyBlockIndex, blockIndex,
	)

	if !inRange(
		wr, config.KeyMdFileIndex,
		mdFileIndex, len(ws.dLoader.RenderedFiles())) {
		return
	}
	mdFile := ws.dLoader.RenderedFiles()[mdFileIndex]

	if !inRange(wr, config.KeyBlockIndex, blockIndex, len(mdFile.Blocks)) {
		return
	}
	block := mdFile.Blocks[blockIndex]

	if _, err := ws.codeWriter.Write([]byte(block.Code())); err != nil {
		slog.Error("codeWriter failed", "err", err)
	}
	_, _ = fmt.Fprintln(wr, "Ok")
}
