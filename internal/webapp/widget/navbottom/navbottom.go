package navbottom

import (
	_ "embed"

	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
)

const (
	TmplName = "tmplNavBottom"
)

var (
	//go:embed navbottom.js
	Js string

	//go:embed navbottom.css
	Css string

	//go:embed navbottom.html
	html string
)

func AsTmpl() string {
	return common.AsTmpl(TmplName, html)
}
