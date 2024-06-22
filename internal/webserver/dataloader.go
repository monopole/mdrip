package webserver

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"strings"
	"time"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/parsren"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/appstate"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/mdrip"
)

// DataLoader is an embarrassment.
// It's a computation cache around FsLoader.
type DataLoader struct {
	ldr         *loader.FsLoader
	pRen        parsren.MdParserRenderer
	paths       []string
	title       string
	folder      *loader.MyFolder
	loadTime    time.Time
	navLeftRoot template.HTML
	appState    *appstate.AppState
}

const maxAge = 5 * time.Minute

func NewDataLoader(
	ldr *loader.FsLoader, paths []string,
	pRen parsren.MdParserRenderer, title string) *DataLoader {
	return &DataLoader{
		ldr:      ldr,
		paths:    paths,
		pRen:     pRen,
		title:    title,
		folder:   nil,
		loadTime: time.Time{},
	}
}

func (dl *DataLoader) RenderedFiles() []*parsren.RenderedMdFile {
	return dl.pRen.RenderedMdFiles()
}

func (dl *DataLoader) FilteredBlocks() []*loader.CodeBlock {
	return dl.pRen.FilteredBlocks(loader.WildCardLabel)
}

func (dl *DataLoader) LoadAndRender() (err error) {
	if len(dl.paths) == 0 {
		return fmt.Errorf("specify some paths to load")
	}
	if time.Since(dl.loadTime) < maxAge {
		slog.Info(
			"Data not old enough to reload",
			"age", time.Since(dl.loadTime))
		return
	}
	dl.pRen.Reset()
	slog.Info("Loading", "paths", dl.paths)
	dl.folder, err = dl.ldr.LoadTrees(dl.paths)
	if err != nil {
		return
	}
	if dl.folder == nil {
		return fmt.Errorf("no markdown found at %s", dl.paths)
	}
	dl.loadTime = time.Now()
	{
		vc := loader.NewVisitorCounter()
		dl.folder.Accept(vc)
		slog.Info("Loaded",
			"top", dl.folder.Path(),
			"numFolders", vc.NumFolders,
			"numFiles", vc.NumFiles)
	}
	dl.navLeftRoot, dl.appState = mdrip.RenderFolder(
		&mdrip.RenderingArgs{
			Pr:         dl.pRen,
			DataSource: dl.getDataSource(),
			Folder:     dl.folder,
			Title:      dl.title,
		},
	)
	return
}

func (dl *DataLoader) makeLastLoadTimeVeryOld() {
	dl.loadTime = time.UnixMicro(0)
}

func (dl *DataLoader) getDataSource() string {
	if len(dl.paths) == 0 {
		return "hardcoded test data"
	}
	return strings.Join(dl.paths, " ")
}

func (dl *DataLoader) makeErrorContent(err error) []byte {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf(`
# Trouble loading %s

%s
`, dl.getDataSource(), err.Error()))
	return b.Bytes()
}
