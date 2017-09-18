package model

import (
	"fmt"
	"github.com/russross/blackfriday"
	"html/template"
	"io"
)

// OldBlock groups OpaqueCode with its labels.
type OldBlock struct {
	labels []Label
	code   OpaqueCode
	// prose is human language documentation for the OpaqueCode
	prose []byte
}

const (
	TmplNameOldBlock = "oldblock"
	TmplBodyOldBlock = `
{{define "` + TmplNameOldBlock + `"}}
<div class="proseblock"> {{.Prose}} </div>
<h3 id="control" class="control">
  <span class="blockButton" onclick="onRunBlockClick(event)">
     {{ .Name }}
  </span>
  <span class="spacer"> &nbsp; </span>
</h3>
<pre class="codeblock">
{{ .Code }}
</pre>
{{end}}
`
)

func NewOldBlock(labels []Label, code string, prose []byte) *OldBlock {
	if len(labels) < 1 {
		// Assure at least one label.
		labels = []Label{MistakeLabel}
	}
	return &OldBlock{labels, OpaqueCode(code), prose}
}

// GetName returns the name of the command block.
//
// It's always the first label, and construction assures there will be
// at least one.
func (x *OldBlock) Name() Label      { return x.labels[0] }
func (x *OldBlock) Path() FilePath   { return FilePath("wutwutwut") }
func (x *OldBlock) Labels() []Label  { return x.labels }
func (x *OldBlock) Code() OpaqueCode { return x.code }
func (x *OldBlock) RawProse() []byte { return x.prose }
func (x *OldBlock) Prose() template.HTML {
	return template.HTML(string(blackfriday.MarkdownCommon(x.prose)))
}

func (x *OldBlock) Print(
	w io.Writer, prefix string, n int, label Label, fileName FilePath) {
	fmt.Fprintf(w, "echo \"%s @%s (block #%d in %s) of %s\"\n\n",
		prefix, x.Name(), n, label, fileName)
	fmt.Fprint(w, x.Code())
}
