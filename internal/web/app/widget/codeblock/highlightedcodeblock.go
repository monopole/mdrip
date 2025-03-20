package codeblock

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

var (
	// kindHighlightedCodeBlock is a NodeKind of the HighlightedCodeBlock node.
	kindHighlightedCodeBlock = ast.NewNodeKind("HighlightedCodeBlock")
)

// HighlightedCodeBlock represents a code block one might be able to run.
//
// Sometimes we want to render a fenced code block using the normal default
// fenced code block rendered, e.g. when a code block is "protected" inside a
// blockquote. Other times we want to do a special rendering, that encourages
// copy/paste and execution.
//
// In the latter case, we replace the AST node instance of a native
// FencedCodeBlock with an instance of HighlightedCodeBlock that has the
// FencedCodeBlock as its lone child in the AST. The HighlightedCodeBlock
// must register its own Kind and renderer with the goldmark infrastructure.
type HighlightedCodeBlock struct {
	ast.BaseBlock
	FileIndex  int
	BlockIndex int
	Title      string
}

// Dump implements Node.dump.
func (n *HighlightedCodeBlock) Dump(source []byte, level int) {
	m := map[string]string{
		"FileIndex":  fmt.Sprintf("%d", n.FileIndex),
		"BlockIndex": fmt.Sprintf("%d", n.BlockIndex),
		"Title":      fmt.Sprintf("%s", n.Title),
	}
	ast.DumpHelper(n, source, level, m, nil)
}

// Kind implements Node.Kind.
func (n *HighlightedCodeBlock) Kind() ast.NodeKind {
	return kindHighlightedCodeBlock
}

// render renders a HighlightedCodeBlock with the id and styling elements needed
// to get something that both looks like a terminal and is properly
// hooked up to the javascript that does a copy and POST back to the server.
func (n *HighlightedCodeBlock) render(
	w util.BufWriter, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString(
			fmt.Sprintf(`<div class='codeBlockContainer' id='codeBlockId%d'>
<div class='codeBlockControl'>
<span class='codeBlockTitle'> %s </span>
</div>
<div class='codeBlockPrompt'> %s </div>
<div class='codeBlockArea'>`, n.BlockIndex, n.Title, CbPrompt))
		return ast.WalkContinue, nil
	}
	_, _ = w.WriteString(`</div></div>`)
	return ast.WalkContinue, nil
}
