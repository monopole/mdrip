package navleftfolder

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/common"
)

const (
	TmplName = "tmplNavLeftFolder"
)

var (
	//go:embed navleftfolder.js
	Js string
	//go:embed navleftfolder.css
	Css string
	//go:embed navleftfolder.html
	html string
)

func AsTmpl() string {
	return common.AsTmpl(TmplName, html)
}
