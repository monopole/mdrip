package model

import (
	"github.com/monopole/mdrip/base"
)

// BlockTut is a part of a LessonTut.
type BlockTut struct {
	BlockParsed
}

const AnonBlockName = "clickToCopy"

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

func (x *BlockTut) Accept(v TutVisitor) { v.VisitBlockTut(x) }
func (x *BlockTut) Title() string {
	return x.Name()
}

func (x *BlockTut) Name() string {
	l := x.firstNiceLabel()
	if l == base.AnonLabel {
		return AnonBlockName
	}
	return string(l)
}

func (x *BlockTut) Path() base.FilePath  { return base.FilePath("notUsingThis") }
func (x *BlockTut) Children() []Tutorial { return []Tutorial{} }
