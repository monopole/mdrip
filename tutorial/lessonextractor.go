package tutorial

// LessonExtractor extracts all Lessons from a Tutorial.
type LessonExtractor struct {
	lessons []*Lesson
}

func NewLessonExtractor() *LessonExtractor {
	return &LessonExtractor{[]*Lesson{}}
}

func (v *LessonExtractor) Lessons() []*Lesson {
	return v.lessons
}

func (v *LessonExtractor) VisitCommandBlock(b *CommandBlock) {
}

func (v *LessonExtractor) VisitLesson(l *Lesson) {
	v.lessons = append(v.lessons, l)
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
