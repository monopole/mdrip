package model

// LessonCounter is a visitor that merely counts lessons.
type LessonCounter struct {
	count int
}

// NewTutorialLessonCounter makes a new LessonCounter.
func NewTutorialLessonCounter() *LessonCounter {
	return &LessonCounter{0}
}

// Count is the reason this visitor exists.
func (v *LessonCounter) Count() int {
	return v.count
}

// VisitBlockTut does nothing.
func (v *LessonCounter) VisitBlockTut(b *BlockTut) {
}

// VisitLessonTut increments the count.
func (v *LessonCounter) VisitLessonTut(l *LessonTut) {
	v.count++
}

// VisitCourse visits children.
func (v *LessonCounter) VisitCourse(c *Course) {
	for _, x := range c.Children() {
		x.Accept(v)
	}
}

// VisitTopCourse visits children.
func (v *LessonCounter) VisitTopCourse(t *TopCourse) {
	for _, x := range t.Children() {
		x.Accept(v)
	}
}
