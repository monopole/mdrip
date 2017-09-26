package model

// BlockBase groups OpaqueCode with prose commentary.
type BlockBase struct {
	// prose is presumably human language documentation
	// for the OpaqueCode.
	prose []byte
	code  OpaqueCode
}

func (x *BlockBase) Prose() []byte                  { return x.prose }
func (x *BlockBase) Code() OpaqueCode               { return x.code }
func NewBlockBase(p []byte, c OpaqueCode) BlockBase { return BlockBase{p, c} }
