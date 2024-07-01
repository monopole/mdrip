package navtop_test

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/appstate"
	burgerbars "github.com/monopole/mdrip/v2/internal/web/app/widget/burgerbars1"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/common"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/helpbutton"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/navtop"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/session"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/testutil"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/timeline"
	"html/template"
	"testing"
)

func TestWidget(t *testing.T) {
	testutil.RenderHtmlToFile(
		t, navtop.AsTmpl()+tmplTestBody,
		makeParams("85", testutil.MakeAppStateTest0()))
}

func makeParams(tlId string, as *appstate.AppState) any {
	atp := struct {
		common.ParamStructJsCss
		AppState    *appstate.AppState
		TimelineId  string
		BurgerBars  template.HTML
		HelpButton  template.HTML
		TimelineRow template.HTML
	}{
		ParamStructJsCss: common.ParamDefaultJsCss,
		AppState:         as,
		TimelineId:       tlId,
	}
	atp.BurgerBars = common.MustRenderHtml(
		burgerbars.AsTmpl(), atp, burgerbars.TmplName)
	atp.HelpButton = common.MustRenderHtml(
		helpbutton.AsTmpl(), atp, helpbutton.TmplName)
	atp.TimelineRow = common.MustRenderHtml(
		timeline.AsTmpl(), atp, timeline.TmplName)
	return atp
}

var (
	tmplTestBody = `
{{define "` + testutil.TmplTestName + `"}}
<html><head>
<style>
` + common.Css + `
` + helpbutton.Css + `
` + burgerbars.Css + `
` + timeline.Css + `
` + navtop.Css + `
</style>
<script type="text/javascript">
` + common.Js + `
` + session.Js + `
` + appstate.Js + `
` + burgerbars.Js + `
` + helpbutton.Js + `
` + timeline.Js + `
` + navtop.Js + `
` + testutil.Js + `
function onLoad() {
  as = tstMakeAppState()
  let hbc = new HelpButtonController()
  hbc.onClick(()=>{console.log("helpButton clicked.");})
  let bbc = new BurgerBarsController()
  bbc.onClick(()=>{console.log("burgerBars clicked.");})

  let tlc = new TimelineController(as, {{.TimelineId}})
  let tnc = new NavTopController(as, tlc);
  window.addEventListener('keydown', function (event) {
    if (event.defaultPrevented) {
      return;
    }
    switch (event.key) {
      case 'n':
        as.toggleNav();
        break;
      case '-':
        as.toggleTitle();
        break;
      default:
    }
  }, false);
  as.zero();
}
</script>
</head>
<body onload='onLoad()'>
{{ template "` + navtop.TmplName + `" . }}
</body></html>
{{end}}
`
)
