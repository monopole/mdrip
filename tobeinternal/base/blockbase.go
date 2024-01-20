package base

// BlockBase groups OpaqueCode with prose commentary.
type BlockBase struct {
	prose MdProse
	code  OpaqueCode
}

// Prose from the block.
func (x *BlockBase) Prose() MdProse { return x.prose }

// Code from the block.
func (x *BlockBase) Code() OpaqueCode { return x.code }

// NewBlockBase is a ctor.
func NewBlockBase(p MdProse, c OpaqueCode) BlockBase { return BlockBase{p, c} }
