package model

// BlockTut is a part of a LessonTut.
type BlockTut struct {
	BlockParsed
}

func NewBlockTut(b *BlockParsed) *BlockTut {
	return &BlockTut{*b}
}

func (x *BlockTut) Accept(v TutVisitor) { v.VisitBlockTut(x) }
func (x *BlockTut) Name() string {
	if len(x.Labels()) > 0 {
		return string(x.Labels()[0])
	}
	return "noName"
}
func (x *BlockTut) Path() FilePath       { return FilePath("notUsingThis") }
func (x *BlockTut) Children() []Tutorial { return []Tutorial{} }
