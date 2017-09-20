package webapp

import (
	"html/template"
	"io"

	"bytes"
	"github.com/monopole/mdrip/model"
	"github.com/monopole/mdrip/tutorial"
)

// A tutorial and associated info for rendering.
type App struct {
	sessId model.TypeSessId
	host   string
	tut    tutorial.Tutorial
	tmpl   *template.Template
}

func (app *App) SessId() model.TypeSessId    { return app.sessId }
func (app *App) Host() string                { return app.host }
func (app *App) Tutorial() tutorial.Tutorial { return app.tut }
func (app *App) Program() *model.Program {
	return tutorial.NewProgramFromTutorial(model.AnyLabel, app.tut)
}

func (app *App) Lessons() []*tutorial.Lesson {
	v := tutorial.NewLessonExtractor()
	app.tut.Accept(v)
	return v.Lessons()
}

func (app *App) AppName() string {
	return app.host
}

func (app *App) Render(w io.Writer) error {
	return app.tmpl.ExecuteTemplate(w, tmplNameWebApp, app)
}

func makeLeftNavBody(tut tutorial.Tutorial) string {
	var b bytes.Buffer
	v := NewTutorialNavPrinter(&b)
	tut.Accept(v)
	return b.String()
}

func makeMasterTemplate(tut tutorial.Tutorial) *template.Template {
	return template.Must(
		template.New("main").Parse(
			tmplBodyLesson +
				tmplBodyCommandBlock +
				tmplBodyLessonList +
				makeAppTemplate(
					makeLeftNavBody(tut))))
}

func NewWebApp(sessId model.TypeSessId, host string, tut tutorial.Tutorial) *App {
	v := tutorial.NewLessonExtractor()
	tut.Accept(v)
	return &App{sessId, host, tut, makeMasterTemplate(tut)}
}

// The logic involved in building the leftnav is much less awkward
// in plain Go than in the Go template language, so slapping the nav
// into the template as a pre-built string.
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
  <div class='leftNav'>
    <div
        class='overallTitleNav'
        onclick='assureActiveLesson(0)'>
      {{ .AppName }}
    </div>
` + leftNavBody + `
  </div>
  <div class='lessonList'>
    <div class='overallTitleLesson'>
      <button onclick='toggleLeftNav()' id='heyho' type='button'>&gt;</button>
      <span id='activeLessonName'></span>
    </div>
    <div >
      {{ template "` + tmplNameLessonList + `" .Lessons }}
    </div>
  </div>
</body>
</html>
{{end}}
`
}

const (
	tmplNameWebApp = "webApp"
)

const (
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
)

const (
	tmplNameLesson = "navlesson"
	tmplBodyLesson = `
{{define "` + tmplNameLesson + `"}}
{{range $i, $c := .Children}}
  <div class="commandBlock" data-id="{{$i}}">
  {{ template "` + tmplNameCommandBlock + `" $c }}
  </div>
{{end}}
{{end}}
`
)

const (
	tmplNameCommandBlock = "navcommandblock"
	tmplBodyCommandBlock = `
{{define "` + tmplNameCommandBlock + `"}}
<div class="proseblock"> {{.Prose}} </div>
<h3 id="control" class="control">
  <span class="blockButton" onclick="onRunBlockClick(event)">
     {{ .Name }}
  </span>
  <span class="spacer"> &nbsp; </span>
</h3>
<pre class="codeblock">
{{ .Code }}
</pre>
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
}

div.overallTitleNav {
  position: fixed;
  top: 0;
  height: 26px;
  padding: 4px 0px 4px 4px;
}

div.leftNav {
  position: fixed;
  z-index: 100;
  top: 26px;
  left: 0;
  /* width: 100px; */
  /* top rig bot lef */
  padding: 4px 0px 4px 4px;
}

div.overallTitleLesson {
  position: fixed;
  z-index: 100;
  padding: 0;
  background-color: #ddd;
  /* opacity: 0.0; */
  top: 0;
  height: 26px;
  left: 200px;
  width: 100%;
  padding: 4px 0px 4px 4px;
}

div.lessonList {
  position: absolute;
  top: 26px;
  left: 200px;
  width: 100%;
}

div.navCourseTitle {
}

div.navCourseTitle:hover {
  color: #06e;
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
  padding: 4px 0px 2px 4px;
}

