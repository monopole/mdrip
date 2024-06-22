package navcontentrow_test

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/appstate"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/session"
	"testing"

	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/navcontentrow"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/testutil"
)

func TestWidget(t *testing.T) {
	testutil.RenderHtmlToFile(t, navcontentrow.AsTmpl()+tmplTestBody, makeParams())
}

func makeParams() any {
	return struct {
		common.ParamStructJsCss
		AppState *appstate.AppState
		navcontentrow.ParamStructContentRow
	}{
		common.ParamDefaultJsCss,
		testutil.MakeAppStateTest0(),
		navcontentrow.ParamStructContentRow{
			ContentTop:    testutil.FillerDiv("The Canopy"),
			ContentLeft:   testutil.FillerDiv("Californians"),
			ContentCenter: testutil.LoremIpsum(20),
			ContentRight:  testutil.FillerDiv("Smart Folks"),
			ContentBottom: testutil.FillerDiv("Davy Jones's Locker"),
		},
	}
}

var (
	tmplTestBody = `
{{define "` + testutil.TmplTestName + `"}}
<html><head>
<style>
` + common.Css + `
` + testutil.Css /* CSS for text Filler div */ + `
` + navcontentrow.Css + `
</style>
<script type="text/javascript">
` + common.Js + `
` + session.Js + `
` + appstate.Js + `
` + navcontentrow.Js + `
` + testutil.Js + `
let as = null;
function onLoad() {
  as = tstMakeAppState()
  let ctr = new NavigatedContentRowController(as)
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
{{ template "` + navcontentrow.TmplName + `" . }}
</body></html>
{{end}}
`
)
