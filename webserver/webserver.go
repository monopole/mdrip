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
	"github.com/monopole/mdrip/model"
	"github.com/monopole/mdrip/tmux"
	"github.com/monopole/mdrip/tutorial"
	"github.com/monopole/mdrip/util"
	"github.com/monopole/mdrip/webapp"
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
	pathArgs     []model.FilePath
	store        sessions.Store
	upgrader     websocket.Upgrader
	connections  map[model.TypeSessId]*myConn
	connReaperCh chan bool
}

const (
	cookieName = "mdrip"
	keySessId  = "sessId"
)

// var keyAuth = securecookie.GenerateRandomKey(16)
var keyAuth = []byte("static-visible-secret")
var keyEncrypt = []byte(nil)

func NewServer(pathArgs []model.FilePath) *Server {
	s := sessions.NewCookieStore(keyAuth, keyEncrypt)
	s.Options = &sessions.Options{
		Domain:   "localhost",
		Path:     "/",
		MaxAge:   3600 * 8, // 8 hours
		HttpOnly: true,
	}
	result := &Server{
		pathArgs,
		s,
		websocket.Upgrader{},
		make(map[model.TypeSessId]*myConn),
		nil}
	result.startConnReaper()
	return result
}

func getSessionId(s *sessions.Session) model.TypeSessId {
	if c, ok := s.Values[keySessId].(string); ok {
		return model.TypeSessId(c)
	}
	return ""
}

func assureSessionId(s *sessions.Session) model.TypeSessId {
	c := getSessionId(s)
	if c == "" {
		c = makeSessionId()
		s.Values[keySessId] = string(c)
	}
	return c
}

func makeSessionId() model.TypeSessId {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return model.TypeSessId(fmt.Sprintf("%X", b))
}

func getSessionIdParam(n string, r *http.Request) (model.TypeSessId, error) {
	v := r.URL.Query().Get(n)
	if v == "" {
		return "", errors.New("no session Id")
	}
	return model.TypeSessId(v), nil
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

func (ws *Server) showControlPage(w http.ResponseWriter, r *http.Request) {
	session, err := ws.store.Get(r, cookieName)
	if err != nil {
		write500(w, err)
		return
	}
	t, err := tutorial.LoadTutorialFromPaths(ws.pathArgs)
	if err != nil {
		write500(w, err)
		return
	}
	app := webapp.NewWebApp(assureSessionId(session), r.Host, t)
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
	t, err := tutorial.LoadTutorialFromPaths(ws.pathArgs)
	if err != nil {
		write500(w, err)
		return
	}
	t.Accept(tutorial.NewTutorialTxtPrinter(w))
	p := tutorial.NewProgramFromTutorial(t)
	fmt.Fprintf(w, "\n\nfile count %d\n\n", len(p.Lessons()))
	for i, lesson := range p.Lessons() {
		fmt.Fprintf(w, "file %d: %s\n", i, lesson.Path())
		for j, b := range lesson.Blocks() {
			fmt.Fprintf(w, "  block %d content: %s\n",
				j, util.SampleString(string(b.Code()), 50))
			fmt.Fprintf(w, "  num labels: %d\n", len(b.Labels()))
			for k, label := range b.Labels() {
				fmt.Fprintf(w, "    label %d:  %s\n", k, string(label))
			}
			fmt.Fprintln(w)
		}
	}
}

// Returns a writer one can write a code block to for execution.
// First tries to find a session socket.  Failing that, try to find
// a locally running instance of tmux.  Failing that, returns a
// writer that discards the code.
func (ws *Server) getCodeRunner(sessId model.TypeSessId) io.Writer {
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
		// TODO(monopole): 404 on bad params
		indexFile := getIntParam("fid", r, -1)
		glog.Info("fid = ", indexFile)
		indexBlock := getIntParam("bid", r, -1)
		glog.Info("bid = ", indexBlock)
		t, err := tutorial.LoadTutorialFromPaths(ws.pathArgs)
		if err != nil {
			write500(w, err)
			return
		}
		p := tutorial.NewProgramFromTutorial(t)
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
		// TODO(monopole): 404 on out of range indices
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
func (ws *Server) oldServe(hostAndPort string) {
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

// Serve offers an http service.
func (ws *Server) Serve(hostAndPort string) {
	r := mux.NewRouter()
	// r.Host(hostAndPort)
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
