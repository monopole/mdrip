package program

import (
	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
)

// LessonPgmExtractor extracts all Lessons in depth first order
// from a Tutorial to create a flat list of lessons.  The lessons
// are edited - only blocks with the given label are carried over
// into the new extracted lessons.  If a lesson has no blocks with
// the given label, it is completely dropped.
type LessonPgmExtractor struct {
	label      base.Label
	firstTitle string
	lessons    []*LessonPgm
	blockAccum []*BlockPgm
}

func NewLessonPgmExtractor(label base.Label) *LessonPgmExtractor {
	return &LessonPgmExtractor{label, "", []*LessonPgm{}, []*BlockPgm{}}
}

func (v *LessonPgmExtractor) Lessons() []*LessonPgm {
	return v.lessons
}

func (v *LessonPgmExtractor) FirstTitle() string {
	return v.firstTitle
}

func (v *LessonPgmExtractor) VisitBlockTut(b *model.BlockTut) {
	if v.label == base.WildCardLabel || b.HasLabel(v.label) {
		v.blockAccum = append(v.blockAccum, NewBlockPgmFromBlockTut(b))
	}
}

func (v *LessonPgmExtractor) VisitLessonTut(l *model.LessonTut) {
	if len(v.firstTitle) == 0 {
		v.firstTitle = l.Title()
	}
	v.blockAccum = []*BlockPgm{}
	for _, x := range l.Children() {
		x.Accept(v)
	}
	if len(v.blockAccum) < 1 {
		return
	}
	id := -1
	for _, b := range v.blockAccum {
		if len(b.Code()) > 0 {
			id++
			b.id = id
		} else {
			b.id = -1
		}
	}
	v.lessons = append(v.lessons, NewLessonPgm(l.Path(), v.blockAccum))
}

func (v *LessonPgmExtractor) VisitCourse(c *model.Course) {
	for _, x := range c.Children() {
		x.Accept(v)
	}
}

func (v *LessonPgmExtractor) VisitTopCourse(t *model.TopCourse) {
	for _, x := range t.Children() {
		x.Accept(v)
	}
}
