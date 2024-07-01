package session

import (
	"crypto/rand"
	_ "embed"
	"encoding/gob"
	"fmt"

	"github.com/gorilla/sessions"
)

var (
	//go:embed session.js
	Js string
)

// TypeSessID represents a session ID.
type TypeSessID string

const forRegistration = TypeSessID("arbitrary")

func init() {
	gob.Register(forRegistration)
}

//go:generate stringer -type=Route -linecomment
type Route int

const (
	RouteUnknown Route = iota
	// RouteJs is the GET endpoint for most of the javascript needed by the webapps.
	RouteJs // js
	// RouteCss is the GET endpoint for all the css needed by the webapps.
	RouteCss // css
	// RouteReload tells the server to reload all data from the file system.
	RouteReload // reload
	// RouteLabelsForFile is the GET endpoint for code block labels of one markdown file.
	RouteLabelsForFile // labelsForFile
	// RouteHtmlForFile is the GET endpoint for HTML of one markdown file.
	RouteHtmlForFile // htmlForFile
	// RouteRunBlock is the POST endpoint to trigger code block execution.
	RouteRunBlock // runCodeBlock
	// RouteSave is the POST endpoint to save application state.
	RouteSave // save
	// RouteImage returns an image.
	RouteImage // image
	// RouteQuit tells the server to quit.
	RouteQuit // quit
	// RouteDebug tells the server to render a debug page.
	RouteDebug // debug
	// RouteWebSocket sets up a socket.
	RouteWebSocket // debug
)

const (
	// dynamicPrefix is the prefix for dynamic, rendering required
	// requests, POST requests and special requests like
	// telling the server to quit.  This distinguishes
	// such paths from paths to static content like images.
	// TODO: put up distinct ports for static vs dynamic?
	dynamicPrefix = "/_/"
)

func Dynamic(r Route) string {
	return dynamicPrefix + r.String()
}

// These URL query parameter and cookie field names
// should all be unique and preferably short.
const (

	// KeyMdSessID is the param name for session ID.
	KeyMdSessID = "sid"
	// KeyIsTitleOn is the param name for is-the-title-on boolean.
	KeyIsTitleOn = "tit"
	// KeyIsNavOn is the param name for the is-the-nav-on boolean.
	KeyIsNavOn = "nav"
	// KeyMdFileIndex is the param name for the markdown file index.
	KeyMdFileIndex = "fix"
	// KeyBlockIndex is the param name for the code block index.
	KeyBlockIndex = "bix"
)

func makeSessionID() TypeSessID {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return TypeSessID(fmt.Sprintf("%x", b))
}

// AssureDefaults inserts default values if values are missing.
func AssureDefaults(s *sessions.Session) {
	var ok bool
	if _, ok = s.Values[KeyMdSessID].(TypeSessID); !ok {
		s.Values[KeyMdSessID] = makeSessionID()
	}
	if _, ok = s.Values[KeyIsTitleOn].(bool); !ok {
		s.Values[KeyIsTitleOn] = true
	}
	if _, ok = s.Values[KeyIsNavOn].(bool); !ok {
		s.Values[KeyIsNavOn] = false
	}
	if _, ok = s.Values[KeyMdFileIndex].(int); !ok {
		s.Values[KeyMdFileIndex] = 0
	}
	if _, ok = s.Values[KeyBlockIndex].(int); !ok {
		s.Values[KeyBlockIndex] = -1
	}
}

// Bucket holds session state data, presumably associated with a cookie.
type Bucket struct {
	// The session ID.
	MdSessID TypeSessID
	// Is the header showing?
	IsHeaderOn bool
	// Is the nav showing?
	IsNavOn bool
	// The active markdown file.
	MdFileIndex int
	// The active block in that file.
	BlockIndex int
}

// ConvertToBucket creates a SessionData instance;
// a copy of the session data but in typesafe fields rather than
// a map of string to any.
func ConvertToBucket(s *sessions.Session) *Bucket {
	return &Bucket{
		MdSessID:    s.Values[KeyMdSessID].(TypeSessID),
		IsHeaderOn:  s.Values[KeyIsTitleOn].(bool),
		IsNavOn:     s.Values[KeyIsNavOn].(bool),
		MdFileIndex: s.Values[KeyMdFileIndex].(int),
		BlockIndex:  s.Values[KeyBlockIndex].(int),
	}
}
