package webserver

import (
	"testing"

	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
)

type tPair struct {
	path      string
	lessIdx   int
	courseIdx []int
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

var pfTests = []pfTest{
	{"bareLesson",
		makeLesson("L0"),
		[]tPair{
			{"L0", 0, []int{0}},
			{"", 0, []int{0}},
			{"zebra", 0, []int{0}},
			{"L1", 0, []int{0}}}},
	{"smallCourse",
		makeCourse("C0", makeLesson("L0"), makeLesson("L1")),
		[]tPair{
			{"", 0, []int{0}},
			{"zebra", 0, []int{0}},
			{"L0", 0, []int{0}},
			{"C0", 0, []int{0, 0}},
			{"C0/L0", 0, []int{0, 0}},
			{"C0/L1", 1, []int{0, 1}}}},
	{"biggerCourse",
		makeCourse("C0",
			makeLesson("L0"),
			makeCourse("C1", makeLesson("L1"), makeLesson("L2")),
			makeLesson("L3"),
			makeCourse("C2", makeLesson("L4"), makeLesson("L5")),
			makeLesson("L6")),
		[]tPair{
			{"", 0, []int{0}},
			{"zebra", 0, []int{0}},
			{"L0", 0, []int{0}},
			{"C0", 0, []int{0, 0}},
			{"C0/L0", 0, []int{0, 0}},
			{"C0/C1/L1", 1, []int{0, 1, 1}},
			{"C0/C1/L2", 2, []int{0, 1, 2}},
			{"C0/L3", 3, []int{0, 3}},
			{"C0/C2/L4", 4, []int{0, 2, 4}},
			{"C0/C2/L5", 5, []int{0, 2, 5}},
			{"C0/L6", 6, []int{0, 6}},
			{"C0/L6/apple", 0, []int{0}}}},
}

func slicesEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestLessonFinder(t *testing.T) {
	for _, test := range pfTests {
		v := newLessonFinder()
		test.input.Accept(v)
		for _, w := range test.results {
			if got := v.getLessonIndex(w.path); got != w.lessIdx {
				t.Errorf("%s %s:\ngot\n\"%d\"\nwant\n\"%d\"\n",
					test.name, w.path, got, w.lessIdx)
			}
			if got := v.getIndices(w.path); !slicesEqual(got, w.courseIdx) {
				t.Errorf("%s %s:\ngot\n\"%v\"\nwant\n\"%v\"\n",
					test.name, w.path, got, w.courseIdx)
			}
		}
		v.print2()
	}
}
