package model

import (
	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/util"
)

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

// A TopCourse is exactly like a Course accept that visitors
// may treat it differently, ignoring everything about it
// except its children.  Its name is special in that it might
// be derived from a URL, from a list of files and directories,
// etc.
type TopCourse struct {
	Course
}

func NewTopCourse(n string, p base.FilePath, c []Tutorial) *TopCourse {
	return &TopCourse{Course{n, p, c}}
}
func (t *TopCourse) Accept(v TutVisitor) { v.VisitTopCourse(t) }

// A Course is a directory - an ordered list of Lessons and Courses.
type Course struct {
	name     string
	path     base.FilePath
	children []Tutorial
}

func NewCourse(p base.FilePath, c []Tutorial) *Course { return &Course{p.Base(), p, c} }
func (c *Course) Accept(v TutVisitor)                 { v.VisitCourse(c) }
func (c *Course) Name() string                        { return util.DropLeadingNumbers(c.name) }
func (c *Course) Path() base.FilePath                 { return c.path }
func (c *Course) Children() []Tutorial                { return c.children }
