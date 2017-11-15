package webserver

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
)

// lessonFinder traverses a tutorial tree to build quick
// data structures (a map and an array) that can answer
// common questions without the need for more traversals.
type lessonFinder struct {
	nextLesson            int
	courseCounter         int
	namePathAccumulator   []string
	coursePathAccumulator []int
	coursePathMap         map[base.FilePath][]int
	coursePathIndex       [][]int
}

func newLessonFinder() *lessonFinder {
	return &lessonFinder{
		0, -1, []string{},
		[]int{}, make(map[base.FilePath][]int), [][]int{}}
}

// getLessonPath returns ordered list of course IDs,
// ending with the lesson ID.  The argument should be
// a path, e.g. benelux/belgium/beer, the result is
// something like [0, 2, 6], where benelux is course #0,
// belgium is course #2 inside benelux, and beer is lesson
// #6 inside belgium.
func (v *lessonFinder) getLessonPath(path string) []int {
	r := v.coursePathMap[base.FilePath(path)]
	if r == nil {
		return []int{0}
	}
	return r
}

// getCoursePaths returns a array of arrays.
// The index is a lesson ID, and the entry at that
// index is an array of course IDs above the lesson.
// In the example provided for getLessonPath, the
// value at index 6 would be [0, 2], i.e. the beer
// lesson is found under benelux/belgium.
func (v *lessonFinder) getCoursePaths() [][]int {
	return v.coursePathIndex
}

// For debugging.
func (v *lessonFinder) print() {
	fmt.Println("-------------")
	for k, v := range v.coursePathMap {
		fmt.Printf("%20s %v\n", k, v)
	}
	fmt.Println("-------------")
	fmt.Println()
}

func (v *lessonFinder) addMapEntry() {
	newSlice := make([]int, len(v.coursePathAccumulator), len(v.coursePathAccumulator)+1)
	copy(newSlice, v.coursePathAccumulator)
	v.coursePathMap[base.FilePath(strings.Join(v.namePathAccumulator, "/"))] =
		append(newSlice, v.nextLesson)
}

func (v *lessonFinder) addIndexEntry() {
	newSlice := make([]int, len(v.coursePathAccumulator))
	copy(newSlice, v.coursePathAccumulator)
	if v.nextLesson != len(v.coursePathIndex) {
		panic(
			fmt.Sprintf(
				"Ordering problem: nextLesson =%d, len(coursePathIndex) = %d",
				v.nextLesson, len(v.coursePathIndex)))
	}
	v.coursePathIndex = append(v.coursePathIndex, newSlice)
}

func (v *lessonFinder) VisitBlockTut(x *model.BlockTut) {
	v.namePathAccumulator = append(v.namePathAccumulator, x.Name())
	v.addMapEntry()
	v.namePathAccumulator = v.namePathAccumulator[:len(v.namePathAccumulator)-1]
}

func (v *lessonFinder) VisitLessonTut(x *model.LessonTut) {
	glog.V(2).Infof("visiting lesson %s \n", x.Name())
	v.addIndexEntry()
	v.namePathAccumulator = append(v.namePathAccumulator, x.Name())
	v.addMapEntry()
	for _, c := range x.Children() {
		c.Accept(v)
	}
	v.namePathAccumulator = v.namePathAccumulator[:len(v.namePathAccumulator)-1]
	v.nextLesson++
}

func (v *lessonFinder) VisitCourse(x *model.Course) {
	v.courseCounter++
	glog.V(2).Infof("visiting course %s \n", x.Name())
	v.namePathAccumulator = append(v.namePathAccumulator, x.Name())
	v.coursePathAccumulator = append(v.coursePathAccumulator, v.courseCounter)
	v.addMapEntry()
	for _, c := range x.Children() {
		c.Accept(v)
	}
	v.namePathAccumulator = v.namePathAccumulator[:len(v.namePathAccumulator)-1]
	v.coursePathAccumulator = v.coursePathAccumulator[:len(v.coursePathAccumulator)-1]
}

func (v *lessonFinder) VisitTopCourse(x *model.TopCourse) {
	glog.V(2).Infof("visiting top %s \n", x.Name())
	v.addMapEntry()
	for _, c := range x.Children() {
		c.Accept(v)
	}
}
