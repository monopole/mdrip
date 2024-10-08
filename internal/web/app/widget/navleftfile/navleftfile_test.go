package navleftfile_test

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/common"
	"testing"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/navleftfile"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/testutil"
)

func TestWidget(t *testing.T) {
	testutil.RenderHtmlToFile(
		t, navleftfile.AsTmpl()+tmplTestBody,
		makeParams("222", loader.NewEmptyFile("File0")))
}

func makeParams(id string, f *loader.MyFile) any {
	return struct {
		common.ParamStructJsCss
		ObjectId string
		FileName string
		FilePath loader.FilePath
	}{
		common.ParamDefaultJsCss,
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
