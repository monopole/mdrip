package webserver

import (
	"testing"

	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
)

type tPair struct {
	path    string
	lessIdx int
}

type pfTest struct {
	name    string
	input   model.Tutorial
	results []tPair
}

func makeLesson(n string) *model.LessonTut {
	return model.NewLessonTut(base.FilePath(n), []*model.BlockTut{})
}

func makeCourse(n string, t ...model.Tutorial) *model.Course {
	return model.NewCourse(base.FilePath(n), t)
}

var pfCourse1 = makeCourse("C1", makeLesson("LA"), makeLesson("LB"))
var pfCourse2 = makeCourse("C2", makeLesson("LC"), makeLesson("LD"))

var pfTests = []pfTest{
	{"bareLesson",
		makeLesson("L1"),
		[]tPair{{"", 0}, {"zebra", 0}, {"L1", 0}}},
	{"smallCourse",
		pfCourse1,
		[]tPair{{"", 0}, {"zebra", 0}, {"L1", 0},
			{"C1", 0}, {"C1/LA", 0}, {"C1/LB", 1}}},
	{"biggerCourse",
		makeCourse("C3",
			makeLesson("Lx"),
			pfCourse1,
			makeLesson("Ly"),
			pfCourse2,
			makeLesson("Lz")),
		[]tPair{{"", 0}, {"zebra", 0}, {"Lx", 0},
			{"C3", 0}, {"C3/Lx", 0}, {"C3/C1/LA", 1}, {"C3/C1/LB", 2},
			{"C3/Ly", 3}, {"C3/C2/LC", 4}, {"C3/C2/LD", 5},
			{"C3/Lz", 6}, {"C3/Lz/apple", 0}}},
}

func TestLessonFinder(t *testing.T) {
	for _, test := range pfTests {
		v := newLessonFinder()
		test.input.Accept(v)
		for _, w := range test.results {
			got := v.getLessonIndex(w.path)
			if got != w.lessIdx {
				t.Errorf("%s %s:\ngot\n\"%d\"\nwant\n\"%d\"\n",
					test.name, w.path, got, w.lessIdx)
			}
		}
	}
}
