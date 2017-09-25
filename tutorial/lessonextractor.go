package tutorial

import (
	"github.com/monopole/mdrip/model"
)

// LessonExtractor extracts all Lessons in depth first order
// from a Tutorial to create a flat list of lessons.  The lessons
// are edited - only blocks with the given label are carried over
// into the new extracted lesson.  If a lesson has no blocks with
// the given label, it is completely dropped.
type LessonExtractor struct {
	label   model.Label
	lessons []*Lesson
}

func NewLessonExtractor(label model.Label) *LessonExtractor {
	return &LessonExtractor{label, []*Lesson{}}
}

func (v *LessonExtractor) Lessons() []*Lesson {
	return v.lessons
}

func (v *LessonExtractor) VisitCommandBlock(b *CommandBlock) {
}

func (v *LessonExtractor) VisitLesson(l *Lesson) {
	if v.label == model.AnyLabel && len(l.Children()) > 0 {
		v.lessons = append(v.lessons, l)
		return
	}
	blocks := l.GetBlocksWithLabel(v.label)
	if len(blocks) > 0 {
		v.lessons = append(v.lessons, NewLesson(l.Path(), blocks))
	}
}

func (v *LessonExtractor) VisitCourse(c *Course) {
	for _, x := range c.children {
		x.Accept(v)
	}
}

func (v *LessonExtractor) VisitTopCourse(t *TopCourse) {
	for _, x := range t.children {
		x.Accept(v)
	}
}
