package navleftfolder_test

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/navleftfile"
	"testing"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/navleftfolder"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/testutil"
)

func TestWidget(t *testing.T) {
	f := loader.NewFolder("DIR_0").
		AddFile(loader.NewEmptyFile("FILE_0")).
		AddFile(loader.NewEmptyFile("FILE_1"))
	testutil.RenderHtmlToFile(
		t, navleftfolder.AsTmpl()+tmplTestBody,
		makeParams("55", f))
}

func makeParams(id string, f *loader.MyFolder) any {
	return struct {
		common.ParamStructJsCss
		ObjectId string
		FileName string
		FilePath loader.FilePath
		Children string
	}{
		common.ParamDefaultJsCss,
		id,
		f.Name(),
		f.Path(),
		"Hi there, we are your children.",
	}
}

var (
	tmplTestBody = `
{{define "` + testutil.TmplTestName + `"}}
<html><head>
<style>
` + common.Css + `
` + navleftfile.Css + `
` + navleftfolder.Css + `
</style>
<script type="text/javascript">
` + common.Js + `
` + navleftfile.Js + `
` + navleftfolder.Js + `
let nlc = null;
function onLoad() {
  nlc = new NavLeftFolderController({{.ObjectId}});
}
</script>
</head>
<body onload='onLoad()'>
{{ template "` + navleftfolder.TmplName + `" . }}
</body></html>
{{end}}
`
)
