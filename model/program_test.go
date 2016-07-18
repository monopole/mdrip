package model

import (
	"strings"
	"testing"
	"time"

	"github.com/monopole/mdrip/scanner"
)

const timeout = 2 * time.Second

var noLabels []Label = []Label{}
var labels = []Label{Label("foo"), Label("bar")}
var emptyCommandBlock *CommandBlock = NewCommandBlock(noLabels, "")

func TestRunnerWithNothing(t *testing.T) {
	if NewProgram(timeout, labels[0]).RunInSubShell().Problem() != nil {
		t.Fail()
	}
}

func doIt(blocks []*CommandBlock) *RunResult {
	p := NewProgram(timeout, labels[0]).Add(NewScript("iAmFileName", blocks))
	return p.RunInSubShell()
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

func checkFail(t *testing.T, got, want *RunResult) {
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

	want := NoCommandsRunResult(
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

	want := NoCommandsRunResult(
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
	want := NoCommandsRunResult(
		NewFailureOutput("dunno"),
		"fileNameTestTimeOut",
		0,
		scanner.MsgTimeout)

	// Insert this sleep in a command block.
	// Arrange to sleep for two seconds longer than the timeout.
	sleep := timeout + (2 * time.Second)

	blocks := []*CommandBlock{
		NewCommandBlock(labels, "date\nsleep "+sleep.String()+"\necho kale"),
		NewCommandBlock(labels, "echo beans\necho cheese\n")}
	checkFail(t, doIt(blocks), want)
}
