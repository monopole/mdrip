package navleftroot

import (
	"bytes"
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/navlefttopfolder"
	"html/template"
	"strings"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/navleftfile"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/navleftfolder"
)

var (
	baseAtp = struct {
		common.ParamStructTransition
		ObjectId int
		FileName string
		FilePath string
		Children template.HTML
	}{
		ParamStructTransition: common.ParamDefaultTransition,
	}
	tmplFile      = common.MustHtmlTemplate(navleftfile.AsTmpl())
	tmplFolder    = common.MustHtmlTemplate(navleftfolder.AsTmpl())
	tmplTopFolder = common.MustHtmlTemplate(navlefttopfolder.AsTmpl())
)

const tryTopFolderHack = false

// Renderer renders left nav HTML to a Writer.
type Renderer struct {
	buff        *bytes.Buffer
	err         error
	indexFolder int
	indexFile   int
	name        []string
}

// NewRenderer returns a new Renderer for the given writer.
func NewRenderer(buff *bytes.Buffer) *Renderer {
	return &Renderer{
		buff:        buff,
		indexFolder: -1,
		indexFile:   -1,
		name:        make([]string, 0),
	}
}

func (v *Renderer) NumFiles() int {
	return v.indexFile + 1
}

func (v *Renderer) NumFolders() int {
	return v.indexFolder + 1
}

func (v *Renderer) Error() error {
	return v.err
}

func (v *Renderer) path() string {
	if len(v.name) > 1 && v.name[0] == string(loader.CurrentDir) {
		return strings.Join(v.name[1:], string(loader.RootSlash))
	}
	return strings.Join(v.name, string(loader.RootSlash))
}

// VisitFile renders a file nav widget, with ID matching the depth first
// file ordering.
func (v *Renderer) VisitFile(x *loader.MyFile) {
	v.indexFile++
	v.name = append(v.name, x.Name())
	atp := baseAtp
	atp.ObjectId = v.indexFile
	atp.FileName = strings.TrimSuffix(x.Name(), ".md")
	atp.FilePath = v.path()

	v.err = tmplFile.ExecuteTemplate(v.buff, navleftfile.TmplName, atp)
	if v.err != nil {
		return
	}

	v.name = v.name[:len(v.name)-1]
}

// VisitTopFolder renders the top-most folder.
func (v *Renderer) VisitTopFolder(x *loader.MyTopFolder) {
	params := struct {
		Children template.HTML
	}{}
	safe := v.buff
	v.buff = &bytes.Buffer{}
	x.VisitChildren(v)
	params.Children = template.HTML(v.buff.String())
	v.buff = safe
	v.err = tmplTopFolder.ExecuteTemplate(v.buff, navlefttopfolder.TmplName, params)
	if v.err != nil {
		return
	}
}

// VisitFolder renders a folder nav widget, with ID matching the depth first
// folder ordering.
func (v *Renderer) VisitFolder(x *loader.MyFolder) {
	v.indexFolder++
	v.name = append(v.name, x.Name())
	atp := baseAtp
	atp.ObjectId = v.indexFolder
	atp.FileName = x.Name()
	atp.FilePath = v.path()
	{
		safe := v.buff
		v.buff = &bytes.Buffer{}
		x.VisitChildren(v)
		atp.Children = template.HTML(v.buff.String())
		v.buff = safe
	}

	v.err = tmplFolder.ExecuteTemplate(v.buff, navleftfolder.TmplName, atp)
	if v.err != nil {
		return
	}
	v.name = v.name[:len(v.name)-1]
}
