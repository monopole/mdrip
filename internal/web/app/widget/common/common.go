package common

import (
	"bytes"
	_ "embed"
	"fmt"
	htmlTmpl "html/template"
	"strings"
	textTmpl "text/template"

	"github.com/monopole/mdrip/v2/internal/loader"
)

var (
	//go:embed common.css
	Css string
	//go:embed common.js
	Js string
)

const irrelevantTemplateName = "thisNameDoesNotMatter"

// MakeFuncMap makes a string to function map for use in Go template rendering.
func MakeFuncMap() map[string]interface{} {
	return map[string]interface{}{
		"toUpper": strings.ToUpper,
		"idAndLabel": func(i int, lab loader.Label) interface{} {
			return &struct {
				Id    int
				Label string
			}{Id: i, Label: string(lab)}
		},
		// It seems like an "em" is about 5/6 of one average character.
		"numCharsToEm": func(i int) string {
			return fmt.Sprintf("%.1fem", 5.0*float32(i)/6.0)
		},
	}
}

func MustRenderHtml(
	tmplBody string, values any, tmplName string) htmlTmpl.HTML {
	var b bytes.Buffer
	tmplParsed, err := ParseAsHtmlTemplate(tmplBody)
	if err != nil {
		panic(fmt.Errorf("unable to parse %s; %w", tmplName, err))
	}
	err = tmplParsed.ExecuteTemplate(&b, tmplName, values)
	if err != nil {
		panic(fmt.Errorf("unable to execute %s; %w", tmplName, err))
	}
	return htmlTmpl.HTML(b.String())
}

func ParseAsHtmlTemplate(s string) (*htmlTmpl.Template, error) {
	return htmlTmpl.New(irrelevantTemplateName).Funcs(MakeFuncMap()).Parse(s)
}

func ParseAsTextTemplate(s string) (*textTmpl.Template, error) {
	return textTmpl.New(irrelevantTemplateName).Funcs(MakeFuncMap()).Parse(s)
}

func MustHtmlTemplate(s string) *htmlTmpl.Template {
	return htmlTmpl.Must(ParseAsHtmlTemplate(s))
}

func AsTmpl(name, body string) string {
	return `
{{define "` + name + `"}}` + body + `{{end}}
`
}
