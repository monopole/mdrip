package parsren

import (
	"github.com/monopole/mdrip/v2/internal/loader"
	"html/template"
)

// MdParserRenderer is a tree visitor that parses and renders markdown.
// The two operations are closely coupled by a shared abstract syntax tree
// and shared raw bytes from the source markdown.
// Usage:
//   - Load markdown files into a tree of treeNode.
//   - Accept this visitor into the tree.
//   - After the call to Accept finishes, consult
//     RenderMdFiles() and/or FilteredBlocks() for whatever purpose.
type MdParserRenderer interface {
	loader.TreeVisitor
	// RenderedMdFiles is a slice of rendered markdown files
	// in depth-first order.
	RenderedMdFiles() []*RenderedMdFile
	// FilteredBlocks returns all blocks with the given label,
	// from all files in the tree.
	FilteredBlocks(loader.Label) []*loader.CodeBlock
	// Reset resets the parser.  Handy if you want to run another visitation,
	// and don't want data to accumulate.
	Reset()
}

type RenderedMdFile struct {
	// Index is a zero-relative unique integer ID of the file for whatever
	// purpose.
	Index int
	// Path is the path to the file.
	Path loader.FilePath
	// Html is the ready-to-rock HTML rendered from the file's markdown.
	Html template.HTML
	// Blocks holds all the code blocks found in the file.
	Blocks []*loader.CodeBlock
}
