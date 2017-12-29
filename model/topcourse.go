package model

import "github.com/monopole/mdrip/base"

// A TopCourse is exactly like a Course accept that visitors
// may treat it differently, ignoring everything about it
// except its children.  Its name is special in that it might
// be derived from a URL, from a list of files and directories,
// etc.  It's usually a list of directories.
type TopCourse struct {
	Course
}

// NewTopCourse makes a new TopCourse.
func NewTopCourse(n string, p base.FilePath, c []Tutorial) *TopCourse {
	return &TopCourse{Course{n, p, c}}
}

// Accept accepts a visitor.
func (t *TopCourse) Accept(v TutVisitor) { v.VisitTopCourse(t) }
