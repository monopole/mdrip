package model

import "github.com/monopole/mdrip/base"

type Tutorial interface {
	Accept(v TutVisitor)
	Name() string
	Path() base.FilePath
	Children() []Tutorial
}

type TutVisitor interface {
	VisitTopCourse(t *TopCourse)
	VisitCourse(c *Course)
	VisitLessonTut(l *LessonTut)
	VisitBlockTut(b *BlockTut)
}

// A TopCourse is a Course with no name - it's the root of the tree, the cover of the book.
type TopCourse struct {
	path     base.FilePath
	children []Tutorial
}

func NewTopCourse(p base.FilePath, c []Tutorial) *TopCourse { return &TopCourse{p, c} }
func (t *TopCourse) Accept(v TutVisitor)                    { v.VisitTopCourse(t) }
func (t *TopCourse) Name() string                           { return "" }
func (t *TopCourse) Path() base.FilePath                    { return t.path }
func (t *TopCourse) Children() []Tutorial                   { return t.children }

// A Course, or directory, has a name but no content, and an ordered list of
// Lessons and Courses. If the list is empty, the Course is dropped (hah!).
type Course struct {
	path     base.FilePath
	children []Tutorial
}

func NewCourse(p base.FilePath, c []Tutorial) *Course { return &Course{p, c} }
func (c *Course) Accept(v TutVisitor)                 { v.VisitCourse(c) }
func (c *Course) Name() string                        { return c.path.Base() }
func (c *Course) Path() base.FilePath                 { return c.path }
func (c *Course) Children() []Tutorial                { return c.children }
