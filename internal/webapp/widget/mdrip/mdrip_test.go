package mdrip_test

import (
	_ "embed"
	"testing"

	"github.com/monopole/mdrip/v2/internal/parsren/usegold"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/mdrip"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/testutil"
)

func TestWidget(t *testing.T) {
	testutil.RenderHtmlToFile(
		t, mdrip.AsTmplHtml()+tmplTestBody,
		mdrip.MakeParams(mdrip.RenderFolder(
			&mdrip.RenderingArgs{
				Pr:         usegold.NewGParser(),
				DataSource: "/my/marky/markdown",
				Folder:     testutil.MakeFolderTreeOfMarkdown(),
				Title:      "You Only Live Twice",
			})))
}

const maxWordLen = 50

func TestJsRendering(t *testing.T) {
	testutil.RenderTextToFile(
		t, mdrip.AsTmplJs()+tmplTestBodyJsOnly,
		mdrip.MakeBaseParams(maxWordLen))
}

func TestCssRendering(t *testing.T) {
	testutil.RenderTextToFile(
		t, mdrip.AsTmplCss()+tmplTestBodyCssOnly,
		mdrip.MakeBaseParams(maxWordLen))
}

var (
	tmplTestBodyJsOnly = `
{{define "` + testutil.TmplTestName + `"}}
{{ template "` + mdrip.TmplNameJs + `" . }}
{{end}}
`
	tmplTestBodyCssOnly = `
{{define "` + testutil.TmplTestName + `"}}
{{ template "` + mdrip.TmplNameCss + `" . }}
{{end}}
`

	tmplTestBody = `
{{define "` + testutil.TmplTestName + `"}}
<!DOCTYPE html>
<html><head>
<style>
` + mdrip.AllCss + `
</style>
<script type="text/javascript">
` + mdrip.AllJs + `
` + testutil.Js + `
let as = null;
function onLoad() {
  as = tstMakeAppState()
  let nac = new MdRipController(as);
  as.zero();
}
</script>
</head>
<body onload='onLoad()'>
{{ template "` + mdrip.TmplNameHtml + `" . }}
</body></html>
{{end}}
`
)
