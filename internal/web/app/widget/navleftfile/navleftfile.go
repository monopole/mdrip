package navleftfile

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/common"
)

const (
	TmplName = "tmplNavLeftFile"
)

var (
	//go:embed navleftfile.js
	Js string
	//go:embed navleftfile.css
	Css string
	//go:embed navleftfile.html
	html string
)

func AsTmpl() string {
	return common.AsTmpl(TmplName, html)
}
