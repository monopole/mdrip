package webapp

import (
	"html/template"
	"io"

	"bytes"
	"strings"

	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
	"github.com/monopole/mdrip/program"
)

type TypeSessId string

// WebApp presents a tutorial to a web browser.
// Not a complex app, so eschewing react, angular2, polymer, etc.
type WebApp struct {
	sessId      TypeSessId
	host        string
	tut         model.Tutorial
	ds          *base.DataSource
	tmpl        *template.Template
	lessonPath  []int
	coursePaths [][]int
}

func NewWebApp(
	sessId TypeSessId, host string,
	tut model.Tutorial, ds *base.DataSource, lp []int, cp [][]int) *WebApp {
	return &WebApp{sessId, host, tut, ds, makeParsedTemplate(tut), lp, cp}
}

func (wa *WebApp) SessId() TypeSessId { return wa.sessId }

func (wa *WebApp) Host() string { return wa.host }

func (wa *WebApp) Lessons() []*program.LessonPgm {
	v := program.NewLessonPgmExtractor(base.WildCardLabel)
	wa.tut.Accept(v)
	return v.Lessons()
}

func (wa *WebApp) AppName() string {
	return wa.ds.Display()
}

func (wa *WebApp) AppLink() template.URL {
	return template.URL(wa.ds.Href())
}

func (wa *WebApp) TrimName() string {
	result := strings.TrimSpace(wa.AppName())
	if len(result) > maxAppNameLen {
		return result[maxAppNameLen-3:] + "..."
	}
	return result
}

const (
	// arbitrary
	maxAppNameLen = len("gh:kubernetes/website/reference") + 10
)

// Return the last element or zero.
func (wa *WebApp) InitialLesson() int {
	if len(wa.lessonPath) == 0 {
		return 0
	}
	return wa.lessonPath[len(wa.lessonPath)-1]
}

// Return everything BUT the last element
// i.e. omit the file, just return the directory path.
func (wa *WebApp) xCoursePath() []int {
	if len(wa.lessonPath) == 0 {
		return []int{}
	}
	return wa.lessonPath[:len(wa.lessonPath)-1]
}

// Emit a javascript 2D array
// with length equal to the number of lessons.
// each entry should contain an array of course indices
// that should be active when the lesson is actice.
func (wa *WebApp) CoursePaths() [][]int {
	return wa.coursePaths
}

const (
	blue700         = "#1976D2" // header
	blue500         = "#2196F3" // hav
	blue200         = "#90CAF9" // help
	deepOrange500   = "#FF5722" // controls
	deepOrange200   = "#FF8A65"
	deepOrange700   = "#E64A19"
	teal            = "#00838f"
	hackerNewsBeige = "#f6f6ef"
	k8sBlue         = "#326DE6"
	grayIsh         = "#c9c9c9"
	whiteIsh        = "#e3e3e3"
	seaBlue         = "#9ad3de"
	darkerBlue      = "#89bdd3"
	hoverBlue       = "#06e"
	terminalGreen   = "#33ff66"
	greenA200       = "#B2FF59"
	greenA400       = "#76FF03"
	greenA700       = "#64DD17"
)

func (wa *WebApp) TransitionSpeed() string { return "0.25s" }
func (wa *WebApp) LayBodyWideWidth() int   { return 1200 }
func (wa *WebApp) LayBodyMediumWidth() int { return 800 }
func (wa *WebApp) LayMinHeaderWidth() int  { return 400 }
func (wa *WebApp) LayNavBoxWidth() int     { return 210 }
func (wa *WebApp) LayMinHeaderHeight() int { return 50 }
func (wa *WebApp) LayHeaderHeight() int    { return 120 }
func (wa *WebApp) LayFooterHeight() int    { return 70 }
func (wa *WebApp) LayNavTopBotPad() int    { return 7 }
func (wa *WebApp) LayNavLeftPad() int      { return 20 }

func (wa *WebApp) ColorBackground() string          { return "white" }
func (wa *WebApp) ColorHelpBackground() string      { return whiteIsh }
func (wa *WebApp) ColorHeader() string              { return blue700 }
func (wa *WebApp) ColorCodeBlockText() string       { return greenA400 }
func (wa *WebApp) ColorCodeBlockBackground() string { return "black" }
func (wa *WebApp) ColorNavBackground() string       { return blue200 }
func (wa *WebApp) ColorNavText() string             { return "black" }
func (wa *WebApp) ColorNavSelected() string         { return wa.ColorBackground() }
func (wa *WebApp) ColorHover() string               { return deepOrange500 }
func (wa *WebApp) ColorCodeHover() string           { return deepOrange700 }
func (wa *WebApp) ColorControls() string            { return greenA200 }
func (wa *WebApp) ColorTitle() string               { return wa.ColorControls() }

//func (wa *WebApp) ColorHeader() string

