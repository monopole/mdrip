package model

import (
	"github.com/monopole/mdrip/base"
)

// A LessonTut has a one to one correspondence to a file.
// It must have a name, and may have blocks.
// An entirely empty file might appear with no blocks.
type LessonTut struct {
	path      base.FilePath
	mdContent *MdContent
	blocks    []*BlockTut
}

func NewLessonTutForTests(p base.FilePath, blocks []*BlockTut) *LessonTut {
	return &LessonTut{p, NewMdContent(), blocks}
}

func NewLessonTutFromMdContent(p base.FilePath, md *MdContent) *LessonTut {
	result := make([]*BlockTut, len(md.Blocks))
	for i, b := range md.Blocks {
		result[i] = NewBlockTut(b)
	}
	return &LessonTut{p, md, result}
}

func (l *LessonTut) Accept(v TutVisitor) { v.VisitLessonTut(l) }
func (l *LessonTut) Title() string {
	if l.mdContent.HasTitle() {
		return l.mdContent.GetTitle()
	}
	return l.Name()
}
func (l *LessonTut) Name() string {
	return l.path.Base()
}
func (l *LessonTut) Path() base.FilePath { return l.path }
func (l *LessonTut) Children() []Tutorial {
	result := []Tutorial{}
	for _, b := range l.blocks {
		result = append(result, b)
	}
	return result
}

func (l *LessonTut) Blocks() []*BlockTut { return l.blocks }
