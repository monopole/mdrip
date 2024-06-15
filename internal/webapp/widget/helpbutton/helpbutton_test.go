package helpbutton_test

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
	"testing"

	"github.com/monopole/mdrip/v2/internal/webapp/widget/helpbutton"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/testutil"
)

func TestWidget(t *testing.T) {
	testutil.RenderHtmlToFile(
		t, helpbutton.AsTmpl()+tmplTestBody, makeParams())
}

func makeParams() any {
	return struct {
		common.ParamStructTransition
	}{
		common.ParamDefaultTransition,
	}
}

var (
	tmplTestBody = `
{{define "` + testutil.TmplTestName + `"}}
<html><head>
<style>
` + common.Css + `
` + helpbutton.Css + `
</style>
<script type="text/javascript">
` + common.Js + `
` + helpbutton.Js + `
function onLoad() {
  let hbc = new HelpButtonController(getDocElByClass("helpButton"))
  hbc.onClick(()=>{console.log("helpButton clicked.");})
}
</script>
</head>
<body onload='onLoad()'>
{{ template "` + helpbutton.TmplName + `" . }}
</body></html>
{{end}}
`
)