func (wa *WebApp) LessonCount() int {
	c := model.NewTutorialLessonCounter()
	wa.tut.Accept(c)
	return c.Count()
}

func (wa *WebApp) Render(w io.Writer) error {
	return wa.tmpl.ExecuteTemplate(w, tmplNameWebApp, wa)
}

func makeParsedTemplate(tut model.Tutorial) *template.Template {
	return template.Must(
		template.New("main").Parse(
			tmplBodyLesson +
				tmplBodyBlockPgm +
				tmplBodyLessonList +
				tmplBodyLessonHead +
				makeAppTemplate(makeLeftNavBody(tut))))
}

// The logic involved in building the leftnav is much less awkward
// in plain Go than in the Go template language, so creating it
// this way rather than writing it out with a bunch of {{if}}s, etc.
func makeLeftNavBody(tut model.Tutorial) string {
	var b bytes.Buffer
	v := NewTutorialNavPrinter(&b)
	tut.Accept(v)
	return b.String()
}

func makeAppTemplate(htmlNavActual string) string {
	return `
{{define "` + tmplNameWebApp + `"}}
<html>
<head>
<style type="text/css">` + cssInHeader + `
</style>
<script type="text/javascript">` + jsInHeader + `
</script>
</head>
<body id='body' onload='onLoad()'>
  <header id='header'>
    <div class='navButtonBox' onclick='nav.toggle()'>
      <div class='navBurger'>
        <div class='burgBar1'></div>
        <div class='burgBar2'></div>
        <div class='burgBar3'></div>
      </div>
    </div>
    <div class='headerColumn'>
      <a target='_blank' href='{{.AppLink}}'> <title id='title'> {{.TrimName}} </title></a>
      <div class='activeLessonName'> Droplet Formation Rates </div>
      ` + htmlLessonNavRow + `
    </div>
    <div class='navButtonBox'> &nbsp; </div>
  </header>

  <div class='navLeftBox navLeftBoxShadow'>
    <nav class='navActual'>
      ` + htmlNavActual + `
    </nav>
  </div>
  <div class='helpBox'>
    <div class='helpActual'>
    ` + htmlHelp + `
    </div>
  </div>

  <div class='navRightBox navRightBoxShadow'>
    <nav class='navActual'>
      <!-- <p> T O C </p> <p> COMING </p> <p> HERE </p> -->
    </nav>
  </div>

  <div class='scrollingColumn'>
    <div class='headerSpacer'> HEADER SPACER </div>
    <div class='proseRow'>
      <div class='navLeftSpacer'> &nbsp; </div>
      <div class='proseColumn'>
					{{ template "` + tmplNameLessonList + `" .Lessons }}
      </div>
      <div class='navRightSpacer'> &nbsp; </div>
    </div>
    <footer>
    ` + htmlLessonNavRow + `
    </footer>
  </div>
</body>
</html>
{{end}}
`
}

const (
	tmplNameWebApp     = "webApp"
	tmplNameLessonList = "lessonList"
	tmplBodyLessonList = `
{{define "` + tmplNameLessonList + `"}}
{{range $i, $c := .}}
  <div class='oneLesson' id='BL{{$i}}' data-id='{{$i}}' >
  {{ template "` + tmplNameLesson + `" $c }}
  </div>
{{end}}
{{end}}
`
	tmplNameLesson = "lessonlist"
	tmplBodyLesson = `
{{define "` + tmplNameLesson + `"}}
{{range $i, $c := .Blocks}}
  <div class="commandBlockBody" data-id="{{$i}}">
  {{ template "` + tmplNameBlockPgm + `" $c }}
  </div>
{{end}}
{{end}}
`
	tmplNameBlockPgm = "blockPgm"
	tmplBodyBlockPgm = `
{{define "` + tmplNameBlockPgm + `"}}
<div class="proseblock"> {{.HtmlProse}} </div>
{{if .Code}}
<div class="codeBox">
  <div class="codeBlockControl">
    <span class="codeBlockButton" onclick="codeBlock.run(event)">
      {{.Name}}
    </span>
    <span class="codeBlockSpacer"> &nbsp; </span>
  </div>
<div class="codeblockBody">
{{ .Code }}
</div>
</div>
{{end}}
{{end}}
`
	tmplNameLessonHead = "lessonhead"
	tmplBodyLessonHead = `
{{define "` + tmplNameLessonHead + `"}}
<p><code> {{.Path}} </code></p>
{{end}}
`
)

const htmlLessonNavRow = `
<div class='lessonNavRow'>
  <div class='lessonPrevClickerRow' onclick='lessonMgr.goPrev()'>
    <div class='lessonPrevTitle'> quantum flux </div>
    <div class='lessonPrevPointer'> &lt; </div>
  </div>
  <div class='helpButtonBox' onclick='help.toggle()'> ? </div>
  <div class='lessonNextClickerRow' onclick='lessonMgr.goNext()'>
    <div class='lessonNextPointer'> &gt; </div>
    <div class='lessonNextTitle'> magnetic flux  </div>
  </div>
</div>
`

