package webserver

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/loader"
	"github.com/monopole/mdrip/model"
	"github.com/monopole/mdrip/program"
	"github.com/monopole/mdrip/tmux"
	"github.com/monopole/mdrip/util"
	"github.com/monopole/mdrip/webapp"
)

type myConn struct {
	conn    *websocket.Conn
	lastUse time.Time
}

func (c *myConn) Write(bytes []byte) (n int, err error) {
	glog.Info("Attempting socket write.")
	c.lastUse = time.Now()
	err = c.conn.WriteMessage(websocket.TextMessage, bytes)
	if err != nil {
		glog.Error("bad socket write:", err)
		return 0, err
	}
	glog.Info("Socket seemed to work.")
	return len(bytes), nil
}

// Server represents a webserver.
type Server struct {
	loader           *loader.Loader
	didFirstRender   bool
	tutorial         model.Tutorial
	store            sessions.Store
	upgrader         websocket.Upgrader
	connections      map[webapp.TypeSessID]*myConn
	connReaperQuitCh chan bool
}

const (
	cookieName = "mdrip"
)

// var keyAuth = securecookie.GenerateRandomKey(16)
var keyAuth = []byte("static-visible-secret")
var keyEncrypt = []byte(nil)

// NewServer returns a new web server configured with the given loader.
func NewServer(l *loader.Loader) (*Server, error) {
	s := sessions.NewCookieStore(keyAuth, keyEncrypt)
	s.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 8, // 8 hours
		HttpOnly: true,
	}
	result := &Server{
		l,
		false,
		nil,
		s,
		websocket.Upgrader{},
		make(map[webapp.TypeSessID]*myConn),
		make(chan bool),
	}
	go result.reapConnections()
	return result, nil
}

func getSessIdParam(r *http.Request) (webapp.TypeSessID, error) {
	v := r.URL.Query().Get(webapp.KeySessID)
	if v == "" {
		return "", errors.New("no session Id")
	}
	return webapp.TypeSessID(v), nil
}

