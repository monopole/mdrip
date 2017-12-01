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
	tmpl        *template.Template
	lessonPath  []int
	coursePaths [][]int
}

func NewWebApp(
	sessId TypeSessId, host string,
	tut model.Tutorial, lp []int, cp [][]int) *WebApp {
	return &WebApp{sessId, host, tut, makeParsedTemplate(tut), lp, cp}
}

func (wa *WebApp) SessId() TypeSessId { return wa.sessId }
func (wa *WebApp) Host() string       { return wa.host }

// func (wa *WebApp) Tutorial() model.Tutorial { return wa.tut }
func (wa *WebApp) Lessons() []*program.LessonPgm {
	v := program.NewLessonPgmExtractor(base.WildCardLabel)
	wa.tut.Accept(v)
	return v.Lessons()
}

// This should probably be some text passed to the ctor instead,
// after pulling it from the command line.
func (wa *WebApp) AppName() string {
	return wa.tut.Name()
}

func (wa *WebApp) TrimName() string {
	result := strings.TrimSpace(wa.AppName())
	if len(result) > maxAppNameLen {
		return result[maxAppNameLen-3:] + "..."
	}
	return result
}

const (
	delta         = 2
	maxAppNameLen = len("gh:kubernetes/kubernetes.github.io")
)

func (wa *WebApp) TransitionSpeed() string { return "0.4s" }

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

func (wa *WebApp) LayBodyWideWidth() int   { return 1200 }
func (wa *WebApp) LayBodyMediumWidth() int { return 800 }
func (wa *WebApp) LayMinHeaderWidth() int  { return 400 }
func (wa *WebApp) LayNavBoxWidth() int     { return 210 }
func (wa *WebApp) LayHeaderHeight() int    { return 120 }
func (wa *WebApp) LayFooterHeight() int    { return 70 }
func (wa *WebApp) LayNavTopBotPad() int    { return 7 }
func (wa *WebApp) LayNavLeftPad() int      { return 20 }

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
  <header>
    <div class='navButtonBox' onclick='nav.toggle()'>
      <div class='navBurger'>
        <div class='burgBar1'></div>
        <div class='burgBar2'></div>
        <div class='burgBar3'></div>
      </div>
    </div>
    <div class='headerColumn'>
      <title> {{ .TrimName }} </title>
      <div class='activeLessonName'> Droplet Formation Rates </div>
      ` + htmlLessonNavRow + `
    </div>
    <div class='navButtonBox'> &nbsp; </div>
  </header>

  <div class='navLeftBox'>
    <nav class='navActual'>
      ` + htmlNavActual + `
    </nav>
  </div>
  <div class='helpBox'>
    <div class='helpActual'>
    ` + htmlHelp + `
    </div>
  </div>

  <div class='navRightBox'>
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
<h3 class="codeBlockControl">
  <span class="codeBlockButton" onclick="codeBlock.run(event)">
     {{.Name}}
  </span>
  <span class="codeBlockSpacer"> &nbsp; </span>
</h3>
<pre class="codeblockBody">
{{ .Code }}
</pre>
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
    <div class='lessonPointer'> &lt; </div>
  </div>
  <div class='helpButtonBox' onclick='help.toggle()'> ? </div>
  <div class='lessonNextClickerRow' onclick='lessonMgr.goNext()'>
    <div class='lessonPointer'> &gt; </div>
    <div class='lessonNextTitle'> electromagnetic  </div>
  </div>
</div>
`

const htmlHelp = `
<p>This is markdown content harvested from</p>
<blockquote>
<code> {{.AppName}} </code>
</blockquote>
<p>Clicking on a code block header copies the block to your clipboard.</p>
<p>
For one-click usage (no need to mouse/aim/paste - nice for demos):
<ul>
<li>
Install <code><a target="_blank"
href="https://github.com/tmux/tmux/wiki">tmux</a></code>
(the terminal multiplexer).</li>
<li>
Install <code><a target="_blank"
href="https://golang.org/doc/install">Go</a></code>
(the language).</li>
<li>Install <code><a target="_blank"
href="https://github.com/monopole/mdrip">mdrip</a></code>
(a <code>tmux</code> websocket adapter):
<pre>
  GOBIN=$TMPDIR go install github.com/monopole/mdrip
</pre>
</li>
<li>Run tmux:
<pre>
  tmux
</pre>
</li>
<li>In some non-tmux shell, run this service:
<pre>
  host=ws://{{.Host}}
  $TMPDIR/mdrip \
      --alsologtostderr --v 0 \
      --stderrthreshold INFO \
      --mode tmux \
      ${host}/_/ws?id={{.SessId}}
