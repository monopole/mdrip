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
	"github.com/gorilla/websocket"
	"github.com/gorilla/sessions"
	"github.com/monopole/mdrip/model"
	"github.com/monopole/mdrip/program"
	"github.com/monopole/mdrip/tmux"
)

type TypeSessId string

// A program and anything that needs to be rendered with it.
type Control struct {
	SessId TypeSessId
	Pgm       *program.Program
}

type Webserver struct {
	store       sessions.Store
	upgrader    websocket.Upgrader
	control     Control
	connections map[TypeSessId]*websocket.Conn
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
		model.TmplBodyCommandBlock + model.TmplBodyScript + program.TmplBodyProgram + tmplBodyControl))

func NewWebserver(p *program.Program) *Webserver {
	s := sessions.NewCookieStore(keyAuth, keyEncrypt)
	s.Options = &sessions.Options{
		Domain:   "localhost",
		Path:     "/",
		MaxAge:   3600 * 8, // 8 hours
		HttpOnly: true,
	}
	return &Webserver{s, websocket.Upgrader{}, Control{"blood", p}, make(map[TypeSessId]*websocket.Conn)}
}

func getSessionId(s *sessions.Session) TypeSessId {
	if c, ok := s.Values[keySessId].(string); ok {
		return TypeSessId(c)
	}
	return ""
}

func assureSessionId(s *sessions.Session) TypeSessId {
	c := getSessionId(s)
	if c == "" {
		c = makeSessionId()
		s.Values[keySessId] = string(c)
	}
	return c
}

func makeSessionId() TypeSessId {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return TypeSessId(fmt.Sprintf("%X", b))
}

func dumpSessionInfo(s *sessions.Session) {
	glog.Infof("    Session Name: %v", s.Name())
	glog.Infof("   Session isNew: %v", s.IsNew)
	glog.Infof("  Session Values: %v", s.Values)
}

func getSessionIdParam(n string, r *http.Request) (TypeSessId, error) {
	v := r.URL.Query().Get(n)
	if v == "" {
		return "", errors.New("no session Id")
	}
	return TypeSessId(v), nil
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
	if ws.connections[sessId] != nil {
		// Already have a session?
		glog.Info("Wut? session already exists: ", sessId)
		return
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
	ws.connections[sessId] = c
}

func (ws *Webserver) showControlPage(w http.ResponseWriter, r *http.Request) {
	session, err := ws.store.Get(r, cookieName)
	if err != nil {
		glog.Errorf("Unable to get session: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.control.SessId = assureSessionId(session)
	dumpSessionInfo(session)
	err = session.Save(r, w)
	if err != nil {
		glog.Errorf("Unable to save session: %v", err)
	}

	ws.control.Pgm.Reload()
	if err := templates.ExecuteTemplate(w, tmplNameControl, ws.control); err != nil {
		glog.Fatal(err)
	}
}

func (ws *Webserver) getCodeRunner() io.Writer {
	t := tmux.NewTmux(tmux.Path)
	if !t.IsUp() {
		glog.Info("tmux not up, will run anyway, discarding scripts.")
		return ioutil.Discard
	}
	return t
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
		indexScript := getIntParam("sid", r, -1)
		indexBlock := getIntParam("bid", r, -1)
		block := ws.control.Pgm.Scripts[indexScript].Blocks()[indexBlock]
		glog.Info("Running ", block.Name())

		/// TODO: bury this behind io.Writer, and return it from getCodeRunner.
		c := ws.connections[sessId]
		if c == nil {
			glog.Infof("No socket found for ID %v", sessId)
		} else {
			glog.Infof("Attempting write to %v", sessId)
			err = c.WriteMessage(websocket.TextMessage, block.Code().Bytes())
			if err != nil {
				glog.Info("bad socket write:", err)
			}
			return
		}

		codeRunner := ws.getCodeRunner()
		_, err = codeRunner.Write(block.Code().Bytes())
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}

		session.Values["script"] = strconv.Itoa(indexScript)
		session.Values["block"] = strconv.Itoa(indexBlock)
		dumpSessionInfo(session)
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
	for _, c := range ws.connections {
		c.Close()
	}
	os.Exit(0)
}

// Serve offers an http service.
func (ws *Webserver) Serve(hostAndPort string) {
	http.HandleFunc("/", ws.showControlPage)
	http.HandleFunc("/runblock", ws.makeBlockRunner())
	http.HandleFunc("/ws", ws.openWebSocket)
	http.HandleFunc("/favicon.ico", ws.favicon)
	http.HandleFunc("/image", ws.image)
	http.HandleFunc("/q", ws.quit)
	fmt.Println("Serving at " + hostAndPort)
	glog.Info("Serving at " + hostAndPort)
	glog.Fatal(http.ListenAndServe(hostAndPort, nil))
}

const instructionsHtml = `
<blockquote>
<p>This a tutorial with command blocks tested to run
on a linux system.</p>
<p>
Clicking on a command block header copies the block into your clipboard,
which you can then paste into a shell.</p>
<p>
For surprisingly pleasant one-click (auto-paste) usage do this:
<ul>
<li>
Install <code><a href="https://github.com/tmux/tmux/wiki">tmux</a></code>.
</li>
<li>In any shell, within or outside <code>tmux</code>, run
<pre>
  GOPATH=/tmp/mdrip go install github.com/monopole/mdrip
  /tmp/mdrip/bin/mdrip --mode tmux ws://localhost:8000/ws?id={{.SessId}}
</pre>
</li>
<li>
Establish focus in any <code>tmux</code> shell.
<li>
Click any command block header below.
</li>
</ul>
The block is then sent over the websocket established above,
then <em>pasted</em> to your focussed <code>tmux</code> pane.
</blockquote>
`

const headerHtml = `
<style type="text/css">
body {
  font-family: "Veranda", Veranda, sans-serif;
  /* background-color: antiquewhite; */
  background-color: white;
}

blockquote {
  font-size: 0.7em;
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
</style>
<script type="text/javascript">
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
    var scriptId = getId(commandBlockDiv.parentNode);
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
    xhttp.open('GET', '/runblock?sid=' + scriptId + '&bid=' + blockId, true);
    xhttp.send();
  }
</script>
`
