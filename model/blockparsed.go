package model

import "github.com/monopole/mdrip/base"

// BlockParsed groups a BlockBase with labels.
type BlockParsed struct {
	base.BlockBase
	labels []base.Label
}

// NewProseOnlyBlock makes a BlockParsed with no code.
func NewProseOnlyBlock(p base.MdProse) *BlockParsed {
	return NewBlockParsed([]base.Label{}, p, base.OpaqueCode(""))
}

// NewBlockParsed returns a BlockParsed with the given content.
func NewBlockParsed(labels []base.Label, p base.MdProse, c base.OpaqueCode) *BlockParsed {
	return &BlockParsed{base.NewBlockBase(p, c), labels}
}

// Labels are the labels found on the block.
func (x *BlockParsed) Labels() []base.Label { return x.labels }

// HasLabel is true if the block has the given label argument.
func (x *BlockParsed) HasLabel(label base.Label) bool {
	for _, l := range x.labels {
		if l == label {
			return true
		}
	}
	return false
}
