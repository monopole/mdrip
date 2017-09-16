package model

import (
	"fmt"
	"github.com/russross/blackfriday"
	"html/template"
	"io"
)

// opaqueCode is an opaque, uninterpreted, unknown block of text that
// is presumably shell commands parsed from markdown.  Fed into a
// shell interpreter, the entire thing either succeeds or fails.
type opaqueCode string

func (c opaqueCode) String() string {
	return string(c)
}

func (c opaqueCode) Bytes() []byte {
	return []byte(c)
}

// CommandBlock groups opaqueCode with its labels.
type CommandBlock struct {
	labels []Label
	code   opaqueCode
	// prose is human language documentation for the opaqueCode
	prose string
}

const (
	TmplNameCommandBlock = "commandblock"
	TmplBodyCommandBlock = `
{{define "` + TmplNameCommandBlock + `"}}
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

func NewCommandBlock(labels []Label, code, prose string) *CommandBlock {
	if len(labels) < 1 {
		// Assure at least one label.
		labels = []Label{Label("unknown")}
	}
	return &CommandBlock{labels, opaqueCode(code), prose}
}

// GetName returns the name of the command block.
//
// It's always the first label, and construction assures there will be
// at least one.
func (x CommandBlock) Name() Label {
	return x.labels[0]
}

func (x CommandBlock) Labels() []Label {
	return x.labels
}

func (x CommandBlock) Code() opaqueCode {
	return x.code
}

func (x CommandBlock) Prose() template.HTML {
	return template.HTML(string(blackfriday.MarkdownCommon([]byte(x.prose))))
}

func (x CommandBlock) Print(
	w io.Writer, prefix string, n int, label Label, fileName FilePath) {
	fmt.Fprintf(w, "echo \"%s @%s (block #%d in %s) of %s\"\n\n",
		prefix, x.Name(), n, label, fileName)
	fmt.Fprint(w, x.Code())
}
