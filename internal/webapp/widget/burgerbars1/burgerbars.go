package burgerbars1

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
)

const (
	TmplName = "tmplBurgerBars"
)

var (
	//go:embed burgerbars.css
	Css string
	//go:embed burgerbars.js
	Js string
	//go:embed burgerbars.html
	html string
)

func AsTmpl() string {
	return common.AsTmpl(TmplName, html)
}
