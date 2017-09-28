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
// Although the fields match Course, it's a different type so that visitors may treat ot differently;
// they typically ignore it and immediately descend into children.
type TopCourse struct {
	Course
}

func NewTopCourse(p base.FilePath, c []Tutorial) *TopCourse { return &TopCourse{Course{p, c}} }
func (t *TopCourse) Accept(v TutVisitor)                    { v.VisitTopCourse(t) }
func (t *TopCourse) Name() string                           { return "" }

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
