package common

import (
	"github.com/monopole/mdrip/v2/internal/webapp/widget/session"
)

type ParamStructJsCss struct {
	MdHost string

	MaxNavWordLength int

	PathRunBlock         string
	PathSave             string
	PathReload           string
	PathGetHtmlForFile   string
	PathGetLabelsForFile string

	KeyMdSessID    string
	KeyMdFileIndex string
	KeyBlockIndex  string
	KeyIsTitleOn   string
	KeyIsNavOn     string

	MdSessID          string
	TransitionSpeedMs int
}

var (
	ParamDefaultJsCss = ParamStructJsCss{
		MdHost: "www.yourmom.com",

		MaxNavWordLength: 43,

		PathSave:             session.Dynamic(session.RouteSave),
		PathReload:           session.Dynamic(session.RouteReload),
		PathGetHtmlForFile:   session.Dynamic(session.RouteHtmlForFile),
		PathGetLabelsForFile: session.Dynamic(session.RouteLabelsForFile),
		PathRunBlock:         session.Dynamic(session.RouteRunBlock),

		KeyMdFileIndex: session.KeyMdFileIndex,
		KeyBlockIndex:  session.KeyBlockIndex,
		KeyIsTitleOn:   session.KeyIsTitleOn,
		KeyIsNavOn:     session.KeyIsNavOn,
		KeyMdSessID:    session.KeyMdSessID,

		MdSessID:          "notARealSessId",
		TransitionSpeedMs: 250,
	}
)
