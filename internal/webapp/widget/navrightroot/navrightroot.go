package navrightroot

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
)

const (
	TmplName = "tmplNavRightRoot"
)

var (
	//go:embed navrightroot.js
	Js string
	//go:embed navrightroot.css
	Css string
	//go:embed navrightroot.html
	html string
)

func AsTmpl() string {
	return common.AsTmpl(TmplName, html)
}
