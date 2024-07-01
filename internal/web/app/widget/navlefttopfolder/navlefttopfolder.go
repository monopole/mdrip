package navlefttopfolder

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/common"
)

const (
	TmplName = "tmplNavLeftTopFolder"
)

var (
	//go:embed navlefttopfolder.css
	Css string
	//go:embed navlefttopfolder.html
	html string
)

func AsTmpl() string {
	return common.AsTmpl(TmplName, html)
}
