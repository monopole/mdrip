package burgerbars1_test

import (
	_ "embed"
	"testing"

	. "github.com/monopole/mdrip/v2/internal/webapp/widget/burgerbars1"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/testutil"
)

func TestWidget(t *testing.T) {
	testutil.RenderHtmlToFile(
		t, AsTmpl()+tmplTestBody, makeParams())
}

var (
	tmplTestBody = `
{{define "` + testutil.TmplTestName + `"}}
<html><head>
<style>
` + common.Css + `
` + Css + `
</style>
<script type="text/javascript">
` + common.Js + `
` + Js + `
function onLoad() {
  let bbc = new BurgerBarsController()
  bbc.onClick(()=>{console.log("burgerBars clicked.");})
}
</script>
</head>
<body onload='onLoad()'>
{{ template "` + TmplName + `" . }}
</body></html>
{{end}}
`
)

func makeParams() any {
	return struct {
		common.ParamStructTransition
	}{
		common.ParamDefaultTransition,
	}
}