div.navItemBox {
  /* top rig bot lef */
  padding: 2px 0px 2px 20px;
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
.topcorner {
  position: fixed;
  top: 6;
  right: 10;
  z-index: 200;
}
div.instructions {
  position: fixed;
  font-size: 0.7em;
  display: none;
  width: 480px;
  margin: auto;
  background-color: #cccccc;
  border: 5px solid #eeeeee;
  top: 23px;
  right: 0px;
  /* top rig bot lef */
  padding: 10px 20px 20px 20px;
}
`

const headerJs = `
function getElByClass(name) {
  var elements = document.getElementsByClassName(name);
  return elements[0];
}
function toggleLeftNav() {
  var ln = getElByClass('leftNav');
  var list = getElByClass('lessonList');
  var e = document.getElementById('heyho')
  if (e.innerHTML == '&gt;') {
    // Show the nav
    e.innerHTML = '&lt;'
    ln.style.display = 'block';
    list.style.left = '200px';
  } else {
    // Hide the nav
    e.innerHTML = '&gt;'
    ln.style.display = 'none';
    list.style.left = '0';
  }
}
function toggleByClass(name) {
  var e = getElByClass(name);
  e.style.display = (e.style.display == 'block') ? 'none' : 'block';
}
function toggleNC(index) {
  var e = document.getElementById('NC' + index.toString());
  e.style.display = (e.style.display == 'block') ? 'none' : 'block';
}
// blockUx, which may cause screen flicker, not needed if write is very fast.
var blockUx = false
var runButtons = []
var requestRunning = false
var activeLesson = -1
function assureNoActiveLesson() {
  if (activeLesson == -1) {
    return
  }
  var index = activeLesson
  document.getElementById('activeLessonName').innerHTML = ''
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
  document.getElementById('activeLessonName').innerHTML = path
  activeLesson = index
}
function onLoad() {
  if (blockUx) {
    runButtons = document.getElementsByTagName('input');
  }
  assureActiveLesson(0)
}
function getId(el) {
  return el.getAttribute("data-id");
}
function setRunButtonsDisabled(value) {
  for (var i = 0; i < runButtons.length; i++) {
    runButtons[i].disabled = value;
  }
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
  if (blockUx) {
    setRunButtonsDisabled(true)
  }
  var b = event.target;
  var commandBlockDiv = b.parentNode.parentNode;
  // Sorry about the fragility here :P
  var codeBody = commandBlockDiv.childNodes[5].firstChild;
  attemptCopyToBuffer(codeBody.textContent)
  var blockId = getId(commandBlockDiv);
  var fileId = getId(commandBlockDiv.parentNode);
  var oldColor = b.style.color;
  var oldValue = b.value;
  if (blockUx) {
     b.style.color = 'red';
     b.value = 'running...';
  }
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (xhttp.readyState == XMLHttpRequest.DONE) {
      if (blockUx) {
        b.style.color = oldColor;
        b.value = oldValue;
      }
      addCheck(b.parentNode)
      requestRunning = false;
      if (blockUx) {
        setRunButtonsDisabled(false);
      }
    }
  };
  xhttp.open('GET', '/runblock?fid=' + fileId + '&bid=' + blockId, true);
  xhttp.send();
}
`
const instructionsHtml = `
<div class="topcorner">
<button onclick="toggleByClass('instructions')" type="button">
Meta Instructions</button>
<div class="instructions" onclick="toggleByClass('instructions')">
<p>You're viewing a tutorial with command blocks tested to run
in bash on a linux system.</p>
<p>Clicking on a command block header
copies the block to your clipboard so you can mouse over
to a shell and click again to paste it for execution.</p>
<p>
For one-click usage (preferred for demos):
<ul>
<li>
Install <code><a target="_blank"
href="https://golang.org/doc/install">Go</a></code>
(the programming language) and
<code><a target="_blank"
href="https://github.com/tmux/tmux/wiki">tmux</a></code>
(the terminal multiplexer).</li>
<li>Install the <code>tmux</code>
websocket adapter
<code><a target="_blank"
href="https://github.com/monopole/mdrip">mdrip</a></code>:
<pre>
  GOPATH=/tmp/mdrip go install github.com/monopole/mdrip
</pre>
</li>
<li>Run (in any shell):
<pre>
  /tmp/mdrip/bin/mdrip --mode tmux ws://{{.Host}}/ws?id={{.SessId}}
</pre>
</li>
<li>
Run <code>tmux</code>.
</ul>
<p>
Now, clicking a command block header sends the block
from this page's server over a websocket to your local
<code>mdrip</code>, which then  'pastes' the block
to your active <code>tmux</code> pane.</p><p>
The socket evaporates after a period of inactivity,
and can be restarted with the same command.</p>
</div>
</div>
`
