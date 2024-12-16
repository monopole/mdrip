package navleftroot_test

import (
	"bytes"
	_ "embed"
	"html/template"
	"os"
	"testing"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/appstate"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/common"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/navleftfile"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/navleftfolder"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/navleftroot"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/session"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/testutil"
	"github.com/spf13/afero"
)

func TestWidget(t *testing.T) {
	testutil.RenderHtmlToFile(
		t, navleftroot.AsTmpl()+tmplTestBody,
		makeParams(loader.NewTopFolder(testutil.MakeFolderTreeOfMarkdown())))
}

func TestWidget2(t *testing.T) {
	if !runTheUnportableLocalFileSystemDependentTests {
		t.Skip("skipping non-portable tests")
	}
	f, err := loader.New(
		afero.NewOsFs(),
		loader.IsMarkDownFile,
		loader.InNotIgnorableFolder).LoadOneTree(
		loader.FilePath(
			"/home/" + os.Getenv("USER") + "/myrepos/github.com/sigs.k8s.io/kustomize"))
	if err != nil {
		t.Fatal(err)
	}
	testutil.RenderHtmlToFile(
		t, navleftroot.AsTmpl()+tmplTestBody, makeParams(f))
}

func makeParams(folder loader.MyTreeNode) any {
	atp := struct {
		common.ParamStructJsCss
		AppState    *appstate.AppState
		NavLeftRoot template.HTML
	}{
		ParamStructJsCss: common.ParamDefaultJsCss,
	}
	numFolders := 0
	{
		var b bytes.Buffer
		v := navleftroot.NewRenderer(&b)
		folder.Accept(v)
		numFolders = v.NumFolders()
		atp.NavLeftRoot = template.HTML(b.String())
	}
	{
		as := testutil.MakeAppStateTest1(folder)
		as.Facts.NumFolders = numFolders
		atp.AppState = as
	}
	return &atp
}

var (
	tmplTestBody = `
{{define "` + testutil.TmplTestName + `"}}
<html><head>
<style>
` + common.Css + `
` + navleftfile.Css + `
` + navleftfolder.Css + `
` + navleftroot.Css + `
</style>
<script type="text/javascript">
` + common.Js + `
` + session.Js + `
` + appstate.Js + `
` + navleftfile.Js + `
` + navleftfolder.Js + `
` + navleftroot.Js + `
` + testutil.Js + `
let as = null;
let nlc = null;
function onLoad() {
  as = tstMakeAppState()
  nlc = new NavLeftRootController(as)
  as.addFileChangeReactor(new TstReactor(as));
}
</script>
</head>
<body onload='onLoad()'>
{{ template "` + navleftroot.TmplName + `" . }}
</body></html>
{{end}}
`
)
