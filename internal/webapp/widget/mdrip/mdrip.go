package mdrip

import (
	"bytes"
	_ "embed"
	"html/template"
	"strings"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/parsren"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/appstate"
	burgerbars "github.com/monopole/mdrip/v2/internal/webapp/widget/burgerbars1"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/codeblock"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/codelabel"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/helpbox"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/helpbutton"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/mdfiles"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/monkey"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/navbottom"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/navcontentrow"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/navleftfile"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/navleftfolder"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/navleftroot"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/navrightroot"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/navtop"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/session"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/timeline"
)

const (
	tmplNameBase = "tmplMdRip"
	TmplNameHtml = tmplNameBase + "Html"
	TmplNameJs   = tmplNameBase + "Js"
	TmplNameCss  = tmplNameBase + "Css"

	// These values just have to differ from each other.
	timelineIdTop    = "333"
	timelineIdBottom = "444"
)

var (
	//go:embed mdrip.css
	Css string

	//go:embed mdrip.js
	Js string

	//go:embed mdrip.html
	html string

	AllCss = strings.Join(
		[]string{
			common.Css,
			burgerbars.Css,
			helpbutton.Css,
			helpbox.Css,
			mdfiles.Css,
			codeblock.Css,
			codelabel.Css,
			timeline.Css,
			navtop.Css,
			navleftfile.Css,
			navleftfolder.Css,
			navleftroot.Css,
			navbottom.Css,
			navcontentrow.Css,
			navbottom.Css,
			navrightroot.Css,
			Css,
		}, "\n")

	AllJs = strings.Join(
		[]string{
			common.Js,
			appstate.Js,
			session.Js,
			codelabel.Js,
			codeblock.Js,
			burgerbars.Js,
			helpbutton.Js,
			helpbox.Js,
			timeline.Js,
			mdfiles.Js,
			navtop.Js,
			navleftfile.Js,
			navleftfolder.Js,
			navleftroot.Js,
			navcontentrow.Js,
			navrightroot.Js,
			navbottom.Js,
			monkey.Js,
			Js,
		}, "\n")
)

func AsTmplHtml() string {
	return common.AsTmpl(TmplNameHtml, html)
}

func AsTmplJs() string {
	return common.AsTmpl(TmplNameJs, AllJs)
}

func AsTmplCss() string {
	return common.AsTmpl(TmplNameCss, AllCss)
}

type TmplParams struct {
	common.ParamStructSession
	common.ParamStructTransition
	TimelineIdTop string
	TimelineIdBot string
	navcontentrow.ParamStructContentRow
	NavLeftRoot   template.HTML
	AppState      *appstate.AppState
	BurgerBars    template.HTML
	HelpButton    template.HTML
	TimelineId    string
	TimelineRow   template.HTML
	NavContentRow template.HTML
	HelpBox       template.HTML
}

type RenderingArgs struct {
	Pr         parsren.MdParserRenderer
	DataSource string
	Folder     *loader.MyFolder
	Title      string
}

// RenderFolder partially renders a folder, and computes an appState
// which feeds into remaining rendering stages.
func RenderFolder(rArgs *RenderingArgs) (
	navLeftRoot template.HTML, appState *appstate.AppState) {
	numFolders := 0
	{
		var b bytes.Buffer
		v := navleftroot.NewRenderer(&b)
		loader.NewTopFolder(rArgs.Folder).Accept(v)
		numFolders = v.NumFolders()
		navLeftRoot = template.HTML(b.String())
	}
	{
		loader.NewTopFolder(rArgs.Folder).Accept(rArgs.Pr)
		appState = appstate.New(
			rArgs.DataSource, rArgs.Pr.RenderedMdFiles(), rArgs.Title)
		appState.Facts.NumFolders = numFolders
	}
	return
}

func MakeBaseParams() *TmplParams {
	return &TmplParams{
		ParamStructSession:    common.ParamDefaultSession,
		ParamStructTransition: common.ParamDefaultTransition,
		TimelineIdTop:         timelineIdTop,
		TimelineIdBot:         timelineIdBottom,
	}
}

func MakeParams(
	lftNavRoot template.HTML, appState *appstate.AppState) *TmplParams {
	tps := MakeBaseParams()
	tps.NavLeftRoot = lftNavRoot
	tps.AppState = appState
	tps.ContentLeft = common.MustRenderHtml(
		navleftroot.AsTmpl(), tps, navleftroot.TmplName)

	tps.ContentRight = common.MustRenderHtml(
		codelabel.AsTmpl()+navrightroot.AsTmpl(), tps, navrightroot.TmplName)
	tps.BurgerBars = common.MustRenderHtml(
		burgerbars.AsTmpl(), tps, burgerbars.TmplName)

	tps.HelpButton = common.MustRenderHtml(
		helpbutton.AsTmpl(), tps, helpbutton.TmplName)

	tps.TimelineId = timelineIdTop
	tps.TimelineRow = common.MustRenderHtml(
		timeline.AsTmpl(), tps, timeline.TmplName)
	tps.ContentTop = common.MustRenderHtml(
		navtop.AsTmpl(), tps, navtop.TmplName)

	tps.TimelineId = timelineIdBottom
	tps.TimelineRow = common.MustRenderHtml(
		timeline.AsTmpl(), tps, timeline.TmplName)
	tps.ContentBottom = common.MustRenderHtml(
		navbottom.AsTmpl(), tps, navbottom.TmplName)

	tps.ContentCenter = common.MustRenderHtml(
		mdfiles.AsTmpl(), tps, mdfiles.TmplName)

	tps.NavContentRow = common.MustRenderHtml(
		navcontentrow.AsTmpl(), tps, navcontentrow.TmplName)

	tps.HelpBox = common.MustRenderHtml(
		helpbox.AsTmpl(), tps, helpbox.TmplName)
	return tps
}
