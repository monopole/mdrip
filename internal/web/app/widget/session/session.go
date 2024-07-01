package session

import (
	"crypto/rand"
	_ "embed"
	"encoding/gob"
	"fmt"

	"github.com/gorilla/sessions"
	"github.com/monopole/mdrip/v2/internal/web/config"
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
	if _, ok = s.Values[config.KeyMdSessID].(TypeSessID); !ok {
		s.Values[config.KeyMdSessID] = makeSessionID()
	}
	if _, ok = s.Values[config.KeyIsTitleOn].(bool); !ok {
		s.Values[config.KeyIsTitleOn] = true
	}
	if _, ok = s.Values[config.KeyIsNavOn].(bool); !ok {
		s.Values[config.KeyIsNavOn] = false
	}
	if _, ok = s.Values[config.KeyMdFileIndex].(int); !ok {
		s.Values[config.KeyMdFileIndex] = 0
	}
	if _, ok = s.Values[config.KeyBlockIndex].(int); !ok {
		s.Values[config.KeyBlockIndex] = -1
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
		MdSessID:    s.Values[config.KeyMdSessID].(TypeSessID),
		IsHeaderOn:  s.Values[config.KeyIsTitleOn].(bool),
		IsNavOn:     s.Values[config.KeyIsNavOn].(bool),
		MdFileIndex: s.Values[config.KeyMdFileIndex].(int),
		BlockIndex:  s.Values[config.KeyBlockIndex].(int),
	}
}
