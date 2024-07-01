package app

import (
	_ "embed"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/common"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/mdrip"
	"github.com/monopole/mdrip/v2/internal/web/config"
)

const (
	TmplName = "tmplWebApp"

	MimeJs  = "application/javascript"
	MimeCss = "text/css"

	// classlessCss = "https://cdn.jsdelivr.net/npm/water.css@2/out/dark.css"
	// classlessCss = "https://raw.githubusercontent.com/raj457036/attriCSS/master/themes/darkforest-green.css"
	//
	//   To load some "classless" css, throw in a line like
	//     <link rel="stylesheet" href="` + classlessCss + `">
	//   There are many classless css examples at
	//      https://github.com/dbohdan/classless-css
	//   Most of them mess with <body> and <pre>, screwing up mdrip's layout,
	//   but maybe copy a subset.
)

var (
	// Don't forget to set the content-type header if you use this.
	cssViaLink = `<link rel='stylesheet' type='` + MimeCss +
		`' href='` + config.Dynamic(config.RouteCss) + `' />`

	// Use this instead of cssViaLink to inject directly into the html response.
	cssInjected = `<style> ` + mdrip.AllCss + ` </style>`
)

var (
	html = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <title>{{.AppState.Title}}</title>
    ` + cssViaLink + `
    <script type='` + MimeJs + `' src='` + config.Dynamic(config.RouteJs) + `'></script>
    <script type='` + MimeJs + `'>
      function makeEmptyCache() {
        let c = new Array({{len .AppState.RenderedFiles}});
        for (let i = 0; i < c.length; i++) {
          c[i] = null;
        }
        return c;
      }
      // Define these outside onLoad to allow console access (debugging).
      let sc = null;
      let as = null;
      let nac = null;
      function onLoad() {
        sc = new SessionController(makeEmptyCache());
        as = new AppState(sc, {{.AppState.InitialRender}});
        nac = new MdRipController(as);
        sc.enable();
        // Load the initial (zeroth) file.
        as.loadCurrentFile(StartAt.Top, ActivateBlock.No);
      }
    </script>
  </head>
  <body onload='onLoad()'>
  {{template "` + mdrip.TmplNameHtml + `" .}}
  </body>
</html>
`
)

func AsTmpl() string {
	return mdrip.AsTmplHtml() + common.AsTmpl(TmplName, html)
}
