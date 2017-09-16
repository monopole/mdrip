package tutorial

import (
	"github.com/monopole/mdrip/model"
)

// ScriptExtractor extracts scripts with a given label from a Tutorial.
type ScriptExtractor struct {
	label   model.Label
	scripts []*model.Script
}

func NewScriptExtractor(l model.Label) *ScriptExtractor {
	return &ScriptExtractor{l, []*model.Script{}}
}

func (v *ScriptExtractor) Scripts() []*model.Script {
	return v.scripts
}

func (v *ScriptExtractor) VisitLesson(l *Lesson) {
	if blocks, ok := l.Structure()[v.label]; ok {
		v.scripts = append(v.scripts, model.NewScript(l.Path(), blocks))
	}
}

func (v *ScriptExtractor) VisitCourse(c *Course) {
	for _, x := range c.children {
		x.Accept(v)
	}
}

func (v *ScriptExtractor) VisitTopCourse(t *TopCourse) {
	for _, x := range t.children {
		x.Accept(v)
	}
}
