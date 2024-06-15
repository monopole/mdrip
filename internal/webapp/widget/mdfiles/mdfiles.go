package mdfiles

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
)

const (
	TmplName = "tmplMdfiles"
)

var (
	//go:embed mdfiles.js
	Js string
	//go:embed mdfiles.css
	Css string
	//go:embed mdfiles.html
	html string
)

func AsTmpl() string {
	return common.AsTmpl(TmplName, html)
}
