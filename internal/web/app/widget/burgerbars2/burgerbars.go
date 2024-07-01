package burgerbars2

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/common"
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