</pre>
</li>
</ul>
<p>
Now, clicking a command block header sends the block
from this page's server over a websocket to your local
<code>mdrip</code>, which 'pastes' the block
to your active <code>tmux</code> pane.<br>
The service self-exits after a period of inactivity,
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
  font-size: 14pt;
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
  background: linear-gradient(0deg, #f6f6ef, #00838f);
  display: flex;
  justify-content: space-between;
  flex-direction: row;
  flex-wrap: nowrap;
  align-items: center;
  box-shadow: 0 2px 2px 2px rgba(0,0,0,.4);
}

.navLeftBox, .navRightBox {
  position: fixed;
  top: calc({{.LayHeaderHeight}}px + 2px);  /* leave room for header drop-shadow */
  height: calc(100vh - ({{.LayFooterHeight}}px + {{.LayHeaderHeight}}px + 4px));
  background-color: #f6f6ef;
  display: inline-block;
  overflow: hidden;  /* initially hideNav */
  width: 0px;  /* initially hideNav */
  min-width: 0px;  /* initially hideNav */
  transition: width 0.2s, min-width 0.2s;
  /* offset-x | offset-y | blur-radius | spread-radius | color */
  /* box-shadow: 2px 2px 2px 0 rgba(0,0,0,.2), 0 3px 1px -2px rgba(0,0,0,.2), 0 1px 5px 0 rgba(0,0,0,.12); */
}
.navRightBox {
  opacity: 0.7;
  /* shadow on bottom, top and left */
  box-shadow: 0 2px 2px 2px rgba(0,0,0,.2), -2px 0px 2px 2px rgba(0,0,0,.2);
}
.navLeftBox {
  /* shadow on bottom, top and right */
  box-shadow: 0 2px 2px 2px rgba(0,0,0,.2), 2px 0px 2px 2px rgba(0,0,0,.2);
}
.navActual {
  padding-left: 1em;
}

.navCourseTitle {
  padding: 0px;
}

.navCourseTitle:hover {
  color: #06e;
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
  background-color: #ddd;
}

.navLessonTitleOff {
}

.navLessonTitleOff:hover {
  color: #06e;
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
  background-color: #f6f6ef;
  width: inherit;
  display: flex;
  flex-direction: row;
}