const htmlHelp = `
<p>
You are viewing a snapshot of markdown content from</p>
<blockquote>
<a target='_blank' href='{{.AppLink}}'> <code> {{.AppName}} </code></a>
</blockquote>

<ul>
<li>Arrow keys navigate,
    <code>'m'</code> toggles menu,
    <code>'h'</code> toggles help.</li>
<li>Click on command block headers to copy blocks to your clipboard.</li>
<li>Check marks track code block execution progress.</li>
<li>Use <code>tmux</code> to get one-click execution.</li>
</ul>

<h3> Serve locally with tmux for one-click code block execution</h3>

<p>
To avoid the need to mouse/aim/paste, serve the content locally:
<pre>
  GOBIN=$TMPDIR go install github.com/monopole/mdrip
  $TMPDIR/mdrip --port 8001 --mode demo {{.AppName}}
</pre>
and run <a target="_blank"
href="https://github.com/tmux/tmux/wiki">tmux</a>:
<pre>
  tmux
</pre>
Then clicking on a code block header in your browser pastes
the command block to the active tmux session.
This is a handy way to drive demos from markdown.
</p>

<h3> Remote server tmux </h3>
<p> <em>A proof of concept
for using tmux over a websocket to remote servers.
Needs better session mgmt to work with load balanced traffic.
The websocket described below not needed in the previous
scenario using a local server. </em></p>
<p>
For one-click usage from a remote server:
<ul>
<li>
Install <code><a target="_blank"
href="https://github.com/monopole/mdrip">mdrip</a></code>
as described above.
</li>
<br>
<li>Run <a target="_blank"
href="https://github.com/tmux/tmux/wiki">tmux</a>:
<pre>
  tmux
</pre>
</li>
<li>In some non-tmux shell, run mdrip in <code>--mode tmux</code>:
<pre>
  host=ws://{{.Host}}
  mdrip \
      --alsologtostderr --v 0 \
      --stderrthreshold INFO \
      --mode tmux \
      ${host}/_/ws?id={{.SessId}}
</pre>
</li>
</ul>
<p>
Now, clicking a command block header sends the block
from this page's server over a websocket to the local
<code>mdrip</code>, which then 'pastes' the block
to your active <code>tmux</code> pane.<p>
<p>
The <code>mdrip</code> service self-exits after a period of inactivity,
and can be restarted with the same command.</p>
`

