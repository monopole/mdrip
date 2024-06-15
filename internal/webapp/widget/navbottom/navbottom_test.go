package navbottom_test

import (
	_ "embed"
	"html/template"
	"testing"

	"github.com/monopole/mdrip/v2/internal/webapp/widget/appstate"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/helpbutton"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/navbottom"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/session"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/testutil"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/timeline"
)

func TestWidget(t *testing.T) {
	testutil.RenderHtmlToFile(
		t, navbottom.AsTmpl()+tmplTestBody,
		makeParams("888", testutil.MakeAppStateTest0()))
}

func makeParams(tlId string, as *appstate.AppState) any {
	atp := struct {
		common.ParamStructSession
		common.ParamStructTransition
		AppState    *appstate.AppState
		TimelineId  string
		HelpButton  template.HTML
		TimelineRow template.HTML
	}{
		ParamStructSession:    common.ParamDefaultSession,
		ParamStructTransition: common.ParamDefaultTransition,
		AppState:              as,
		TimelineId:            tlId,
	}
	atp.HelpButton = common.MustRenderHtml(
		helpbutton.AsTmpl(), atp, helpbutton.TmplName)
	atp.TimelineRow = common.MustRenderHtml(
		timeline.AsTmpl(), atp, timeline.TmplName)
	return &atp
}

var (
	tmplTestBody = `
{{define "` + testutil.TmplTestName + `"}}
<html><head>
<style>
` + common.Css + `
` + helpbutton.Css + `
` + timeline.Css + `
` + navbottom.Css + `
</style>
<script type="text/javascript">
` + common.Js + `
` + session.Js + `
` + appstate.Js + `
` + helpbutton.Js + `
` + timeline.Js + `
` + navbottom.Js + `
` + testutil.Js + `
function onLoad() {
  as = tstMakeAppState()
  let hbc = new HelpButtonController()
  hbc.onClick(()=>{console.log("helpButton clicked.");})
  let tlc = new TimelineController(as, {{.TimelineId}})
  as.zero();
}
</script>
</head>
<body onload='onLoad()'>
{{ template "` + navbottom.TmplName + `" . }}
</body></html>
{{end}}
`
)
