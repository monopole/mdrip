package codeblock_test

import (
	_ "embed"
	"html/template"
	"testing"

	"github.com/monopole/mdrip/v2/internal/webapp/widget/codeblock"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/testutil"
)

func TestWidget(t *testing.T) {
	testutil.RenderHtmlToFile(
		t, codeblock.AsTmpl()+tmplTestBody, makeParams())
}

func makeParams() any {
	return struct {
		common.ParamStructSession
		common.ParamStructTransition
		CbPrompt template.HTML
		Id       int
		Title    string
		Code     string
	}{
		common.ParamDefaultSession,
		common.ParamDefaultTransition,
		codeblock.CbPrompt,
		700,
		"Your Mom",
		`
echo "your mom"
date
which ls
cal
cat /etc/hosts
echo "the rain in Spain"
cat /etc/hosts | wc -l
echo "falls mainly on the plain"
`[1:],
	}
}

var (
	tmplTestBody = `
{{define "` + testutil.TmplTestName + `"}}
<html><head>
<style>
body {
  background-color: #202020;
}
` + common.Css + `
` + codeblock.Css + `
</style>
<script type="text/javascript">
` + common.Js + `
` + codeblock.Js + `
let cbc = null;
function onLoad() {
  cbc = new CodeBlockController({{.Id}});
  cbc.reset();
  cbc.addOnClick(()=>{ cbc.toggle(); });
}
</script>
</head>
<body onload='onLoad()'>
<p> before code block </p>
{{ template "` + codeblock.TmplName + `" . }}
<p> after code block </p>
</body></html>
{{end}}
`
)
