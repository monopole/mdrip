package codeblock

import (
	_ "embed"
	"html/template"

	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
)

const (
	CbPrompt = template.HTML("&nbsp;►")
	TmplName = "tmplCodeBlock"
)

var (
	//go:embed codeblock.js
	Js string
	//go:embed codeblock.css
	Css string
	//go:embed codeblock.html
	myHtml string
)

func AsTmpl() string {
	return common.AsTmpl(TmplName, myHtml)
}
