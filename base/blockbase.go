package base

// BlockBase groups OpaqueCode with prose commentary.
type BlockBase struct {
	prose MdProse
	code  OpaqueCode
}

func (x *BlockBase) Prose() MdProse                  { return x.prose }
func (x *BlockBase) Code() OpaqueCode                { return x.code }
func NewBlockBase(p MdProse, c OpaqueCode) BlockBase { return BlockBase{p, c} }
