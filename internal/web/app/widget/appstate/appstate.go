package appstate

import (
	_ "embed"
	"html/template"
	"strconv"
	"strings"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/parsren"
)

var (
	//go:embed appstate.js
	Js string
)

const BadId = -1

type Facts struct {
	InitialFileIndex      int
	InitialCodeBlockIndex int
	NumFolders            int
	MaxCodeBlocksInAFile  int
	MaxNavWordLength      int
	IsNavVisible          bool
	IsTitleVisible        bool
}

type InitialRender struct {
	Title        string
	DataSource   string
	OrderedPaths []loader.FilePath
	Facts        Facts
}

// HtmlAndLabels has the rendered HTML from a markdown file, plus
// a list of code labels, one label (the main label) for each
// code block in the file.
type HtmlAndLabels struct {
	Html            template.HTML
	CodeBlockLabels loader.LabelList
}

type AppState struct {
	InitialRender
	RenderedFiles []HtmlAndLabels
}

func (as *AppState) InitialLabels() loader.LabelList {
	labels := make(loader.LabelList, as.Facts.MaxCodeBlocksInAFile)
	for j := range labels {
		labels[j] = loader.Label("label" + strconv.Itoa(j))
	}
	return labels
}

func (as *AppState) SetInitialFileIndex(p string) {
	as.Facts.InitialFileIndex = 0
	if strings.HasPrefix(p, "/") {
		p = p[1:]
	}
	if p == "" {
		return
	}
	for i := range as.OrderedPaths {
		if p == string(as.OrderedPaths[i]) {
			as.Facts.InitialFileIndex = i
		}
	}
}

func New(
	dSource string, files []*parsren.RenderedMdFile, title string) *AppState {
	var as AppState
	as.DataSource = dSource
	as.Title = title
	as.RenderedFiles = make([]HtmlAndLabels, len(files))
	as.OrderedPaths = make([]loader.FilePath, len(files))
	maxCodeBlocksInOneFile := 0
	for i, f := range files {
		as.OrderedPaths[i] = f.Path
		if len(f.Blocks) > maxCodeBlocksInOneFile {
			maxCodeBlocksInOneFile = len(f.Blocks)
		}
		as.RenderedFiles[i] = HtmlAndLabels{
			Html:            f.Html,
			CodeBlockLabels: loader.NewLabelList(f.Blocks),
		}
	}
	as.Facts.IsNavVisible = false
	as.Facts.IsTitleVisible = true
	as.Facts.MaxCodeBlocksInAFile = maxCodeBlocksInOneFile
	as.Facts.InitialFileIndex = BadId
	as.Facts.InitialCodeBlockIndex = BadId
	return &as
}
