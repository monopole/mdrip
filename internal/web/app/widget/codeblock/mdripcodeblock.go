package codeblock

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/monopole/mdrip/v2/internal/web/app/widget/common"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

var (
	// kindMdRipCodeBlock is a NodeKind of the MdRipCodeBlock node.
	kindMdRipCodeBlock = ast.NewNodeKind("MdRipCodeBlock")

	codeBlockTemplate = common.MustHtmlTemplate(AsTmpl())
)

// A MdRipCodeBlock struct represents a fenced code block of Markdown text.
//
// Sometimes we want to render a fenced code block using the normal default
// fenced code block rendered, e.g. when a code block is "protected" inside a
// blockquote.
//
// But sometimes we want to do a special rendering, that encourages copy/paste,
// execution, etc.
//
// When we do want a special rendering, we replace the AST node instance of a
// native FencedCodeBlock with an instance of MdRipCodeBlock.  The latter
// must register its own Kind and renderer with the goldmark infrastructure.
//
// FWIW, in goldMark's interfaces, there doesn't seem to be a way to call a
// "native" renderer (in particular, the FencedCodeBlock renderer)
// from a custom renderer function, because a customer renderer
// function cannot see
//   - its own encapsulating renderer at goldmark/renderer/renderer.go:L135
//   - the global renderer function lookup table
//     held inside the encapsulating renderer
//   - the private "renderFencedCodeBlock" function
//     near goldmark/renderer/html/html.go:L386
//
// Annoyingly, since we replace the FencedCodeBlock with MdRipCodeBlock,
// the goldmark code highlighter won't be invoked.
//
// TODO: rebuild the AST so that instead of replacing a FencedCodeBLock
//
//	with a MdRipCodeBlock, we nest the FencedCodeBlock inside the
//	MdRipCodeBlock.
type MdRipCodeBlock struct {
	ast.FencedCodeBlock
}

// NewMdRipCodeBlock return a new MdRipCodeBlock AST node.
func NewMdRipCodeBlock(fcb *ast.FencedCodeBlock) *MdRipCodeBlock {
	return &MdRipCodeBlock{*fcb}
}

// Kind implements Node.Kind.
func (n *MdRipCodeBlock) Kind() ast.NodeKind {
	return kindMdRipCodeBlock
}

// render renders a MdRipCodeBlock with the id and styling elements needed
// get something that both looks like a terminal and is properly
// hooked up to the "copy and post for execution" javascript.
func (n *MdRipCodeBlock) render(w util.BufWriter, source []byte) {
	id, err := n.GetBlockIndex()
	if err != nil {
		_, _ = w.WriteString(err.Error())
		return
	}
	title, err := n.GetTitle()
	if err != nil {
		_, _ = w.WriteString(err.Error())
		return
	}
	templateInput := &struct {
		CbPrompt template.HTML
		Id       int
		Title    string
		Code     string
	}{
		CbPrompt: CbPrompt,
		Id:       id,
		Title:    title,
		Code:     n.grabNodeText(source),
	}
	if err = codeBlockTemplate.ExecuteTemplate(
		w, TmplName, templateInput); err != nil {
		panic(err)
	}
}

func (n *MdRipCodeBlock) grabNodeText(source []byte) string {
	var buff strings.Builder
	numLines := n.Lines().Len()
	for i := 0; i < numLines; i++ {
		line := n.Lines().At(i)
		_, _ = buff.Write(line.Value(source))
	}
	return buff.String()
}

// These fields are used by the parser to store parsed data into the
// abstract syntax tree during parsing.  A later rendering pass through
// the AST can then use the values in these fields.
const (
	attrBlockIdx   = "attrBlockIdx"
	attrFileIdx    = "attrFileIdx"
	attrBlockTitle = "attrBlockTitle"
)

func (n *MdRipCodeBlock) SetFileIndex(i int) {
	n.SetAttributeString(attrFileIdx, i)
}

func (n *MdRipCodeBlock) SetBlockIndex(i int) {
	n.SetAttributeString(attrBlockIdx, i)
}

func (n *MdRipCodeBlock) GetBlockIndex() (int, error) {
	return n.getIntAttribute(attrBlockIdx)
}

func (n *MdRipCodeBlock) SetTitle(s string) {
	n.SetAttributeString(attrBlockTitle, s)
}

func (n *MdRipCodeBlock) GetTitle() (string, error) {
	return n.getStrAttribute(attrBlockTitle)
}

func (n *MdRipCodeBlock) getIntAttribute(name string) (int, error) {
	tmp, err := n.getRawAttribute(name)
	if err != nil {
		return 0, err
	}
	result, ok := tmp.(int)
	if !ok {
		return 0, fmt.Errorf(
			"unable to cast int fencedCodeBlock attr %q", name)
	}
	return result, nil
}

func (n *MdRipCodeBlock) getStrAttribute(name string) (string, error) {
	tmp, err := n.getRawAttribute(name)
	if err != nil {
		return "", err
	}
	result, ok := tmp.(string)
	if !ok {
		return "", fmt.Errorf(
			"unable to cast string fencedCodeBlock attr %q", name)
	}
	return result, nil
}

func (n *MdRipCodeBlock) getRawAttribute(name string) (any, error) {
	tmp, ok := n.AttributeString(name)
	if !ok {
		return "", fmt.Errorf(
			"unable to parse fencedCodeBlock attr %q", name)
	}
	return tmp, nil
}
