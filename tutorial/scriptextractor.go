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

func (v *ScriptExtractor) VisitCommandBlock(b *CommandBlock) {
}

func (v *ScriptExtractor) VisitLesson(l *Lesson) {
	otherblocks, ok := l.Structure()[v.label]
	if !ok {
		return
	}
	myblocks := []*model.OldBlock{}
	for _, x := range otherblocks {
		myblocks = append(myblocks, model.NewOldBlock(x.Labels(), string(x.Code()), x.RawProse()))
	}
	v.scripts = append(v.scripts, model.NewScript(l.Path(), myblocks))

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
