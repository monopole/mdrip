package model

import "github.com/monopole/mdrip/base"

// BlockParsed groups a BlockBase with labels.
type BlockParsed struct {
	base.BlockBase
	labels []base.Label
}

func NewBlockParsed(labels []base.Label, p []byte, c string) *BlockParsed {
	return &BlockParsed{base.NewBlockBase(p, base.OpaqueCode(c)), labels}
}

func (x *BlockParsed) Labels() []base.Label { return x.labels }
func (x *BlockParsed) HasLabel(label base.Label) bool {
	for _, l := range x.labels {
		if l == label {
			return true
		}
	}
	return false
}
