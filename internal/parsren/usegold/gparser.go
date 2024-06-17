package usegold

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"strings"

	"go.abhg.dev/goldmark/mermaid"

	"github.com/monopole/mdrip/v2/internal/loader"
	"github.com/monopole/mdrip/v2/internal/parsren"
	"github.com/monopole/mdrip/v2/internal/webapp/widget/codeblock"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// GParser implements MdParserRenderer
var _ parsren.MdParserRenderer = &GParser{}

// GParser is a MyFolder tree visitor that both parses and renders markdown.
// It uses the goldmark parser/renderer to do both.
type GParser struct {
	// currentFile is an ephemeral state variable used during folder visitation.
	currentFile *loader.MyFile

	// p holds the actual parser and rendered, capable of handling
	// one file at a time.
	p goldmark.Markdown

	// err is the error encountered while parsing.
	err error

	// renderMdFiles holds all the HTML rendered markdown files.
	// The renderings have <h>, <p> etc. but no <html>,
	// <head> or <body> tags; such structure must be provided
	// by some containing web application.
	// The renderMdFiles also contain any extracted code blocks.
	renderMdFiles []*parsren.RenderedMdFile
}

func NewGParser() *GParser {
	gp := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			&mermaid.Extender{},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			// html.WithHardWraps(),
			// html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
	const priority = 100
	gp.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&codeblock.MyFcbRenderer{}, priority)),
	)
	return &GParser{
		p: gp,
	}
}

func (v *GParser) Reset() {
	v.err = nil
	v.renderMdFiles = nil
}

func (v *GParser) Error() error {
	return v.err
}

func (v *GParser) RenderedMdFiles() []*parsren.RenderedMdFile {
	return v.renderMdFiles
}

// FilteredBlocks returns a slice of filtered code blocks from the entire tree.
func (v *GParser) FilteredBlocks(
	l loader.Label) (result []*loader.CodeBlock) {
	for _, file := range v.renderMdFiles {
		for _, b := range file.Blocks {
			if b.HasLabel(l) {
				result = append(result, b)
			}
		}
	}
	return result
}

func (v *GParser) VisitTopFolder(fl *loader.MyTopFolder) {
	fl.VisitChildren(v)
}

func (v *GParser) VisitFolder(fl *loader.MyFolder) {
	fl.VisitChildren(v)
}

func (v *GParser) VisitFile(fi *loader.MyFile) {
	v.currentFile = fi
	// node is the root of an abstract syntax tree discovered by
	// parsing the file content.
	// node cannot be used alone; it holds pointers into the
	// file's byte array, rather than actually holding a copy
	// of the bytes.
	node := v.p.Parser().Parse(text.NewReader(fi.C()))

	fencedBlocks, err := gatherMyFencedCodeBlocks(node)
	if err != nil && v.err == nil {
		v.err = err
	}

	var blocks []*loader.CodeBlock
	for i, fcb := range fencedBlocks {
		cb := v.convertFcbToCb(fcb, i)
		// Store the zero-relatives index as node attributes
		// in the syntax tree for later use in rendering
		// div 'id' or 'data-' attributes.
		fcb.SetFileIndex(len(v.renderMdFiles))
		fcb.SetBlockIndex(len(blocks))
		fcb.SetTitle(string(cb.FirstLabel()))
		blocks = append(blocks, cb)
	}
	rf := &parsren.RenderedMdFile{
		Index: len(v.renderMdFiles),
		// One cannot render the file until _after_ the above loop that
		// sets attributes on the fenced code blocks.
		Html:   v.renderMdFile(fi, node),
		Path:   fi.Path(),
		Blocks: blocks,
	}
	v.renderMdFiles = append(v.renderMdFiles, rf)
}

func gatherMyFencedCodeBlocks(n ast.Node) (
	result []*codeblock.MyFencedCodeBlock, err error) {
	var originals []*ast.FencedCodeBlock
	err = ast.Walk(
		n,
		func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if !entering {
				return ast.WalkContinue, nil
			}
			if n.Kind() == ast.KindFencedCodeBlock {
				fcb, ok := n.(*ast.FencedCodeBlock)
				if !ok {
					return ast.WalkStop, fmt.Errorf(
						"ast.Kind() appears to be confused")
				}
				if !parentIsBlockQuote(n) {
					originals = append(originals, fcb)
				}
			}
			return ast.WalkContinue, nil
		})
	result = make([]*codeblock.MyFencedCodeBlock, len(originals))
	for i := range originals {
		// The following messes with the AST, so we don't want to
		// do it *during* the Walk, only *after* the Walk.
		result[i] = swapOutTheirCodeBlockForMine(originals[i])
	}
	return
}

func swapOutTheirCodeBlockForMine(n *ast.FencedCodeBlock) *codeblock.MyFencedCodeBlock {
	mine := codeblock.NewMyFencedCodeBlock(n)
	n.Parent().ReplaceChild(n.Parent(), n, mine)
	return mine
}

func parentIsBlockQuote(n ast.Node) bool {
	return n.Parent() != nil && n.Parent().Kind() == ast.KindBlockquote
}

func (v *GParser) renderMdFile(
	fi *loader.MyFile, node ast.Node) template.HTML {
	var buf bytes.Buffer
	if err := v.p.Renderer().Render(&buf, fi.C(), node); err != nil {
		slog.Error("render fail", "file", fi.Path(), "err", err.Error())
		if v.err == nil {
			// Save the first error, but keep going.
			v.err = err
		}
	}
	return template.HTML(buf.String())
}

func (v *GParser) convertFcbToCb(
	fcb *codeblock.MyFencedCodeBlock, index int) *loader.CodeBlock {
	cb := loader.NewCodeBlock(
		v.currentFile, v.nodeText(fcb), index,
		string(fcb.Language(v.currentFile.C())))
	if prev := fcb.PreviousSibling(); prev != nil && prev.Kind() == ast.KindHTMLBlock {
		if block, ok := prev.(*ast.HTMLBlock); ok {
			// We have a preceding HTML block.
			// If it's an HTML comment, try to extract labels.
			// If no labels found, the label array remains empty,
			// i.e. no label defaults are actually stored here.
			cb.AddLabels(
				loader.ParseLabels(loader.CommentBody(v.nodeText(block))))
		}
	}
	return cb
}

// TODO: Could change this to preserve lines?
func (v *GParser) nodeText(n ast.Node) string {
	var buff strings.Builder
	for i := 0; i < n.Lines().Len(); i++ {
		s := n.Lines().At(i)
		buff.Write(v.currentFile.C()[s.Start:s.Stop])
	}
	return buff.String()
}
