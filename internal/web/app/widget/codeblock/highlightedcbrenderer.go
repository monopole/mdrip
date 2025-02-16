package codeblock

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// HighlightedCbRenderer is an implementation of renderer.NodeRenderer, but
// despite the name of the interface, instances of this interface don't actually
// render anything.
//
// All an instance does is provide a method that registers a "Kind" with a
// rendering function.  The rendering function can live anywhere, and in this
// case it lives in an instance of HighlightedCodeBlock.
type HighlightedCbRenderer struct {
	Writer html.Writer
}

// Proof of interface implementation.
var _ renderer.NodeRenderer = &HighlightedCbRenderer{}

// RegisterFuncs implements NodeRenderer.RegisterFuncs.
func (r *HighlightedCbRenderer) RegisterFuncs(
	reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindHighlightedCodeBlock, renderHighlightedCodeBlock)
}

// Proof of interface implementation.
var _ renderer.NodeRendererFunc = renderHighlightedCodeBlock

func renderHighlightedCodeBlock(
	w util.BufWriter, _ []byte,
	node ast.Node, entering bool) (ast.WalkStatus, error) {
	return node.(*HighlightedCodeBlock).render(w, entering)
}
