package webserver

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/lexer"
	"github.com/monopole/mdrip/model"
	"github.com/monopole/mdrip/program"
	"github.com/monopole/mdrip/tmux"
	"github.com/monopole/mdrip/util"
	"github.com/monopole/mdrip/webapp"
	"strings"
)

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

type Server struct {
	pathArgs     []base.FilePath
	tutorial     model.Tutorial
	store        sessions.Store
	upgrader     websocket.Upgrader
	connections  map[webapp.TypeSessId]*myConn
	connReaperCh chan bool
}

const (
	cookieName = "mdrip"
	keySessId  = "sessId"
)

// var keyAuth = securecookie.GenerateRandomKey(16)
var keyAuth = []byte("static-visible-secret")
var keyEncrypt = []byte(nil)

func NewServer(pathArgs []base.FilePath, tut model.Tutorial) *Server {
	s := sessions.NewCookieStore(keyAuth, keyEncrypt)
	s.Options = &sessions.Options{
		Domain:   "localhost",
		Path:     "/",
		MaxAge:   3600 * 8, // 8 hours
		HttpOnly: true,
	}
	result := &Server{
		pathArgs,
		tut,
		s,
		websocket.Upgrader{},
		make(map[webapp.TypeSessId]*myConn),
		nil}
	result.startConnReaper()
	return result
}

func getSessionId(s *sessions.Session) webapp.TypeSessId {
	if c, ok := s.Values[keySessId].(string); ok {
		return webapp.TypeSessId(c)
	}
	return ""
}

func assureSessionId(s *sessions.Session) webapp.TypeSessId {
	c := getSessionId(s)
	if c == "" {
		c = makeSessionId()
		s.Values[keySessId] = string(c)
	}
	return c
}

func makeSessionId() webapp.TypeSessId {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return webapp.TypeSessId(fmt.Sprintf("%X", b))
}

func getSessionIdParam(n string, r *http.Request) (webapp.TypeSessId, error) {
	v := r.URL.Query().Get(n)
	if v == "" {
		return "", errors.New("no session Id")
	}
	return webapp.TypeSessId(v), nil
}

// Pull session Id out of request, create a socket connection,
// store connection in a map.  The block runner will attempt to
// find the connection and write to it, else fall back to its
// other behaviors.
func (ws *Server) openWebSocket(w http.ResponseWriter, r *http.Request) {
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

func write500(w http.ResponseWriter, e error) {
	http.Error(w, e.Error(), http.StatusInternalServerError)
}

const github = "https://github.com/"

func (ws *Server) reload(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("q")
	var t model.Tutorial
	var err error
	if len(url) > 0 {
		if strings.HasPrefix(url, github) {
			t, err = lexer.LoadTutorialFromGitHub(url)
			if err != nil {
				http.Error(w,
					fmt.Sprintf("Unable to read from url %s",
						url), http.StatusBadRequest)
				return
			}
			url = url[len(github):]
		} else {
			t, err = lexer.LoadTutorialFromPath(base.FilePath(url), url)
			if err != nil {
				write500(w, err)
				return
			}
		}
		ws.pathArgs = []base.FilePath{base.FilePath(url)}
	} else {
		t, err = lexer.LoadTutorialFromPaths(ws.pathArgs)
		if err != nil {
			write500(w, err)
			return
		}
	}
	ws.tutorial = t
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (ws *Server) showControlPage(w http.ResponseWriter, r *http.Request) {
	session, err := ws.store.Get(r, cookieName)
	if err != nil {
		write500(w, err)
		return
	}
	app := webapp.NewWebApp(
		assureSessionId(session), string(ws.pathArgs[0]), r.Host, ws.tutorial)
	err = session.Save(r, w)
	if err != nil {
		write500(w, err)
		return
	}
	if err := app.Render(w); err != nil {
		write500(w, err)
		return
	}
}

func (ws *Server) showDebugPage(w http.ResponseWriter, r *http.Request) {
	ws.tutorial.Accept(model.NewTutorialTxtPrinter(w))
	p := program.NewProgramFromTutorial(base.AnyLabel, ws.tutorial)
	fmt.Fprintf(w, "\n\nfile count %d\n\n", len(p.Lessons()))
	for i, lesson := range p.Lessons() {
		fmt.Fprintf(w, "file %d: %s\n", i, lesson.Path())
		for j, b := range lesson.Blocks() {
			fmt.Fprintf(w, "  block %d, content: %s\n",
				j, util.SampleString(b.Code().String(), 50))
		}
	}
}

// Returns a writer one can write a code block to for execution.
// First tries to find a session socket.  Failing that, try to find
// a locally running instance of tmux.  Failing that, returns a
// writer that discards the code.
func (ws *Server) getCodeRunner(sessId webapp.TypeSessId) io.Writer {
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

func (ws *Server) makeBlockRunner() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := ws.store.Get(r, cookieName)
		if err != nil {
			write500(w, err)
			return
		}
		sessId := assureSessionId(session)
		indexFile := getIntParam("fid", r, -1)
		glog.Info("fid = ", indexFile)
		indexBlock := getIntParam("bid", r, -1)
		glog.Info("bid = ", indexBlock)
		p := program.NewProgramFromTutorial(base.AnyLabel, ws.tutorial)
		limit := len(p.Lessons()) - 1
		if indexFile < 0 || indexFile > limit {
			http.Error(w,
				fmt.Sprintf("fid %d out of range 0-%d",
					indexFile, limit), http.StatusBadRequest)
			return
		}
		limit = len(p.Lessons()[indexFile].Blocks()) - 1
		if indexBlock < 0 || indexBlock > limit {
			http.Error(w,
				fmt.Sprintf("bid %d out of range 0-%d",
					indexBlock, limit), http.StatusBadRequest)
			return
		}
		block := p.Lessons()[indexFile].Blocks()[indexBlock]
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

func (ws *Server) favicon(w http.ResponseWriter, r *http.Request) {
	util.Lissajous(w, 7, 3, 1)
}

func (ws *Server) image(w http.ResponseWriter, r *http.Request) {
	session, _ := ws.store.Get(r, cookieName)
	session.Save(r, w)
	util.Lissajous(w,
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

func (ws *Server) quit(w http.ResponseWriter, r *http.Request) {
	close(ws.connReaperCh)
	os.Exit(0)
}

// Periodically look for and close idle websockets.
func (ws *Server) startConnReaper() {
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
func (ws *Server) Serve(hostAndPort string) {
	r := mux.NewRouter()
	// r.Host(hostAndPort)
	r.HandleFunc("/reload", ws.reload)
	r.HandleFunc("/runblock", ws.makeBlockRunner())
	r.HandleFunc("/debug", ws.showDebugPage)
	r.HandleFunc("/ws", ws.openWebSocket)
	r.HandleFunc("/favicon.ico", ws.favicon)
	r.HandleFunc("/image", ws.image)
	r.HandleFunc("/q", ws.quit)
	r.HandleFunc("/", ws.showControlPage)
	fmt.Println("Serving at " + hostAndPort)
	glog.Info("Serving at " + hostAndPort)
	//http.Handle("/", r)
	glog.Fatal(http.ListenAndServe(hostAndPort, r))
}
