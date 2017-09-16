package webserver

import (
	"crypto/rand"
	"fmt"

	"html/template"
	"io"
	"net/http"

	"errors"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/golang/glog"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/monopole/mdrip/model"
	"github.com/monopole/mdrip/program"
	"github.com/monopole/mdrip/util"

	"github.com/monopole/mdrip/tmux"
	"time"
)

type typeSessId string

// A program and associated info for rendering.
type Control struct {
	SessId typeSessId
	Host   string
	Pgm    *program.Program
}

type myConn struct {
	conn    *websocket.Conn
	lastUse time.Time
}

func (c myConn) Write(bytes []byte) (n int, err error) {
	glog.Info("Attempting socket write.")
	c.lastUse = time.Now()
	err = c.conn.WriteMessage(websocket.TextMessage, bytes)
	if err != nil {
		glog.Error("bad socket write:", err)
	}
	return len(bytes), err
}

type Webserver struct {
	store        sessions.Store
	upgrader     websocket.Upgrader
	control      Control
	connections  map[typeSessId]*myConn
	connReaperCh chan bool
}

const (
	cookieName = "mdrip"
	keySessId  = "sessId"
)

const (
	tmplNameControl = "control"
	tmplBodyControl = `
{{define "` + tmplNameControl + `"}}
<html>
<head>` + headerHtml + `</head>
<body onload="onLoad()"> ` + instructionsHtml + `
{{ template "` + program.TmplNameProgram + `" .Pgm }}
</body>
</html>
{{end}}
`
)

// var keyAuth = securecookie.GenerateRandomKey(16)
var keyAuth = []byte("static-visible-secret")
var keyEncrypt = []byte(nil)

var templates = template.Must(
	template.New("main").Parse(
		model.TmplBodyCommandBlock + model.TmplBodyParsedFile + program.TmplBodyProgram + tmplBodyControl))

func NewWebserver(p *program.Program) *Webserver {
	s := sessions.NewCookieStore(keyAuth, keyEncrypt)
	s.Options = &sessions.Options{
		Domain:   "localhost",
		Path:     "/",
		MaxAge:   3600 * 8, // 8 hours
		HttpOnly: true,
	}
	result := &Webserver{
		s,
		websocket.Upgrader{},
		Control{"blood", "example.com", p},
		make(map[typeSessId]*myConn),
		nil}
	result.startConnReaper()
	return result
}

func getSessionId(s *sessions.Session) typeSessId {
	if c, ok := s.Values[keySessId].(string); ok {
		return typeSessId(c)
	}
	return ""
}

func assureSessionId(s *sessions.Session) typeSessId {
	c := getSessionId(s)
	if c == "" {
		c = makeSessionId()
		s.Values[keySessId] = string(c)
	}
	return c
}

func makeSessionId() typeSessId {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return typeSessId(fmt.Sprintf("%X", b))
}

func getSessionIdParam(n string, r *http.Request) (typeSessId, error) {
	v := r.URL.Query().Get(n)
	if v == "" {
		return "", errors.New("no session Id")
	}
	return typeSessId(v), nil
}

// Pull session Id out of request, create a socket connection,
// store connection in a map.  The block runner will attempt to
// find the connection and write to it, else fall back to its
// other behaviors.
func (ws *Webserver) openWebSocket(w http.ResponseWriter, r *http.Request) {
	sessId, err := getSessionIdParam("id", r)
	if err != nil {
		glog.Errorf("no session Id: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if c := ws.connections[sessId]; c != nil {
		glog.Info("Wut? session found: ", sessId)
		// Possibly the other side shutdown and restarted.
		// Close and make new one.
		c.conn.Close()
		ws.connections[sessId] = nil
	}
	c, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		glog.Info("upgrade:", err)
		return
	}
	glog.Info("established websocket")
	go func() {
		_, message, err := c.ReadMessage()
		if err == nil {
			glog.Info("handshake: ", string(message))
		} else {
			glog.Info("websocket err:", err)
		}
	}()
	ws.connections[sessId] = &myConn{c, time.Now()}
}

