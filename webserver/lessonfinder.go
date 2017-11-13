package webserver

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
	"sort"
	"strings"
)

type lessonFinder struct {
	indexMap   map[base.FilePath]int
	coursePathMap   map[base.FilePath][]int
	nextLesson int
	courseCounter int
	name       []string
	coursePath  []int
}

// newLessonFinder builds a path -> lesson index map.
// For full paths to files this is simple, but for
// paths to directories one always wants the first lesson.
func newLessonFinder() *lessonFinder {
	return &lessonFinder{
		make(map[base.FilePath]int),
		make(map[base.FilePath][]int),
		0, -1, []string{}, []int{}}
}

// Returns 0 if not found (0 is index of first lesson).
func (v *lessonFinder) getLessonIndex(path string) int {
	return v.indexMap[base.FilePath(path)]
}

// Returns ordered list of course IDs
func (v *lessonFinder) getIndices(path string) []int {
	r := v.coursePathMap[base.FilePath(path)]
	if r == nil {
		return []int{0}
	}
	return r
}

// for debugging
func (v *lessonFinder) print1() {
	indexSet := make(map[int]bool)
	hoser := make(map[int][]string, 0)
	for k, v := range v.indexMap {
		a := hoser[v]
		if a == nil {
			a = []string{}
		}
		hoser[v] = append(a, string(k))
		indexSet[v] = true
	}
	var allIndices []int
	for k := range indexSet {
		allIndices = append(allIndices, k)
	}
	sort.Ints(allIndices)
	for _, i := range allIndices {
		for _, j := range hoser[i] {
			fmt.Printf("%3d %s\n", i, j)
		}
	}
}

func (v *lessonFinder) print2() {
	fmt.Println("-------------")
	for k, v := range v.coursePathMap {
		fmt.Printf("%20s %v\n", k, v)
	}
	fmt.Println()
	for k, v := range v.indexMap {
		fmt.Printf("%20s %v\n", k, v)
	}
	fmt.Println("-------------")
	fmt.Println()
	fmt.Println()
}

func (v *lessonFinder) addMapEntry() {
	newSlice := make([]int, len(v.coursePath), len(v.coursePath) + 1)
	copy(newSlice, v.coursePath)
	x := append(newSlice, v.nextLesson)
	k := base.FilePath(strings.Join(v.name, "/"))
	fmt.Printf("  adding entry %20s %d %v\n", string(k), v.nextLesson, x)
	v.indexMap[k] = v.nextLesson
	v.coursePathMap[k] = x
}

func (v *lessonFinder) VisitBlockTut(x *model.BlockTut) {
	v.name = append(v.name, x.Name())
	v.addMapEntry()
	v.name = v.name[:len(v.name)-1]
}

func (v *lessonFinder) VisitLessonTut(x *model.LessonTut) {
	glog.V(2).Infof("visiting lesson %s \n", x.Name())
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
	v.coursePath = append(v.coursePath, v.courseCounter)
	v.addMapEntry()
	for _, c := range x.Children() {
		c.Accept(v)
	}
	v.name = v.name[:len(v.name)-1]
	v.coursePath = v.coursePath[:len(v.coursePath)-1]
}

func (v *lessonFinder) VisitTopCourse(x *model.TopCourse) {
	glog.V(2).Infof("visiting top %s \n", x.Name())
	v.addMapEntry()
	for _, c := range x.Children() {
		c.Accept(v)
	}
}
