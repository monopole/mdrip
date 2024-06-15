package navtop

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
)

const (
	TmplName = "tmplNavTop"
)

var (
	//go:embed navtop.js
	Js string

	//go:embed navtop.css
	Css string

	//go:embed navtop.html
	html string
)

func AsTmpl() string {
	return common.AsTmpl(TmplName, html)
}