footer {
  background: linear-gradient(0deg, #00838f, #f6f6ef);
  height: {{.LayFooterHeight}}px;
}

.helpButtonBox, .navButtonBox {
  display: flex;
  flex-direction: row;
  justify-content: center;
  align-items: center;
  min-height: 3em;
  min-width: 2em;
  cursor: pointer;
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
  font-size: larger;
  font-weight: bold;
  color: #ff6e40;;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  min-height: 2em;
}

.activeLessonName {
  color: #ff6e40;;
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
  background: #00838f;
  color: white;
  transition: height 0.2s;
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
  transition: width 0.2s;
}

.proseActual {
  /* top right bottom left */
  padding: 0 1em 0 1em;
}

.lessonPrevClickerRow, .lessonNextClickerRow {
  height: 100%;
  cursor: pointer;
  display: flex;
  flex-basis: 45%;
  flex-direction: row;
  flex-wrap: nowrap;
  align-items: center;
}
.lessonPrevClickerRow {
  justify-content: flex-end;
}
.lessonNextClickerRow {
}

.lessonPointer {
  /* top right bottom left */
  padding: 0 1em 0 1em;
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
.lessonNextTitle {
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
  color: #06e;
}

.codeBlockSpacer {
  height: 100%;
  width: 5px;
}

pre.codeblockBody {
  font-family: "Lucida Console", Monaco, monospace;
  font-size: 0.9em;
  color: #33ff66;
  /* color: orange; */
  background-color: black;
  /* top rig bot lef */
  padding: 10px 20px 0px 20px;
  margin: 0px 0px 0px 20px;
  border: 0px;
  overflow-x: auto;
  max-width: calc({{.LayBodyMediumWidth}}px - 20px)
}

.codeBlockControl {
  /* font-family: "Courier New", Courier, monospace; */
  font-family: "Lucida Console", Monaco, monospace;
  font-size: 1.0em;
  /* font-weight: bold; */
  /* font-style: oblique; */
  margin: 20px 10px 12px 20px;
  padding: 0px;
}

.proseblock {
  /* font-size: 1.2em; */
  /* top rig bot lef */
  padding: 10px 20px 0px 0px;
}

.oneLesson {
  display: none;
  padding: 0 1em 0 1em;
  width: inherit;
}

.navBurger {
  display: inline-block;
  cursor: pointer;
}
.burgBar1, .burgBar2, .burgBar3 {
  background-color: #ff6e40;
  width: 14px;
  height: 2px;
  /* background-color: #333; */
  /* top rig bot lef */
  margin: 2px 0 2px 2px;
  transition: 0.2s;
}
.burgIsAnX .burgBar1 {
  -webkit-transform: translate(-3px, 0px) rotate(-45deg);
  transform: translate(-3px, 0px) rotate(-45deg);
}
.burgIsAnX .burgBar2 {
}
.burgIsAnX .burgBar3 {
  -webkit-transform: translate(-3px, 0px) rotate(45deg);
  transform: translate(-3px, 0px) rotate(45deg);
}
`

const jsInHeader = `
function getElByClass(n) {
  return document.getElementsByClassName(n)[0];
}

function getDataId(el) {
  return el.getAttribute("data-id");
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
    hideABox(theLeftBox);
    hideABox(theRightBox);
  }
  var showBoxes = function() {
    showABox(theLeftBox);
    showABox(theRightBox);
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
    theRightBox.right = '0px'
    showBoxes()
    showBurger()
  }
  var showMedium = function() {
    showNarrow()
    squeezeCenter()
  }
  var showWide = function() {
    theRightBox.right = 'calc(100vw - (' + bodyWideWidth + ' + 12px))';
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
  var handleWidthChange = function(discard) {
    if (mqWide.matches) {
      theBody.width = bodyWideWidth;
      theHelpBox.left = navBoxWidth;
      theHelpBox.right =
        'calc(100vw - ' + bodyWideWidth + ' + ' + navBoxWidth + ' - 12px)';
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
    theLeftBox = getElByClass('navLeftBox').style;
    theRightBox = getElByClass('navRightBox').style;
    theLeftSpacer = getElByClass('navLeftSpacer').style;
    theRightSpacer = getElByClass('navRightSpacer').style;
    theBody = document.getElementById('body').style;
    theProseColumn = getElByClass('proseColumn').style;

    mqWide = window.matchMedia('(min-width: ' + bodyWideWidth + ')');
    mqMedium = window.matchMedia(
        '(min-width: ' + bodyMediumWidth
        + ') and (max-width: ' + bodyWideWidth + ')');
    mqWide.addListener(handleWidthChange);
    mqMedium.addListener(handleWidthChange)
    handleWidthChange('whatever');
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
  }
  var showIt = function() {
    box.height = 'calc(100vh - ({{.LayFooterHeight}}px + {{.LayHeaderHeight}}px))';
    box.overflow = 'auto';
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
    var commandBlockDiv = b.parentNode.parentNode;
    // Fragile, but brief!
    var codeBody = commandBlockDiv.childNodes[5].firstChild;
    attemptCopyToBuffer(codeBody.textContent)
    var blockId = getDataId(commandBlockDiv);
    var fileId = getDataId(commandBlockDiv.parentNode);
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

var lessonMgr = new function() {
  var activeIndex = -1;
  var coursePaths = null;
  var theLessonName = null;
  var thePrevName = null;
  var theNextName = null;

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

  this.assureActiveLesson = function(index) {
    if (activeIndex == index) {
      return
    }
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

    e = getNavLesson(nextIndex(index))
    path = e.getAttribute('data-path')
    theNextName[0].innerHTML = path;
    theNextName[1].innerHTML = path;

    e = getNavLesson(prevIndex(index))
    path = e.getAttribute('data-path')
    thePrevName[0].innerHTML = path;
    thePrevName[1].innerHTML = path;

    e = getNavLesson(index)
    e.className = 'navLessonTitleOn'

    path = e.getAttribute('data-path')
    theLessonName.innerHTML = path

    activeIndex = index
    if (history.pushState) {
      window.history.pushState("not using data yet", "someTitle", "/" + path);
    } else {
      document.location.href = path;
    }
    smoothScroll()
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
  }
}

function onLoad() {
  help.initialize();
  nav.initialize({{.LessonCount}} > 1);
  var coursePaths = [
    [],
    [0], [0], [0], [0],
    [1], [1], [1], [1], [1], [1], [1], [1], [1], [1], [1],
    [], []
  ];
  lessonMgr.initialize({{.CoursePaths}});
  lessonMgr.assureActiveLesson({{.InitialLesson}});
  window.addEventListener('keydown', function (event) {
    if (event.defaultPrevented) {
      return;
    }
    switch (event.key) {
      case 'ArrowLeft':
        lessonMgr.goPrev();
        break;
      case 'ArrowRight':
        lessonMgr.goNext();
        break;
      default:
    }
  }, true);
}
`
