package webserver

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
	"sort"
	"strings"
)

type LessonFinder struct {
	indexMap   map[base.FilePath]int
	nextLesson int
	name       []string
}

// NewLessonFinder builds a path -> lesson index map.
// For full paths to files this is simple, but for
// paths to directories one always wants the first lesson.
func NewLessonFinder() *LessonFinder {
	return &LessonFinder{
		make(map[base.FilePath]int), 0, make([]string, 0)}
}

// Returns 0 if not found (0 is index of first lesson).
func (v *LessonFinder) GetLessonIndex(path string) int {
	return v.indexMap[base.FilePath(path)]
}

// for debugging
func (v *LessonFinder) print() {
	indexSet := make(map[int]bool)
	hoser := make(map[int][]string, 0)
	for k, v := range v.indexMap {
		bozo := hoser[v]
		if bozo == nil {
			bozo = []string{}
		}
		bozo = append(bozo, string(k))
		hoser[v] = bozo
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

func (v *LessonFinder) addMapEntry() {
	glog.V(2).Infof("  adding entry %s %d\n", strings.Join(v.name, "/"), v.nextLesson)
	v.indexMap[base.FilePath(strings.Join(v.name, "/"))] = v.nextLesson
}

func (v *LessonFinder) VisitBlockTut(x *model.BlockTut) {
	v.name = append(v.name, x.Name())
	v.addMapEntry()
	v.name = v.name[:len(v.name)-1]
}

func (v *LessonFinder) VisitLessonTut(x *model.LessonTut) {
	glog.V(2).Infof("visiting lesson %s \n", x.Name())
	v.name = append(v.name, x.Name())
	v.addMapEntry()
	for _, c := range x.Children() {
		c.Accept(v)
	}
	v.name = v.name[:len(v.name)-1]
	v.nextLesson++
}

func (v *LessonFinder) VisitCourse(x *model.Course) {
	glog.V(2).Infof("visiting course %s \n", x.Name())
	v.name = append(v.name, x.Name())
	v.addMapEntry()
	for _, c := range x.Children() {
		c.Accept(v)
	}
	v.name = v.name[:len(v.name)-1]
}

func (v *LessonFinder) VisitTopCourse(x *model.TopCourse) {
	glog.V(2).Infof("visiting top %s \n", x.Name())
	v.addMapEntry()
	for _, c := range x.Children() {
		c.Accept(v)
	}
}
