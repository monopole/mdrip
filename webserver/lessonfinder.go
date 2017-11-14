package webserver

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
)

type lessonFinder struct {
	coursePathMap   map[base.FilePath][]int
	nextLesson      int
	courseCounter   int
	name            []string
	coursePathIndex []int
	superIndex      [][]int
}

// newLessonFinder builds a path -> lesson index map.
// For full paths to files this is simple, but for
// paths to directories one always wants the first lesson.
func newLessonFinder() *lessonFinder {
	return &lessonFinder{
		make(map[base.FilePath][]int),
		0, -1, []string{},
		[]int{}, [][]int{}}
}

func (v *lessonFinder) getCoursePaths() [][]int {
	return v.superIndex
}

// Returns ordered list of course IDs, ending with the lesson ID.
// Represents a directory path followed by a filename.
func (v *lessonFinder) getLessonPath(path string) []int {
	r := v.coursePathMap[base.FilePath(path)]
	if r == nil {
		return []int{0}
	}
	return r
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
	newSlice := make([]int, len(v.coursePathIndex), len(v.coursePathIndex)+1)
	copy(newSlice, v.coursePathIndex)
	x := append(newSlice, v.nextLesson)
	k := base.FilePath(strings.Join(v.name, "/"))
	glog.V(2).Infof("  adding entry %20s %d %v\n", string(k), v.nextLesson, x)
	v.coursePathMap[k] = x
}

func (v *lessonFinder) addIndexEntry() {
	newSlice := make([]int, len(v.coursePathIndex))
	copy(newSlice, v.coursePathIndex)
	if v.nextLesson != len(v.superIndex) {
		panic(fmt.Sprintf("Ordering problem: nextLesson =%d, len(superIndex) = %d",
			v.nextLesson, len(v.superIndex)))
	}
	glog.V(2).Infof("  adding lesson entry %5d  %v\n", v.nextLesson, v.nextLesson, newSlice)
	v.superIndex = append(v.superIndex, newSlice)
}

func (v *lessonFinder) VisitBlockTut(x *model.BlockTut) {
	v.name = append(v.name, x.Name())
	v.addMapEntry()
	v.name = v.name[:len(v.name)-1]
}

func (v *lessonFinder) VisitLessonTut(x *model.LessonTut) {
	glog.V(2).Infof("visiting lesson %s \n", x.Name())
	v.addIndexEntry()
	v.name = append(v.name, x.Name())
	v.addMapEntry()
	for _, c := range x.Children() {
		c.Accept(v)
	}
	v.name = v.name[:len(v.name)-1]
	v.nextLesson++
}

func (v *lessonFinder) VisitCourse(x *model.Course) {
	v.courseCounter++
	glog.V(2).Infof("visiting course %s \n", x.Name())
	v.name = append(v.name, x.Name())
	v.coursePathIndex = append(v.coursePathIndex, v.courseCounter)
	v.addMapEntry()
	for _, c := range x.Children() {
		c.Accept(v)
	}
	v.name = v.name[:len(v.name)-1]
	v.coursePathIndex = v.coursePathIndex[:len(v.coursePathIndex)-1]
}

func (v *lessonFinder) VisitTopCourse(x *model.TopCourse) {
	glog.V(2).Infof("visiting top %s \n", x.Name())
	v.addMapEntry()
	for _, c := range x.Children() {
		c.Accept(v)
	}
}
