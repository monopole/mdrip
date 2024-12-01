package codeblock

import (
	"fmt"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

var (
	// kindJeffCodeBlock is a NodeKind of the JeffCodeBlock node.
	kindJeffCodeBlock = ast.NewNodeKind("JeffCodeBlock")
)

type JeffCodeBlock struct {
	ast.BaseBlock
	FileIndex  int
	BlockIndex int
	Title      string
}

// NewJeffCodeBlock return a new JeffCodeBlock AST node.
func NewJeffCodeBlock() *JeffCodeBlock {
	return &JeffCodeBlock{}
}

// Dump implements Node.Dump.
func (n *JeffCodeBlock) Dump(source []byte, level int) {
	m := map[string]string{
		"FileIndex":  fmt.Sprintf("%d", n.FileIndex),
		"BlockIndex": fmt.Sprintf("%d", n.BlockIndex),
		"Title":      fmt.Sprintf("%s", n.Title),
	}
	ast.DumpHelper(n, source, level, m, nil)
}

// Kind implements Node.Kind.
func (n *JeffCodeBlock) Kind() ast.NodeKind {
	return kindJeffCodeBlock
}

// render renders a JeffCodeBlock with the id and styling elements needed
// to get something that both looks like a terminal and is properly
// hooked up to the "copy and post for execution" javascript.
func (n *JeffCodeBlock) render(
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
