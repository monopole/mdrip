package webapp

import (
	"crypto/rand"
	"encoding/gob"
	"html/template"
	"io"

	"bytes"

	"fmt"

	"github.com/gorilla/sessions"
	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
	"github.com/monopole/mdrip/program"
)

// TypeSessID represents a session ID.
type TypeSessID string

const forRegistration = TypeSessID("arbitrary")

func init() {
	gob.Register(forRegistration)
}

// SessionData holds session state data, presumably associated with a cookie.
type SessionData struct {
	// The session ID.
	SessID TypeSessID
	// Is the header showing?
	IsHeaderOn bool
	// Is the nav showing?
	IsNavOn bool
	// The active lesson.
	LessonIndex int
	// The active block.
	BlockIndex int
}

// These must all be unique, and preferably short.
// They are used as URL query param and cookie field names.
const (
	// KeySessID is the param name for session ID.
	KeySessID = "sid"
	// KeyIsHeaderOn is the param name for is the header on boolean.
	KeyIsHeaderOn = "hed"
	// KeyIsNavOn is the param name for the is the nav on boolean.
	KeyIsNavOn = "nav"
	// KeyLessonIndex is the param name for the lesson index.
	KeyLessonIndex = "lix"
	// KeyBlockIndex is the param name for the block index.
	KeyBlockIndex = "bix"
)

func makeSessionID() TypeSessID {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return TypeSessID(fmt.Sprintf("%x", b))
}

// AssureSessionData tries to recover session data, saving defaults for missing data.
func AssureSessionData(s *sessions.Session) *SessionData {
	r := &SessionData{}
	var ok bool
	r.SessID, ok = s.Values[KeySessID].(TypeSessID)
	if !ok {
		r.SessID = makeSessionID()
		s.Values[KeySessID] = r.SessID
	}
	r.IsHeaderOn, ok = s.Values[KeyIsHeaderOn].(bool)
	if !ok {
		r.IsHeaderOn = true
		s.Values[KeyIsHeaderOn] = r.IsHeaderOn
	}
	r.IsNavOn, ok = s.Values[KeyIsNavOn].(bool)
	if !ok {
		r.IsNavOn = false
		s.Values[KeyIsNavOn] = r.IsNavOn
	}
	r.LessonIndex, ok = s.Values[KeyLessonIndex].(int)
	if !ok {
		r.LessonIndex = 0
		s.Values[KeyLessonIndex] = r.LessonIndex
	}
	r.BlockIndex, ok = s.Values[KeyBlockIndex].(int)
	if !ok {
		r.BlockIndex = 0
		s.Values[KeyBlockIndex] = r.BlockIndex
	}
	return r
}

// WebApp presents a tutorial to a web browser.
type WebApp struct {
	sessionData *SessionData
	host        string
	tut         model.Tutorial
	ds          *base.DataSource
	tmpl        *template.Template
	rawLessons  []*program.LessonPgm
	title       string
	lessonPath  []int
	coursePaths [][]int
}

// NewWebApp makes a new web app.
func NewWebApp(
	sessionData *SessionData, host string,
	tut model.Tutorial, ds *base.DataSource, lp []int, cp [][]int) *WebApp {
	v := program.NewLessonPgmExtractor(base.WildCardLabel)
	tut.Accept(v)
	title := v.FirstTitle()
	if len(title) > maxTitleLength {
		title = title[maxTitleLength-3:] + "..."
	}
	return &WebApp{
		sessionData, host, tut, ds, makeParsedTemplate(tut),
		v.Lessons(), title, lp, cp}
}

// SessID is the id of the session returned
func (wa *WebApp) SessID() TypeSessID { return wa.sessionData.SessID }

// Host is the webapp's host.
func (wa *WebApp) Host() string { return wa.host }

// Lessons is the list of lessons known to the webapp.
func (wa *WebApp) Lessons() []*program.LessonPgm {
	return wa.rawLessons
}

// DataSourceName is the source of the data.
func (wa *WebApp) DataSourceName() string {
	return wa.ds.Display()
}

// DataSourceLink lets the user find the original data.
func (wa *WebApp) DataSourceLink() template.URL {
	return template.URL(wa.ds.Href())
}

// DocTitle is the name of the document or web page.
func (wa *WebApp) DocTitle() string {
	return wa.title
}

const (
	// arbitrary
	maxTitleLength = len("gh:kubernetes/website/reference") + 10
)

// InitialHeaderOn is should the header be on?
func (wa *WebApp) InitialHeaderOn() bool {
	return wa.sessionData.IsHeaderOn
}

// InitialNavOn is should the nav be on?
func (wa *WebApp) InitialNavOn() bool {
	return wa.sessionData.IsNavOn
}

