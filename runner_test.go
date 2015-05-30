package main

import (
	"strings"
	"testing"
)

func TestRunnerWithNothing(t *testing.T) {
	if RunInSubShell([]*ScriptBucket{}).err != nil {
		t.Fail()
	}
}

func TestRunnerWithGoodStuff(t *testing.T) {
	labels := []string{"foo", "bar"}
	blocks := []*codeBlock{
		&codeBlock{labels, "ls\ndate\n"},
		&codeBlock{labels, "echo beans\necho cheese\n"},
		&codeBlock{labels, "echo hasta\necho la vista\n"}}
	result := RunInSubShell([]*ScriptBucket{&ScriptBucket{"iAmFileName", blocks}})
	if result.err != nil {
		t.Fail()
	}
}

func checkFail(t *testing.T, got, want *ErrorBucket) {
	if got.err == nil {
		t.Fail()
	}
	if got.index != want.index {
		t.Errorf("%s got\n\t%v\nwant\n\t%v", "script", got.index, want.index)
	}
	if strings.TrimSpace(got.message) != want.message {
		t.Errorf("%s got\n\t%v\nwant\n\t%v", "message", got.message, want.message)
	}
}

func TestWithBadStuff1(t *testing.T) {
	want := &ErrorBucket{
		textBucket{false, "dunno"},
		"iAmAFileName",
		0,
		emptyCodeBlock,
		nil,
		"bash: line 1: notagoodcommand: command not found"}

	labels := []string{"foo", "bar"}
	blocks := []*codeBlock{&codeBlock{labels, "notagoodcommand\n"}}
	got := RunInSubShell([]*ScriptBucket{&ScriptBucket{"iAmFileName", blocks}})
	checkFail(t, got, want)
}

func TestWithBadStuff2(t *testing.T) {
	want := &ErrorBucket{
		textBucket{false, "dunno"},
		"iAmAFileName",
		2,
		emptyCodeBlock,
		nil,
		"bash: line 9: lochNessMonster: command not found"}

	labels := []string{"foo", "bar"}

	blocks := []*codeBlock{
		&codeBlock{labels, "ls\ndate\n"},
		&codeBlock{labels, "echo beans\necho cheese\n"},
		&codeBlock{labels, "lochNessMonster\n"},
		&codeBlock{labels, "echo hasta\necho la vista\n"}}

	got := RunInSubShell([]*ScriptBucket{&ScriptBucket{"iAmFileName", blocks}})
	checkFail(t, got, want)
}
