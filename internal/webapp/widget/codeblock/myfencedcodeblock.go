package codeblock

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/monopole/mdrip/v2/internal/webapp/widget/common"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

var (
	// kindMyFencedCodeBlock is a NodeKind of the MyFencedCodeBlock node.
	kindMyFencedCodeBlock = ast.NewNodeKind("MyFencedCodeBlock")

	codeBlockTemplate = common.MustHtmlTemplate(AsTmpl())
)

// A MyFencedCodeBlock struct represents a fenced code block of Markdown text.
//
// Sometimes we want to render a fenced code block using the normal default
// fenced code block rendered, e.g. when a code block is "protected" inside a
// blockquote.
//
// But more often we want to do a special rendering, that encourages copy/paste,
// execution, etc.
//
// When we do want a special rendering, we replace the AST node instance of a
// native FencedCodeBlock with an instance of MyFencedCodeBlock.  The latter
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
type MyFencedCodeBlock struct {
	ast.FencedCodeBlock
}

// NewMyFencedCodeBlock return a new MyFencedCodeBlock node.
func NewMyFencedCodeBlock(fcb *ast.FencedCodeBlock) *MyFencedCodeBlock {
	return &MyFencedCodeBlock{*fcb}
}

// Kind implements Node.Kind.
func (n *MyFencedCodeBlock) Kind() ast.NodeKind {
	return kindMyFencedCodeBlock
}

func (n *MyFencedCodeBlock) render(w util.BufWriter, source []byte) {
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

func (n *MyFencedCodeBlock) grabNodeText(source []byte) string {
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

func (n *MyFencedCodeBlock) SetFileIndex(i int) {
	n.SetAttributeString(attrFileIdx, i)
}

func (n *MyFencedCodeBlock) SetBlockIndex(i int) {
	n.SetAttributeString(attrBlockIdx, i)
}

func (n *MyFencedCodeBlock) GetBlockIndex() (int, error) {
	return n.getIntAttribute(attrBlockIdx)
}

func (n *MyFencedCodeBlock) SetTitle(s string) {
	n.SetAttributeString(attrBlockTitle, s)
}

func (n *MyFencedCodeBlock) GetTitle() (string, error) {
	return n.getStrAttribute(attrBlockTitle)
}

func (n *MyFencedCodeBlock) getIntAttribute(name string) (int, error) {
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

func (n *MyFencedCodeBlock) getStrAttribute(name string) (string, error) {
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

func (n *MyFencedCodeBlock) getRawAttribute(name string) (any, error) {
	tmp, ok := n.AttributeString(name)
	if !ok {
		return "", fmt.Errorf(
			"unable to parse fencedCodeBlock attr %q", name)
	}
	return tmp, nil
}
