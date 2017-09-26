package model

// BlockParsed groups a BlockBase with labels.
type BlockParsed struct {
	BlockBase
	labels []Label
}

func NewBlockParsed(labels []Label, p []byte, c string) *BlockParsed {
	return &BlockParsed{BlockBase{p, OpaqueCode(c)}, labels}
}

func (x *BlockParsed) Labels() []Label { return x.labels }
func (x *BlockParsed) HasLabel(label Label) bool {
	for _, l := range x.labels {
		if l == label {
			return true
		}
	}
	return false
}
