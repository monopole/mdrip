package codelabel_test

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/common"
	"testing"

	"github.com/monopole/mdrip/v2/internal/web/app/widget/codelabel"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/testutil"
)

func TestWidget(t *testing.T) {
	testutil.RenderHtmlToFile(
		t, codelabel.AsTmpl()+tmplTestBody, makeParams(222))
}

func makeParams(id int) any {
	return struct {
		common.ParamStructJsCss
		Id    int
		Label string
	}{
		common.ParamDefaultJsCss,
		id,
		"leetCode",
	}
}

var (
	tmplTestBody = `
{{define "` + testutil.TmplTestName + `"}}
<html><head>
<style>
` + common.Css + `
` + codelabel.Css + `
body {
  background-color: var(--color-lr-nav-background);
}
</style>
<script type="text/javascript">
` + common.Js + `
` + codelabel.Js + `
let clc = null;
function onLoad() {
  clc = new CodeLabelController({{.Id}});
  clc.onClick(()=>{clc.toggle()})
}
</script>
</head>
<body onload='onLoad()'>
{{ template "` + codelabel.TmplName + `" . }}
</body></html>
{{end}}
`
)
