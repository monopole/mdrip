package mdfiles_test

import (
	"testing"

	"github.com/monopole/mdrip/v2/internal/webapp/widget/appstate"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/codeblock"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/mdfiles"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/session"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/testutil"
)

func TestWidget(t *testing.T) {
	testutil.RenderHtmlToFile(
		t, mdfiles.AsTmpl()+tmplTestBody,
		makeParams(testutil.MakeAppStateTest0()))
}

func makeParams(as *appstate.AppState) any {
	return struct {
		common.ParamStructJsCss
		AppState *appstate.AppState
	}{
		common.ParamDefaultJsCss,
		as,
	}
}

var (
	tmplTestBody = `
{{define "` + testutil.TmplTestName + `"}}
<html><head>
<style>
` + common.Css + `
` + codeblock.Css + `
` + mdfiles.Css + `
</style>
<script type="text/javascript">
` + common.Js + `
` + appstate.Js + `
` + session.Js + `
` + codeblock.Js + `
` + mdfiles.Js + `
` + testutil.Js + `
let as = null;
let mfc = null;
function onLoad() {
  as = tstMakeAppState()
  mfc = new MdFilesController(as);
  as.zero();
}
</script>
</head>
<body onload='onLoad()'>
{{ template "` + mdfiles.TmplName + `" . }}
</body></html>
{{end}}
`
)
