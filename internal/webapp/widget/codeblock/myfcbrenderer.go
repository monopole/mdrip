package codeblock

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// MyFcbRenderer is an implementation of renderer.NodeRenderer, but despite
// the name of the interface, instances of this don't actually render
// anything.  All an instance does is provide a method that registers
// a "Kind" with a rendering function.  The rendering function can live
// anywhere, and in this case it lives in an instance of MyFencedCodeBlock.
// This all seems odd, but it appears to work.
type MyFcbRenderer struct {
	Writer html.Writer
}

// Proof of interface implementation.
var _ renderer.NodeRenderer = &MyFcbRenderer{}

// RegisterFuncs implements NodeRenderer.RegisterFuncs.
func (r *MyFcbRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindMyFencedCodeBlock, callMyFencedCodeBlockRender)
}

func callMyFencedCodeBlockRender(
	w util.BufWriter, source []byte,
	node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		node.(*MyFencedCodeBlock).render(w, source)
	}
	return ast.WalkContinue, nil
}