func (ws *Webserver) showControlPage(w http.ResponseWriter, r *http.Request) {
	session, err := ws.store.Get(r, cookieName)
	if err != nil {
		glog.Errorf("Unable to get session: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ws.control.SessId = assureSessionId(session)
	ws.control.Host = r.Host
	err = session.Save(r, w)
	if err != nil {
		glog.Errorf("Unable to save session: %v", err)
	}
	ws.control.Pgm.Reload()
	if err := templates.ExecuteTemplate(w, tmplNameControl, ws.control); err != nil {
		glog.Fatal(err)
	}
}

func (ws *Webserver) showDebugPage(w http.ResponseWriter, r *http.Request) {
	ws.control.Pgm.Reload()
	ws.control.Pgm.GetTutorial().Accept(program.NewTutorialPrinter(w))

	fmt.Fprintf(w, "file count %d\n\n", ws.control.Pgm.ParsedFileCount())
	for i, s := range ws.control.Pgm.AllParsedFiles() {
		fmt.Fprintf(w, "file %d: %s\n", i, s.Path())
		for j, b := range s.Blocks() {
			fmt.Fprintf(w, "  block %d content: %s\n", j, util.SampleString(string(b.Code()), 50))
			fmt.Fprintf(w, "  num labels: %d\n", len(b.Labels()))
			for k, l := range b.Labels() {
				fmt.Fprintf(w, "    label %d:  %s\n", k, string(l))
			}
			fmt.Fprintln(w)
		}
	}
}

// Returns a writer one can write a code block to for execution.
// First tries to find a session socket.  Failing that, try to find
// a locally running instance of tmux.  Failing that, returns a
// writer that discards the code.
func (ws *Webserver) getCodeRunner(sessId typeSessId) io.Writer {
	c := ws.connections[sessId]
	if c != nil {
		glog.Infof("Socket found for ID %v", sessId)
		return c
	}
	t := tmux.NewTmux(tmux.Path)
	if t.IsUp() {
		glog.Info("No socket, writing to local tmux.")
		return t
	}
	glog.Info("No sockets, tmux not up, discarding code.")
	return ioutil.Discard
}

func (ws *Webserver) makeBlockRunner() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := ws.store.Get(r, cookieName)
		if err != nil {
			glog.Errorf("Unable to get session: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sessId := assureSessionId(session)
		// TODO(jregan): 404 on bad params
		indexFile := getIntParam("fid", r, -1)
		glog.Info("fid = ", indexFile)
		indexBlock := getIntParam("bid", r, -1)
		glog.Info("bid = ", indexBlock)
		block := ws.control.Pgm.ParsedFiles[indexFile].Blocks()[indexBlock]
		_, err = ws.getCodeRunner(sessId).Write(block.Code().Bytes())
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		session.Values["file"] = strconv.Itoa(indexFile)
		session.Values["block"] = strconv.Itoa(indexBlock)
		err = session.Save(r, w)
		if err != nil {
			glog.Errorf("Unable to save session: %v", err)
		}
		fmt.Fprintln(w, "Ok")
	}
}

func (ws *Webserver) favicon(w http.ResponseWriter, r *http.Request) {
	model.Lissajous(w, 7, 3, 1)
}

func (ws *Webserver) image(w http.ResponseWriter, r *http.Request) {
	session, _ := ws.store.Get(r, cookieName)
	session.Save(r, w)
	model.Lissajous(w,
		getIntParam("s", r, 300),
		getIntParam("c", r, 30),
		getIntParam("n", r, 100))
}

func getIntParam(n string, r *http.Request, d int) int {
	v, err := strconv.Atoi(r.URL.Query().Get(n))
	if err != nil {
		return d
	}
	return v
}

func (ws *Webserver) quit(w http.ResponseWriter, r *http.Request) {
	close(ws.connReaperCh)
	os.Exit(0)
}

// Periodically look for and close idle websockets.
func (ws *Webserver) startConnReaper() {
	if ws.connReaperCh != nil {
		glog.Fatal("Already have a reaper?")
	}
	ws.connReaperCh = make(chan bool)
	go func() {
		for {
			for _, c := range ws.connections {
				if time.Since(c.lastUse) > 10*time.Minute {
					c.conn.Close()
				}
			}
			select {
			case <-time.After(time.Minute):
			case <-ws.connReaperCh:
				for _, c := range ws.connections {
					c.conn.Close()
				}
				return
			}
		}
	}()
}

// Serve offers an http service.
func (ws *Webserver) Serve(hostAndPort string) {
	http.HandleFunc("/", ws.showControlPage)
	http.HandleFunc("/runblock", ws.makeBlockRunner())
	http.HandleFunc("/debug", ws.showDebugPage)
	http.HandleFunc("/ws", ws.openWebSocket)
	http.HandleFunc("/favicon.ico", ws.favicon)
	http.HandleFunc("/image", ws.image)
	http.HandleFunc("/q", ws.quit)
	fmt.Println("Serving at " + hostAndPort)
	glog.Info("Serving at " + hostAndPort)
	glog.Fatal(http.ListenAndServe(hostAndPort, nil))
}

const instructionsHtml = `
<div class="topcorner">
<button onclick="showhide('instructions')" type="button">
Meta Instructions</button>
<div class="instructions" onclick="showhide('instructions')">
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

const headerHtml = `
<style type="text/css">
body {
  font-family: "Veranda", Veranda, sans-serif;
  /* background-color: antiquewhite; */
  background-color: white;
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
  top: 0;
  right: 10;
  z-index: 100;
}
div.instructions {
  position: absolute;
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
</style>
<script type="text/javascript">
  function showhide(name) {
    var elements = document.getElementsByClassName(name);
    var e = elements[0];
    e.style.display = (e.style.display == 'block') ? 'none' : 'block';
  }
  // blockUx, which may cause screen flicker, not needed if write is very fast.
  var blockUx = false
  var runButtons = []
  var requestRunning = false
  function onLoad() {
    if (blockUx) {
      runButtons = document.getElementsByTagName('input');
    }
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
</script>
`
