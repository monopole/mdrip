package model

import (
	"github.com/monopole/mdrip/tobeinternal/base"
)

// BlockTut is a part of a LessonTut - one block of code, maybe with prose.
type BlockTut struct {
	BlockParsed
}

// AnonBlockName used for blocks that have no explicit name.
const AnonBlockName = "clickToCopy"

// NewBlockTut is a ctor.
func NewBlockTut(b *BlockParsed) *BlockTut {
	return &BlockTut{*b}
}

func (x *BlockTut) firstNiceLabel() base.Label {
	for _, l := range x.labels {
		if l != base.WildCardLabel && l != base.AnonLabel {
			return l
		}
	}
	return base.AnonLabel
}

// Accept accepts a visitor.
func (x *BlockTut) Accept(v TutVisitor) { v.VisitBlockTut(x) }

// Title is what appears to be the title of the block.
func (x *BlockTut) Title() string {
	return x.Name()
}

// Name attempts to return a decent name for the block.
func (x *BlockTut) Name() string {
	l := x.firstNiceLabel()
	if l == base.AnonLabel {
		return AnonBlockName
	}
	return string(l)
}

// Path to the file containing the block.
func (x *BlockTut) Path() base.FilePath { return base.FilePath("notUsingThis") }

// Children of the block - there aren't any at this time.
// One could imagine each line of code in a code block as a child
// if that were useful somehow.
func (x *BlockTut) Children() []Tutorial { return []Tutorial{} }
