package navleftfile_test

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
	"testing"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/navleftfile"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/testutil"
)

func TestWidget(t *testing.T) {
	testutil.RenderHtmlToFile(
		t, navleftfile.AsTmpl()+tmplTestBody,
		makeParams("222", loader.NewEmptyFile("File0")))
}

func makeParams(id string, f *loader.MyFile) any {
	return struct {
		ObjectId string
		FileName string
		FilePath loader.FilePath
	}{
		id,
		f.Name(),
		f.Path(),
	}
}

var (
	tmplTestBody = `
{{define "` + testutil.TmplTestName + `"}}
<html><head>
<style>
` + common.Css + `
` + navleftfile.Css + `
</style>
<script type="text/javascript">
` + common.Js + `
` + navleftfile.Js + `
let nlc = null;
function onLoad() {
  nlc = new NavLeftFileController({{.ObjectId}});
}
</script>
</head>
<body onload='onLoad()'>
{{ template "` + navleftfile.TmplName + `" . }}
</body></html>
{{end}}
`
)
