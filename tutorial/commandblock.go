package tutorial

import (
	"fmt"
	"html/template"
	"io"

	"github.com/monopole/mdrip/model"
	"github.com/russross/blackfriday"
)

// CommandBlock groups opaqueCode with its labels.
type CommandBlock struct {
	labels []model.Label
	// prose is human language documentation for the opaqueCode
	prose []byte
	code  model.OpaqueCode
}

func NewCommandBlock(labels []model.Label, prose []byte, code model.OpaqueCode) *CommandBlock {
	//if len(labels) < 1 {
	//	// Assure at least one label.
	//	labels = []model.Label{model.MistakeLabel}
	//}
	return &CommandBlock{labels, prose, code}
}

func (x *CommandBlock) Accept(v Visitor)       { v.VisitCommandBlock(x) }
func (x *CommandBlock) Name() string           { return string(x.labels[0]) }
func (x *CommandBlock) Path() model.FilePath   { return model.FilePath("notUsingThis") }
func (x *CommandBlock) Labels() []model.Label  { return x.labels }
func (x *CommandBlock) Code() model.OpaqueCode { return x.code }
func (x *CommandBlock) Children() []Tutorial   { return []Tutorial{} }
func (x *CommandBlock) RawProse() []byte       { return x.prose }
func (x *CommandBlock) Prose() template.HTML {
	return template.HTML(string(blackfriday.MarkdownCommon(x.prose)))
}
func (x *CommandBlock) HasLabel(label model.Label) bool {
	for _, l := range x.Labels() {
		if l == label {
			return true
		}
	}
	return false
}
func (x *CommandBlock) Print(
	w io.Writer, prefix string, n int, label model.Label, fileName model.FilePath) {
	fmt.Fprintf(w, "echo \"%s @%s (block #%d in %s) of %s\"\n\n",
		prefix, x.Name(), n, label, fileName)
	fmt.Fprint(w, x.Code())
}
