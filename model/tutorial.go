package model

import (
	"github.com/monopole/mdrip/base"
)

// Tutorial represents a book in tree / hierarchical form.
type Tutorial interface {
	Accept(v TutVisitor)
	Title() string
	Name() string
	Path() base.FilePath
	Children() []Tutorial
}

// TutVisitor has the ability to visit the items specified in its methods.
type TutVisitor interface {
	VisitTopCourse(t *TopCourse)
	VisitCourse(c *Course)
	VisitLessonTut(l *LessonTut)
	VisitBlockTut(b *BlockTut)
}
