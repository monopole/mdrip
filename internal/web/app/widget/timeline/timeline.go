package timeline

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/common"
)

const (
	TmplName = "tmplTimeLine"
)

var (
	//go:embed timeline.js
	Js string

	//go:embed timeline.css
	Css string

	//go:embed timeline.html
	html string
)

func AsTmpl() string {
	return common.AsTmpl(TmplName, html)
}
