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
// Not using react, angular2, polymer, etc. because
// want to keep it simple and shippable as a single binary.
type WebApp struct {
	sessId TypeSessId
	host   string
	tut    model.Tutorial
	tmpl   *template.Template
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

func (wa *WebApp) LayMainWidth() int         { return 950 }
func (wa *WebApp) LayNavWidth() int          { return 250 }
func (wa *WebApp) LayLessonWidth() int       { return wa.LayMainWidth() - wa.LayNavWidth() }
func (wa *WebApp) LayInstructionsWidth() int { return wa.LayLessonWidth() - 30 }

func (wa *WebApp) LayNavTopBotPad() int      { return 7 }
func (wa *WebApp) LayNavLeftPad() int        { return 20 }
func (wa *WebApp) LayNavWidthPlusDelta() int { return wa.LayNavWidth() + delta }

func (wa *WebApp) LayTitleHeight() int          { return 30 }
func (wa *WebApp) LayTitleHeightPlusDelta() int { return wa.LayTitleHeight() + delta }

func (wa *WebApp) LessonCount() int {
	c := model.NewTutorialLessonCounter()
	wa.tut.Accept(c)
	return c.Count()
}

func (wa *WebApp) Render(w io.Writer) error {
	return wa.tmpl.ExecuteTemplate(w, tmplNameWebApp, wa)
}

func NewWebApp(sessId TypeSessId, host string, tut model.Tutorial) *WebApp {
	return &WebApp{sessId, host, tut, makeParsedTemplate(tut)}
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

func makeAppTemplate(leftNavBody string) string {
	return `
{{define "` + tmplNameWebApp + `"}}
<html>
<head>
<style type="text/css">` + headerCss + `
</style>
<script type="text/javascript">` + headerJs + `
</script>
</head>
<body onload="onLoad()">
<div class='main'>
` + instructionsHtml + `
  <div class='titleBar'>
    <span class='titleNav' onclick='assureActiveLesson(0)'> {{ .TrimName }} </span>
    <button class='navToggle' type='button' onclick='toggleLeftNav()'
        id='navToggle' >&lt;</button>
    <button type='button' onclick="toggleByClass('instructions')">?</button>
    <span class='activeLessonName'>lesson name</span>
    </span>
  </div>
  <div class='leftNav'>
` + leftNavBody + `
  </div>
  <div class='lessonList'>
    {{ template "` + tmplNameLessonList + `" .Lessons }}
  </div>
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
  <div class="commandBlock" data-id="{{$i}}">
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
<h3 id="control" class="control">
  <span class="blockButton" onclick="onRunBlockClick(event)">
     {{.Name}}
  </span>
  <span class="spacer"> &nbsp; </span>
</h3>
<pre class="codeblock">
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

const headerCss = `
body {
  font-family: "Veranda", Veranda, sans-serif;
  /* background-color: antiquewhite; */
  background-color: white;
  margin: 0;
  padding: 0;
}

div.main {
  position: relative;
  width: {{.LayMainWidth}}px;
  min-width: {{.LayMainWidth}}px;
}

div.titleBar {
  position: fixed;
  z-index: 100;
  top: 0;
  width: 100%;
  background-color: #ddd;
  height: {{.LayTitleHeight}}px;
  /* top rig bot lef */
}

span.titleNav {
  display: inline-block;
  width: {{.LayNavWidth}}px;
  min-width: {{.LayNavWidth}}px;
  padding: 4px 0px 4px {{.LayNavLeftPad}}px;
}

span.activeLessonName {
  padding: 4px 0px 4px 6px;
}

.navToggle {
  /* float: left; */
  /* position: fixed; */
  /* left: {{.LayNavWidthPlusDelta}}px; */
}

div.leftNav {
  position: fixed;
  z-index: 100;
  top: {{.LayTitleHeightPlusDelta}}px;
  left: 0;
  /* top rig bot lef */
  padding: 20px 0px 4px {{.LayNavLeftPad}}px;
}

div.lessonList {
  position: absolute;
  top: {{.LayTitleHeightPlusDelta}}px;
  left: {{.LayNavWidthPlusDelta}}px;
  width: {{.LayLessonWidth}}px;
  /* top rig bot lef */
  padding: 0px 0px 4px {{.LayNavLeftPad}}px;
}

div.instructions {
  position: fixed;
  display: none;
  width: {{.LayInstructionsWidth}}px;
  z-index: 100;
  margin: auto;
  background-color: #cccccc;
  border: 5px solid #eeeeee;
  top: {{.LayTitleHeightPlusDelta}}px;
  right: {{.LayTitleHeightPlusDelta}}px;
  /* top rig bot lef */
  padding: 10px 20px 20px 20px;
}

div.navCourseTitle {
  /* top rig bot lef */
  padding: 0px 0px 0px 0px;
}

div.navCourseTitle:hover {
  color: #06e;
}

div.navCourseContent {
  /* top rig bot lef */
  padding: {{.LayNavTopBotPad}}px 0px 0px 0px;
}

div.navLessonTitleOn {
  background-color: #ddd;
}

div.navLessonTitleOff {
}

div.navLessonTitleOff:hover {
  color: #06e;
}

div.navItemTop {
  /* top rig bot lef */
  padding: {{.LayNavTopBotPad}}px 0px {{.LayNavTopBotPad}}px 4px;
}

div.navItemBox {
  /* top rig bot lef */
  padding: {{.LayNavTopBotPad}}px 0px {{.LayNavTopBotPad}}px {{.LayNavLeftPad}}px;
}

div.commandBlock {
  margin: 0px;
  border: 0px;
  padding: 0px;
}

.blockButton {
  height: 100%;
  cursor: pointer;
}

.blockButton:hover {
  color: #06e;
}

.spacer {
  height: 100%;
  width: 5px;
}

div.proseblock {
  font-size: 1.2em;
  /* top rig bot lef */
  padding: 10px 20px 0px 0px;
}

div.oneLesson {
  display: none;
  padding: 2px 2px 2px 2px;
}

.control {
  /* font-family: "Courier New", Courier, monospace; */
  font-family: "Lucida Console", Monaco, monospace;
  font-size: 1.0em;
  /* font-weight: bold; */
  /* font-style: oblique; */
  margin: 20px 10px 12px 20px;
  padding: 0px;
}

pre.codeblock {
  font-family: "Lucida Console", Monaco, monospace;
  font-size: 0.9em;
  color: #33ff66;
  /* color: orange; */
  background-color: black;
  /* top rig bot lef */
  padding: 10px 20px 0px 20px;
  margin: 0px 0px 0px 20px;
  border: 0px;
}

.didit {
  display: inline-block;
  width: 24px;
  height: 15px;
  background-repeat: no-repeat;
  background-size: contain;
  background-image: url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABgAAAAWCAMAAADto6y6AAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAAgY0hSTQAAeiYAAICEAAD6AAAAgOgAAHUwAADqYAAAOpgAABdwnLpRPAAAAQtQTFRFAAAAAH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//AH//////BQzC2AAAAFd0Uk5TAAADLy4QZVEHKp8FAUnHbeJ3BAh68IYGC4f4nQyM/LkYCYnXf/rvAm/2/oFY7rcTPuHkOCEky3YjlW4Pqbww0MVTfUZA96p061Xs3mz1e4P70R2aHJYf2KM0AgAAAAFiS0dEWO21xI4AAAAJcEhZcwAAEysAABMrAbkohUIAAADTSURBVCjPbdDZUsJAEAXQXAgJIUDCogHBkbhFEIgCsqmo4MImgij9/39iUT4Qkp63OV0zfbsliTkIhWWOEVHUKOdaTNER9HgiaYQY1xUzlWY8kz04tBjP5Y8KRc6PxUmJcftUnMkIFGCdX1yqjDtX5cp1MChQrVHd3Xn8/y1wc0uNpuejZmt7Ae7aJDreBt1e3wVw/0D06HobYPD0/GI7Q0G10V4i4NV8e/8YE/V8KwImUxJEM82fFM78k4gW3MhfS1p9B3ckobgWBpiChJ/fjc//AJIfFr4X0swAAAAAJXRFWHRkYXRlOmNyZWF0ZQAyMDE2LTA3LTMwVDE0OjI3OjUxLTA3OjAwUzMirAAAACV0RVh0ZGF0ZTptb2RpZnkAMjAxNi0wNy0zMFQxNDoyNzo0NC0wNzowMLz8tSkAAAAZdEVYdFNvZnR3YXJlAHd3dy5pbmtzY2FwZS5vcmeb7jwaAAAAFXRFWHRUaXRsZQBibHVlIENoZWNrIG1hcmsiA8jIAAAAAElFTkSuQmCC);
}
`

const headerJs = `
function getElByClass(name) {
  var elements = document.getElementsByClassName(name);
  return elements[0];
}
function openLeftNav(e) {
  e.innerHTML = '&lt;'
  var ln = getElByClass('leftNav');
  ln.style.display = 'block';
  var list = getElByClass('lessonList');
  list.style.left = '{{.LayNavWidth}}px';
}
function closeLeftNav(e) {
  e.innerHTML = '&gt;'
  var ln = getElByClass('leftNav');
  ln.style.display = 'none';
  var list = getElByClass('lessonList');
  list.style.left = '0';
}
function assureLeftNavOpen() {
  var e = document.getElementById('navToggle')
  if (e.innerHTML == '&gt;') {
    openLeftNav(e)
  }
}
function assureLeftNavClosed() {
  var e = document.getElementById('navToggle')
  if (e.innerHTML == '&lt;') {
    closeLeftNav(e)
  }
}
function toggleLeftNav() {
  var e = document.getElementById('navToggle')
  if (e.innerHTML == '&gt;') {
    openLeftNav(e)
  } else {
    closeLeftNav(e)
  }
}
function toggleByClass(name) {
  dToggle(getElByClass(name))
}
function toggleNC(index) {
  dToggle(document.getElementById('NC' + index.toString()))
}
function dToggle(e) {
  e.style.display = (e.style.display == 'block') ? 'none' : 'block'
}
var requestRunning = false
var activeLesson = -1
function assureNoActiveLesson() {
  if (activeLesson == -1) {
    return
  }
  var index = activeLesson
  getElByClass('activeLessonName').innerHTML = ''
  // hide lesson body.
  var e = document.getElementById('BL' + index.toString())
  e.style.display = 'none'
  // hide lesson nav.
  var e = document.getElementById('NL' + index.toString())
  e.className = 'navLessonTitleOff'
  activeLesson = -1
}
function assureActiveLesson(index) {
  if (activeLesson == index) {
    return
  }
  if (activeLesson != -1) {
    assureNoActiveLesson()
  }
  // show lesson body.
  var e = document.getElementById('BL' + index.toString())
  e.style.display = 'block'

  // show lesson nav.
  var e = document.getElementById('NL' + index.toString())
  e.className = 'navLessonTitleOn'

  path = e.getAttribute('data-path')
  getElByClass('activeLessonName').innerHTML = path
  activeLesson = index
}
function onLoad() {
  if ({{.LessonCount}} > 1) {
    assureLeftNavOpen()
  } else {
    assureLeftNavClosed()
  }
  assureActiveLesson(0)
}
function getDataId(el) {
  return el.getAttribute("data-id");
}
function addCheck(el) {
  var t = 'span';
  var c = document.createElement(t);
  c.setAttribute('class', 'didit');
  el.appendChild(c);
}
function attemptCopyToBuffer(text) {
  // https://stackoverflow.com/questions/400212
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
function onRunBlockClick(event) {
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
  xhr.open('GET', '/runblock?fid=' + fileId + '&bid=' + blockId + '&sid={{.SessId}}', true);
  xhr.send();
}
`
const instructionsHtml = `
<div class='instructions' onclick="toggleByClass('instructions')">
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
  TMP_DIR=$(mktemp -d)
  GOPATH=$TMP_DIR go install github.com/monopole/mdrip
</pre>
</li>
<li>Run tmux:
<pre>
  tmux
</pre>
</li>
<li>In some non-tmux shell, run this service:
<pre>
  $TMP_DIR/bin/mdrip \
      --alsologtostderr --v 0 \
      --stderrthreshold INFO \
      --mode tmux ws://{{.Host}}/ws?id={{.SessId}}
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
</div>
`
