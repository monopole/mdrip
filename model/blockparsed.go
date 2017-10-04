package model

import "github.com/monopole/mdrip/base"

// BlockParsed groups a BlockBase with labels.
type BlockParsed struct {
	base.BlockBase
	labels []base.Label
}

func NewProseOnlyBlock(p base.MdProse) *BlockParsed {
	return NewBlockParsed([]base.Label{}, p, base.OpaqueCode(""))
}

func NewBlockParsed(labels []base.Label, p base.MdProse, c base.OpaqueCode) *BlockParsed {
	return &BlockParsed{base.NewBlockBase(p, c), labels}
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
