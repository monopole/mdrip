package model

import (
	"github.com/monopole/mdrip/base"
)

// LessonTut has a one to one correspondence to a file.
// It must have a name, and may have blocks.
// An entirely empty file might appear with no blocks.
type LessonTut struct {
	path      base.FilePath
	mdContent *MdContent
	blocks    []*BlockTut
}

// NewLessonTutForTests makes one for tests.
func NewLessonTutForTests(p base.FilePath, blocks []*BlockTut) *LessonTut {
	return &LessonTut{p, NewMdContent(), blocks}
}

// NewLessonTutFromMdContent converts MdContent to a LessonTut.
func NewLessonTutFromMdContent(p base.FilePath, md *MdContent) *LessonTut {
	result := make([]*BlockTut, len(md.Blocks))
	for i, b := range md.Blocks {
		result[i] = NewBlockTut(b)
	}
	return &LessonTut{p, md, result}
}

// Accept accepts a visitor.
func (l *LessonTut) Accept(v TutVisitor) { v.VisitLessonTut(l) }

// Title is the purported title of the LessonTut.
func (l *LessonTut) Title() string {
	if l.mdContent.HasTitle() {
		return l.mdContent.GetTitle()
	}
	return l.Name()
}

// Name is the purported name of the lesson.
func (l *LessonTut) Name() string {
	return l.path.Base()
}

// Path to the lesson.  A lesson has a 1:1 correspondence with a path.
func (l *LessonTut) Path() base.FilePath { return l.path }

// Children of the lesson - the code blocks.
func (l *LessonTut) Children() []Tutorial {
	result := []Tutorial{}
	for _, b := range l.blocks {
		result = append(result, b)
	}
	return result
}

// Blocks in the lesson.
func (l *LessonTut) Blocks() []*BlockTut { return l.blocks }
