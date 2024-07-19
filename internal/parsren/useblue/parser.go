package useblue

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type gomark struct {
	doMyStuff bool
	p         *parser.Parser
	doc       ast.Node
}

func (gm *gomark) Load(_ []byte) error {
	// panic("TODO: get this to work")
	gm.doc = gm.p.Parse(nil)
	myWalk(gm.doc)
	return nil
}

func (gm *gomark) Render() (string, error) {
	opts := html.RendererOptions{
		Flags: html.CommonFlags | html.HrefTargetBlank,
	}
	if gm.doMyStuff {
		opts.RenderNodeHook = myRenderHook
	}
	renderer := html.NewRenderer(opts)
	return string(markdown.Render(gm.doc, renderer)), nil
}

func (gm *gomark) Dump() {
	ast.PrintWithPrefix(os.Stdout, gm.doc, "  ")
}

func NewMarker(doMyStuff bool) *gomark {
	p := parser.NewWithExtensions(parser.CommonExtensions |
		parser.AutoHeadingIDs |
		parser.NoEmptyLineBeforeBlock |
		parser.Attributes)
	if doMyStuff {
		p.Opts.ParserHook = parserHook
	}
	return &gomark{p: p}
}

// parserHook is a custom parsren.
// If successful it returns an ast.Node containing the results of the parsing,
// a buffer that should be parsed as a block and added to the document (see below),
// and the number of bytes consumed (the guts of the parsren will skip over this).
// The buffer returned could be anything - e.g. data pulled from the web.
// Any nodes parsed from it will follow
// the ast.Node returned here at the same level, and not be a child to it.
// I think this is normally nil?
// It seems to be a way to inject data into the document.
func parserHook(data []byte) (ast.Node, []byte, int) {
	if node, d, n := attemptToParseGallery(data); node != nil {
		return node, d, n
	}
	return nil, nil, 0
}

func myWalk(doc ast.Node) {
	slog.Info("Walking...")
	ast.Walk(doc, &nodeVisitor{})
	slog.Info("Done Walking.")
}

type nodeVisitor struct {
	indent string
}

func (v *nodeVisitor) Visit(n ast.Node, entering bool) ast.WalkStatus {
	if !entering {
		return ast.GoToNext
	}
	// ast.Print recurses its argument, instead of just visiting
	// only the argument, so it's not what you want.
	// ast.Print(os.Stdout, n)
	leafLiteral := ""
	if n.AsLeaf() != nil {
		leafLiteral = string(n.AsLeaf().Literal)
	}
	slog.Info("visit", "nodeType", nodeType(n), "literal", leafLiteral)
	return ast.GoToNext
}

// get a short name of the type of v which excludes package name
// and strips "()" from the end
func nodeType(node ast.Node) string {
	s := fmt.Sprintf("%T", node)
	s = strings.TrimSuffix(s, "()")
	if idx := strings.Index(s, "."); idx != -1 {
		return s[idx+1:]
	}
	return s
}

func myRenderHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	switch node.(type) {
	case *ast.CodeBlock:
		if entering {
			_, _ = io.WriteString(w, "code_replacement\n")
		}
		return ast.GoToNext, true
	case *Gallery:
		if entering {
			_, _ = io.WriteString(w, "\n<gallery></gallery>\n\n")
		}
		return ast.GoToNext, true
	default:
		return ast.GoToNext, false
	}
}
