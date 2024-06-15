package helpbox

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
)

const (
	TmplName = "tmplHelpBox"
)

var (
	//go:embed helpbox.html
	html string

	//go:embed helpbox.css
	Css string

	//go:embed helpbox.js
	Js string
)

func AsTmpl() string {
	return common.AsTmpl(TmplName, html)
}
