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
	model.Block
}

func NewCommandBlock(labels []model.Label, prose []byte, code string) *CommandBlock {
	if !hasLabel(labels, model.AnyLabel) {
		labels = append(labels, model.AnyLabel)
	}
	return &CommandBlock{*model.NewBlock(labels, prose, code)}
}

func (x *CommandBlock) Accept(v Visitor)     { v.VisitCommandBlock(x) }
func (x *CommandBlock) Name() string         { return string(x.Labels()[0]) }
func (x *CommandBlock) Path() model.FilePath { return model.FilePath("notUsingThis") }
func (x *CommandBlock) Children() []Tutorial { return []Tutorial{} }
func (x *CommandBlock) HtmlProse() template.HTML {
	return template.HTML(string(blackfriday.MarkdownCommon(x.Prose())))
}
func (x *CommandBlock) HasLabel(label model.Label) bool {
	return hasLabel(x.Labels(), label)
}

func hasLabel(labels []model.Label, label model.Label) bool {
	for _, l := range labels {
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
	// If the command block has a 'sleep' label, add a brief sleep at the end.
	// This hack gives servers placed in the background time to start, assuming
	// they can do so in the time added!  Yeah, bad.
	if x.HasLabel(model.SleepLabel) {
		fmt.Fprint(w, "sleep 3s # Added by mdrip\n")
	}
}
