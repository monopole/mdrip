package session_test

import (
	_ "embed"
	"testing"

	"github.com/monopole/mdrip/v2/internal/webapp/widget/appstate"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/session"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/testutil"
)

func TestWidget(t *testing.T) {
	item := struct {
		common.ParamStructJsCss
		AppState *appstate.AppState
	}{
		common.ParamDefaultJsCss,
		testutil.MakeAppStateTest0(),
	}
	testutil.RenderHtmlToFile(t, tmplTestBody, item)
}

var (
	tmplTestBody = `
{{define "` + testutil.TmplTestName + `"}}
<html><head>
<script type="text/javascript">
` + session.Js + `
let sc = new SessionController({{.AppState.RenderedFiles}});
</script>
</head>
<body>
<p>Hello.</p>
</body></html>
{{end}}
`
)
