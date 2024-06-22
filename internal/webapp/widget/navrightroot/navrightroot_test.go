package navrightroot_test

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/appstate"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/session"
	"testing"

	"github.com/monopole/mdrip/v2/internal/webapp/widget/codelabel"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/navrightroot"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/testutil"
)

func TestWidget(t *testing.T) {
	testutil.RenderHtmlToFile(
		t, codelabel.AsTmpl()+navrightroot.AsTmpl()+tmplTestBody,
		makeParams(testutil.MakeAppStateTest0()))
}

func makeParams(as *appstate.AppState) any {
	return struct {
		common.ParamStructJsCss
		AppState *appstate.AppState
	}{
		ParamStructJsCss: common.ParamDefaultJsCss,
		AppState:         as,
	}
}

var (
	tmplTestBody = `
{{define "` + testutil.TmplTestName + `"}}
<html><head>
<style>
` + common.Css + `
` + codelabel.Css + `
` + navrightroot.Css + `
</style>
<script type="text/javascript">
` + common.Js + `
` + session.Js + `
` + appstate.Js + `
` + codelabel.Js + `
` + navrightroot.Js + `
` + testutil.Js + `
let as = null;
let nrc = null;
function onLoad() {
  as = tstMakeAppState()
  nrc = new NavRightRootController(as)
  as.zero();
}
</script>
</head>
<body onload='onLoad()'>
{{ template "` + navrightroot.TmplName + `" . }}
</body></html>
{{end}}
`
)
