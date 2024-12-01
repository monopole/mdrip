package codeblock

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// JeffCbRenderer is an implementation of renderer.NodeRenderer, but despite
// the name of the interface, instances of this don't actually render
// anything.  All an instance does is provide a method that registers
// a "Kind" with a rendering function.  The rendering function can live
// anywhere, and in this case it lives in an instance of MdRipCodeBlock.
// This all seems odd, but it appears to work.
type JeffCbRenderer struct {
	Writer html.Writer
}

// Proof of interface implementation.
var _ renderer.NodeRenderer = &JeffCbRenderer{}

// RegisterFuncs implements NodeRenderer.RegisterFuncs.
func (r *JeffCbRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindJeffCodeBlock, callJeffCodeBlockRender)
}

// Proof of interface implementation.
var _ renderer.NodeRendererFunc = callJeffCodeBlockRender

func callJeffCodeBlockRender(
	w util.BufWriter, _ []byte,
	node ast.Node, entering bool) (ast.WalkStatus, error) {
	return node.(*JeffCodeBlock).render(w, entering)
}
