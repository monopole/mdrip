package navleftroot

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/common"
)

const (
	TmplName = "tmplNavLeftRoot"
)

var (
	//go:embed navleftroot.js
	Js string
	//go:embed navleftroot.css
	Css string
	//go:embed navleftroot.html
	html string
)

func AsTmpl() string {
	return common.AsTmpl(TmplName, html)
}
