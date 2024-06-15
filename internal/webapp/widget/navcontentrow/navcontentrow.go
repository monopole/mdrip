package navcontentrow

import (
	_ "embed"
	"html/template"

	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
)

const (
	TmplName = "tmplNavContentRow"
)

type ParamStructContentRow struct {
	ContentTop    template.HTML
	ContentLeft   template.HTML
	ContentCenter template.HTML
	ContentRight  template.HTML
	ContentBottom template.HTML
}

var (
	//go:embed navcontentrow.css
	Css string

	//go:embed navcontentrow.js
	Js string

	//go:embed navcontentrow.html
	html string
)

func AsTmpl() string {
	return common.AsTmpl(TmplName, html)
}
