package helpbox_test

import (
	_ "embed"
	"testing"

	"github.com/monopole/mdrip/v2/internal/webapp/widget/appstate"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/helpbox"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/testutil"
)

func TestWidget(t *testing.T) {
	testutil.RenderHtmlToFile(
		t, helpbox.AsTmpl()+tmplTestBody, makeParams())
}

func makeParams() any {
	return struct {
		common.ParamStructJsCss
		AppState *appstate.AppState
	}{
		common.ParamDefaultJsCss,
		&appstate.AppState{
			InitialRender: appstate.InitialRender{
				DataSource: "/the/root/of/all/markdown",
			},
		},
	}
}

var (
	tmplTestBody = `
{{define "` + testutil.TmplTestName + `"}}
<html><head>
<style>
` + common.Css + `
` + helpbox.Css + `
</style>
<script type="text/javascript">
` + common.Js + `
` + helpbox.Js + `
class NavTopController {
  get height() {
    return '12rem';
  }
}
function onLoad() {
  let hbc = new HelpBoxController(new NavTopController())
  window.addEventListener('keydown', function (event) {
    if (event.defaultPrevented) {
      return;
    }
    switch (event.key) {
      case '/':
      case '?':
        hbc.toggle();
        break;
      default:
    }
  }, false);
}
</script>
</head>
<body onload='onLoad()'>
{{ template "` + helpbox.TmplName + `" . }}
</body></html>
{{end}}
`
)
