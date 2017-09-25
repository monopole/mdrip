package model

import (
	"fmt"
	"html/template"
	"io"

	"github.com/russross/blackfriday"
)

// CommandBlock groups opaqueCode with its labels.
type CommandBlock struct {
	labels []Label
	prose []byte
	code  OpaqueCode
}

func NewCommandBlock(labels []Label, prose []byte, code OpaqueCode) *CommandBlock {
	if !hasLabel(labels, AnyLabel) {
		labels = append(labels, AnyLabel)
	}
	return &CommandBlock{labels, prose, code}
}

func (x *CommandBlock) Accept(v TutVisitor)  { v.VisitCommandBlock(x) }
func (x *CommandBlock) Name() string         { return string(x.Labels()[0]) }
func (x *CommandBlock) Path() FilePath { return FilePath("notUsingThis") }
func (x *CommandBlock) Children() []Tutorial { return []Tutorial{} }
func (x *CommandBlock) HtmlProse() template.HTML {
	return template.HTML(string(blackfriday.MarkdownCommon(x.Prose())))
}
func (x *CommandBlock) Labels() []Label  { return x.labels }
func (x *CommandBlock) Prose() []byte    { return x.prose }
func (x *CommandBlock) Code() OpaqueCode { return x.code }

func (x *CommandBlock) HasLabel(label Label) bool {
	return xhasLabel(x.Labels(), label)
}

func xhasLabel(labels []Label, label Label) bool {
	for _, l := range labels {
		if l == label {
			return true
		}
	}
	return false
}
func (x *CommandBlock) Print(
	w io.Writer, prefix string, n int, label Label, fileName FilePath) {
	fmt.Fprintf(w, "echo \"%s @%s (block #%d in %s) of %s\"\n\n",
		prefix, x.Name(), n, label, fileName)
	fmt.Fprint(w, x.Code())
	// If the command block has a 'sleep' label, add a brief sleep at the end.
	// This hack gives servers placed in the background time to start, assuming
	// they can do so in the time added!  Yeah, bad.
	if x.HasLabel(SleepLabel) {
		fmt.Fprint(w, "sleep 3s # Added by mdrip\n")
	}
}
