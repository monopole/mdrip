package webserver

import (
	"testing"

	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
)

type tPair struct {
	path      string
	courseIdx []int
}

type pfTest1 struct {
	name    string
	input   model.Tutorial
	results []tPair
}

type pfTest2 struct {
	name    string
	input   model.Tutorial
	results [][]int
}

func makeLesson(n string) *model.LessonTut {
	return model.NewLessonTutForTests(base.FilePath(n), []*model.BlockTut{})
}

func makeCourse(n string, t ...model.Tutorial) *model.Course {
	return model.NewCourse(base.FilePath(n), t)
}

var tut1 = makeLesson("L0")
var tut2 = makeCourse(
	"C0",
	makeLesson("L0"), makeLesson("L1"))
var tut3 = makeCourse(
	"C0",
	makeLesson("L0"),
	makeCourse(
		"C1",
		makeLesson("L1"), makeLesson("L2")),
	makeLesson("L3"),
	makeCourse(
		"C2",
		makeLesson("L4"), makeLesson("L5")),
	makeLesson("L6"))

var pfTests1 = []pfTest1{
	{"bareLesson",
		tut1,
		[]tPair{
			{"L0", []int{0}},
			{"", []int{0}},
			{"zebra", []int{0}},
			{"L1", []int{0}}}},
	{"smallCourse",
		tut2,
		[]tPair{
			{"", []int{0}},
			{"zebra", []int{0}},
			{"L0", []int{0}},
			{"C0", []int{0, 0}},
			{"C0/L0", []int{0, 0}},
			{"C0/L1", []int{0, 1}}}},
	{"biggerCourse",
		tut3,
		[]tPair{
			{"", []int{0}},
			{"zebra", []int{0}},
			{"L0", []int{0}},
			{"C0", []int{0, 0}},
			{"C0/L0", []int{0, 0}},
			{"C0/C1/L1", []int{0, 1, 1}},
			{"C0/C1/L2", []int{0, 1, 2}},
			{"C0/L3", []int{0, 3}},
			{"C0/C2/L4", []int{0, 2, 4}},
			{"C0/C2/L5", []int{0, 2, 5}},
			{"C0/L6", []int{0, 6}},
			{"C0/L6/apple", []int{0}}}},
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

func TestGetLessonPath(t *testing.T) {
	for _, test := range pfTests1 {
		v := newLessonFinder()
		test.input.Accept(v)
		for _, w := range test.results {
			if got := v.getLessonPath(w.path); !slicesEqual(got, w.courseIdx) {
				t.Errorf("%s %s:\ngot\n\"%v\"\nwant\n\"%v\"\n",
					test.name, w.path, got, w.courseIdx)
			}
		}
	}
}

var pfTests2 = []pfTest2{
	{"bareLesson2",
		tut1,
		[][]int{{}}, /* wrong */
	},
	{"smallCourse2",
		tut2,
		[][]int{{0}, {0}},
	},
	{"biggerCourse2",
		tut3,
		[][]int{{0}, {0, 1}, {0, 1}, {0}, {0, 2}, {0, 2}, {0}},
	},
}

func TestGetCoursePaths(t *testing.T) {
	for _, test := range pfTests2 {
		v := newLessonFinder()
		test.input.Accept(v)
		result := v.getCoursePaths()
		if len(result) != len(test.results) {
			t.Errorf("%s length test : got %d, want %d\n",
				test.name, len(result), len(test.results))
		} else {
			for i := range test.results {
				if !slicesEqual(result[i], test.results[i]) {
					t.Errorf("%s slice test :\ngot\n\"%v\"\nwant\n\"%v\"\n",
						test.name, result[i], test.results[i])
				}
			}
		}
	}
}
