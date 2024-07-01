package minify

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	textTmpl "text/template"

	"github.com/monopole/mdrip/v2/internal/web/app"
	"github.com/monopole/mdrip/v2/internal/web/app/widget/common"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
)

type TmplArgs struct {
	Name   string
	Body   string
	Params any
}

type Args struct {
	MimeType string
	Tmpl     TmplArgs
}

type Minifier struct {
	minifier *minify.M
	doMinify bool
}

func MakeMinifier() *Minifier {
	result := &Minifier{
		minifier: minify.New(),
		doMinify: true, /* for debugging */
	}
	result.minifier.AddFunc(app.MimeJs, js.Minify)
	result.minifier.AddFunc(app.MimeCss, css.Minify)
	return result
}

func (mn *Minifier) Write(wr http.ResponseWriter, args *Args) {
	var (
		err  error
		tmpl *textTmpl.Template
	)
	// Parsing Js (or CSS) as 'html' replaces "i < 2" with "i &lt; 2".
	// Parse as 'text' to avoid this.
	// This isn't solvable with template.Js, because we're _inflating_ a
	// template full of Js, not _injecting_ known Js into some template.
	tmpl, err = common.ParseAsTextTemplate(args.Tmpl.Body)
	if err != nil {
		write500(wr, fmt.Errorf("%s parse fail; %w", args.Tmpl.Name, err))
		return
	}
	wr.Header().Set("Content-Type", args.MimeType)
	if mn.doMinify {
		if err = mn.minify(wr, tmpl, args); err != nil {
			write500(wr, err)
			return
		}
		slog.Info(args.Tmpl.Name + " minified success")
		return
	}
	err = tmpl.ExecuteTemplate(wr, args.Tmpl.Name, args.Tmpl.Params)
	if err != nil {
		write500(wr, fmt.Errorf("failed to inflate %s; %w", args.Tmpl.Name, err))
		return
	}
	slog.Info(args.Tmpl.Name + " success")
}

func (mn *Minifier) minify(
	wr http.ResponseWriter, tmpl *textTmpl.Template, args *Args) error {
	// There's probably some man-in-the-middle way to do this to skip
	// using "buff" and "ugly".
	var (
		buff bytes.Buffer
		ugly []byte
	)
	err := tmpl.ExecuteTemplate(&buff, args.Tmpl.Name, args.Tmpl.Params)
	if err != nil {
		return fmt.Errorf("tmpl %s inflate fail; %w", args.MimeType, err)
	}
	ugly, err = mn.minifier.Bytes(args.MimeType, buff.Bytes())
	if err != nil {
		return fmt.Errorf("%s minification fail; %w", args.MimeType, err)
	}
	if _, err = wr.Write(ugly); err != nil {
		return fmt.Errorf("write of %s failed; %w", args.MimeType, err)
	}
	return nil
}

func write500(w http.ResponseWriter, e error) {
	slog.Error(e.Error())
	http.Error(w, e.Error(), http.StatusInternalServerError)
}
