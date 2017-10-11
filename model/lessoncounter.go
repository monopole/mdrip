package model

type LessonCounter struct {
	count int
}

func NewTutorialLessonCounter() *LessonCounter {
	return &LessonCounter{0}
}

func (v *LessonCounter) Count() int {
	return v.count
}

func (v *LessonCounter) VisitBlockTut(b *BlockTut) {
}

func (v *LessonCounter) VisitLessonTut(l *LessonTut) {
	v.count++
}

func (v *LessonCounter) VisitCourse(c *Course) {
	for _, x := range c.Children() {
		x.Accept(v)
	}
}

func (v *LessonCounter) VisitTopCourse(t *TopCourse) {
	for _, x := range t.Children() {
		x.Accept(v)
	}
}