// InitialBlock is where the user should start.
func (wa *WebApp) InitialBlock() int {
	return wa.sessionData.BlockIndex
}

// InitialLesson is where the user should start.
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

// CoursePaths emits a javascript 2D array holding lesson paths.
// The length equals the number of lessons.
// Each entry should contain an array of course indices
// that should be active when the lesson is active.
func (wa *WebApp) CoursePaths() [][]int {
	return wa.coursePaths
}

// Named color specifications to use on the web page.
const (
	blue700         = "#1976D2"
	blue500         = "#2196F3"
	blue200         = "#90CAF9"
	deepOrange500   = "#FF5722"
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

// ColorBackground is just that.
func (wa *WebApp) ColorBackground() string { return "white" }

// ColorHelpBackground is just that.
func (wa *WebApp) ColorHelpBackground() string { return whiteIsh }

// ColorHeader is just that.
func (wa *WebApp) ColorHeader() string { return blue700 }

// ColorCodeBlockText is just that.
func (wa *WebApp) ColorCodeBlockText() string { return greenA400 }

// ColorCodeBlockBackground is just that.
func (wa *WebApp) ColorCodeBlockBackground() string { return "black" }

// ColorNavBackground is just that.
func (wa *WebApp) ColorNavBackground() string { return blue200 }

// ColorNavText is just that.
func (wa *WebApp) ColorNavText() string { return "black" }

// ColorNavSelected is just that.
func (wa *WebApp) ColorNavSelected() string { return wa.ColorBackground() }

// ColorHover is just that.
func (wa *WebApp) ColorHover() string { return deepOrange500 }

// ColorCodeHover is just that.
func (wa *WebApp) ColorCodeHover() string { return deepOrange700 }

// ColorControls is just that.
func (wa *WebApp) ColorControls() string { return greenA200 }

// ColorTitle is just that.
func (wa *WebApp) ColorTitle() string { return wa.ColorControls() }

// TransitionSpeedMs is speed of css transitions in milliseconds.
func (wa *WebApp) TransitionSpeedMs() int { return 250 }

// LayBodyWideWidth is the min body width of "wide" mode.
func (wa *WebApp) LayBodyWideWidth() int { return 1200 }

// LayBodyMediumWidth is the min body width of medium mode.
// Small mode (presumably phones) is anything thinner.
func (wa *WebApp) LayBodyMediumWidth() int { return 800 }

// LayMinHeaderWidth is just that.
func (wa *WebApp) LayMinHeaderWidth() int { return 400 }

// LayNavBoxWidth is just that.
func (wa *WebApp) LayNavBoxWidth() int { return 210 }

// LayHeaderHeight is just that.
func (wa *WebApp) LayHeaderHeight() int { return 120 }

// LayFooterHeight is just that.
func (wa *WebApp) LayFooterHeight() int { return 50 }

// LayMinimizedHeaderHeight is just that.
func (wa *WebApp) LayMinimizedHeaderHeight() int { return wa.LayFooterHeight() }

// LayNavTopBotPad is just that.
func (wa *WebApp) LayNavTopBotPad() int { return 7 }

// LayNavLeftPad is just that.
func (wa *WebApp) LayNavLeftPad() int { return 20 }

// KeyLessonIndex delivers the corresponding const to a template.
func (wa *WebApp) KeyLessonIndex() string { return KeyLessonIndex }

// KeyBlockIndex delivers the corresponding const to a template.
func (wa *WebApp) KeyBlockIndex() string { return KeyBlockIndex }

// KeyIsHeaderOn delivers the corresponding const to a template.
func (wa *WebApp) KeyIsHeaderOn() string { return KeyIsHeaderOn }

// KeyIsNavOn delivers the corresponding const to a template.
func (wa *WebApp) KeyIsNavOn() string { return KeyIsNavOn }

// KeySessID delivers the corresponding const to a template.
func (wa *WebApp) KeySessID() string { return KeySessID }

// LessonCount is just that.
func (wa *WebApp) LessonCount() int {
	c := model.NewTutorialLessonCounter()
	wa.tut.Accept(c)
	return c.Count()
}

// Render writes a web page to the given writer.
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
    <div class='navButtonBox'>
      <div class='navBurger'>
        <div class='burgBar1'></div>
        <div class='burgBar2'></div>
        <div class='burgBar3'></div>
      </div>
    </div>
    <div class='headerColumn'>
      <a target='_blank' href='{{.DataSourceLink}}'>
        <title id='title'> {{.DocTitle}} </title>
      </a>
      <div class='activeLessonName'> Droplet Formation Rates </div>
      ` + htmlLessonNavRow + `
    </div>
    <div class='navButtonBox'> &nbsp; </div>
  </header>

  <div class='navLeftBox navLeftBoxShadow' tabindex='-1'>
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
	tmplNameLesson = "oneLesson"
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
<div class='proseblock'> {{.HTMLProse}} </div>
{{if .Code}}
<div class='codeBox' data-id='{{.ID}}'>
  <div class='codeBlockControl'>
    <span class='codePrompt'> &nbsp;&gt;&nbsp; </span>
    <span class='codeBlockButton' onclick='codeBlockController.setAndRun({{.ID}})'>
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
  <div class='lessonPrevClickerRow'>
    <div class='lessonPrevTitle'> quantum flux </div>
    <div class='lessonPrevPointer'> &lt; </div>
  </div>
  <div class='helpButtonBox'> ? </div>
  <div class='lessonNextClickerRow'>
    <div class='lessonNextPointer'> &gt; </div>
    <div class='lessonNextTitle'> magnetic flux  </div>
  </div>
</div>
`

const htmlHelp = `
<p>
Snapshot of markdown from
<a target='_blank' href='{{.DataSourceLink}}'><code>{{.DataSourceName}}</code></a>.

<h3>Keys</h3>
<p>
<table>
  <tr>
     <td class='kind'> help </td>
     <td> ? &nbsp; / </td>
  </tr>
  <tr>
    <td class='kind'> activate (previous, next) code block </td>
    <td> w, s &nbsp; j, k </td>
  </tr>
  <tr>
    <td class='kind'> scroll to active code block </td>
    <td> x </td>
  </tr>
  <tr>
     <td class='kind'> copy/execute activated block </td>
     <td> &crarr; </td>
  </tr>
  <tr>
    <td class='kind'> (previous, next) lesson </td>
    <td> a, d &nbsp; h, l &nbsp; &larr;, &rarr; </td>
  </tr>
  <tr>
    <td class='kind'> minimize header </td>
    <td> - </td>
  </tr>
  <tr>
    <td class='kind'> nav sidebar </td>
    <td> n </td>
  </tr>
  <tr>
    <td class='kind'> monkey </td>
    <td> ! </td>
  </tr>
</table>
</p>

<h3> Serve locally with tmux for no-mouse execution</h3>

<p>
Serve the content locally with
<code><a target="_blank"
href="https://github.com/monopole/mdrip">mdrip</a></code>:
<pre>
  GOBIN=$TMPDIR go install github.com/monopole/mdrip
  $TMPDIR/mdrip --port 8001 --mode demo {{.DataSourceName}}
</pre>
and run <a target="_blank"
href="https://github.com/tmux/tmux/wiki">tmux</a>:
<pre>
  tmux
</pre>
Then, at <a target="_blank"
href="http://localhost:8001">localhost:8001</a>,
whatever action copies a code block (&crarr; or mouse click)
also pastes the block to the active tmux pane for
immediate execution.
</p>

<h3> Remote server tmux </h3>
<p> <em>Proof of concept
for using tmux over a websocket to remote servers.
Currently lacks the session mgmt necessary
to reliably work with LB traffic.
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
    ws://{{.Host}}/_/ws?{{.KeySessID}}={{.SessID}}
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

td.kind {
  text-align: right;
  padding-right: 2em;
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
  /*            x   y blur spread color             x   y blur spread color */
  box-shadow: 0px 2px  1px    0px rgba(0,0,0,.3), 2px 0px 1px 0px rgba(0,0,0,.3);
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

function myAddListener(n, f) {
  var btn = document.getElementsByClassName(n);
  for (var i = 0; i < btn.length; i++) {
    btn[i].addEventListener('click', f);
  }
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
  this.IsVisible = function() {
    return isVisible()
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
    saveSession();
  }
  this.initialize = function() {
    styleHeader = document.getElementById('header').style;
    styleHeadSpacer = getElByClass('headSpacer').style;
    styleTitle = document.getElementById('title').style;
    styleLessonNavRow = document.getElementsByClassName('lessonNavRow')[0].style;
  }
  this.reset = function() {
    if ({{.InitialHeaderOn}}) {
      showIt();
    } else {
      hideIt();
    }
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
  this.IsVisible = function() {
    return isVisible()
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
    saveSession();
  }
  var keyHandler = function(event) {
    switch (event.key) {
      case 'w':
      case 'k':
      case 'ArrowUp':
        event.preventDefault();
        lessonController.goPrev();
        break;
      case 'j':
      case 's':
      case 'ArrowDown':
        event.preventDefault();
        lessonController.goNext();
        break;
      default:
    }
  }
  this.initialize = function() {
    elBurger = getElByClass('navBurger');
    elNavLeft = getElByClass('navLeftBox');
    elNavLeft.addEventListener('keydown', keyHandler, false);
    elNavLeft.onmouseover = function() { elNavLeft.focus(); }
    elNavLeft.onmouseout = function() { elNavLeft.blur(); }
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
    mqMedium.addListener(this.handleWidthChange);
    myAddListener('navButtonBox', this.toggle);
    this.handleWidthChange('whatever');
  }
  this.reset = function() {
    if ({{.InitialNavOn}}) {
      showIt();
    } else {
      hideIt();
    }
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
    myAddListener('helpButtonBox', this.toggle);
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
    } catch (err) {
      console.log('Oops, unable to copy');
    }
    document.body.removeChild(tA);
  }
  this.goPrev = function() {
    this.deActivateCurrent();
    if (cbIndex < 0) {
      // Already -1
      return;
    }
    cbIndex--;
    activateCurrent();
    saveSession();
  }
  this.goCurrent = function() {
    activateCurrent();
  }
  this.goNext = function() {
    this.deActivateCurrent();
    if (cbIndex >= blocks.length) {
      // Do nothing, not even modulo wrap.
      return;
    }
    cbIndex++;
    activateCurrent();
    saveSession();
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
         console.log('Counting problem')
       }
    }
  }
  // For monkeyController
  this.toggle = function() {
    this.setCurrent(randomInt(blocks.length));
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
  this.getActiveBlock = function() {
    return cbIndex;
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
        'POST',
        '/_/runblock'
            + '?{{.KeyLessonIndex}}=' + fileId
            + '&{{.KeyBlockIndex}}=' + cbIndex
            + '&{{.KeySessID}}={{.SessID}}',
        true);
    xhr.send();
  }
  this.initialize = function() {
    requestRunning = false;
    cbIndex = -1;
  }
  this.reset = function() {
    this.setCurrent({{.InitialBlock}});
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

  var goodIndex = function(i) {
    return i > -1 && i < coursePaths.length
  }
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
    if (!goodIndex(activeIndex)) {
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
    if (!goodIndex(lesson)) {
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
    if (!goodIndex(index)) {
      return
    }
    var prevState = bodyController.isVertScrollBarVisible();
    if (goodIndex(activeIndex)) {
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
  this.goPrev = function() {
    if (activeIndex < 0) {
      // Already -1
      return;
    }
    this.assureActiveLesson(activeIndex - 1)
    saveSession();
  }
  this.goNext = function() {
    if (activeIndex >= coursePaths.length) {
      // Do nothing, not even modulo wrap.
      return;
    }
    this.assureActiveLesson(activeIndex + 1)
    saveSession();
  }
  this.getActiveLesson = function() {
    return activeIndex
  }
  this.initialize = function(cp) {
    coursePaths = cp;
    activeIndex = -1;
    myAddListener('lessonPrevClickerRow', this.goPrev.bind(this));
    myAddListener('lessonNextClickerRow', this.goNext.bind(this));
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

var suppressSessionSave = false

function saveSession() {
  if (suppressSessionSave) {
    return
  }
  var xhr = new XMLHttpRequest();
  xhr.onreadystatechange = function() {
    if (xhr.readyState == XMLHttpRequest.DONE) {
      // console.log('saved session')
    }
  };
  xhr.open(
      'POST',
      '/_/s'
          + '?{{.KeyIsHeaderOn}}=' + headerController.IsVisible()
          + '&{{.KeyIsNavOn}}=' + navController.IsVisible()
          + '&{{.KeyLessonIndex}}=' + lessonController.getActiveLesson()
          + '&{{.KeyBlockIndex}}=' + codeBlockController.getActiveBlock(),
      true);
  xhr.send();
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
      suppressSessionSave = false;
      return;
    }
    suppressSessionSave = true;
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
          lessonController, navController, codeBlockController));
  monkeyController.reset();
  window.addEventListener('keydown', function (event) {
    if (event.defaultPrevented) {
      return;
    }
    switch (event.key) {
      case 'Enter':
      case 'r':
        codeBlockController.runCurrent();
        codeBlockController.goNext();
        break;
      case 'x':
        codeBlockController.goCurrent();
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
      case 'w':
      case 'k':
        codeBlockController.goPrev();
        break;
      case 'j':
      case 's':
        codeBlockController.goNext();
        break;
      case 'a':
      case 'h':
      case 'ArrowLeft':
        lessonController.goPrev();
        break;
      case 'd':
      case 'l':
      case 'ArrowRight':
        lessonController.goNext();
        break;
      case '!':
        monkeyController.toggle();
        break;
      default:
    }
  }, false);
  window.setTimeout(codeBlockController.goCurrent, 700);
}
`
