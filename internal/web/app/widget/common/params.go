package common

import (
	"github.com/monopole/mdrip/v2/internal/web/config"
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

		PathSave:             config.Dynamic(config.RouteSave),
		PathReload:           config.Dynamic(config.RouteReload),
		PathGetHtmlForFile:   config.Dynamic(config.RouteHtmlForFile),
		PathGetLabelsForFile: config.Dynamic(config.RouteLabelsForFile),
		PathRunBlock:         config.Dynamic(config.RouteRunBlock),

		KeyMdFileIndex: config.KeyMdFileIndex,
		KeyBlockIndex:  config.KeyBlockIndex,
		KeyIsTitleOn:   config.KeyIsTitleOn,
		KeyIsNavOn:     config.KeyIsNavOn,
		KeyMdSessID:    config.KeyMdSessID,

		MdSessID:          "notARealSessId",
		TransitionSpeedMs: 250,
	}
)
