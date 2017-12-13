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

func (wa *WebApp) TransitionSpeedMs() int        { return 250 }
func (wa *WebApp) LayBodyWideWidth() int         { return 1200 }
func (wa *WebApp) LayBodyMediumWidth() int       { return 800 }
func (wa *WebApp) LayMinHeaderWidth() int        { return 400 }
func (wa *WebApp) LayNavBoxWidth() int           { return 210 }
func (wa *WebApp) LayHeaderHeight() int          { return 120 }
func (wa *WebApp) LayFooterHeight() int          { return 50 }
func (wa *WebApp) LayMinimizedHeaderHeight() int { return wa.LayFooterHeight() }
func (wa *WebApp) LayNavTopBotPad() int          { return 7 }
func (wa *WebApp) LayNavLeftPad() int            { return 20 }

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
<body onload='onLoad()'>

  <header id='header'>
    <div class='navButtonBox' onclick='navController.toggle()'>
      <div class='navBurger'>
        <div class='burgBar1'></div>
        <div class='burgBar2'></div>
        <div class='burgBar3'></div>
      </div>
    </div>
    <div class='headerColumn'>
      <a target='_blank' href='{{.AppLink}}'>
        <title id='title'> {{.TrimName}} </title>
      </a>
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
    <div class='headSpacer'> HEADER SPACER </div>
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
  <div class="commandBlockBody">
  {{ template "` + tmplNameBlockPgm + `" $c }}
  </div>
{{end}}
{{end}}
`
	tmplNameBlockPgm = "blockPgm"
	tmplBodyBlockPgm = `
{{define "` + tmplNameBlockPgm + `"}}
<div class='proseblock'> {{.HtmlProse}} </div>
{{if .Code}}
<div class='codeBox' data-id='{{.Id}}'>
  <div class='codeBlockControl'>
    <span class='codePrompt'> &nbsp;&gt;&nbsp; </span>
    <span class='codeBlockButton' onclick='codeBlockController.setAndRun({{.Id}})'>
      {{.Name}}
    </span>
    <span class='codeBlockSpacer'> &nbsp; </span>
  </div>
<div class='codeblockBody'>
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
  <div class='lessonPrevClickerRow' onclick='lessonController.goPrev()'>
    <div class='lessonPrevTitle'> quantum flux </div>
    <div class='lessonPrevPointer'> &lt; </div>
  </div>
  <div class='helpButtonBox' onclick='helpController.toggle()'> ? </div>
  <div class='lessonNextClickerRow' onclick='lessonController.goNext()'>
    <div class='lessonNextPointer'> &gt; </div>
    <div class='lessonNextTitle'> magnetic flux  </div>
  </div>
</div>
`

const htmlHelp = `
<p>
Markdown snapshot from
<a target='_blank' href='{{.AppLink}}'> <code> {{.AppName}} </code></a>

<h3>Keys</h3>
<ul>
  <li>&larr;, &rarr;, h, l: prev/next lesson </li>
  <li>j, k: activate prev/next code block </li>
  <li>&crarr;: copy activated block (or mouse click)</li>
  <li>?, /: help</li>
  <li>-: header</li>
  <li>n: nav sidebar</li>
  <li>m: monkey</li>
</ul>

Check marks track block execution progress.

<h3> Serve locally with tmux for no-mouse code block execution</h3>

<p>
Serve the content locally:
<pre>
  GOBIN=$TMPDIR go install github.com/monopole/mdrip
  $TMPDIR/mdrip --port 8001 --mode demo {{.AppName}}
</pre>
and run <a target="_blank"
href="https://github.com/tmux/tmux/wiki">tmux</a>:
<pre>
  tmux
</pre>
Then whatever action copies a code block (hitting &crarr; or mouse click)
also pastes the block to the active tmux session for
immediate execution.
</p>

<h3> Remote server tmux </h3>
<p> <em>Proof of concept
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
as described above, and run <a target="_blank"
href="https://github.com/tmux/tmux/wiki">tmux</a>.
</li>
<li>In some non-tmux shell, run mdrip in <em>tmux</em> mode with a session arg:
<pre>
  mdrip \
    --alsologtostderr --v 0 \
    --stderrthreshold INFO \
    --mode tmux \
    ws://{{.Host}}/_/ws?id={{.SessId}}
</pre>
</li>
</ul>
<p>
Now, a copy action sends the block
from this page's server over a websocket to the local
<code>mdrip</code>, which then pastes the block
to the active <code>tmux</code> pane.<p>
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

