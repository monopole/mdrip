package common

import (
	"github.com/monopole/mdrip/v2/internal/webapp/widget/session"
)

type ParamStructSession struct {
	MdHost string

	PathRunBlock         string
	PathSave             string
	PathGetHtmlForFile   string
	PathGetLabelsForFile string

	KeyMdSessID    string
	KeyMdFileIndex string
	KeyBlockIndex  string
	KeyIsTitleOn   string
	KeyIsNavOn     string

	MdSessID string
}

type ParamStructTransition struct {
	TransitionSpeedMs int
}

var (
	ParamDefaultSession = ParamStructSession{
		MdHost: "www.yourmom.com",

		PathSave:             session.PathSave,
		PathGetHtmlForFile:   session.PathGetHtmlForFile,
		PathGetLabelsForFile: session.PathGetLabelsForFile,
		PathRunBlock:         session.PathRunBlock,

		KeyMdFileIndex: session.KeyMdFileIndex,
		KeyBlockIndex:  session.KeyBlockIndex,
		KeyIsTitleOn:   session.KeyIsTitleOn,
		KeyIsNavOn:     session.KeyIsNavOn,
		KeyMdSessID:    session.KeyMdSessID,

		MdSessID: "notARealSessId",
	}
	ParamDefaultTransition = ParamStructTransition{
		TransitionSpeedMs: 250,
	}
)