// Pull session Id out of request, create a socket connection,
// store connection in a map.  The block runner will attempt to
// find the connection and write to it, else fall back to its
// other behaviors.
func (ws *Server) openWebSocket(w http.ResponseWriter, r *http.Request) {
	sessID, err := getSessIdParam(r)
	if err != nil {
		glog.Errorf("no session Id: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	existingConn := ws.connections[sessID]
	var c *websocket.Conn
	if existingConn != nil {
		glog.Infof("Reusing live session %v found when asking for new session.", sessID)
		// Possibly the other side shutdown and restarted.
		// Could close and make new one,
		//  c.conn.Close()
		//  delete(ws.connections, sessID)
		// but try to reuse
		c = existingConn.conn
	} else {
		glog.Infof("Attempting to upgrade session %v to a websocket.", sessID)
		c, err = ws.upgrader.Upgrade(w, r, nil)
	}
	if err != nil {
		glog.Errorf("unable to upgrade for session %v: %v", sessID, err)
		write500(w, err)
		return
	}
	glog.Infof("established websocket for session %v", sessID)
	go func() {
		_, message, err := c.ReadMessage()
		if err == nil {
			glog.Info("handshake: ", string(message))
		} else {
			glog.Info("websocket err: ", err)
		}
	}()
	ws.connections[sessID] = &myConn{c, time.Now()}
}

func write500(w http.ResponseWriter, e error) {
	http.Error(w, e.Error(), http.StatusInternalServerError)
}

func (ws *Server) reload(w http.ResponseWriter, r *http.Request) {
	session, err := ws.store.Get(r, cookieName)
	if err != nil {
		write500(w, err)
		return
	}
	value := mux.Vars(r)["gitclone"]
	if len(value) < 1 {
		value = r.URL.Query().Get("q")
	}
	var t model.Tutorial
	if len(value) > 0 {
		// Load data from new source.
		ds, err := base.NewDataSet([]string{value})
		if err != nil {
			http.Error(w,
				fmt.Sprintf("Bad value %s", value), http.StatusBadRequest)
			return
		}
		l := loader.NewLoader(ds)
		glog.Infof("Loading from new source.")
		t, err = l.Load()
		if err != nil {
			http.Error(w,
				fmt.Sprintf("Unable to load from %s: %v", ds, err),
				http.StatusBadRequest)
			return
		}
		ws.loader = l
	} else {
		// reload from same source, presumably changed.
		glog.Infof("Reloading.")
		t, err = ws.loader.Load()
		if err != nil {
			write500(w, err)
			return
		}
	}
	err = session.Save(r, w)
	if err != nil {
		glog.Errorf("Unable to save session: %v", err)
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
	sessionData := webapp.AssureSessionData(session)
	glog.Infof("Main page render in sessID: %v", sessionData.SessID)
	if ws.didFirstRender {
		// Consider reloading data on all renders beyond the first.
		glog.Infof("Already did first render.")
		if !ws.loader.SmellsLikeGithub() {
			t, err := ws.loader.Load()
			if err == nil {
				ws.tutorial = t
				glog.Info("Reloaded data.")
			} else {
				glog.Errorf("Trouble reloading local data: %v", err)
			}
		}
	}
	err = session.Save(r, w)
	if err != nil {
		write500(w, err)
		return
	}

	app := ws.makeWebApp(sessionData, r.Host, r.URL.Path)
	ws.didFirstRender = true
	if err := app.Render(w); err != nil {
		write500(w, err)
		return
	}
}

func (ws *Server) makeWebApp(sessionData *webapp.SessionData, host, path string) *webapp.WebApp {
	v := newLessonFinder()
	ws.tutorial.Accept(v)
	var lessonPath []int
	if len(path) > 0 && path[0] == '/' {
		lessonPath = v.getLessonPath(path[1:])
	} else {
		lessonPath = v.getLessonPath(path)
	}
	return webapp.NewWebApp(
		sessionData, host,
		ws.tutorial, ws.loader.DataSet().FirstArg(),
		lessonPath, v.getCoursePaths())
}

func (ws *Server) showDebugPage(w http.ResponseWriter, r *http.Request) {
	session, err := ws.store.Get(r, cookieName)
	if err != nil {
		write500(w, err)
		return
	}
	session.Save(r, w)
	ws.tutorial.Accept(model.NewTutorialTxtPrinter(w))
	p := program.NewProgramFromTutorial(base.WildCardLabel, ws.tutorial)
	fmt.Fprintf(w, "\n\nfile count %d\n\n", len(p.Lessons()))
	for i, lesson := range p.Lessons() {
		fmt.Fprintf(w, "file %d: %s\n", i, lesson.Path())
		for j, b := range lesson.Blocks() {
			fmt.Fprintf(w, "  block %d, content: %s\n",
				j, util.SampleString(b.Code().String(), 50))
		}
	}
}

func (ws *Server) attemptTmuxWrite(b *program.BlockPgm) error {
	t := tmux.NewTmux(tmux.Path)
	if !t.IsUp() {
		return errors.New("no local tmux to write to")
	}
	_, err := t.Write(b.Code().Bytes())
	return err
}

func inRange(w http.ResponseWriter, name string, arg, n int) bool {
	if arg >= 0 || arg < n {
		return true
	}
	http.Error(w,
		fmt.Sprintf("%s %d out of range 0-%d",
			name, arg, n-1), http.StatusBadRequest)
	return false
}

func (ws *Server) makeBlockRunner() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		arg := r.URL.Query().Get(webapp.KeySessID)
		if len(arg) == 0 {
			http.Error(w, "No session id for block runner", http.StatusBadRequest)
			return
		}
		sessID := webapp.TypeSessID(arg)
		glog.Infof("%s = %s", webapp.KeySessID, sessID)
		indexFile := getIntParam(webapp.KeyLessonIndex, r, -1)
		glog.Infof("%s = %s", webapp.KeyLessonIndex, indexFile)
		indexBlock := getIntParam(webapp.KeyBlockIndex, r, -1)
		glog.Infof("%s = %s", webapp.KeyBlockIndex, indexBlock)

		p := program.NewProgramFromTutorial(base.WildCardLabel, ws.tutorial)
		if !inRange(w, webapp.KeyLessonIndex, indexFile, len(p.Lessons())) {
			return
		}
		lesson := p.Lessons()[indexFile]
		if !inRange(w, webapp.KeyBlockIndex, indexBlock, len(lesson.Blocks())) {
			return
		}
		block := lesson.Blocks()[indexBlock]

		var err error

		c := ws.connections[sessID]
		if c == nil {
			glog.Infof("no socket for session %v", sessID)
		} else {
			_, err := c.Write(block.Code().Bytes())
			if err != nil {
				glog.Infof("socket write failed: %v", err)
				delete(ws.connections, sessID)
			}
		}
		if c == nil || err != nil {
			err = ws.attemptTmuxWrite(block)
			if err != nil {
				glog.Infof("tmux write failed: %v", err)
				// nothing more to try
			}
		}
		fmt.Fprintln(w, "Ok")
	}
}

func (ws *Server) saveSession(w http.ResponseWriter, r *http.Request) {
	session, err := ws.store.Get(r, cookieName)
	if err != nil {
		write500(w, err)
		return
	}
	session.Values[webapp.KeyIsNavOn] = getBoolParam(webapp.KeyIsNavOn, r, false)
	session.Values[webapp.KeyIsHeaderOn] = getBoolParam(webapp.KeyIsHeaderOn, r, false)
	session.Values[webapp.KeyLessonIndex] = getIntParam(webapp.KeyLessonIndex, r, 0)
	session.Values[webapp.KeyBlockIndex] = getIntParam(webapp.KeyBlockIndex, r, 0)
	err = session.Save(r, w)
	if err != nil {
		glog.Errorf("Unable to save session: %v", err)
	}
	fmt.Fprintln(w, "Ok")
	glog.Info("Saved session.")
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

func getBoolParam(n string, r *http.Request, d bool) bool {
	v, err := strconv.ParseBool(r.URL.Query().Get(n))
	if err != nil {
		return d
	}
	return v
}

func (ws *Server) quit(w http.ResponseWriter, r *http.Request) {
	close(ws.connReaperQuitCh)
	fmt.Fprint(w, "\nbye bye\n")
	go func() {
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}()
}

const (
	maxConnectionIdleTime    = 30 * time.Minute
	connectionScanWaitPeriod = 5 * time.Minute
)

// Look for and close idle websockets.
func (ws *Server) closeStaleConnections() {
	for s, c := range ws.connections {
		if time.Since(c.lastUse) > maxConnectionIdleTime {
			glog.Infof(
				"Time since last use in session %v exceeds %v; closing.",
				s, maxConnectionIdleTime)
			c.conn.Close()
			delete(ws.connections, s)
		}
	}
}

// reapConnections periodically scans websockets for idleness.
// It also closes everything and quits scanning if quit signal received.
func (ws *Server) reapConnections() {
	for {
		ws.closeStaleConnections()
		select {
		case <-time.After(connectionScanWaitPeriod):
		case <-ws.connReaperQuitCh:
			glog.Info("Received quit, reaping all connections.")
			for s, c := range ws.connections {
				c.conn.Close()
				delete(ws.connections, s)
			}
			return
		}
	}
}

// Serve offers an http service.
func (ws *Server) Serve(hostAndPort string) error {
	r := mux.NewRouter()
	r.HandleFunc("/_/r", ws.reload)
	r.HandleFunc("/_/r/", ws.reload)
	r.HandleFunc("/_/r/{gitclone:.*}", ws.reload)
	r.HandleFunc("/_/runblock", ws.makeBlockRunner())
	r.HandleFunc("/_/s", ws.saveSession)
	r.HandleFunc("/_/debug", ws.showDebugPage)
	r.HandleFunc("/_/ws", ws.openWebSocket)
	r.HandleFunc("/_/image", ws.image)
	r.HandleFunc("/_/q", ws.quit)
	r.HandleFunc("/favicon.ico", ws.favicon)
	r.PathPrefix("/").HandlerFunc(ws.showControlPage)
	var err error
	fmt.Printf("Loading from %s\n", ws.loader.DataSet())
	ws.tutorial, err = ws.loader.Load()
	if err != nil {
		return err
	}
	fmt.Println("Serving at " + hostAndPort)
	glog.Fatal(http.ListenAndServe(hostAndPort, r))
	return nil
}
