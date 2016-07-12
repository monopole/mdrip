package model

import (
	"github.com/monopole/mdrip/scanner"
	"strconv"
	"strings"
	"testing"
	"time"
)

var noLabels []Label = []Label{}
var labels = []Label{Label("foo"), Label("bar")}
var emptyCommandBlock *CommandBlock = NewCommandBlock(noLabels, "")

const timeoutSeconds = 1

func TestRunnerWithNothing(t *testing.T) {
	if NewProgram().RunInSubShell(timeoutSeconds*time.Second).Problem() != nil {
		t.Fail()
	}
}

func doIt(blocks []*CommandBlock) *ScriptResult {
	p := NewProgram().Add(NewScript("iAmFileName", blocks))
	return p.RunInSubShell(timeoutSeconds * time.Second)
}

func TestRunnerWithGoodStuff(t *testing.T) {
	blocks := []*CommandBlock{
		NewCommandBlock(labels, "echo kale\ndate\n"),
		NewCommandBlock(labels, "echo beans\necho apple\n"),
		NewCommandBlock(labels, "echo hasta\necho la vista\n")}
	result := doIt(blocks)
	if result.Problem() != nil {
		t.Fail()
	}
}

func checkFail(t *testing.T, got, want *ScriptResult) {
	if got.Problem() == nil {
		t.Fail()
	}
	if got.Index() != want.Index() {
		t.Errorf("%s got\n\t%v\nwant\n\t%v", "script", got.Index(), want.Index())
	}
	if !strings.Contains(got.Message(), want.Message()) {
		t.Errorf("%s got\n\t%v\nwant\n\t%v", "message", got.Message(), want.Message())
	}
}

func TestStartWithABadCommand(t *testing.T) {

	want := NoCommandsScriptResult(
		NewFailureOutput("dunno"),
		"fileNameTestStartWithABadCommand",
		0,
		"line 1: notagoodcommand: command not found")

	blocks := []*CommandBlock{
		NewCommandBlock(labels, "notagoodcommand\ndate\n"),
		NewCommandBlock(labels, "echo beans\necho cheese\n")}
	checkFail(t, doIt(blocks), want)
}

func TestBadCommandInTheMiddle(t *testing.T) {

	want := NoCommandsScriptResult(
		NewFailureOutput("dunno"),
		"fileNameTestBadCommandInTheMiddle",
		2,
		"line 9: lochNessMonster: command not found")

	blocks := []*CommandBlock{
		NewCommandBlock(labels, "echo tofu\ndate\n"),
		NewCommandBlock(labels, "echo beans\necho kale\n"),
		NewCommandBlock(labels, "lochNessMonster\n"),
		NewCommandBlock(labels, "echo hasta\necho la vista\n")}

	checkFail(t, doIt(blocks), want)
}

func TestTimeOut(t *testing.T) {
	want := NoCommandsScriptResult(
		NewFailureOutput("dunno"),
		"fileNameTestTimeOut",
		0,
		scanner.MsgTimeout)

	// Go to sleep for twice the length of the timeout.
	blocks := []*CommandBlock{
		NewCommandBlock(labels, "date\nsleep "+strconv.Itoa(timeoutSeconds+2)+"\necho kale"),
		NewCommandBlock(labels, "echo beans\necho cheese\n")}
	checkFail(t, doIt(blocks), want)
}
