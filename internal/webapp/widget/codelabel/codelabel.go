package codelabel

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
)

const (
	TmplName = "tmplCodeLabel"
)

var (
	//go:embed codelabel.js
	Js string
	//go:embed codelabel.css
	Css string
	//go:embed codelabel.html
	html string
)

func AsTmpl() string {
	return common.AsTmpl(TmplName, html)
}
