package main

import (
	"testing"
)

func TestRunnerWithNothing(t *testing.T) {
	if RunInSubShell([]string{}).err != nil {
		t.Fail()
	}
}

func TestRunnerWithGoodStuff(t *testing.T) {
	result := RunInSubShell([]string{
		"ls\ndate\n",
		"echo beans\necho cheese\n",
		"echo hasta\necho la vista\n"})
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
	if got.message != want.message {
		t.Errorf("%s got\n\t%v\nwant\n\t%v", "message", got.message, want.message)
	}
}

func TestWithBadStuff1(t *testing.T) {
	want := &ErrorBucket{
		textBucket{false, "dunno"},
		0,
		"",
		nil,
		"bash: line 1: notagoodcommand: command not found"}

	got := RunInSubShell([]string{
		"notagoodcommand\n"})
	checkFail(t, got, want)
}

func TestWithBadStuff2(t *testing.T) {
	want := &ErrorBucket{
		textBucket{false, "dunno"},
		2,
		"",
		nil,
		"bash: line 9: lochNessMonster: command not found"}

	got := RunInSubShell([]string{
		"ls\ndate\n",
		"echo beans\necho cheese\n",
		"lochNessMonster\n",
		"echo hasta\necho la vista\n"})
	checkFail(t, got, want)
}
