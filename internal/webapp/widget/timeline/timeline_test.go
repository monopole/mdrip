package timeline_test

import (
	_ "embed"
	"html/template"
	"testing"

	"github.com/monopole/mdrip/v2/internal/webapp/widget/appstate"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/helpbutton"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/session"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/testutil"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/timeline"
)

func TestWidget(t *testing.T) {
	testutil.RenderHtmlToFile(
		t, timeline.AsTmpl()+tmplTestBody,
		makeParams("888", testutil.MakeAppStateTest0()))
}

func makeParams(tlId string, as *appstate.AppState) any {
	atp := struct {
		common.ParamStructJsCss
		AppState   *appstate.AppState
		TimelineId string
		HelpButton template.HTML
	}{
		ParamStructJsCss: common.ParamDefaultJsCss,
		AppState:         as,
		TimelineId:       tlId,
	}
	atp.HelpButton = common.MustRenderHtml(
		helpbutton.AsTmpl(), atp, helpbutton.TmplName)
	return atp
}

var (
	tmplTestBody = `
{{define "` + testutil.TmplTestName + `"}}
<html><head>
<style>
` + common.Css + `
` + helpbutton.Css + `
` + timeline.Css + `
</style>
<script type="text/javascript">
` + common.Js + `
` + session.Js + `
` + appstate.Js + `
` + helpbutton.Js + `
` + timeline.Js + `
` + testutil.Js + `
function onLoad() {
  let as = tstMakeAppState();
  let tlc = new TimelineController(as, {{.TimelineId}})
  as.zero();
}
</script>
</head>
<body onload='onLoad()'>
<p>
{{ template "` + timeline.TmplName + `" . }}
</p>
</body></html>
{{end}}
`
)