const cssInHeader = `
body {
  padding: 0;
  margin: 0;
  background-color: darkgray;
  /* font-family: "Roboto", sans-serif; */
  /* font-family: "Veranda", Veranda, sans-serif; */
  font-family: Verdana, Geneva, sans-serif;
  position: relative;
  font-size: 12pt;
  line-height: 1.4;
  -webkit-font-smoothing: antialiased;
  width: 100%;
}

header, .headerSpacer {
  height: {{.LayHeaderHeight}}px;
  width: inherit;
}

header {
  position: fixed;
  top: 0;
  background: {{.ColorHeader}};
  /* background: linear-gradient(0deg, {{.ColorBackground}}, {{.ColorHeader}}); */
  display: flex;
  justify-content: space-between;
  flex-direction: row;
  flex-wrap: nowrap;
  align-items: center;
  transition: height {{.TransitionSpeed}};
  box-shadow: 0 2px 2px 2px rgba(0,0,0,.4);
}

.navLeftBox, .navRightBox {
  position: fixed;
  top: calc({{.LayHeaderHeight}}px + 2px);  /* leave room for header drop-shadow */
  height: calc(100vh - ({{.LayFooterHeight}}px + {{.LayHeaderHeight}}px + 4px));
  background-color: {{.ColorNavBackground}};
  color: {{.ColorNavText}};
  display: inline-block;
  overflow: hidden;  /* initially hideNav */
  width: 0px;  /* initially hideNav */
  min-width: 0px;  /* initially hideNav */
  transition: width {{.TransitionSpeed}}, min-width {{.TransitionSpeed}};
}
.navRightBoxShadow {
  /* shadow on bottom, top and left */
  box-shadow: 0 2px 2px 2px rgba(0,0,0,.2), -2px 0px 2px 2px rgba(0,0,0,.2);
}
.navLeftBoxShadow {
  /* shadow on bottom, top and right */
  box-shadow: 0 2px 2px 2px rgba(0,0,0,.2), 2px 0px 2px 2px rgba(0,0,0,.2);
}
.navActual {
  padding-left: 1em;
}

a {
  text-decoration: none;
}
a:hover, a:visited, a:link, a:active {
  text-decoration: none;
}

.navCourseTitle {
  padding: 0px;
}

.navCourseTitle:hover {
  color: {{.ColorHover}};
  font-weight: bold;
}

.navItemTop {
  /* top rig bot lef */
  padding: {{.LayNavTopBotPad}}px 0px {{.LayNavTopBotPad}}px 4px;
}

.navItemBox {
  /* top rig bot lef */
  padding: {{.LayNavTopBotPad}}px 0px {{.LayNavTopBotPad}}px {{.LayNavLeftPad}}px;
}

.navCourseContent {
  /* top rig bot lef */
  padding: {{.LayNavTopBotPad}}px 0px 0px 0px;
}

.navLessonTitleOn {
  background-color: {{.ColorNavSelected}};
}

.navLessonTitleOff {
}

.navLessonTitleOff:hover {
  color: {{.ColorHover}};
  font-weight: bold;
}

.scrollingColumn {
  width: inherit;
}

.navLeftSpacer, .navRightSpacer {
   width: {{.LayNavBoxWidth}}px;
   min-width: {{.LayNavBoxWidth}}px;
   display: none;  /* initially hideNav */
}

.proseRow {
  background-color: {{.ColorBackground}};
  width: inherit;
  display: flex;
  flex-direction: row;
}

footer {
  background: {{.ColorHeader}};
  /* background: linear-gradient(0deg, {{.ColorHeader}}, {{.ColorBackground}}); */
  height: {{.LayFooterHeight}}px;
}

.helpButtonBox, .navButtonBox {
  font-size: larger;
  display: flex;
  flex-direction: row;
  justify-content: center;
  align-items: center;
  min-height: 2.8em;
  min-width: 2em;
  cursor: pointer;
  color: {{.ColorControls}};
  font-weight: bold;
}
.helpButtonBox:hover {
  color: {{.ColorHover}};
}
.navButtonBox {
  min-width: 6em;
}

.headerColumn {
  width: 80%;
  min-width: {{.LayMinHeaderWidth}}px;
  height: inherit;
  display: flex;
  justify-content: center;
  flex-direction: column;
  flex-wrap: nowrap;
  align-items: center;
}


title {
  font-size: 2em;
  font-weight: bold;
  color: {{.ColorTitle}};
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}
title:hover {
  color: {{.ColorHover}};
  font-weight: bold;
}

.activeLessonName {
  font-weight: bold;
}

.lessonNavRow {
  height: inherit;
  display: flex;
  width: 100%;
  justify-content: center;
  flex-direction: row;
  flex-wrap: nowrap;
  align-items: center;
}

.helpBox {
  position: fixed;
  top: {{.LayHeaderHeight}}px;
  left: {{.LayNavBoxWidth}}px;
  right: {{.LayNavBoxWidth}}px;
  height: 0px;  /* initially hideHelp */
	z-index: 3;

	/* border: solid 1px #555; */
	/* border-radius: 4px; */
           /*   x   y blur spread color             x   y blur spread color */
  /* box-shadow: 0px 2px  2px    1px rgba(0,0,0,.3), 2px 0px 2px 1px rgba(0,0,0,.3); */

  background-color: {{.ColorHelpBackground}};
  color: {{.ColorNavText}};
  transition: height {{.TransitionSpeed}};
  overflow: auto;
}

.helpActual {
  padding: 1em;
}

.proseColumn {
  display: flex;
  justify-content: flex-start;
  flex-direction: column;
  overflow-x: hidden;
  overflow-y: auto;
  align-items: flex-start;
  transition: width {{.TransitionSpeed}};
}

.proseActual {
  /* top right bottom left */
  padding: 0 1em 0 1em;
}

.lessonPrevClickerRow, .lessonNextClickerRow {
  height: 100%;
  color: {{.ColorControls}};
  cursor: pointer;
  display: flex;
  flex-basis: 45%;
  flex-direction: row;
  flex-wrap: nowrap;
  align-items: center;
}
.lessonPrevClickerRow:hover, .lessonNextClickerRow:hover {
  color: {{.ColorHover}};
  font-weight: bold;
}
.lessonPrevClickerRow {
  justify-content: flex-end;
}
.lessonNextClickerRow {
}

.lessonPrevPointer, .lessonNextPointer {
  /* top right bottom left */
  padding: 0 1em 0 1em;
  font-weight: bold;
  font-size: larger;
}

.lessonPrevTitle, .lessonNextTitle {
  font-style: oblique;
  width: 100%;
  cursor: pointer;
  display: inline-block;
}
.lessonPrevTitle {
  text-align: right;
}

.commandBlockBody {
  margin: 0px;
  border: 0px;
  padding: 0px;
}

.codeBlockCheckOff {
  display: inline-block;
  width: 24px;
  height: 15px;
  background-repeat: no-repeat;
  background-size: contain;
  background-image: url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABgAAAAWCAMAAADto6y6AAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAAgY0hSTQAAeiYAAICEAAD6AAAAgOgAAHUwAADqYAAAOpgAABdwnLpRPAAAAQtQTFRFAAAAAH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//////BQzC2AAAAFd0Uk5TAAADLy4QZVEHKp8FAUnHbeJ3BAh68IYGC4f4nQyM/LkYCYnXf/rvAm/2/oFY7rcTPuHkOCEky3YjlW4Pqbww0MVTfUZA96p061Xs3mz1e4P70R2aHJYf2KM0AgAAAAFiS0dEWO21xI4AAAAJcEhZcwAAEysAABMrAbkohUIAAADTSURBVCjPbdDZUsJAEAXQXAgJIUDCogHBkbhFEIgCsqmo4MImgij9/39iUT4Qkp63OV0zfbsliTkIhWWOEVHUKOdaTNER9HgiaYQY1xUzlWY8kz04tBjP5Y8KRc6PxUmJcftUnMkIFGCdX1yqjDtX5cp1MChQrVHd3Xn8/y1wc0uNpuejZmt7Ae7aJDreBt1e3wVw/0D06HobYPD0/GI7Q0G10V4i4NV8e/8YE/V8KwImUxJEM82fFM78k4gW3MhfS1p9B3ckobgWBpiChJ/fjc//AJIfFr4X0swAAAAAJXRFWHRkYXRlOmNyZWF0ZQAyMDE2LTA3LTMwVDE0OjI3OjUxLTA3OjAwUzMirAAAACV0RVh0ZGF0ZTptb2RpZnkAMjAxNi0wNy0zMFQxNDoyNzo0NC0wNzowMLz8tSkAAAAZdEVYdFNvZnR3YXJlAHd3dy5pbmtzY2FwZS5vcmeb7jwaAAAAFXRFWHRUaXRsZQBibHVlIENoZWNrIG1hcmsiA8jIAAAAAElFTkSuQmCC);
}

.codeBlockButton {
  height: 100%;
  cursor: pointer;
}

.codeBlockButton:hover {
  color: {{.ColorCodeHover}};
}

.codeBlockSpacer {
  height: 100%;
  width: 5px;
}

.codeBox {
  padding-top: 10px;
  padding-left: 20px;
}

.codeBlockControl {
  font-family: "Lucida Console", Monaco, monospace;
  font-weight: bold;
}

.codeblockBody {
  white-space: pre;
  font-family: "Lucida Console", Monaco, monospace;
  color: {{.ColorCodeBlockText}};
  background-color: {{.ColorCodeBlockBackground}};
  margin-top: 5px;
  padding-left: 10px;
  overflow-x: auto;
	border: solid 1px #555;
	border-radius: 4px;
           /*   x   y blur spread color             x   y blur spread color */
  box-shadow: 0px 2px  2px    1px rgba(0,0,0,.3), 2px 0px 2px 1px rgba(0,0,0,.3);

  /* This is hard to get right with current structure. */
  min-width: {{.LayMinHeaderWidth}};
  max-width: calc(100% - 40px);
}

.proseblock {
}

.oneLesson {
  display: none;
  padding: 0 1em 0 1em;
  width: '100%';
  padding-bottom: 1em;
}

.navBurger {
  display: inline-block;
  cursor: pointer;
}

.burgBar1, .burgBar2, .burgBar3 {
  width: 28px;
  height: 4px;
  /* top rig bot lef */
  margin: 6px 0 6px 0px;
  transition: {{.TransitionSpeed}};
	box-shadow: 0px 1px 1px 1px rgba(0,0,0,0.4);
  border: solid 1px #555;
	background-color: {{.ColorControls}};
	border-radius:25px;
}
.burgBar1:hover, .burgBar2:hover, .burgBar3:hover {
  background-color: {{.ColorHover}};
}

.burgIsAnX .burgBar1 {
  -webkit-transform: translate(-3px, 0px) rotate(-45deg);
  transform: translate(-4px, 0px) rotate(-45deg);
}
.burgIsAnX .burgBar2 {
}
.burgIsAnX .burgBar3 {
  -webkit-transform: translate(-3px, 0px) rotate(45deg);
  transform: translate(-4px, 0px) rotate(45deg);
}
`

