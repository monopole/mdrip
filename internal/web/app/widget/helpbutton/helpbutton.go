package helpbutton

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/common"
)

const (
	TmplName = "tmplHelpButton"
)

var (
	//go:embed helpbutton.html
	html string

	//go:embed helpbutton.css
	Css string

	//go:embed helpbutton.js
	Js string
)

func AsTmpl() string {
	return common.AsTmpl(TmplName, html)
}
