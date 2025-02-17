package parsren

import (
	"github.com/monopole/mdrip/v2/internal/loader"
	"html/template"
)

// BlockFilter is a function that returns true or false based on the
// incoming CodeBlock.
type BlockFilter func(*loader.CodeBlock) bool

// MdParserRenderer is a tree visitor that parses and renders markdown.
// The two operations are closely coupled by a shared abstract syntax tree
// and shared raw bytes from the source markdown.
// Usage:
//   - Load markdown files into a tree of treeNode.
//   - Accept this visitor into the tree.
//   - After the call to Accept finishes, consult
//     RenderMdFiles() and/or Filter() for whatever purpose.
type MdParserRenderer interface {
	loader.TreeVisitor
	// RenderedMdFiles is a slice of rendered markdown files
	// in depth-first order.
	RenderedMdFiles() []*RenderedMdFile
	// Filter returns all blocks that pass the filter.
	Filter(BlockFilter) []*loader.CodeBlock
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