const jsInHeader = `
function getElByClass(n) {
  return document.getElementsByClassName(n)[0];
}

function getDataId(el) {
  return el.getAttribute("data-id");
}

function isVertScrollBarVisible() {
  return document.body.scrollHeight > document.body.clientHeight;
}

var nav = new function() {
  var theHelpBox = null;
  var theBurger = null;
  var theLeftBox = null;
  var theRightBox = null;
  var theLeftSpacer = null;
  var theRightSpacer = null;
  var theBody = null;
  var theProseColumn = null;
  var mqWide = null;
  var mqMedium = null;
  var navBoxWidth = '{{.LayNavBoxWidth}}px'
  var bodyWideWidth = '{{.LayBodyWideWidth}}px';
  var bodyMediumWidth = '{{.LayBodyMediumWidth}}px';

  var fudge = function() {
    // approximate scroll bar width
    return isVertScrollBarVisible() ? '14px' : '0px';
  }
  var showBurger = function() {
    theBurger.classList.add('burgIsAnX');
  }
  var hideBurger = function() {
    theBurger.classList.remove('burgIsAnX');
  }
  var hideABox = function(x) {
    x.width = '0px';
    x.minWidth = '0px';
    x.overflow = 'hidden';
  }
  var showABox = function(x) {
    x.width = navBoxWidth;
    x.minWidth = navBoxWidth;
    x.overflow = 'auto';
  }
  var hideBoxes = function() {
    theLeftBox.classList.remove('navLeftBoxShadow');
    theRightBox.classList.remove('navRightBoxShadow');
    hideABox(theLeftBox.style);
    hideABox(theRightBox.style);
  }
  var showBoxes = function() {
    theLeftBox.classList.add('navLeftBoxShadow');
    theRightBox.classList.add('navRightBoxShadow');
    showABox(theLeftBox.style);
    showABox(theRightBox.style);
  }
  var expandCenter = function() {
    theLeftSpacer.display = 'none';
    theRightSpacer.display = 'none';
    theProseColumn.width = 'inherit'
  }
  var squeezeCenter = function() {
    theLeftSpacer.display = 'inline-block';
    theRightSpacer.display = 'inline-block';
    theProseColumn.width = '100%';
  }
  var showNarrow = function() {
    theRightBox.style.right = '0px'
    showBoxes()
    showBurger()
  }
  var showMedium = function() {
    showNarrow()
    squeezeCenter()
  }
  var showWide = function() {
    theRightBox.style.right =
       'calc(100vw - (' + bodyWideWidth + ' + ' + fudge() + '))';
    showBoxes()
    showBurger()
    squeezeCenter()
  }
  var hideNarrow = function() {
    hideBoxes()
    hideBurger()
  }
  var hideMedium = function() {
    expandCenter()
    hideNarrow()
  }
  var hideWide = function() {
    hideMedium()
  }
  var showIt = function() {
    alert('show nav not set')
  }
  var hideIt = function(){
    alert('hide nav not set')
  }
  var isVisible = function() {
    return theBurger.classList.contains('burgIsAnX')
  }
  this.handleWidthChange = function(discard) {
    if (mqWide.matches) {
      theBody.width = bodyWideWidth;
      theHelpBox.left = navBoxWidth;
      theHelpBox.right =
        'calc(100vw - ' + bodyWideWidth + ' + '
        + navBoxWidth + ' - ' + fudge() + ')';
      showIt = showWide
      hideIt = hideWide
    } else if (mqMedium.matches) {
      theBody.width = '100%';
      theHelpBox.left = '0px';
      theHelpBox.right = '0px';
      showIt = showMedium
      hideIt = hideMedium
    } else {
      theBody.width = '100%';
      theHelpBox.left = '0px';
      theHelpBox.right = '0px';
      expandCenter();
      showIt = showNarrow
      hideIt = hideNarrow
    }
    if (isVisible()) {
      showIt()
    } else {
      hideIt()
    }
  }
  this.toggle = function() {
    if (isVisible()) {
      hideIt()
    } else {
      showIt()
    }
  }
  this.initialize = function(showNav) {
    theBurger = getElByClass('navBurger');
    theHelpBox = getElByClass('helpBox').style;
    theLeftBox = getElByClass('navLeftBox');
    theRightBox = getElByClass('navRightBox');
    theLeftSpacer = getElByClass('navLeftSpacer').style;
    theRightSpacer = getElByClass('navRightSpacer').style;
    theBody = document.getElementById('body').style;
    theProseColumn = getElByClass('proseColumn').style;

    if ({{.LessonCount}} < 2) {
      theBurger.style.display = 'none';
    }

    mqWide = window.matchMedia('(min-width: ' + bodyWideWidth + ')');
    mqMedium = window.matchMedia(
        '(min-width: ' + bodyMediumWidth
        + ') and (max-width: ' + bodyWideWidth + ')');
    mqWide.addListener(this.handleWidthChange);
    mqMedium.addListener(this.handleWidthChange)
    this.handleWidthChange('whatever');
    if (showNav) {
      if (!isVisible()) {
        showIt()
      }
    } else {
      if (isVisible()) {
        hideIt()
      }
    }
  }
}

var help = new function() {
  var box = null
  var hideIt = function() {
    box.height = '0px';
    box.overflow = 'hidden';
    box.removeProperty('border');
    box.removeProperty('border-radius');
    box.removeProperty('box-shadow');
  }
  var showIt = function() {
    box.height = 'calc(100vh - ({{.LayFooterHeight}}px + {{.LayHeaderHeight}}px))';
    box.overflow = 'auto';
    box.border =  'solid 1px #555';
    box.borderRadius = '4px';
    box.boxShadow = '0px 2px  2px    1px rgba(0,0,0,.3), 2px 0px 2px 1px rgba(0,0,0,.3)';
  }
  var isVisible = function() {
    return (box.height != '0px')
  }
  this.toggle = function() {
    if (isVisible()) {
      hideIt()
    } else {
      showIt()
    }
  }
  this.initialize = function() {
    box = getElByClass('helpBox').style;
    hideIt()
  }
}

var codeBlock = new function() {
  var requestRunning = false

  var addCheck = function(el) {
    var t = 'span';
    var c = document.createElement(t);
    c.setAttribute('class', 'codeBlockCheckOff');
    el.appendChild(c);
  }

  // https://stackoverflow.com/questions/400212
  var attemptCopyToBuffer = function(text) {
    var tA = document.createElement("textarea");
    tA.style.position = 'fixed';
    tA.style.top = 0;
    tA.style.left = 0;
    tA.style.width = '2em';
    tA.style.height = '2em';
    tA.style.padding = 0;
    tA.style.border = 'none';
    tA.style.outline = 'none';
    tA.style.boxShadow = 'none';
    tA.style.background = 'transparent';
    tA.value = text;
    document.body.appendChild(tA);
    tA.select();
    try {
      var successful = document.execCommand('copy');
      var msg = successful ? 'successful' : 'unsuccessful';
      console.log('Copying text command was ' + msg);
    } catch (err) {
      console.log('Oops, unable to copy');
    }
    document.body.removeChild(tA);
  }

  this.run = function(event) {
    if (!(event && event.target)) {
      alert('no event!');
      return
    }
    if (requestRunning) {
      alert('busy!');
      return
    }
    requestRunning = true;
    var b = event.target;
    var codeBox = b.parentNode.parentNode;
    // Fragile, but brief!
    var codeBody = codeBox.childNodes[3].firstChild;
    attemptCopyToBuffer(codeBody.textContent)
    var blockId = getDataId(codeBox.parentNode);
    var fileId = getDataId(codeBox.parentNode.parentNode);
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function() {
      if (xhr.readyState == XMLHttpRequest.DONE) {
        addCheck(b.parentNode)
        requestRunning = false;
      }
    };
    xhr.open('GET', '/_/runblock?fid=' + fileId + '&bid=' + blockId + '&sid={{.SessId}}', true);
    xhr.send();
  }
}

var header = new function() {
  var theHeader = null;
  var theActiveLessonName = null;
  var theLessonNavRow = null;
  var theTitle = null;

  var hideIt = function() {
    theHeader.height = '{{.LayMinHeaderHeight}}px';
    theLessonNavRow.display = 'none';
    theTitle.removeProperty('min-height');
    theTitle.fontSize = '1em';
    /*
        Also have to change 'top' and 'height' in
           navLeftBox, navRightBox, helpbox
        the most elegant way is to exxpress the top and height
        as js functions in these latter objects, and have that
        function check the visibility of the header.
        The both depend on the header, but the header does not
        depend on them

     */
  }
  var showIt = function() {
    theHeader.height = '{{.LayHeaderHeight}}px';
    theLessonNavRow.display = 'flex';
    theTitle.minHeight = '2em';
    theTitle.fontSize = '2em';
  }
  var isVisible = function() {
    return (theHeader.height == '{{.LayHeaderHeight}}px');
  }
  this.h = function() {
    return theHeader;
  }
  this.r = function() {
    return theLessonNavRow;
  }
  this.v = function() {
    return isVisible();
  }
  this.toggle = function() {
    if (isVisible()) {
      hideIt()
    } else {
      showIt()
    }
  }

  this.initialize = function() {
    theHeader = document.getElementById('header').style;
    theTitle = document.getElementById('title').style;
    theLessonNavRow = document.getElementsByClassName('lessonNavRow')[0].style;
    showIt();
  }
}


var lessonMgr = new function() {
  var activeIndex = -1;
  var coursePaths = null;
  var theLessonName = null;
  var thePrevName = null;
  var theNextName = null;
  var theLessonPrevPointer = null;
  var theLessonNextPointer = null;

  var nextIndex = function(i) {
    return (i + 1) % coursePaths.length;
  }
  var prevIndex = function(i) {
    if (i < 1) {
      return coursePaths.length - 1;
    }
    return i - 1;
  }
  var getBodyLesson = function(i) {
    return document.getElementById('BL' + i.toString())
  }
  var getNavLesson = function(i) {
    return document.getElementById('NL' + i.toString())
  }
  var getNavCourse = function(i) {
    return document.getElementById('NC' + i.toString())
  }
  var assureNoActiveLesson = function() {
    if (activeIndex == -1) {
      return
    }
    theLessonName.innerHTML = ''
    for (i = 0; i < 2; i++) {
      thePrevName[i].innerHTML = '';
      theNextName[i].innerHTML = '';
    }
    getBodyLesson(activeIndex).style.display = 'none'
    getNavLesson(activeIndex).className = 'navLessonTitleOff'
    activeIndex = -1
  }
  var assureNoActiveCourse = function() {
    for (id = 0; id < 100; id++) {
      el = getNavCourse(id)
      if (el == null) {
        return;
      }
      el.style.display = 'none'
    }
  }
  var assureActiveCourse = function(id) {
    el = getNavCourse(id)
    if (el == null) {
      return;
    }
    el.style.display = 'block'
  }
  var assureActivePath = function(lesson) {
    if (lesson < 0 || lesson > coursePaths.length) {
      console.log("lesson out of lessonsPaths range " + lesson.toString())
      return
    }
    courses = coursePaths[lesson]
    for (i = 0; i < courses.length; i++) {
      assureActiveCourse(courses[i]);
    }
  }

  var dToggle = function(e) {
    e.style.display = (e.style.display == 'block') ? 'none' : 'block'
  }

  this.toggleNC = function(index) {
    dToggle(getNavCourse(index));
  }

  var smoothScroll = function() {
    var currentScroll =
        document.documentElement.scrollTop || document.body.scrollTop;
    if (currentScroll > 0) {
      window.requestAnimationFrame(smoothScroll);
      window.scrollTo(0,currentScroll - (currentScroll/5));
    }
  }

  var updateUrl = function(path) {
    theLessonName.innerHTML = path
    if (history.pushState) {
      window.history.pushState("not using data yet", "someTitle", "/" + path);
    } else {
      document.location.href = path;
    }
  }

  var updateHeader = function(index) {
    var path = '';
    var ptr = '';
    if (index < coursePaths.length - 1) {
      var e = getNavLesson(nextIndex(index));
      path = e.getAttribute('data-path');
      ptr = '&gt;';
    }
    theNextName[0].innerHTML = path;
    theNextName[1].innerHTML = path;
    theLessonNextPointer[0].innerHTML = ptr;
    theLessonNextPointer[1].innerHTML = ptr;

    path = '';
    ptr = '';
    if (index > 0) {
      var e = getNavLesson(prevIndex(index));
      path = e.getAttribute('data-path');
      ptr = '&lt;';
    }
    thePrevName[0].innerHTML = path;
    thePrevName[1].innerHTML = path;
    theLessonPrevPointer[0].innerHTML = ptr;
    theLessonPrevPointer[1].innerHTML = ptr;

    var e = getNavLesson(index);
    e.className = 'navLessonTitleOn'
    updateUrl(e.getAttribute('data-path'))
  }

  this.assureActiveLesson = function(index) {
    if (activeIndex == index) {
      return
    }
    var prevState = isVertScrollBarVisible();
    if (activeIndex > -1) {
      assureNoActiveLesson()
      assureNoActiveCourse()
    }
    assureActivePath(index)
    var e = getBodyLesson(index)
    if (e == null) {
      console.log("missing lesson " + index);
      return;
    }
    e.style.display = 'block'
    updateHeader(index);
    smoothScroll()
    if (prevState != isVertScrollBarVisible()) {
      nav.handleWidthChange('whatever');
    }
    activeIndex = index;
  }

  this.goNext = function() {
    this.assureActiveLesson(nextIndex(activeIndex))
  }

  this.goPrev = function() {
    this.assureActiveLesson(prevIndex(activeIndex))
  }

  this.initialize = function(cp) {
    coursePaths = cp;
    activeIndex = -1;
    theLessonName = getElByClass('activeLessonName');
    thePrevName = document.getElementsByClassName('lessonPrevTitle');
    theNextName = document.getElementsByClassName('lessonNextTitle');
    theLessonPrevPointer = document.getElementsByClassName('lessonPrevPointer');
    theLessonNextPointer = document.getElementsByClassName('lessonNextPointer');
  }
}

function onLoad() {
  help.initialize();
  header.initialize();
  nav.initialize(false /* {{.LessonCount}} > 1 */);
  lessonMgr.initialize({{.CoursePaths}});
  lessonMgr.assureActiveLesson({{.InitialLesson}});
  window.addEventListener('keydown', function (event) {
    if (event.defaultPrevented) {
      return;
    }
    switch (event.key) {
      case 'm':
        header.toggle();
        break;
      case 'n':
        nav.toggle();
        break;
      case '/':
      case '?':
        help.toggle();
        break;
      case 'j':
        alert('impl PREV block');
        break;
      case 'k':
        alert('impl NEXT block');
        break;
      case 'h':
      case 'ArrowLeft':
        lessonMgr.goPrev();
        break;
      case 'l':
      case 'ArrowRight':
        lessonMgr.goNext();
        break;
      default:
    }
  }, true);
}
`
