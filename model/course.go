package model

import "github.com/monopole/mdrip/base"

// Course is a directory - an ordered list of Lessons and Courses.
type Course struct {
	name     string
	path     base.FilePath
	children []Tutorial
}

// NewCourse makes a Course.
func NewCourse(p base.FilePath, c []Tutorial) *Course { return &Course{p.Base(), p, c} }

// Accept accepts a visitor.
func (c *Course) Accept(v TutVisitor) { v.VisitCourse(c) }

// Title is the purported Course title.
func (c *Course) Title() string { return c.Name() }

// Name is the purported Course name.
func (c *Course) Name() string { return c.name }

// Path is where the course came from.
func (c *Course) Path() base.FilePath { return c.path }

// Children are the parts of the course.
func (c *Course) Children() []Tutorial { return c.children }