header, .headSpacer {
  height: {{.LayHeaderHeight}}px;
  width: inherit;
  transition: height {{.TransitionSpeedMs}}ms;
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
  transition: height {{.TransitionSpeedMs}}ms;
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
  transition: width {{.TransitionSpeedMs}}ms, min-width {{.TransitionSpeedMs}}ms;
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
  transition: height {{.TransitionSpeedMs}}ms;
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
  transition: width {{.TransitionSpeedMs}}ms;
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

.codePrompt {
  background-color: {{.ColorCodeHover}};
  display: none;
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
  transition: {{.TransitionSpeedMs}}ms;
	box-shadow: 0px 1px 1px 1px rgba(0,0,0,0.4);
  border: solid 1px #555;
	background-color: {{.ColorControls}};
	border-radius:25px;
}
.burgBar1:hover, .burgBar2:hover, .burgBar3:hover {
  background-color: {{.ColorHover}};
}

.burgIsAnX .burgBar1 {
  transform: translate(-6px, 1px) rotate(-40deg);
}
.burgIsAnX .burgBar2 {
  transform: translate(4px, 0px);
}
.burgIsAnX .burgBar3 {
  transform: translate(-6px, -1px) rotate(40deg);
}
`

const jsInHeader = `
function getElByClass(n) {
  return document.getElementsByClassName(n)[0];
}

function getDataId(el) {
  return el.getAttribute("data-id");
}

function coinFlip() {
  return (Math.random() >= 0.5);
}

// return int in range 0..(n-1)
function randomInt(n) {
  return Math.floor(Math.random() * n)
}

var bodyController = new function() {
  var styleBody = null;

  this.isVertScrollBarVisible = function() {
    return document.body.scrollHeight > document.body.clientHeight;
  }
  this.vertScrollBarWidth = function() {
    return this.isVertScrollBarVisible() ? '14px' : '0px';
  }
  this.getWideWidth = function() {
    return '{{.LayBodyWideWidth}}px';
  }
  this.getMediumWidth = function() {
    return '{{.LayBodyMediumWidth}}px';
  }
  this.enterModeWide = function() {
    styleBody.width = this.getWideWidth();
  }
  this.enterModeMedium = function() {
    this.enterModeNarrow();
  }
  this.enterModeNarrow = function() {
    styleBody.width = '100%';
  }
  this.initialize = function() {
    styleBody = document.body.style;
  }
}

var headerController = new function() {
  var styleHeader = null;
  var styleHeadSpacer = null;
  var styleLessonNavRow = null;
  var styleTitle = null;

  var setHeight = function(x) {
    styleHeader.height = x;
    styleHeadSpacer.height = x;
  }
  var hideIt = function() {
    setHeight('{{.LayMinimizedHeaderHeight}}px');
    styleLessonNavRow.display = 'none';
    styleTitle.removeProperty('min-height');
    styleTitle.fontSize = '1em';
  }
  var showIt = function() {
    setHeight('{{.LayHeaderHeight}}px');
    styleLessonNavRow.display = 'flex';
    styleTitle.minHeight = '2em';
    styleTitle.fontSize = '2em';
  }
  var isVisible = function() {
    return (styleHeader.height == '{{.LayHeaderHeight}}px');
  }
  this.height = function() {
    return styleHeader.height;
  }
  this.toggle = function() {
    if (isVisible()) {
      hideIt()
    } else {
      showIt()
    }
    navController.render();
    helpController.render();
  }
  this.initialize = function() {
    styleHeader = document.getElementById('header').style;
    styleHeadSpacer = getElByClass('headSpacer').style;
    styleTitle = document.getElementById('title').style;
    styleLessonNavRow = document.getElementsByClassName('lessonNavRow')[0].style;
  }
  this.reset = function() {
    showIt()
  }
}

var navController = new function() {
  var elBurger = null;
  var elNavLeft = null;
  var elNavRight = null;
  var styleSpacerLeft = null;
  var styleSpacerRight = null;
  var styleProseColumn = null;
  var mqWide = null;
  var mqMedium = null;

  var showBurger = function() {
    elBurger.classList.add('burgIsAnX');
  }
  var hideBurger = function() {
    elBurger.classList.remove('burgIsAnX');
  }
  var hideABox = function(x) {
    x.width = '0px';
    x.minWidth = '0px';
    x.overflow = 'hidden';
  }
  var showABox = function(x) {
    x.width = '{{.LayNavBoxWidth}}px';
    x.minWidth = '{{.LayNavBoxWidth}}px';
    x.overflow = 'auto';
  }
  var hideBoxes = function() {
    elNavLeft.classList.remove('navLeftBoxShadow');
    elNavRight.classList.remove('navRightBoxShadow');
    hideABox(elNavLeft.style);
    hideABox(elNavRight.style);
  }
  var setTopAndHeight = function(b) {
    // leave room for header drop-shadow
    b.top = 'calc(' + headerController.height() + ' + 2px)';
    b.height = 'calc(100vh - ({{.LayFooterHeight}}px + '
        + headerController.height() + ' + 4px))';
  }
  var showBoxes = function() {
    setTopAndHeight(elNavLeft.style);
    setTopAndHeight(elNavRight.style);
    elNavLeft.classList.add('navLeftBoxShadow');
    elNavRight.classList.add('navRightBoxShadow');
    showABox(elNavLeft.style);
    showABox(elNavRight.style);
  }
  var expandCenter = function() {
    styleSpacerLeft.display = 'none';
    styleSpacerRight.display = 'none';
    styleProseColumn.width = 'inherit'
  }
  var squeezeCenter = function() {
    styleSpacerLeft.display = 'inline-block';
    styleSpacerRight.display = 'inline-block';
    styleProseColumn.width = '100%';
  }
  var showNarrow = function() {
    elNavRight.style.right = '0px'
    showBoxes()
    showBurger()
  }
  var showMedium = function() {
    showNarrow()
    squeezeCenter()
  }
  var showWide = function() {
    elNavRight.style.right =
       'calc(100vw - (' + bodyController.getWideWidth()
       + ' + ' + bodyController.vertScrollBarWidth() + '))';
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
    return elBurger.classList.contains('burgIsAnX')
  }
  var myRender = function() {
    if (isVisible()) {
      showIt()
    } else {
      hideIt()
    }
  }
  this.render = function() {
    myRender()
  }
  this.handleWidthChange = function(discard) {
    if (mqWide.matches) {
      bodyController.enterModeWide();
      helpController.enterModeWide();
      showIt = showWide
      hideIt = hideWide
    } else if (mqMedium.matches) {
      bodyController.enterModeMedium();
      helpController.enterModeMedium();
      showIt = showMedium
      hideIt = hideMedium
    } else {
      bodyController.enterModeNarrow();
      helpController.enterModeNarrow();
      expandCenter();
      showIt = showNarrow
      hideIt = hideNarrow
    }
    myRender();
  }
  this.toggle = function() {
    if (isVisible()) {
      hideIt()
    } else {
      showIt()
    }
  }
  this.initialize = function() {
    elBurger = getElByClass('navBurger');
    elNavLeft = getElByClass('navLeftBox');
    elNavRight = getElByClass('navRightBox');
    styleSpacerLeft = getElByClass('navLeftSpacer').style;
    styleSpacerRight = getElByClass('navRightSpacer').style;
    styleProseColumn = getElByClass('proseColumn').style;
    if ({{.LessonCount}} < 2) {
      elBurger.style.display = 'none';
    }
    mqWide = window.matchMedia(
        '(min-width: ' + bodyController.getWideWidth() + ')');
    mqMedium = window.matchMedia(
        '(min-width: ' + bodyController.getMediumWidth()
        + ') and (max-width: ' + bodyController.getWideWidth() + ')');
    mqWide.addListener(this.handleWidthChange);
    mqMedium.addListener(this.handleWidthChange)
    this.handleWidthChange('whatever');
  }
  this.reset = function() {
    hideIt();
  }
}

var helpController = new function() {
  var style = null
  var hideIt = function() {
    style.top = headerController.height();
    style.height = '0px';
    style.overflow = 'hidden';
    style.removeProperty('border');
    style.removeProperty('border-radius');
    style.removeProperty('box-shadow');
  }
  var showIt = function() {
    style.top = headerController.height();
    style.height = 'calc(100vh - ({{.LayFooterHeight}}px + '
        + headerController.height() + '))';
    style.overflow = 'auto';
    style.border =  'solid 1px #555';
    style.borderRadius = '4px';
    style.boxShadow = '0px 2px 2px 1px rgba(0,0,0,.3), 2px 0px 2px 1px rgba(0,0,0,.3)';
  }
  var isVisible = function() {
    return (style.height != '0px')
  }
  this.enterModeWide = function() {
    style.left = '{{.LayNavBoxWidth}}px';
    style.right =
        'calc(100vw - ' + bodyController.getWideWidth() + ' + '
        + '{{.LayNavBoxWidth}}px' + ' - '
        + bodyController.vertScrollBarWidth() + ')';
  }
  this.enterModeMedium = function() {
    this.enterModeNarrow();
  }
  this.enterModeNarrow = function() {
    style.left = '0px';
    style.right = '0px';
  }
  this.render = function() {
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
  this.initialize = function() {
    style = getElByClass('helpBox').style;
  }
  this.reset = function() {
    hideIt()
  }
}

var codeBlockController = new function() {
  var blocks = null;
  var requestRunning = false;
  var cbIndex = -1;
  var addCheck = function(el) {
    var t = 'span';
    var c = document.createElement(t);
    c.setAttribute('class', 'codeBlockCheckOff');
    el.appendChild(c);
  }
  var hideIt = function(s) {
    s.position = 'fixed';
    s.top = 0;
    s.left = 0;
    s.width = '2em';
    s.height = '2em';
    s.padding = 0;
    s.border = 'none';
    s.outline = 'none';
    s.boxShadow = 'none';
    s.background = 'transparent';
  }
  // https://stackoverflow.com/questions/400212
  var attemptCopyToBuffer = function(text) {
    var tA = document.createElement("textarea");
    hideIt(tA.style);
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
  this.goPrev = function() {
    if (cbIndex < 1) {
      // Do nothing, not even modulo wrap.
      // Behave like an editor.
      return;
    }
    this.deActivateCurrent();
    cbIndex--;
    activateCurrent();
  }
  this.goNext = function() {
    if (cbIndex >= blocks.length - 1) {
      // Do nothing, not even modulo wrap.
      return;
    }
    this.deActivateCurrent();
    cbIndex++;
    activateCurrent();
  }
  var controlBar = function() {
    return blocks[cbIndex].firstElementChild;
  }
  var prompt = function() {
    return controlBar().firstElementChild;
  }
  var goodIndex = function(i) {
    return i >= 0 && i < blocks.length
  }
  this.deActivateCurrent = function() {
    if (!goodIndex(cbIndex)) {
      return;
    }
    prompt().style.display = 'none';
  }
  var activateCurrent = function() {
    if (!goodIndex(cbIndex)) {
      return;
    }
    prompt().style.display = 'inline-block';
    blocks[cbIndex].scrollIntoView(
      {behavior: 'smooth', block: 'center', inline: 'nearest'});
  }
  this.currentBlock = function() {
    if (goodIndex(cbIndex)) {
      return blocks[cbIndex]
    }
  }
  this.initLesson = function(elLesson) {
    cbIndex = -1;
    blocks = elLesson.querySelectorAll('[data-id]');
    for (i = 0; i < blocks.length; i++) {
       var b = blocks[i];
       var id = parseInt(b.getAttribute('data-id'));
       if (i != id) {
         console.log("Counting problem")
       }
    }
  }
  // For monkeyController
  this.toggle = function() {
    this.setCurrent(randomInt(blocks.length));
  }
  this.reset = function() {
    this.deActivateCurrent()
  }
  this.setCurrent = function(id) {
    if (!goodIndex(id)) {
      return false;
    }
    if (cbIndex != id) {
      this.deActivateCurrent()
    }
    cbIndex = id;
    activateCurrent();
    return true;
  }
  this.setAndRun = function(id) {
    if (!this.setCurrent(id)) {
      alert('bad id: ' + id);
      return
    }
    this.runCurrent();
  }
  this.runCurrent = function() {
    if (!goodIndex(cbIndex)) {
      console.log("cannot run block " + cbIndex);
      return;
    }
    if (requestRunning) {
      alert('busy!');
      return;
    }
    requestRunning = true;
    var codeBox = blocks[cbIndex];
    // Fragile, but brief!
    var codeBody = codeBox.childNodes[3].firstChild;
    attemptCopyToBuffer(codeBody.textContent)
    var fileId = getDataId(codeBox.parentNode.parentNode);
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function() {
      if (xhr.readyState == XMLHttpRequest.DONE) {
        addCheck(codeBox.childNodes[1])
        requestRunning = false;
      }
    };
    xhr.open(
        'GET',
        '/_/runblock?fid=' + fileId
            + '&bid=' + cbIndex
            + '&sid={{.SessId}}',
        true);
    xhr.send();
  }
  this.initialize = function() {
    requestRunning = false;
    cbIndex = -1;
  }
}

var lessonController = new function() {
  var activeIndex = -1;
  var coursePaths = null;
  var elLessonName = null;
  var elPrevName = null;
  var elNextName = null;
  var elLessonPrevPointer = null;
  var elLessonNextPointer = null;

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
    elLessonName.innerHTML = ''
    for (i = 0; i < 2; i++) {
      elPrevName[i].innerHTML = '';
      elNextName[i].innerHTML = '';
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
  this.ncToggle = function(index) {
    dToggle(getNavCourse(index));
  }
  // For monkeyController
  this.toggle = function() {
    this.assureActiveLesson(randomInt(coursePaths.length));
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
    elLessonName.innerHTML = '/' + path
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
      path = '/' + e.getAttribute('data-path');
      ptr = '&gt;';
    }
    elNextName[0].innerHTML = path;
    elNextName[1].innerHTML = path;
    elLessonNextPointer[0].innerHTML = ptr;
    elLessonNextPointer[1].innerHTML = ptr;

    path = '';
    ptr = '';
    if (index > 0) {
      var e = getNavLesson(prevIndex(index));
      path = '/' + e.getAttribute('data-path');
      ptr = '&lt;';
    }
    elPrevName[0].innerHTML = path;
    elPrevName[1].innerHTML = path;
    elLessonPrevPointer[0].innerHTML = ptr;
    elLessonPrevPointer[1].innerHTML = ptr;

    var e = getNavLesson(index);
    e.className = 'navLessonTitleOn'
    updateUrl(e.getAttribute('data-path'))
  }
  this.assureActiveLesson = function(index) {
    if (activeIndex == index) {
      return
    }
    var prevState = bodyController.isVertScrollBarVisible();
    if (activeIndex > -1) {
      codeBlockController.deActivateCurrent();
      assureNoActiveLesson()
      assureNoActiveCourse()
    }
    assureActivePath(index)
    var elLesson = getBodyLesson(index)
    if (elLesson == null) {
      console.log("missing lesson " + index);
      return;
    }
    elLesson.style.display = 'block'
    updateHeader(index);
    codeBlockController.initLesson(elLesson);
    smoothScroll()
    if (prevState != bodyController.isVertScrollBarVisible()) {
      navController.handleWidthChange('whatever');
    }
    activeIndex = index;
  }
  this.goNext = function() {
    codeBlockController.deActivateCurrent();
    this.assureActiveLesson(nextIndex(activeIndex))
  }
  this.goPrev = function() {
    codeBlockController.deActivateCurrent();
    this.assureActiveLesson(prevIndex(activeIndex))
  }
  this.initialize = function(cp) {
    coursePaths = cp;
    activeIndex = -1;
    elLessonName = getElByClass('activeLessonName');
    elPrevName = document.getElementsByClassName('lessonPrevTitle');
    elNextName = document.getElementsByClassName('lessonNextTitle');
    elLessonPrevPointer = document.getElementsByClassName('lessonPrevPointer');
    elLessonNextPointer = document.getElementsByClassName('lessonNextPointer');
  }
  this.reset = function() {
    this.assureActiveLesson({{.InitialLesson}});
  }
}

var monkeyController = new function() {
  var items = null;
  var on = false;
  var interval = null;
  var itemReset = function(item, i) {
    item.reset();
  }
  var run = function() {
    items[randomInt(items.length)].toggle();
  }
  this.toggle = function() {
    if (on) {
      window.clearInterval(interval);
      interval = null;
      this.reset();
      on = false;
      return;
    }
    interval = window.setInterval(run, {{.TransitionSpeedMs}} + 50);
    on = true;
  }
  this.initialize = function(x) {
    items = x;
    on = false;
  }
  this.reset = function() {
    items.forEach(itemReset);
  }
}

function onLoad() {
  headerController.initialize();
  bodyController.initialize();
  helpController.initialize();
  navController.initialize();
  lessonController.initialize({{.CoursePaths}});
  codeBlockController.initialize();
  monkeyController.initialize(
      new Array(
          headerController, helpController,
          codeBlockController, navController, lessonController));
  monkeyController.reset();
  window.addEventListener('keydown', function (event) {
    if (event.defaultPrevented) {
      return;
    }
    switch (event.key) {
      case 'Enter':
      case 'r':
        codeBlockController.runCurrent();
        break;
      case '-':
        headerController.toggle();
        break;
      case 'n':
        navController.toggle();
        break;
      case '/':
      case '?':
        helpController.toggle();
        break;
      case 'k':
        codeBlockController.goPrev();
        break;
      case 'j':
        codeBlockController.goNext();
        break;
      case 'h':
      case 'ArrowLeft':
        lessonController.goPrev();
        break;
      case 'l':
      case 'ArrowRight':
        lessonController.goNext();
        break;
      case 'm':
        monkeyController.toggle();
        break;
      default:
    }
  }, true);
}
`
