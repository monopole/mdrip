package model

// A LessonTut has a one to one correspondence to a file.
// It must have a name, and may have blocks.
// An entirely empty file might appear with no blocks.
type LessonTut struct {
	path   FilePath
	blocks []*BlockTut
}

func NewLessonTut(p FilePath, blocks []*BlockTut) *LessonTut {
	return &LessonTut{p, blocks}
}

func NewLessonTutFromBlockParsed(p FilePath, blocks []*BlockParsed) *LessonTut {
	result := make([]*BlockTut, len(blocks))
	for i, b := range blocks {
		result[i] = NewBlockTut(b)
	}
	return NewLessonTut(p, result)
}

func (l *LessonTut) Accept(v TutVisitor) { v.VisitLessonTut(l) }
func (l *LessonTut) Name() string        { return l.path.Base() }
func (l *LessonTut) Path() FilePath      { return l.path }
func (l *LessonTut) Children() []Tutorial {
	result := []Tutorial{}
	for _, b := range l.blocks {
		result = append(result, b)
	}
	return result
}

func (l *LessonTut) Blocks() []*BlockTut { return l.blocks }
